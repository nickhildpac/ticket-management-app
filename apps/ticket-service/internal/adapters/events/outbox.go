// Package events provides the transactional-outbox publisher and the relay that
// drains it to the message broker. See docs/adr/0002-service-topology.md.
package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

// OutboxPublisher persists domain events to the event_outbox table. Delivery to
// the broker happens later, out-of-band, via Relay — giving at-least-once
// semantics decoupled from the request path.
type OutboxPublisher struct {
	db *sql.DB
}

func NewOutboxPublisher(db *sql.DB) *OutboxPublisher {
	return &OutboxPublisher{db: db}
}

func (p *OutboxPublisher) Publish(ctx context.Context, event domain.OutboxEvent) error {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}
	const q = `INSERT INTO event_outbox (event_type, aggregate_id, payload)
	           VALUES ($1, $2, $3)`
	if _, err := p.db.ExecContext(ctx, q, event.Type, event.AggregateID, payload); err != nil {
		return fmt.Errorf("insert outbox row: %w", err)
	}
	return nil
}
