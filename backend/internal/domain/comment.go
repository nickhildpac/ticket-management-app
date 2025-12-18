package domain

import (
	"time"
)

type Comment struct {
	ID          int64     `json:"id"`
	TicketID    int64     `json:"ticket_id"`
	CreatedBy   string    `json:"created_by"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
