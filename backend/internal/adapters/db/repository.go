package db

import (
	"context"
	"database/sql"
	"time"

	sqlc "github.com/nickhildpac/ticket-management-app/internal/adapters/db/sqlc"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/internal/usecase/port"
)

type UserRepository struct {
	store sqlc.Store
}

func NewUserRepository(store sqlc.Store) *UserRepository {
	return &UserRepository{store: store}
}

func (r *UserRepository) GetUser(ctx context.Context, username string) (*domain.User, error) {
	user, err := r.store.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}
	return mapUser(user), nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user domain.User) (*domain.User, error) {
	created, err := r.store.CreateUser(ctx, sqlc.CreateUserParams{
		Username:       user.Username,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
	})
	if err != nil {
		return nil, err
	}
	return mapUser(created), nil
}

type TicketRepository struct {
	store sqlc.Store
}

func NewTicketRepository(store sqlc.Store) *TicketRepository {
	return &TicketRepository{store: store}
}

func (r *TicketRepository) ListAll(ctx context.Context, limit, offset int32) ([]domain.Ticket, error) {
	rows, err := r.store.ListAllTickets(ctx, sqlc.ListAllTicketsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return mapTickets(rows), nil
}

func (r *TicketRepository) ListByCreator(ctx context.Context, username string, limit, offset int32) ([]domain.Ticket, error) {
	rows, err := r.store.ListTickets(ctx, sqlc.ListTicketsParams{
		CreatedBy: username,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, err
	}
	return mapTickets(rows), nil
}

func (r *TicketRepository) Get(ctx context.Context, id int64) (*domain.Ticket, error) {
	ticket, err := r.store.GetTicket(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapTicket(ticket), nil
}

func (r *TicketRepository) Create(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
	created, err := r.store.CreateTicket(ctx, sqlc.CreateTicketParams{
		Title:       ticket.Title,
		Description: ticket.Description,
		CreatedBy:   ticket.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapTicket(created), nil
}

type CommentRepository struct {
	store sqlc.Store
}

func NewCommentRepository(store sqlc.Store) *CommentRepository {
	return &CommentRepository{store: store}
}

func (r *CommentRepository) ListByTicket(ctx context.Context, ticketID int64, limit, offset int32) ([]domain.Comment, error) {
	rows, err := r.store.ListComment(ctx, sqlc.ListCommentParams{
		TicketID: ticketID,
		Offset:   offset,
		Limit:    limit,
	})
	if err != nil {
		return nil, err
	}
	return mapComments(rows), nil
}

func (r *CommentRepository) Get(ctx context.Context, id int64) (*domain.Comment, error) {
	comment, err := r.store.GetComment(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapComment(comment), nil
}

func (r *CommentRepository) Create(ctx context.Context, comment domain.Comment) (*domain.Comment, error) {
	created, err := r.store.CreateComment(ctx, sqlc.CreateCommentParams{
		TicketID:    comment.TicketID,
		Description: comment.Description,
		CreatedBy:   comment.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return mapComment(created), nil
}

func mapUser(u sqlc.User) *domain.User {
	return &domain.User{
		Username:          u.Username,
		HashedPassword:    u.HashedPassword,
		FirstName:         u.FirstName,
		LastName:          u.LastName,
		Email:             u.Email,
		Role:              u.Role.String,
		PasswordChangedAt: u.PasswordChangedAt,
		CreatedAt:         u.CreatedAt,
	}
}

func mapTicket(t sqlc.Ticket) *domain.Ticket {
	return &domain.Ticket{
		ID:          t.ID,
		CreatedBy:   t.CreatedBy,
		AssignedTo:  t.AssignedTo.String,
		Title:       t.Title,
		Description: t.Description,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   nullableTime(t.UpdatedAt),
	}
}

func mapTickets(rows []sqlc.Ticket) []domain.Ticket {
	out := make([]domain.Ticket, 0, len(rows))
	for _, t := range rows {
		out = append(out, *mapTicket(t))
	}
	return out
}

func mapComment(c sqlc.Comment) *domain.Comment {
	return &domain.Comment{
		ID:          c.ID,
		TicketID:    c.TicketID,
		CreatedBy:   c.CreatedBy,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   nullableTime(c.UpdatedAt),
	}
}

func mapComments(rows []sqlc.Comment) []domain.Comment {
	out := make([]domain.Comment, 0, len(rows))
	for _, c := range rows {
		out = append(out, *mapComment(c))
	}
	return out
}

func nullableTime(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}

var (
	_ port.UserRepository    = (*UserRepository)(nil)
	_ port.TicketRepository  = (*TicketRepository)(nil)
	_ port.CommentRepository = (*CommentRepository)(nil)
)
