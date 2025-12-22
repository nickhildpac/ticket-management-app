package service

import (
	"context"

	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/internal/ports"
)

type CommentService struct {
	repo ports.CommentRepository
}

func NewCommentService(r ports.CommentRepository) *CommentService {
	return &CommentService{repo: r}
}

func (s *CommentService) ListByTicket(ctx context.Context, ticketID int64, limit, offset int32) ([]domain.Comment, error) {
	return s.repo.ListByTicket(ctx, ticketID, limit, offset)
}

func (s *CommentService) GetComment(ctx context.Context, id int64) (*domain.Comment, error) {
	return s.repo.Get(ctx, id)
}

func (s *CommentService) CreateComment(ctx context.Context, comment domain.Comment) (*domain.Comment, error) {
	return s.repo.Create(ctx, comment)
}
