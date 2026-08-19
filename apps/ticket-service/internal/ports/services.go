package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

type UserService interface {
	GetUser(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetUsersByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.User, error)
	UpdateMySkills(ctx context.Context, skills []string) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
	GetAllUsersForAssignment(ctx context.Context) ([]domain.User, error)
	UpdateUserRole(ctx context.Context, id uuid.UUID, role domain.UserRole) (*domain.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type TicketService interface {
	ListAll(ctx context.Context, limit, offset, sortVal int32) ([]domain.Ticket, error)
	ListTicketsWithFilters(ctx context.Context, params domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error)
	// ListAssignedToCurrentUser lists tickets assigned to the authenticated user with optional filters (limit/offset/state/priority/ticket_number). Ignores created_by / assignee query intent; scope is always the current user as assignee.
	ListAssignedToCurrentUser(ctx context.Context, params domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error)
	ListByCreator(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error)
	ListByAssignee(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error)
	GetTicket(ctx context.Context, id uuid.UUID) (*domain.Ticket, error)
	GetTicketByNumber(ctx context.Context, ticketNumber int64) (*domain.Ticket, error)
	GetTicketStats(ctx context.Context) (domain.TicketListStats, error)
	CreateTicket(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error)
	UpdateTicket(ctx context.Context, id uuid.UUID, patch domain.TicketPatch) (*domain.Ticket, error)
	DeleteTicket(ctx context.Context, id uuid.UUID) error
}

type CommentService interface {
	ListByTicket(ctx context.Context, ticketID uuid.UUID, limit, offset int32) ([]domain.Comment, error)
	GetComment(ctx context.Context, id uuid.UUID) (*domain.Comment, error)
	CreateComment(ctx context.Context, comment domain.Comment) (*domain.Comment, error)
}
