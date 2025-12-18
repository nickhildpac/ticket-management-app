package domain

import (
	"time"
)

type Ticket struct {
	ID          int64     `json:"id"`
	CreatedBy   string    `json:"created_by"`
	AssignedTo  string    `json:"assigned_to"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
