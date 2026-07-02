package domain

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID          uuid.UUID `json:"id"`
	TicketID    uuid.UUID `json:"ticket_id"`
	CreatedBy   uuid.UUID `json:"created_by"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
