package handlers

import (
	"time"

	"github.com/google/uuid"
)

type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
}

type TicketResponse struct {
	TicketID    uuid.UUID   `json:"id"`
	CreatedBy   uuid.UUID   `json:"created_by"`
	Creator     UserInfo    `json:"creator"`
	AssignedTo  []uuid.UUID `json:"assigned_to"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	State       string      `json:"state"`
	Priority    string      `json:"priority"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type CommentResponse struct {
	ID        uuid.UUID `json:"id"`
	TicketID  uuid.UUID `json:"ticket_id"`
	CreatedBy uuid.UUID `json:"created_by"`
	Creator   UserInfo  `json:"creator"`
	Description string `json:"description"`
	CreatedAt time.Time `json:"created_at"`
}
