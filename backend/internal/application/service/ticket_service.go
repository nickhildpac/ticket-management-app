package service

import (
	"context"
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

func (s *TicketService) GetTicket(ctx context.Context, id int64) (*domain.Ticket, error) {
	return s.repo.Get(ctx, id)
}

func (s *TicketService) CreateTicket(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
	ticket.State = domain.TicketStateOpen
	ticket.Priority = domain.TicketPriorityLow
	return s.repo.Create(ctx, ticket)
}

func (s *TicketService) UpdateTicket(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
	ticket.UpdatedAt = time.Now()
	return s.repo.Update(ctx, ticket)
}
