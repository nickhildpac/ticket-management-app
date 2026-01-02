package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/application/authorization"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/internal/ports"
)

type TicketService struct {
	repo ports.TicketRepository
}

func NewTicketService(repo ports.TicketRepository) *TicketService {
	return &TicketService{repo: repo}
}

func (s *TicketService) ListAll(ctx context.Context, limit, offset int32) ([]domain.Ticket, error) {
	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	// Admins can see all tickets
	if auth.Role == domain.RoleAdmin {
		return s.repo.ListAll(ctx, limit, offset)
	}

	// Users can only see their own tickets
	if auth.Role == domain.RoleUser {
		return s.repo.ListByCreator(ctx, auth.UserID, limit, offset)
	}

	// Agents can only see assigned tickets
	if auth.Role == domain.RoleAgent {
		return s.repo.ListByAssignee(ctx, auth.UserID, limit, offset)
	}

	return nil, authorization.ErrAccessDenied
}

func (s *TicketService) ListByCreator(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error) {
	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	// Only admins or the user themselves can list by creator
	if auth.Role != domain.RoleAdmin && auth.UserID != id {
		return nil, authorization.ErrAccessDenied
	}

	return s.repo.ListByCreator(ctx, id, limit, offset)
}

func (s *TicketService) ListByAssignee(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error) {
	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	// Only admins or the agent themselves can list by assignee
	if auth.Role != domain.RoleAdmin && auth.UserID != id {
		return nil, authorization.ErrAccessDenied
	}

	return s.repo.ListByAssignee(ctx, id, limit, offset)
}

func (s *TicketService) GetTicket(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	ticket, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if !authorization.CanViewTicket(auth, ticket) {
		return nil, authorization.ErrAccessDenied
	}

	return ticket, nil
}

func (s *TicketService) CreateTicket(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
	ticket.State = domain.TicketStateOpen
	ticket.Priority = domain.TicketPriorityLow
	ticket.UpdatedAt = time.Now()
	return s.repo.Create(ctx, ticket)
}

func (s *TicketService) UpdateTicket(ctx context.Context, ticket domain.Ticket, updatedFields []string) (*domain.Ticket, error) {
	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	prev, err := s.repo.Get(ctx, ticket.ID)
	if err != nil {
		return nil, err
	}

	// Check if user can update this ticket at all
	if !authorization.CanUpdateTicket(auth, prev) {
		return nil, authorization.ErrAccessDenied
	}

	// Field-level authorization
	for _, field := range updatedFields {
		switch field {
		case "state":
			if !authorization.CanUpdateTicketState(auth, prev) {
				return nil, authorization.ErrAccessDenied
			}
		case "priority":
			if !authorization.CanUpdateTicketPriority(auth, prev) {
				return nil, authorization.ErrAccessDenied
			}
		case "assigned_to":
			if !authorization.CanAssignTicket(auth, prev) {
				return nil, authorization.ErrAccessDenied
			}
		}
	}

	// State transition validation
	if ticket.State != prev.State {
		log.Printf("Attempting state transition from %s to %s", prev.State, ticket.State)
		if ok := domain.CanTransition(prev.State, ticket.State); !ok {
			return nil, domain.ErrInvalidStatusTransition
		}
	}

	// Auto-transition to pending when assigned
	if len(ticket.AssignedTo) > 0 && len(prev.AssignedTo) == 0 {
		ticket.State = domain.TicketStatePending
	}

	ticket.CreatedAt = prev.CreatedAt
	ticket.UpdatedAt = time.Now()
	return s.repo.Update(ctx, ticket)
}

func (s *TicketService) DeleteTicket(ctx context.Context, id uuid.UUID) error {
	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return err
	}

	ticket, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	if !authorization.CanDeleteTicket(auth, ticket) {
		return authorization.ErrAccessDenied
	}

	return s.repo.Delete(ctx, id)
}
