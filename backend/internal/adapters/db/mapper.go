package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	sqlc "github.com/nickhildpac/ticket-management-app/internal/adapters/db/sqlc"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

func mapUser(u sqlc.User) *domain.User {
	var role domain.UserRole
	if u.Role.Valid {
		role = domain.UserRole(u.Role.String)
	} else {
		role = domain.RoleUser // default role
	}

	return &domain.User{
		ID:             u.ID,
		HashedPassword: u.HashedPassword,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		Email:          u.Email,
		Role:           role,
		UpdatedAt:      u.UpdatedAt,
		CreatedAt:      u.CreatedAt,
	}
}

func mapTicket(t sqlc.Ticket) *domain.Ticket {
	return &domain.Ticket{
		ID:          nullableUUID(t.ID),
		CreatedBy:   t.CreatedBy,
		AssignedTo:  nullableUUID(t.AssignedTo),
		Title:       t.Title,
		Description: t.Description,
		State:       domain.TicketState(t.State),
		Priority:    domain.TicketPriority(t.Priority),
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
		ID:          nullableUUID(c.ID),
		TicketID:    nullableUUID(c.TicketID),
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

func nullableUUID(u uuid.NullUUID) uuid.UUID {
	if u.Valid {
		return u.UUID
	}
	return uuid.UUID{}
}
