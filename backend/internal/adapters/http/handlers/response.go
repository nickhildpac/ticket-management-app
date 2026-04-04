package handlers

import (
	"time"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

// UserInfo represents basic user information
type UserInfo struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" format:"uuid"`
	FirstName string    `json:"first_name" example:"John"`
	LastName  string    `json:"last_name" example:"Doe"`
	Email     string    `json:"email" example:"john.doe@example.com" format:"email"`
}

// AssignmentUserResponse represents a user entry for assignment selection.
type AssignmentUserResponse struct {
	ID        uuid.UUID       `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" format:"uuid"`
	FirstName string          `json:"first_name" example:"John"`
	LastName  string          `json:"last_name" example:"Doe"`
	Email     string          `json:"email" example:"john.doe@example.com" format:"email"`
	Role      domain.UserRole `json:"role" example:"agent" enum:"user,agent,admin"`
}

// UserResponse represents a user payload returned by user-facing endpoints.
type UserResponse struct {
	ID        uuid.UUID       `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" format:"uuid"`
	FirstName string          `json:"first_name" example:"John"`
	LastName  string          `json:"last_name" example:"Doe"`
	Email     string          `json:"email" example:"john.doe@example.com" format:"email"`
	Role      domain.UserRole `json:"role" example:"agent" enum:"user,agent,admin"`
	Skills    []string        `json:"skills" example:"incident-management,log-analysis"`
	CreatedAt time.Time       `json:"created_at" example:"2025-01-22T10:30:00Z" format:"date-time"`
	UpdatedAt time.Time       `json:"updated_at" example:"2025-01-22T11:45:00Z" format:"date-time"`
}

// TicketResponse represents a complete ticket with all details
type TicketResponse struct {
	TicketID     uuid.UUID   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" format:"uuid"`
	TicketNumber int64       `json:"ticket_number" example:"1001"`
	CreatedBy    uuid.UUID   `json:"created_by" example:"550e8400-e29b-41d4-a716-446655440000" format:"uuid"`
	Creator      UserInfo    `json:"creator"`
	AssignedTo   []uuid.UUID `json:"assigned_to" example:"550e8400-e29b-41d4-a716-446655440000"`
	Skills       []string    `json:"skills" example:"golang,postgresql"`
	Title        string      `json:"title" example:"Fix login bug"`
	Description  string      `json:"description" example:"Users cannot login with SSO"`
	State        string      `json:"state" example:"open" enum:"open,pending,resolved,closed,cancelled"`
	Priority     string      `json:"priority" example:"high" enum:"critical,high,medium,low"`
	CreatedAt    time.Time   `json:"created_at" example:"2025-01-22T10:30:00Z" format:"date-time"`
	UpdatedAt    time.Time   `json:"updated_at" example:"2025-01-22T11:45:00Z" format:"date-time"`
}

// TicketSummaryResponse is a lightweight response for ticket list endpoints
type TicketSummaryResponse struct {
	ID           uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" format:"uuid"`
	TicketNumber int64     `json:"ticket_number" example:"1001"`
	Title        string    `json:"title" example:"Fix login bug"`
	Description  string    `json:"description" example:"Users cannot login with SSO"`
	State        string    `json:"state" example:"open" enum:"open,pending,resolved,closed,cancelled"`
	Priority     string    `json:"priority" example:"high" enum:"critical,high,medium,low"`
	CreatedAt    time.Time `json:"created_at" example:"2025-01-22T10:30:00Z" format:"date-time"`
	UpdatedAt    time.Time `json:"updated_at" example:"2025-01-22T11:45:00Z" format:"date-time"`
}

// CommentResponse represents a comment on a ticket
type CommentResponse struct {
	ID          uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" format:"uuid"`
	TicketID    uuid.UUID `json:"ticket_id" example:"550e8400-e29b-41d4-a716-446655440000" format:"uuid"`
	CreatedBy   uuid.UUID `json:"created_by" example:"550e8400-e29b-41d4-a716-446655440000" format:"uuid"`
	Creator     UserInfo  `json:"creator"`
	Description string    `json:"description" example:"I'm working on this issue"`
	CreatedAt   time.Time `json:"created_at" example:"2025-01-22T12:00:00Z" format:"date-time"`
}

func newUserInfo(user *domain.User) UserInfo {
	return UserInfo{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}
}

func newAssignmentUserResponse(user *domain.User) AssignmentUserResponse {
	return AssignmentUserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      user.Role,
	}
}

func newUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      user.Role,
		Skills:    user.Skills.ToSlice(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
