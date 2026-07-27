package triage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nickhildpac/ticket-management-ai-service/internal/rag"
)

var ticketNumber = 42

var testTicket = TicketContext{
	TicketID:     "11111111-1111-1111-1111-111111111111",
	TicketNumber: &ticketNumber,
	Title:        "How do I reset my password?",
	Description:  "I forgot my password and can't log in.",
	State:        "open",
	Priority:     "low",
}

// stubStore is a one-passage KB so search_docs/rerank_results run against the
// real retrieval code rather than a mock of it.
type stubStore struct {
	semanticCalls []call
	keywordCalls  []call
}

type call struct {
	query string
	k     int
}

const (
	stubSource  = "kb/faq.md"
	stubContent = "Use the reset link on the login page."
)

func (s *stubStore) SearchSemantic(_ context.Context, query string, k int) ([]rag.KBChunk, error) {
	s.semanticCalls = append(s.semanticCalls, call{query, k})
	return []rag.KBChunk{{ID: 1, Content: stubContent, Source: stubSource, Distance: 0.1}}, nil
}

func (s *stubStore) SearchKeyword(_ context.Context, query string, k int) ([]rag.KBChunk, error) {
	s.keywordCalls = append(s.keywordCalls, call{query, k})
	return []rag.KBChunk{{ID: 1, Content: stubContent, Source: stubSource, Distance: 0.2}}, nil
}

type stubReranker struct{}

func (stubReranker) Score(_ context.Context, _ string, passages []string) ([]float64, error) {
	out := make([]float64, len(passages))
	for i := range out {
		out[i] = 1.0
	}
	return out, nil
}

// step scripts one assistant turn: either a refusal, or a tool call the agent's
// real handlers will execute.
type step struct {
	stopReason anthropic.StopReason
	tool       string
	input      map[string]any
}

// fakeClient stands in for the Anthropic Messages API, replaying scripted
// assistant turns. Once the script is exhausted it returns a plain end_turn
// message, mirroring a model that stops calling tools.
type fakeClient struct {
	script []step
	err    error
	calls  int
}

func (f *fakeClient) New(_ context.Context, _ anthropic.MessageNewParams,
	_ ...option.RequestOption) (*anthropic.Message, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.calls >= len(f.script) {
		f.calls++
		return buildMessage(anthropic.StopReasonEndTurn, "", nil)
	}
	s := f.script[f.calls]
	f.calls++
	if s.stopReason == anthropic.StopReasonRefusal {
		return buildMessage(anthropic.StopReasonRefusal, "", nil)
	}
	return buildMessage(anthropic.StopReasonToolUse, s.tool, s.input)
}

// buildMessage assembles an anthropic.Message through its JSON decoder so the
// union content blocks are populated exactly as a real API response would be.
func buildMessage(stop anthropic.StopReason, tool string, input map[string]any) (*anthropic.Message, error) {
	content := []map[string]any{}
	if tool != "" {
		if input == nil {
			input = map[string]any{}
		}
		content = append(content, map[string]any{
			"type": "tool_use", "id": "toolu_" + tool, "name": tool, "input": input,
		})
	}
	raw, err := json.Marshal(map[string]any{
		"id": "msg_test", "type": "message", "role": "assistant", "model": "test-model",
		"stop_reason": string(stop),
		"content":     content,
		"usage":       map[string]any{"input_tokens": 1, "output_tokens": 1},
	})
	if err != nil {
		return nil, err
	}
	var msg anthropic.Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, fmt.Errorf("build fake message: %w", err)
	}
	return &msg, nil
}

func newTestAgent(t *testing.T, client messageClient, opts Options) *Agent {
	t.Helper()
	if opts.Model == "" {
		opts.Model = "claude-opus-4-8"
	}
	if opts.ConfidenceThreshold == 0 {
		opts.ConfidenceThreshold = 0.75
	}
	return newAgent(client, &stubStore{}, stubReranker{}, opts)
}

func draftScript(input map[string]any) []step {
	return []step{{tool: toolDraftReply, input: input}}
}

func triageWith(t *testing.T, script []step) TriageResult {
	t.Helper()
	return newTestAgent(t, &fakeClient{script: script}, Options{}).
		Triage(context.Background(), testTicket)
}

func TestAutoAnswerWhenConfidentAndClean(t *testing.T) {
	result := triageWith(t, draftScript(map[string]any{
		"reply": "Use the reset link.", "confidence": 0.9,
	}))

	assert.Equal(t, ActionAutoAnswer, result.Action)
	assert.Equal(t, "Use the reset link.", result.DraftReply)
	assert.Empty(t, result.SafetyFlags)
}

func TestEscalatesWhenBelowConfidenceThreshold(t *testing.T) {
	result := triageWith(t, draftScript(map[string]any{
		"reply": "Maybe try resetting.", "confidence": 0.72,
	}))

	assert.Equal(t, ActionEscalate, result.Action)
	// The score stays on the result for logging...
	assert.InDelta(t, 0.72, result.Confidence, 1e-9)
	// ...but the customer-facing reason must not leak the metric or threshold.
	assert.NotContains(t, result.EscalationReason, "0.72")
	assert.NotContains(t, result.EscalationReason, "0.75")
	assert.NotContains(t, strings.ToLower(result.EscalationReason), "threshold")
}

func TestGateEscalationReasonHidesSafetyFlagSlugs(t *testing.T) {
	result := triageWith(t, draftScript(map[string]any{
		"reply": "I refunded your account.", "confidence": 0.99,
		"safety_flags": []string{"refund_or_cancellation"},
	}))

	assert.Equal(t, ActionEscalate, result.Action)
	// Flags are preserved on the result (for logging) but never in the reason text.
	assert.Equal(t, []string{"refund_or_cancellation"}, result.SafetyFlags)
	assert.NotContains(t, result.EscalationReason, "refund_or_cancellation")
}

func TestEscalatesWhenSafetyFlagsPresent(t *testing.T) {
	result := triageWith(t, draftScript(map[string]any{
		"reply": "I refunded your account.", "confidence": 0.99,
		"safety_flags": []string{"refund_or_cancellation"},
	}))

	assert.Equal(t, ActionEscalate, result.Action)
	assert.Equal(t, []string{"refund_or_cancellation"}, result.SafetyFlags)
}

func TestDocumentedSelfServiceAutoAnswersWithoutFlags(t *testing.T) {
	// A 403-style ticket whose fix is documented self-service configuration:
	// no sensitive action, a cited draft, no safety flags -> auto_answer.
	result := triageWith(t, draftScript(map[string]any{
		"reply": "Enable the Analytics API Access toggle and confirm the key has the " +
			"read:analytics scope [1]. A human will follow up if that doesn't resolve it.",
		"confidence":   0.85,
		"safety_flags": []string{},
	}))

	assert.Equal(t, ActionAutoAnswer, result.Action)
	assert.Empty(t, result.SafetyFlags)
	assert.Contains(t, result.DraftReply, "[1]")
}

func TestAccountLockoutEscalatesDespiteHighConfidence(t *testing.T) {
	result := triageWith(t, draftScript(map[string]any{
		"reply": "Regaining access requires an account action.", "confidence": 0.98,
		"safety_flags": []string{"account_access_lockout"},
	}))

	assert.Equal(t, ActionEscalate, result.Action)
	assert.Equal(t, []string{"account_access_lockout"}, result.SafetyFlags)
}

func TestEscalatesWhenAutoAnswerHasNoDraft(t *testing.T) {
	result := triageWith(t, draftScript(map[string]any{
		"reply": "  ", "confidence": 0.95,
	}))

	assert.Equal(t, ActionEscalate, result.Action)
}

func TestEscalateTicketToolHandsOffToHuman(t *testing.T) {
	result := triageWith(t, []step{{
		tool:  toolEscalateTicket,
		input: map[string]any{"reason": "A teammate will review this refund."},
	}})

	assert.Equal(t, ActionEscalate, result.Action)
	assert.Equal(t, "A teammate will review this refund.", result.EscalationReason)
}

func TestSearchThenRerankThenDraft(t *testing.T) {
	result := triageWith(t, []step{
		{tool: toolSearchDocs, input: map[string]any{"query": "reset password"}},
		{tool: toolRerankResults, input: map[string]any{"query": "reset password"}},
		{tool: toolDraftReply, input: map[string]any{
			"reply": "Use the reset link [1].", "confidence": 0.9,
		}},
	})

	assert.Equal(t, ActionAutoAnswer, result.Action)
	assert.Equal(t, "Use the reset link [1].", result.DraftReply)
}

// Escalation is sticky: a later draft_reply must not downgrade a recorded
// human handoff to an auto-answer.
func TestEscalationIsStickyAgainstALaterDraft(t *testing.T) {
	agent := newTestAgent(t, &fakeClient{script: []step{
		{tool: toolEscalateTicket, input: map[string]any{"reason": "Needs a human."}},
	}}, Options{})
	sess := &session{ticketID: testTicket.TicketID, candidates: map[int]rag.KBChunk{}}

	escalateTicket(sess, escalateTicketInput{Reason: "Needs a human."})
	out := draftReply(sess, draftReplyInput{Reply: "Actually here you go.", Confidence: 0.99})

	assert.Contains(t, out, "keeping the human handoff")
	require.NotNil(t, sess.decision)
	assert.Equal(t, ActionEscalate, sess.decision.Action)

	result := agent.ApplySafetyGate(testTicket, *sess.decision)
	assert.Equal(t, ActionEscalate, result.Action)
}

func TestRefusalStopReasonEscalates(t *testing.T) {
	result := triageWith(t, []step{{stopReason: anthropic.StopReasonRefusal}})

	assert.Equal(t, ActionEscalate, result.Action)
	assert.Contains(t, result.SafetyFlags, "model_refusal")
}

func TestModelErrorFailsSafeToEscalate(t *testing.T) {
	agent := newTestAgent(t, &fakeClient{err: errors.New("boom")}, Options{})

	result := agent.Triage(context.Background(), testTicket)

	assert.Equal(t, ActionEscalate, result.Action)
	assert.Contains(t, result.SafetyFlags, "model_error")
}

func TestNoTerminalDecisionFailsSafeToEscalate(t *testing.T) {
	// The model searches but never makes a terminal decision.
	result := triageWith(t, []step{
		{tool: toolSearchDocs, input: map[string]any{"query": "reset password"}},
	})

	assert.Equal(t, ActionEscalate, result.Action)
	assert.Contains(t, result.SafetyFlags, "no_decision")
}

func TestMaxIterationsWithoutDecisionEscalates(t *testing.T) {
	// More search turns than the iteration cap, no terminal decision -> fail safe.
	script := make([]step, 10)
	for i := range script {
		script[i] = step{tool: toolSearchDocs, input: map[string]any{"query": "reset password"}}
	}
	client := &fakeClient{script: script}
	agent := newTestAgent(t, client, Options{MaxIterations: 3})

	result := agent.Triage(context.Background(), testTicket)

	assert.Equal(t, ActionEscalate, result.Action)
	assert.Contains(t, result.SafetyFlags, "no_decision")
	assert.Equal(t, 3, client.calls, "the iteration cap must bound the API calls")
}

// A terminal tool must end the loop immediately rather than paying for one more
// model round-trip — the reason this is a manual loop and not the SDK runner.
func TestTerminalToolStopsWithoutAnExtraModelCall(t *testing.T) {
	client := &fakeClient{script: []step{
		{tool: toolSearchDocs, input: map[string]any{"query": "reset password"}},
		{tool: toolDraftReply, input: map[string]any{"reply": "Use the link [1].", "confidence": 0.9}},
	}}
	agent := newTestAgent(t, client, Options{})

	result := agent.Triage(context.Background(), testTicket)

	assert.Equal(t, ActionAutoAnswer, result.Action)
	assert.Equal(t, 2, client.calls)
}

func TestSearchDocsUsesCandidateKOnBothLanes(t *testing.T) {
	store := &stubStore{}
	agent := newAgent(
		&fakeClient{script: []step{
			{tool: toolSearchDocs, input: map[string]any{"query": "reset password"}},
			{tool: toolEscalateTicket, input: map[string]any{"reason": "A teammate will follow up."}},
		}},
		store, stubReranker{},
		Options{Model: "test-model", ConfidenceThreshold: 0.75, CandidateK: 3},
	)

	agent.Triage(context.Background(), testTicket)

	require.Len(t, store.semanticCalls, 1)
	require.Len(t, store.keywordCalls, 1)
	assert.Equal(t, call{"reset password", 3}, store.semanticCalls[0])
	assert.Equal(t, call{"reset password", 3}, store.keywordCalls[0])
}

func TestSearchDocsLogsRetrievedPassages(t *testing.T) {
	logs := captureLogs(t)

	triageWith(t, []step{
		{tool: toolSearchDocs, input: map[string]any{"query": "reset password"}},
		{tool: toolEscalateTicket, input: map[string]any{"reason": "A teammate will follow up."}},
	})

	out := logs.String()
	assert.Contains(t, out, "search_docs")
	assert.Contains(t, out, `"passages":1`)
	assert.Contains(t, out, stubSource)
	assert.Contains(t, out, stubContent)
}

// captureLogs redirects the default slog logger into a buffer for the duration
// of the test.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return buf
}

func TestUserContentIncludesTicketFields(t *testing.T) {
	out := userContent(testTicket)

	assert.Contains(t, out, "Ticket #42")
	assert.Contains(t, out, "priority=low")
	assert.Contains(t, out, "state=open")
	assert.Contains(t, out, testTicket.Title)
	assert.Contains(t, out, testTicket.Description)
}

func TestClamp01(t *testing.T) {
	assert.Equal(t, 0.0, clamp01(-1))
	assert.Equal(t, 1.0, clamp01(2))
	assert.Equal(t, 0.5, clamp01(0.5))
}
