package domain

import "github.com/google/uuid"

// Domain event types published to the outbox. Wire names mirror
// contracts/events/*.json — keep the two in sync.
const (
	EventTicketCreated = "ticket.created"
	EventTicketUpdated = "ticket.updated"
)

// OutboxEvent is a domain event queued for asynchronous delivery to the
// AI/RAG service via the transactional outbox + relay.
type OutboxEvent struct {
	Type        string
	AggregateID uuid.UUID
	Payload     map[string]any
}

// NewTicketEvent builds an outbox event carrying the fields the AI triage
// agent needs (title/description/state) plus routing metadata.
func NewTicketEvent(eventType string, t *Ticket) OutboxEvent {
	assigned := make([]string, 0, len(t.AssignedTo))
	for _, a := range t.AssignedTo {
		assigned = append(assigned, a.String())
	}
	return OutboxEvent{
		Type:        eventType,
		AggregateID: t.ID,
		Payload: map[string]any{
			"ticket_id":     t.ID.String(),
			"ticket_number": t.TicketNumber,
			"title":         t.Title,
			"description":   t.Description,
			"state":         t.State.String(),
			"priority":      t.Priority.String(),
			"created_by":    t.CreatedBy.String(),
			"assigned_to":   assigned,
		},
	}
}
