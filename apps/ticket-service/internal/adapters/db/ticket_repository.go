package db

import (
	"context"
	"database/sql"
	"encoding/json"

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

func (r *TicketRepository) ListAll(ctx context.Context, limit, offset, sortVal int32) ([]domain.Ticket, error) {
	rows, err := r.store.ListAllTickets(ctx, sqlc.ListAllTicketsParams{Limit: limit, Offset: offset, SortVal: sortVal})
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
		SortVal:            params.SortVal,
		OffsetVal:          params.OffsetVal,
		LimitVal:           params.LimitVal,
	})
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return mapTickets(rows), nil
}

func (r *TicketRepository) ListByCreator(ctx context.Context, id uuid.UUID, limit, offset, sortVal int32) ([]domain.Ticket, error) {
	user, err := r.store.GetUser(ctx, id)
	if err != nil {
		return nil, normalizeDBError(err)
	}
	rows, err := r.store.ListTickets(ctx, sqlc.ListTicketsParams{
		CreatedBy: user.ID,
		SortVal:   sortVal,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return mapTickets(rows), nil
}

func (r *TicketRepository) ListByAssignee(ctx context.Context, id uuid.UUID, limit, offset, sortVal int32) ([]domain.Ticket, error) {
	user, err := r.store.GetUser(ctx, id)
	if err != nil {
		return nil, normalizeDBError(err)
	}
	rows, err := r.store.ListTicketsAssigned(ctx, sqlc.ListTicketsAssignedParams{
		AssigneeIds: []uuid.UUID{user.ID},
		SortVal:     sortVal,
		Limit:       limit,
		Offset:      offset,
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
		Priority:    int32(ticket.Priority),
	})
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return mapTicket(created), nil
}

// CreateWithEvent inserts a ticket and its domain event in a single transaction,
// so the outbox row commits atomically with the ticket write (a real
// transactional outbox — the event can never be silently lost). The event is
// built from the inserted row and carries the given eventType.
func (r *TicketRepository) CreateWithEvent(ctx context.Context, ticket domain.Ticket, eventType string) (*domain.Ticket, error) {
	var created *domain.Ticket
	err := r.store.ExecTx(ctx, func(tx *sql.Tx, q *sqlc.Queries) error {
		row, err := q.CreateTicket(ctx, sqlc.CreateTicketParams{
			Title:       ticket.Title,
			Description: ticket.Description,
			CreatedBy:   ticket.CreatedBy,
			UpdatedAt:   ticket.UpdatedAt,
			Skills:      ticket.Skills.ToSlice(),
			Priority:    int32(ticket.Priority),
		})
		if err != nil {
			return err
		}
		created = mapTicket(row)
		return insertOutbox(ctx, tx, domain.NewTicketEvent(eventType, created))
	})
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return created, nil
}

// UpdateWithEvent updates a ticket and writes its domain event atomically. See
// CreateWithEvent.
func (r *TicketRepository) UpdateWithEvent(ctx context.Context, ticket domain.Ticket, eventType string) (*domain.Ticket, error) {
	var updated *domain.Ticket
	err := r.store.ExecTx(ctx, func(tx *sql.Tx, q *sqlc.Queries) error {
		row, err := q.UpdateTicket(ctx, sqlc.UpdateTicketParams{
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
			return err
		}
		updated = mapTicket(row)
		return insertOutbox(ctx, tx, domain.NewTicketEvent(eventType, updated))
	})
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return updated, nil
}

// insertOutbox writes a domain event to the outbox within the given transaction.
func insertOutbox(ctx context.Context, tx *sql.Tx, event domain.OutboxEvent) error {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO event_outbox (event_type, aggregate_id, payload) VALUES ($1, $2, $3)`,
		event.Type, event.AggregateID, payload)
	return err
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

func (r *TicketRepository) CountTicketStatsAll(ctx context.Context, currentUserID uuid.UUID) (domain.TicketListStats, error) {
	row, err := r.store.CountTicketStatsAll(ctx, []uuid.UUID{currentUserID})
	if err != nil {
		return domain.TicketListStats{}, normalizeDBError(err)
	}
	return domain.TicketListStats{
		Total: row.Total, Open: row.Open, Pending: row.Pending, Resolved: row.Resolved, Mine: row.Mine,
	}, nil
}

func (r *TicketRepository) CountTicketStatsByCreator(ctx context.Context, creatorID, currentUserID uuid.UUID) (domain.TicketListStats, error) {
	row, err := r.store.CountTicketStatsByCreator(ctx, sqlc.CountTicketStatsByCreatorParams{
		CreatedBy: creatorID,
		Column2:   []uuid.UUID{currentUserID},
	})
	if err != nil {
		return domain.TicketListStats{}, normalizeDBError(err)
	}
	return domain.TicketListStats{
		Total: row.Total, Open: row.Open, Pending: row.Pending, Resolved: row.Resolved, Mine: row.Mine,
	}, nil
}

func (r *TicketRepository) CountTicketStatsByAssignee(ctx context.Context, assigneeID, currentUserID uuid.UUID) (domain.TicketListStats, error) {
	row, err := r.store.CountTicketStatsByAssignee(ctx, sqlc.CountTicketStatsByAssigneeParams{
		Column1: []uuid.UUID{assigneeID},
		Column2: []uuid.UUID{currentUserID},
	})
	if err != nil {
		return domain.TicketListStats{}, normalizeDBError(err)
	}
	return domain.TicketListStats{
		Total: row.Total, Open: row.Open, Pending: row.Pending, Resolved: row.Resolved, Mine: row.Mine,
	}, nil
}
