package db

import (
	"context"
	"fmt"
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
	skills := `{"incident-management"}`

	rows := sqlmock.NewRows([]string{"id", "created_by", "assigned_to", "title", "description", "state", "priority", "created_at", "updated_at", "ticket_number", "skills"}).
		AddRow(ticketID, creatorID, fmt.Sprintf("{%s}", assigned.String()), "title", "desc", int32(domain.TicketStateOpen), int32(domain.TicketPriorityHigh), now, now, int64(1001), skills)

	mock.ExpectQuery(`(?s)SELECT id, created_by, assigned_to, title, description, state, priority, created_at, updated_at, ticket_number, skills FROM tickets`).
		WithArgs(int32(domain.TicketListSortCreatedDesc), int32(0), int32(5)).
		WillReturnRows(rows)

	tickets, err := repo.ListAll(context.Background(), 5, 0, domain.TicketListSortCreatedDesc)
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
	skills := `{"incident-management"}`

	mock.ExpectQuery(`(?s).*UPDATE tickets\s+SET\s+title = \$2,\s+description = \$3,\s+state = \$4,\s+priority = \$5,\s+assigned_to = \$6,\s+updated_at = \$7,\s+skills = \$8\s+WHERE id = \$1\s+RETURNING id, created_by, assigned_to, title, description, state, priority, created_at, updated_at, ticket_number, skills`).
		WithArgs(ticketID, "new title", "new desc", int32(domain.TicketStatePending), int32(domain.TicketPriorityCritical), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_by", "assigned_to", "title", "description", "state", "priority", "created_at", "updated_at", "ticket_number", "skills"}).
			AddRow(ticketID, creatorID, fmt.Sprintf("{%s}", assigned.String()), "new title", "new desc", int32(domain.TicketStatePending), int32(domain.TicketPriorityCritical), now, now, int64(1001), skills))

	updated, err := repo.Update(context.Background(), domain.Ticket{
		ID:          ticketID,
		Title:       "new title",
		Description: "new desc",
		State:       domain.TicketStatePending,
		Priority:    domain.TicketPriorityCritical,
		AssignedTo:  []uuid.UUID{assigned},
		Skills:      domain.NewSkillsFromSlice([]string{"incident-management"}),
		UpdatedAt:   now,
	})
	require.NoError(t, err)
	require.Equal(t, domain.TicketPriorityCritical, updated.Priority)
	require.Equal(t, domain.TicketStatePending, updated.State)
	require.Equal(t, "new title", updated.Title)
	require.Contains(t, updated.AssignedTo, assigned)

	require.NoError(t, mock.ExpectationsWereMet())
}
