package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

type UserService interface {
	GetUser(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	CreateUser(ctx context.Context, user domain.User) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
}

type TicketService interface {
	ListAll(ctx context.Context, limit, offset int32) ([]domain.Ticket, error)
	ListByCreator(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error)
	ListByAssignee(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error)
	GetTicket(ctx context.Context, id uuid.UUID) (*domain.Ticket, error)
	CreateTicket(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error)
	UpdateTicket(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error)
}

type CommentService interface {
	ListByTicket(ctx context.Context, ticketID uuid.UUID, limit, offset int32) ([]domain.Comment, error)
	GetComment(ctx context.Context, id uuid.UUID) (*domain.Comment, error)
	CreateComment(ctx context.Context, comment domain.Comment) (*domain.Comment, error)
}
