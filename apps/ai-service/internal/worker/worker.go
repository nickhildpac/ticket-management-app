// Package worker consumes the ticket-events Redis stream, triages each ticket
// and applies the result by commenting back on the ticket via the Go ticket API.
package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/nickhildpac/ticket-management-ai-service/internal/triage"
)

const (
	// StreamKey is the Redis stream the ticket-service outbox relay writes to.
	StreamKey = "ticket-events"
	// DLQStream holds messages that exceeded the delivery ceiling.
	DLQStream = "ticket-events-dlq"

	// processedPrefix keys the idempotency markers. After a successful handling
	// we record the event id so a re-delivered event (the outbox relay is
	// at-least-once) is dropped instead of triaged twice.
	processedPrefix = "ai-triage:processed:"
	processedTTL    = 24 * time.Hour

	// reclaimIdle is how long a message must sit un-acked before another
	// consumer may claim it; maxDeliveries is the dead-letter ceiling.
	reclaimIdle   = 60 * time.Second
	maxDeliveries = 5

	// defaultReadBlock is how long XREADGROUP waits for new messages before
	// looping. It also bounds shutdown latency: a blocking read only returns
	// when the window expires or a message arrives, so the worker can take up
	// to this long to notice a cancelled context.
	defaultReadBlock = 5 * time.Second
	readCount        = 10
)

// terminalStates are ticket states where triage is pointless — the ticket is
// finished, so re-triaging a transition into one of these would only spend
// model tokens and post a comment on a closed ticket.
var terminalStates = map[string]struct{}{
	"resolved":  {},
	"closed":    {},
	"cancelled": {},
}

// triagedEventTypes are the outbox event types the worker acts on.
var triagedEventTypes = map[string]struct{}{
	"ticket.created": {},
	"ticket.updated": {},
}

// Triager produces a gated triage decision for a ticket.
type Triager interface {
	Triage(ctx context.Context, ticket triage.TicketContext) triage.TriageResult
}

// Commenter applies a triage result by commenting on the ticket.
type Commenter interface {
	AddComment(ctx context.Context, ticketID, description string) error
}

// Worker is the Redis Streams consumer.
type Worker struct {
	rdb      *redis.Client
	agent    Triager
	tickets  Commenter
	group    string
	consumer string
	// readBlock is the XREADGROUP block window; tests shorten it.
	readBlock time.Duration
}

// New builds a worker bound to a consumer group and this replica's consumer
// name.
func New(rdb *redis.Client, agent Triager, tickets Commenter, group, consumer string) *Worker {
	return &Worker{
		rdb:       rdb,
		agent:     agent,
		tickets:   tickets,
		group:     group,
		consumer:  consumer,
		readBlock: defaultReadBlock,
	}
}

// Run blocks, consuming the stream until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) error {
	if err := w.ensureGroup(ctx); err != nil {
		return err
	}

	slog.Info("triage worker consuming",
		"stream", StreamKey, "group", w.group, "consumer", w.consumer)

	for {
		if ctx.Err() != nil {
			return nil
		}
		w.reclaimStale(ctx)

		streams, err := w.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    w.group,
			Consumer: w.consumer,
			Streams:  []string{StreamKey, ">"},
			Count:    readCount,
			Block:    w.readBlock,
		}).Result()
		switch {
		case errors.Is(err, redis.Nil):
			// Block window elapsed with no new messages.
			continue
		case ctx.Err() != nil:
			return nil
		case err != nil:
			slog.Error("xreadgroup failed; retrying", "error", err)
			// Back off briefly so a persistent Redis failure doesn't spin.
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(time.Second):
			}
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				w.handleAndAck(ctx, msg)
			}
		}
	}
}

// ensureGroup creates the consumer group, tolerating an existing one.
func (w *Worker) ensureGroup(ctx context.Context) error {
	err := w.rdb.XGroupCreateMkStream(ctx, StreamKey, w.group, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("create consumer group %q: %w", w.group, err)
	}
	return nil
}

// handleAndAck processes one message, acking it only when it was handled.
// Failures leave the message in the pending-entries list so reclaimStale
// retries it (and eventually dead-letters it).
func (w *Worker) handleAndAck(ctx context.Context, msg redis.XMessage) {
	fields := stringFields(msg.Values)
	acked, err := w.Handle(ctx, fields)
	if err != nil {
		slog.Error("failed to process message; will retry", "message_id", msg.ID, "error", err)
		return
	}
	if acked {
		if err := w.rdb.XAck(ctx, StreamKey, w.group, msg.ID).Err(); err != nil {
			slog.Error("failed to ack message", "message_id", msg.ID, "error", err)
		}
	}
}

// Handle triages one event and records its id for deduplication. It reports
// whether the message should be acked; an error leaves it pending for retry.
func (w *Worker) Handle(ctx context.Context, fields map[string]string) (bool, error) {
	eventID := fields["event_id"]
	if eventID != "" {
		seen, err := w.rdb.Exists(ctx, processedPrefix+eventID).Result()
		if err != nil {
			return false, err
		}
		if seen > 0 {
			slog.Info("dropping already-processed event", "event_id", eventID)
			return true, nil // duplicate delivery — ack and move on
		}
	}

	if err := ProcessMessage(ctx, w.agent, w.tickets, fields); err != nil {
		return false, err
	}

	if eventID != "" {
		if err := w.rdb.Set(ctx, processedPrefix+eventID, "1", processedTTL).Err(); err != nil {
			// The work is already applied; failing here would re-triage the
			// ticket on redelivery, which is worse than a missing marker.
			slog.Error("failed to record processed event", "event_id", eventID, "error", err)
		}
	}
	return true, nil
}

// reclaimStale recovers messages left pending by a crashed or failed handler.
// Anything past the delivery ceiling is dead-lettered so it can't block the
// group forever.
func (w *Worker) reclaimStale(ctx context.Context) {
	start := "0-0"
	for {
		messages, next, err := w.rdb.XAutoClaim(ctx, &redis.XAutoClaimArgs{
			Stream:   StreamKey,
			Group:    w.group,
			Consumer: w.consumer,
			MinIdle:  reclaimIdle,
			Start:    start,
			Count:    10,
		}).Result()
		if err != nil {
			if ctx.Err() == nil {
				slog.Error("xautoclaim failed", "error", err)
			}
			return
		}

		for _, msg := range messages {
			deliveries := w.deliveryCount(ctx, msg.ID)
			if deliveries > maxDeliveries {
				slog.Error("dead-lettering message",
					"message_id", msg.ID, "deliveries", deliveries)
				if err := w.rdb.XAdd(ctx, &redis.XAddArgs{
					Stream: DLQStream, Values: msg.Values,
				}).Err(); err != nil {
					slog.Error("failed to dead-letter message", "message_id", msg.ID, "error", err)
					continue
				}
				if err := w.rdb.XAck(ctx, StreamKey, w.group, msg.ID).Err(); err != nil {
					slog.Error("failed to ack dead-lettered message", "message_id", msg.ID, "error", err)
				}
				continue
			}

			acked, err := w.Handle(ctx, stringFields(msg.Values))
			if err != nil {
				// Leave it pending for the next reclaim.
				slog.Error("reclaim: failed to process message",
					"message_id", msg.ID, "delivery", deliveries, "error", err)
				continue
			}
			if acked {
				if err := w.rdb.XAck(ctx, StreamKey, w.group, msg.ID).Err(); err != nil {
					slog.Error("reclaim: failed to ack message", "message_id", msg.ID, "error", err)
				}
			}
		}

		if next == "0-0" || next == "" {
			return
		}
		start = next
	}
}

// deliveryCount reads how many times a pending message has been delivered.
// A lookup failure falls back to 1 so a transient Redis error can't
// dead-letter a message prematurely.
func (w *Worker) deliveryCount(ctx context.Context, messageID string) int64 {
	pending, err := w.rdb.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: StreamKey,
		Group:  w.group,
		Start:  messageID,
		End:    messageID,
		Count:  1,
	}).Result()
	if err != nil || len(pending) == 0 {
		return 1
	}
	return pending[0].RetryCount
}

// ticketEventPayload is the outbox payload shape (contracts/events/ticket_events.json).
type ticketEventPayload struct {
	TicketID     string `json:"ticket_id"`
	TicketNumber *int   `json:"ticket_number"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	State        string `json:"state"`
	Priority     string `json:"priority"`
}

// ProcessMessage triages one ticket event and applies the result. It returns an
// error on failure so the caller can leave the message un-acked for retry.
func ProcessMessage(ctx context.Context, agent Triager, tickets Commenter, fields map[string]string) error {
	eventType := fields["event_type"]
	if _, ok := triagedEventTypes[eventType]; !ok {
		slog.Debug("ignoring event", "event_type", eventType)
		return nil
	}

	var payload ticketEventPayload
	if err := json.Unmarshal([]byte(fields["payload"]), &payload); err != nil {
		return fmt.Errorf("decode event payload: %w", err)
	}

	ticket := triage.TicketContext{
		TicketID:     payload.TicketID,
		TicketNumber: payload.TicketNumber,
		Title:        payload.Title,
		Description:  payload.Description,
		State:        payload.State,
		Priority:     payload.Priority,
	}
	if ticket.Priority == "" {
		ticket.Priority = "low"
	}
	if _, terminal := terminalStates[strings.ToLower(ticket.State)]; terminal {
		slog.Info("skipping triage for ticket in terminal state",
			"ticket_id", ticket.TicketID, "state", ticket.State)
		return nil
	}

	result := agent.Triage(ctx, ticket)
	slog.Info("triage decision",
		"ticket_id", ticket.TicketID, "action", result.Action, "confidence", result.Confidence)
	slog.Info("triage output",
		"ticket_id", ticket.TicketID,
		"draft_reply", result.DraftReply,
		"escalation_reason", result.EscalationReason)

	return ApplyResult(ctx, tickets, result)
}

// ApplyResult applies a triage decision via the ticket API. We only post
// comments (valid from any state); we deliberately avoid state transitions to
// respect the ticket state machine.
func ApplyResult(ctx context.Context, tickets Commenter, result triage.TriageResult) error {
	if result.Action == triage.ActionAutoAnswer && result.DraftReply != "" {
		return tickets.AddComment(ctx, result.TicketID,
			"[AI-suggested reply]\n\n"+result.DraftReply)
	}
	// Customer-facing comment: the reason is already customer-safe (internal
	// confidence/threshold metrics and raw safety-flag slugs stay in the logs
	// above, never in the ticket comment the end user can see).
	reason := result.EscalationReason
	if reason == "" {
		reason = "A teammate will follow up on this ticket."
	}
	return tickets.AddComment(ctx, result.TicketID,
		"[AI triage] Escalated to a human: "+reason)
}

// stringFields flattens a Redis stream entry's values, which arrive as
// map[string]any, into the string map the handlers expect.
func stringFields(values map[string]any) map[string]string {
	out := make(map[string]string, len(values))
	for k, v := range values {
		if s, ok := v.(string); ok {
			out[k] = s
			continue
		}
		out[k] = fmt.Sprint(v)
	}
	return out
}
