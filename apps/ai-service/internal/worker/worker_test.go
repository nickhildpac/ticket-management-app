package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nickhildpac/ticket-management-ai-service/internal/triage"
)

const testTicketID = "11111111-1111-1111-1111-111111111111"

// stubAgent records the tickets it was asked to triage and returns a fixed
// result.
type stubAgent struct {
	result  triage.TriageResult
	triaged []string
}

func (s *stubAgent) Triage(_ context.Context, ticket triage.TicketContext) triage.TriageResult {
	s.triaged = append(s.triaged, ticket.TicketID)
	return s.result
}

// stubCommenter records the comments posted back to the ticket API.
type stubCommenter struct {
	fail     bool
	comments [][2]string
}

func (s *stubCommenter) AddComment(_ context.Context, ticketID, description string) error {
	if s.fail {
		return errors.New("ticket API unavailable")
	}
	s.comments = append(s.comments, [2]string{ticketID, description})
	return nil
}

func escalation() triage.TriageResult {
	return triage.TriageResult{
		TicketID:         testTicketID,
		Action:           triage.ActionEscalate,
		EscalationReason: "needs human review",
	}
}

func fields(t *testing.T, eventID, state string) map[string]string {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"ticket_id":     testTicketID,
		"ticket_number": 42,
		"title":         "Password reset",
		"description":   "I forgot my password.",
		"state":         state,
		"priority":      "low",
	})
	require.NoError(t, err)
	return map[string]string{
		"event_id":   eventID,
		"event_type": "ticket.created",
		"payload":    string(payload),
	}
}

// newTestWorker wires a worker against an in-process Redis.
func newTestWorker(t *testing.T, agent Triager, tickets Commenter) (*Worker, *miniredis.Miniredis, *redis.Client) {
	t.Helper()
	srv := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	w := New(rdb, agent, tickets, "ai-triage", "worker-test")
	// A blocking read only returns when its window expires, so shorten it to
	// keep the shutdown path in TestRunConsumesStreamAndAcks fast.
	w.readBlock = 50 * time.Millisecond
	return w, srv, rdb
}

func TestHandleSuccessMarksProcessedAndAcks(t *testing.T) {
	agent := &stubAgent{result: escalation()}
	tickets := &stubCommenter{}
	w, _, rdb := newTestWorker(t, agent, tickets)

	acked, err := w.Handle(context.Background(), fields(t, "evt-1", "open"))

	require.NoError(t, err)
	assert.True(t, acked)
	assert.Equal(t, int64(1), rdb.Exists(context.Background(), processedPrefix+"evt-1").Val())
	assert.Len(t, tickets.comments, 1)
}

func TestHandleDropsDuplicateEventWithoutRetriaging(t *testing.T) {
	agent := &stubAgent{result: escalation()}
	tickets := &stubCommenter{}
	w, _, rdb := newTestWorker(t, agent, tickets)
	require.NoError(t, rdb.Set(context.Background(), processedPrefix+"evt-1", "1", time.Minute).Err())

	acked, err := w.Handle(context.Background(), fields(t, "evt-1", "open"))

	require.NoError(t, err)
	assert.True(t, acked, "a duplicate delivery must still be acked")
	assert.Empty(t, agent.triaged)
	assert.Empty(t, tickets.comments)
}

func TestHandleFailureDoesNotMarkProcessed(t *testing.T) {
	agent := &stubAgent{result: escalation()}
	tickets := &stubCommenter{fail: true}
	w, _, rdb := newTestWorker(t, agent, tickets)

	acked, err := w.Handle(context.Background(), fields(t, "evt-1", "open"))

	require.Error(t, err)
	assert.False(t, acked)
	// Not marked processed: the message stays pending and a retry re-triages.
	assert.Equal(t, int64(0), rdb.Exists(context.Background(), processedPrefix+"evt-1").Val())
}

func TestProcessMessageSkipsTerminalStates(t *testing.T) {
	for _, state := range []string{"resolved", "closed", "cancelled"} {
		t.Run(state, func(t *testing.T) {
			agent := &stubAgent{result: escalation()}
			tickets := &stubCommenter{}

			require.NoError(t, ProcessMessage(context.Background(), agent, tickets, fields(t, "e", state)))
			assert.Empty(t, agent.triaged)
			assert.Empty(t, tickets.comments)
		})
	}
}

func TestProcessMessageIgnoresUnknownEventTypes(t *testing.T) {
	agent := &stubAgent{result: escalation()}
	tickets := &stubCommenter{}

	err := ProcessMessage(context.Background(), agent, tickets,
		map[string]string{"event_type": "ticket.deleted", "payload": "{}"})

	require.NoError(t, err)
	assert.Empty(t, agent.triaged)
}

func TestProcessMessageRejectsMalformedPayload(t *testing.T) {
	agent := &stubAgent{result: escalation()}
	tickets := &stubCommenter{}

	err := ProcessMessage(context.Background(), agent, tickets,
		map[string]string{"event_type": "ticket.created", "payload": "not json"})

	require.Error(t, err)
	assert.Empty(t, agent.triaged)
}

func TestApplyResultPostsAIReplyForAutoAnswer(t *testing.T) {
	tickets := &stubCommenter{}

	require.NoError(t, ApplyResult(context.Background(), tickets, triage.TriageResult{
		TicketID: "t-1", Action: triage.ActionAutoAnswer, Confidence: 0.9,
		DraftReply: "Use the reset link.",
	}))

	require.Len(t, tickets.comments, 1)
	assert.Equal(t, "t-1", tickets.comments[0][0])
	assert.Contains(t, tickets.comments[0][1], "AI-suggested reply")
	assert.Contains(t, tickets.comments[0][1], "Use the reset link.")
}

func TestApplyResultEscalationCommentHidesInternalFlags(t *testing.T) {
	tickets := &stubCommenter{}

	require.NoError(t, ApplyResult(context.Background(), tickets, triage.TriageResult{
		TicketID: "t-1", Action: triage.ActionEscalate,
		EscalationReason: "This request needs review by our team before we can respond.",
		SafetyFlags:      []string{"refund_or_cancellation"},
	}))

	comment := tickets.comments[0][1]
	assert.Contains(t, comment, "Escalated to a human")
	// The end user must not see raw safety-flag slugs in the comment.
	assert.NotContains(t, comment, "refund_or_cancellation")
	assert.NotContains(t, comment, "flags:")
}

// An auto_answer whose draft is empty must still produce a comment rather than
// silently dropping the ticket.
func TestApplyResultFallsBackToEscalationCommentWithoutADraft(t *testing.T) {
	tickets := &stubCommenter{}

	require.NoError(t, ApplyResult(context.Background(), tickets, triage.TriageResult{
		TicketID: "t-1", Action: triage.ActionAutoAnswer, Confidence: 0.9,
	}))

	require.Len(t, tickets.comments, 1)
	assert.Contains(t, tickets.comments[0][1], "A teammate will follow up on this ticket.")
}

// End-to-end through the stream: an event published to Redis is consumed,
// triaged and acked.
func TestRunConsumesStreamAndAcks(t *testing.T) {
	agent := &stubAgent{result: triage.TriageResult{
		TicketID: testTicketID, Action: triage.ActionAutoAnswer,
		Confidence: 0.95, DraftReply: "Use the reset link [1].",
	}}
	tickets := &stubCommenter{}
	w, _, rdb := newTestWorker(t, agent, tickets)

	ctx := context.Background()
	values := map[string]any{}
	for k, v := range fields(t, "evt-stream", "open") {
		values[k] = v
	}
	require.NoError(t, rdb.XAdd(ctx, &redis.XAddArgs{Stream: StreamKey, Values: values}).Err())

	runCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- w.Run(runCtx) }()

	require.Eventually(t, func() bool {
		return len(tickets.comments) == 1
	}, 5*time.Second, 20*time.Millisecond, "the worker should consume and apply the event")

	cancel()
	require.NoError(t, <-done)

	assert.Equal(t, []string{testTicketID}, agent.triaged)
	assert.Contains(t, tickets.comments[0][1], "AI-suggested reply")

	pending, err := rdb.XPending(ctx, StreamKey, "ai-triage").Result()
	require.NoError(t, err)
	assert.Zero(t, pending.Count, "a handled message must be acked, not left pending")
}

// The consumer group must be created idempotently — restarting the worker
// against an existing group is normal.
func TestEnsureGroupIsIdempotent(t *testing.T) {
	w, _, _ := newTestWorker(t, &stubAgent{}, &stubCommenter{})

	require.NoError(t, w.ensureGroup(context.Background()))
	require.NoError(t, w.ensureGroup(context.Background()))
}

func TestStringFieldsCoercesNonStringValues(t *testing.T) {
	out := stringFields(map[string]any{"a": "x", "b": 7})

	assert.Equal(t, map[string]string{"a": "x", "b": "7"}, out)
}
