// Package ports defines interfaces for services and repositories used across the application.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

type UserRepository interface {
	GetUser(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetUsersByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.User, error)
	CreateUser(ctx context.Context, user domain.User) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
	GetAllAgents(ctx context.Context) ([]domain.User, error)
	GetAutoAssignmentCandidates(ctx context.Context, requiredSkills []string, activeStates []domain.TicketState) ([]domain.AutoAssignmentCandidate, error)
	UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type TicketRepository interface {
	ListAll(ctx context.Context, limit, offset, sortVal int32) ([]domain.Ticket, error)
	ListAllFiltered(ctx context.Context, params domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error)
	ListByCreator(ctx context.Context, id uuid.UUID, limit, offset, sortVal int32) ([]domain.Ticket, error)
	ListByAssignee(ctx context.Context, id uuid.UUID, limit, offset, sortVal int32) ([]domain.Ticket, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.Ticket, error)
	GetByNumber(ctx context.Context, ticketNumber int64) (*domain.Ticket, error)
	GetActiveTickets(ctx context.Context) ([]domain.Ticket, error)
	Create(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error)
	Update(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error)
	Delete(ctx context.Context, id uuid.UUID) error

	CountTicketStatsAll(ctx context.Context, currentUserID uuid.UUID) (domain.TicketListStats, error)
	CountTicketStatsByCreator(ctx context.Context, creatorID, currentUserID uuid.UUID) (domain.TicketListStats, error)
	CountTicketStatsByAssignee(ctx context.Context, assigneeID, currentUserID uuid.UUID) (domain.TicketListStats, error)
}

type CommentRepository interface {
	ListByTicket(ctx context.Context, ticketID uuid.UUID, limit, offset int32) ([]domain.Comment, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.Comment, error)
	Create(ctx context.Context, comment domain.Comment) (*domain.Comment, error)
}
