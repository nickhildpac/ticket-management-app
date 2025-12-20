package usecase

import (
	"context"

	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/internal/usecase/port"
)

type TicketService struct {
	repo port.TicketRepository
}

func NewTicketService(r port.TicketRepository) *TicketService {
	return &TicketService{repo: r}
}

func (s *TicketService) ListAll(ctx context.Context, limit, offset int32) ([]domain.Ticket, error) {
	return s.repo.ListAll(ctx, limit, offset)
}

func (s *TicketService) ListByCreator(ctx context.Context, username string, limit, offset int32) ([]domain.Ticket, error) {
	return s.repo.ListByCreator(ctx, username, limit, offset)
}

func (s *TicketService) GetTicket(ctx context.Context, id int64) (*domain.Ticket, error) {
	return s.repo.Get(ctx, id)
}

func (s *TicketService) CreateTicket(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
	return s.repo.Create(ctx, ticket)
}

var _ port.TicketService = (*TicketService)(nil)
