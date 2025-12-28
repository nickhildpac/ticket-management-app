package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	sqlc "github.com/nickhildpac/ticket-management-app/internal/adapters/db/sqlc"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

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

func (r *TicketRepository) ListByCreator(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error) {
	user, err := r.store.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	rows, err := r.store.ListTickets(ctx, sqlc.ListTicketsParams{
		CreatedBy: user.ID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, err
	}
	return mapTickets(rows), nil
}

func (r *TicketRepository) ListByAssignee(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error) {
	user, err := r.store.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	rows, err := r.store.ListTicketsAssigned(ctx, sqlc.ListTicketsAssignedParams{
		AssignedTo: uuid.NullUUID{UUID: user.ID, Valid: true},
		Limit:      limit,
		Offset:     offset,
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

func (r *TicketRepository) Update(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
	updated, err := r.store.UpdateTicket(ctx, sqlc.UpdateTicketParams{
		ID:          ticket.ID,
		Title:       ticket.Title,
		Description: ticket.Description,
		State:       int32(ticket.State),
		Priority:    int32(ticket.Priority),
		AssignedTo: uuid.NullUUID{
			UUID:  ticket.AssignedTo,
			Valid: ticket.AssignedTo != uuid.UUID{},
		},
		UpdatedAt: sql.NullTime{
			Time:  ticket.UpdatedAt,
			Valid: !ticket.UpdatedAt.IsZero(),
		},
	})
	if err != nil {
		return nil, err
	}
	return mapTicket(updated), nil
}
