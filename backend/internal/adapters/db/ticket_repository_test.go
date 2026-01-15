package db

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	sqlcdb "github.com/nickhildpac/ticket-management-app/internal/adapters/db/sqlc"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestTicketRepository_ListAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := sqlcdb.NewStore(db)
	repo := NewTicketRepository(store)

	now := time.Now()
	ticketID := uuid.New()
	creatorID := uuid.New()
	assigned := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "created_by", "assigned_to", "title", "description", "state", "priority", "created_at", "updated_at"}).
		AddRow(ticketID, creatorID, fmt.Sprintf("{%s}", assigned.String()), "title", "desc", int32(domain.TicketStateOpen), int32(domain.TicketPriorityHigh), now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, created_by, assigned_to, title, description, state, priority, created_at, updated_at FROM tickets ORDER BY id LIMIT $1 OFFSET $2")).
		WithArgs(int32(5), int32(0)).
		WillReturnRows(rows)

	tickets, err := repo.ListAll(context.Background(), 5, 0)
	require.NoError(t, err)
	require.Len(t, tickets, 1)
	require.Equal(t, ticketID, tickets[0].ID)
	require.Equal(t, creatorID, tickets[0].CreatedBy)
	require.Contains(t, tickets[0].AssignedTo, assigned)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTicketRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := sqlcdb.NewStore(db)
	repo := NewTicketRepository(store)

	now := time.Now()
	ticketID := uuid.New()
	creatorID := uuid.New()
	assigned := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE tickets
SET 
    title = $2,
    description = $3,
    state = $4,
    priority = $5,
    assigned_to = $6,
    updated_at = $7
WHERE id = $1
RETURNING id, created_by, assigned_to, title, description, state, priority, created_at, updated_at`)).
		WithArgs(ticketID, "new title", "new desc", int32(domain.TicketStatePending), int32(domain.TicketPriorityCritical), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_by", "assigned_to", "title", "description", "state", "priority", "created_at", "updated_at"}).
			AddRow(ticketID, creatorID, fmt.Sprintf("{%s}", assigned.String()), "new title", "new desc", int32(domain.TicketStatePending), int32(domain.TicketPriorityCritical), now, now))

	updated, err := repo.Update(context.Background(), domain.Ticket{
		ID:          ticketID,
		Title:       "new title",
		Description: "new desc",
		State:       domain.TicketStatePending,
		Priority:    domain.TicketPriorityCritical,
		AssignedTo:  []uuid.UUID{assigned},
		UpdatedAt:   now,
	})
	require.NoError(t, err)
	require.Equal(t, domain.TicketPriorityCritical, updated.Priority)
	require.Equal(t, domain.TicketStatePending, updated.State)
	require.Equal(t, "new title", updated.Title)
	require.Contains(t, updated.AssignedTo, assigned)

	require.NoError(t, mock.ExpectationsWereMet())
}
