package ports

import (
	"context"

	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

type UserService interface {
	GetUser(ctx context.Context, username string) (*domain.User, error)
	CreateUser(ctx context.Context, user domain.User) (*domain.User, error)
}

type TicketService interface {
	ListAll(ctx context.Context, limit, offset int32) ([]domain.Ticket, error)
	ListByCreator(ctx context.Context, username string, limit, offset int32) ([]domain.Ticket, error)
	GetTicket(ctx context.Context, id int64) (*domain.Ticket, error)
	CreateTicket(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error)
}

type CommentService interface {
	ListByTicket(ctx context.Context, ticketID int64, limit, offset int32) ([]domain.Comment, error)
	GetComment(ctx context.Context, id int64) (*domain.Comment, error)
	CreateComment(ctx context.Context, comment domain.Comment) (*domain.Comment, error)
}
