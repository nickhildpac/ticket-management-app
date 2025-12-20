package port

import (
	"context"

	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

type CommentRepository interface {
	ListByTicket(ctx context.Context, ticketID int64, limit, offset int32) ([]domain.Comment, error)
	Get(ctx context.Context, id int64) (*domain.Comment, error)
	Create(ctx context.Context, comment domain.Comment) (*domain.Comment, error)
}
