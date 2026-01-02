package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/application/authorization"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/internal/ports"
)

type CommentService struct {
	repo       ports.CommentRepository
	ticketRepo ports.TicketRepository
}

func NewCommentService(r ports.CommentRepository, tr ports.TicketRepository) *CommentService {
	return &CommentService{repo: r, ticketRepo: tr}
}

func (s *CommentService) ListByTicket(ctx context.Context, ticketID uuid.UUID, limit, offset int32) ([]domain.Comment, error) {
	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check if user can view ticket
	ticket, err := s.ticketRepo.Get(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if !authorization.CanViewTicket(auth, ticket) {
		return nil, authorization.ErrAccessDenied
	}

	return s.repo.ListByTicket(ctx, ticketID, limit, offset)
}

func (s *CommentService) GetComment(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	return s.repo.Get(ctx, id)
}

func (s *CommentService) CreateComment(ctx context.Context, comment domain.Comment) (*domain.Comment, error) {
	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check if user can comment on ticket
	ticket, err := s.ticketRepo.Get(ctx, comment.TicketID)
	if err != nil {
		return nil, err
	}

	if !authorization.CanCommentOnTicket(auth, ticket) {
		return nil, authorization.ErrAccessDenied
	}
	comment.UpdatedAt = time.Now()
	return s.repo.Create(ctx, comment)
}
