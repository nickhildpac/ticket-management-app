package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
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
	return s.repo.ListAll(ctx, limit, offset)
}

func (s *TicketService) ListByCreator(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error) {
	return s.repo.ListByCreator(ctx, id, limit, offset)
}

func (s *TicketService) ListByAssignee(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error) {
	return s.repo.ListByAssignee(ctx, id, limit, offset)
}

func (s *TicketService) GetTicket(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	return s.repo.Get(ctx, id)
}

func (s *TicketService) CreateTicket(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
	ticket.State = domain.TicketStateOpen
	ticket.Priority = domain.TicketPriorityLow
	return s.repo.Create(ctx, ticket)
}

func (s *TicketService) UpdateTicket(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
	prev, err := s.repo.Get(ctx, ticket.ID)
	if err != nil {
		return nil, err
	}

	if len(ticket.AssignedTo) > 0 && len(prev.AssignedTo) == 0 {
		ticket.State = domain.TicketStatePending
	}
	// Only validate state transition if state is being changed
	if ticket.State != prev.State {
		log.Printf("Attempting state transition from %s to %s", prev.State, ticket.State)
		if ok := domain.CanTransition(prev.State, ticket.State); !ok {
			return nil, domain.ErrInvalidStatusTransition
		}
	}

	// Update the ticket
	ticket.CreatedAt = prev.CreatedAt // Preserve original creation time
	ticket.UpdatedAt = time.Now()
	return s.repo.Update(ctx, ticket)
}
