package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/application/apperrors"
	"github.com/nickhildpac/ticket-management-app/internal/application/authorization"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/internal/ports"
)

type TicketService struct {
	repo              ports.TicketRepository
	autoAssignmentSvc *AutoAssignmentService
}

// NewTicketService constructs the service. Domain events (ticket.created /
// ticket.updated) are written by the repository in the same transaction as the
// ticket write via CreateWithEvent/UpdateWithEvent — a true transactional
// outbox, so an event is never silently lost and a failed event write fails
// the request before anything commits.
func NewTicketService(repo ports.TicketRepository, autoAssignmentSvc *AutoAssignmentService) *TicketService {
	return &TicketService{
		repo:              repo,
		autoAssignmentSvc: autoAssignmentSvc,
	}
}

func (s *TicketService) ListAll(ctx context.Context, limit, offset, sortVal int32) ([]domain.Ticket, error) {
	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	// Admins can see all tickets
	if auth.Role == domain.RoleAdmin {
		return s.repo.ListAll(ctx, limit, offset, sortVal)
	}

	// Users can only see their own tickets
	if auth.Role == domain.RoleUser {
		return s.repo.ListByCreator(ctx, auth.UserID, limit, offset, sortVal)
	}

	// Agents can only see assigned tickets
	if auth.Role == domain.RoleAgent {
		return s.repo.ListByAssignee(ctx, auth.UserID, limit, offset, sortVal)
	}

	return nil, authorization.ErrAccessDenied
}

// ListTicketsWithFilters lists tickets using SQL filters. Role rules:
//   - Admin: all filter fields honored.
//   - Agent: scoped to tickets assigned to the current user; rejects created_by or a different assignee filter.
//   - User: scoped to tickets created by the current user; rejects assignee or a different created_by filter.
func (s *TicketService) ListTicketsWithFilters(ctx context.Context, params domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error) {
	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	p := params

	switch auth.Role {
	case domain.RoleAdmin:
		return s.repo.ListAllFiltered(ctx, p)
	case domain.RoleAgent:
		if p.FilterCreatedBy.Valid {
			return nil, authorization.ErrAccessDenied
		}
		if p.FilterAssignee.Valid && p.FilterAssignee.UUID != auth.UserID {
			return nil, authorization.ErrAccessDenied
		}
		p.FilterAssignee = uuid.NullUUID{UUID: auth.UserID, Valid: true}
		return s.repo.ListAllFiltered(ctx, p)
	case domain.RoleUser:
		if p.FilterAssignee.Valid {
			return nil, authorization.ErrAccessDenied
		}
		if p.FilterCreatedBy.Valid && p.FilterCreatedBy.UUID != auth.UserID {
			return nil, authorization.ErrAccessDenied
		}
		p.FilterCreatedBy = uuid.NullUUID{UUID: auth.UserID, Valid: true}
		return s.repo.ListAllFiltered(ctx, p)
	default:
		return nil, authorization.ErrAccessDenied
	}
}

// ListAssignedToCurrentUser returns tickets where assigned_to contains the authenticated user,
// with optional state/priority/ticket_number and pagination. Strips creator filter; overwrites assignee with the current user.
func (s *TicketService) ListAssignedToCurrentUser(ctx context.Context, params domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error) {
	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}
	p := params
	p.FilterCreatedBy = uuid.NullUUID{}
	p.FilterAssignee = uuid.NullUUID{UUID: auth.UserID, Valid: true}
	return s.repo.ListAllFiltered(ctx, p)
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

	return s.repo.ListByCreator(ctx, id, limit, offset, domain.TicketListSortCreatedDesc)
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

	return s.repo.ListByAssignee(ctx, id, limit, offset, domain.TicketListSortCreatedDesc)
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

func (s *TicketService) GetTicketByNumber(ctx context.Context, ticketNumber int64) (*domain.Ticket, error) {
	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	ticket, err := s.repo.GetByNumber(ctx, ticketNumber)
	if err != nil {
		return nil, err
	}

	if !authorization.CanViewTicket(auth, ticket) {
		return nil, authorization.ErrAccessDenied
	}

	return ticket, nil
}

// GetTicketStats returns aggregate counts scoped like ListAll (admin: all tickets, user: created by self, agent: assigned to self).
func (s *TicketService) GetTicketStats(ctx context.Context) (domain.TicketListStats, error) {
	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return domain.TicketListStats{}, err
	}

	switch auth.Role {
	case domain.RoleAdmin:
		return s.repo.CountTicketStatsAll(ctx, auth.UserID)
	case domain.RoleUser:
		return s.repo.CountTicketStatsByCreator(ctx, auth.UserID, auth.UserID)
	case domain.RoleAgent:
		return s.repo.CountTicketStatsByAssignee(ctx, auth.UserID, auth.UserID)
	default:
		return domain.TicketListStats{}, authorization.ErrAccessDenied
	}
}

func (s *TicketService) CreateTicket(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
	ticket.State = domain.TicketStateOpen
	if ticket.Priority == 0 {
		ticket.Priority = domain.TicketPriorityLow
	}
	ticket.UpdatedAt = time.Now()

	// Auto-assignment only needs the ticket's skills, so resolve it before the
	// insert: ticket, assignment, and the ticket.created event then commit in a
	// single transaction, and the event carries the final (possibly assigned)
	// state. Assignment failures are logged and the ticket proceeds unassigned.
	if len(ticket.Skills.ToSlice()) > 0 && s.autoAssignmentSvc != nil {
		bestAgent, err := s.autoAssignmentSvc.FindBestAgentForTicket(ctx, &ticket)
		if err != nil {
			log.Printf("Auto-assignment failed: %v", err)
		} else if bestAgent != nil {
			ticket.AssignedTo = []uuid.UUID{bestAgent.ID}
			ticket.State = domain.TicketStatePending
		}
	}

	return s.repo.CreateWithEvent(ctx, ticket, domain.EventTicketCreated)
}

func (s *TicketService) UpdateTicket(ctx context.Context, id uuid.UUID, patch domain.TicketPatch) (*domain.Ticket, error) {
	if patch.IsEmpty() {
		return nil, fmt.Errorf("%w: no fields provided to update", apperrors.ErrBadRequest)
	}

	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	prev, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user can update this ticket at all
	if !authorization.CanUpdateTicket(auth, prev) {
		return nil, authorization.ErrAccessDenied
	}

	ticket := *prev
	if patch.Title != nil {
		ticket.Title = *patch.Title
	}
	if patch.Description != nil {
		ticket.Description = *patch.Description
	}
	if patch.State != nil {
		ticket.State = *patch.State
	}
	if patch.Priority != nil {
		ticket.Priority = *patch.Priority
	}
	if patch.AssignedTo != nil {
		ticket.AssignedTo = *patch.AssignedTo
	}
	if patch.Skills != nil {
		ticket.Skills = *patch.Skills
	}

	if patch.State != nil {
		// Pass the updated/new ticket state so state-level authorization can
		// enforce allowed target states (e.g., user may close/cancel).
		if !authorization.CanUpdateTicketState(auth, &ticket) {
			return nil, authorization.ErrAccessDenied
		}
	}
	if patch.Priority != nil && !authorization.CanUpdateTicketPriority(auth, prev) {
		return nil, authorization.ErrAccessDenied
	}
	if patch.AssignedTo != nil && !authorization.CanAssignTicket(auth, prev) {
		return nil, authorization.ErrAccessDenied
	}

	// Auto-transition to pending when assigned
	if len(ticket.AssignedTo) > 0 && len(prev.AssignedTo) == 0 {
		ticket.State = domain.TicketStatePending
	}

	// State transition validation
	if ticket.State != prev.State {
		log.Printf("Attempting state transition from %s to %s", prev.State, ticket.State)
		if ok := domain.CanTransition(prev.State, ticket.State); !ok {
			return nil, domain.ErrInvalidStatusTransition
		}
	}

	ticket.CreatedAt = prev.CreatedAt
	ticket.UpdatedAt = time.Now()

	// Write ticket.updated only when the lifecycle state actually changed, so the
	// AI/RAG service re-triages on meaningful transitions (not on title edits).
	// The event commits in the same transaction as the update.
	if ticket.State != prev.State {
		return s.repo.UpdateWithEvent(ctx, ticket, domain.EventTicketUpdated)
	}
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
