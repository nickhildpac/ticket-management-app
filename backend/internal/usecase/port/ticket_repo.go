package port

import (
	"context"

	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

type TicketRepository interface {
	ListAll(ctx context.Context, limit, offset int32) ([]domain.Ticket, error)
	ListByCreator(ctx context.Context, username string, limit, offset int32) ([]domain.Ticket, error)
	Get(ctx context.Context, id int64) (*domain.Ticket, error)
	Create(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error)
}
