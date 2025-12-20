package port

import (
	"context"

	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

type TicketService interface {
	ListAll(ctx context.Context, limit, offset int32) ([]domain.Ticket, error)
	ListByCreator(ctx context.Context, username string, limit, offset int32) ([]domain.Ticket, error)
	GetTicket(ctx context.Context, id int64) (*domain.Ticket, error)
	CreateTicket(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error)
}
