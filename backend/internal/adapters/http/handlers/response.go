package handlers

import (
	"time"

	"github.com/google/uuid"
)

type TicketResponse struct {
	TicketID    uuid.UUID   `json:"id"`
	CreatedBy   uuid.UUID   `json:"created_by"`
	AssignedTo  []uuid.UUID `json:"assigned_to"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	State       string      `json:"state"`
	Priority    string      `json:"priority"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
