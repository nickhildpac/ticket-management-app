package handlers

import "time"

type TicketResponse struct {
	TicketID    int64     `json:"id"`
	CreatedBy   string    `json:"created_by"`
	AssignedTo  string    `json:"assigned_to"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	Priority    string    `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
