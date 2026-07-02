package ports

import (
	"context"

	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

// EventPublisher records a domain event for asynchronous delivery. The default
// implementation writes to the transactional outbox; the relay drains it.
type EventPublisher interface {
	Publish(ctx context.Context, event domain.OutboxEvent) error
}
