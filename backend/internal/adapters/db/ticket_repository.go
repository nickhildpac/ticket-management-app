package db

import (
	"context"

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
		return nil, normalizeDBError(err)
	}
	return mapTickets(rows), nil
}

func (r *TicketRepository) ListAllFiltered(ctx context.Context, params domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error) {
	rows, err := r.store.ListAllTicketsByStatePriority(ctx, sqlc.ListAllTicketsByStatePriorityParams{
		FilterState:        params.FilterState,
		FilterPriority:     params.FilterPriority,
		FilterCreatedBy:    params.FilterCreatedBy,
		FilterAssignee:     params.FilterAssignee,
		FilterTicketNumber: params.FilterTicketNumber,
		OffsetVal:          params.OffsetVal,
		LimitVal:           params.LimitVal,
	})
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return mapTickets(rows), nil
}

func (r *TicketRepository) ListByCreator(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error) {
	user, err := r.store.GetUser(ctx, id)
	if err != nil {
		return nil, normalizeDBError(err)
	}
	rows, err := r.store.ListTickets(ctx, sqlc.ListTicketsParams{
		CreatedBy: user.ID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return mapTickets(rows), nil
}

func (r *TicketRepository) ListByAssignee(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error) {
	user, err := r.store.GetUser(ctx, id)
	if err != nil {
		return nil, normalizeDBError(err)
	}
	rows, err := r.store.ListTicketsAssigned(ctx, sqlc.ListTicketsAssignedParams{
		Column1: []uuid.UUID{user.ID},
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return mapTickets(rows), nil
}

func (r *TicketRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	ticket, err := r.store.GetTicket(ctx, id)
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return mapTicket(ticket), nil
}

func (r *TicketRepository) GetByNumber(ctx context.Context, ticketNumber int64) (*domain.Ticket, error) {
	ticket, err := r.store.GetTicketByNumber(ctx, ticketNumber)
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return mapTicket(ticket), nil
}

func (r *TicketRepository) Create(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
	created, err := r.store.CreateTicket(ctx, sqlc.CreateTicketParams{
		Title:       ticket.Title,
		Description: ticket.Description,
		CreatedBy:   ticket.CreatedBy,
		UpdatedAt:   ticket.UpdatedAt,
		Skills:      ticket.Skills.ToSlice(),
	})
	if err != nil {
		return nil, normalizeDBError(err)
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
		AssignedTo:  ticket.AssignedTo,
		UpdatedAt:   ticket.UpdatedAt,
		Skills:      ticket.Skills.ToSlice(),
	})
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return mapTicket(updated), nil
}

func (r *TicketRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return normalizeDBError(r.store.DeleteTicket(ctx, id))
}

func (r *TicketRepository) GetActiveTickets(ctx context.Context) ([]domain.Ticket, error) {
	rows, err := r.store.GetActiveTickets(ctx)
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return mapTickets(rows), nil
}
