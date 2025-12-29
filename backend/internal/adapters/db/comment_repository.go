package db

import (
	"context"

	"github.com/google/uuid"
	sqlc "github.com/nickhildpac/ticket-management-app/internal/adapters/db/sqlc"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

type CommentRepository struct {
	store sqlc.Store
}

func NewCommentRepository(store sqlc.Store) *CommentRepository {
	return &CommentRepository{store: store}
}

func (r *CommentRepository) ListByTicket(ctx context.Context, ticketID uuid.UUID, limit, offset int32) ([]domain.Comment, error) {
	rows, err := r.store.ListComment(ctx, sqlc.ListCommentParams{
		TicketID: uuid.NullUUID{UUID: ticketID, Valid: true},
		Offset:   offset,
		Limit:    limit,
	})
	if err != nil {
		return nil, err
	}
	return mapComments(rows), nil
}

func (r *CommentRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	comment, err := r.store.GetComment(ctx, uuid.NullUUID{UUID: id, Valid: true})
	if err != nil {
		return nil, err
	}
	return mapComment(comment), nil
}

func (r *CommentRepository) Create(ctx context.Context, comment domain.Comment) (*domain.Comment, error) {
	created, err := r.store.CreateComment(ctx, sqlc.CreateCommentParams{
		TicketID:    uuid.NullUUID{UUID: comment.TicketID, Valid: true},
		Description: comment.Description,
		CreatedBy:   comment.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapComment(created), nil
}
