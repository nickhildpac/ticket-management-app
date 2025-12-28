package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

type UserRepository interface {
	GetUser(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	CreateUser(ctx context.Context, user domain.User) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
}

type TicketRepository interface {
	ListAll(ctx context.Context, limit, offset int32) ([]domain.Ticket, error)
	ListByCreator(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error)
	ListByAssignee(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error)
	Get(ctx context.Context, id int64) (*domain.Ticket, error)
	Create(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error)
	Update(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error)
}

type CommentRepository interface {
	ListByTicket(ctx context.Context, ticketID int64, limit, offset int32) ([]domain.Comment, error)
	Get(ctx context.Context, id int64) (*domain.Comment, error)
	Create(ctx context.Context, comment domain.Comment) (*domain.Comment, error)
}
