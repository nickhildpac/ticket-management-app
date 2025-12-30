package authorization

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
)

var (
	ErrAccessDenied   = errors.New("access denied: insufficient permissions")
	ErrNotTicketOwner = errors.New("access denied: not ticket owner")
	ErrNotAssignee    = errors.New("access denied: not assigned to ticket")
)

type AuthContext struct {
	UserID uuid.UUID
	Role   domain.UserRole
}

// GetAuthContext extracts user info from context
func GetAuthContext(ctx context.Context) (AuthContext, error) {
	userIDStr, ok := ctx.Value(configs.UserIDKey).(string)
	if !ok {
		return AuthContext{}, errors.New("user ID not found in context")
	}

	roleStr, ok := ctx.Value(configs.UserRoleKey).(string)
	if !ok {
		return AuthContext{}, errors.New("user role not found in context")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return AuthContext{}, err
	}

	return AuthContext{
		UserID: userID,
		Role:   domain.UserRole(roleStr),
	}, nil
}

// CanViewTicket determines if user can view ticket
func CanViewTicket(auth AuthContext, ticket *domain.Ticket) bool {
	switch auth.Role {
	case domain.RoleAdmin:
		return true
	case domain.RoleAgent:
		return isUserInList(auth.UserID, ticket.AssignedTo)
	case domain.RoleUser:
		return ticket.CreatedBy == auth.UserID
	default:
		return false
	}
}

// CanUpdateTicket determines if user can update ticket
func CanUpdateTicket(auth AuthContext, ticket *domain.Ticket) bool {
	switch auth.Role {
	case domain.RoleAdmin:
		return true
	case domain.RoleAgent:
		return isUserInList(auth.UserID, ticket.AssignedTo)
	case domain.RoleUser:
		return ticket.CreatedBy == auth.UserID
	default:
		return false
	}
}

// CanUpdateTicketState determines if user can change ticket state
func CanUpdateTicketState(auth AuthContext, ticket *domain.Ticket) bool {
	switch auth.Role {
	case domain.RoleAdmin:
		return true
	case domain.RoleAgent:
		return isUserInList(auth.UserID, ticket.AssignedTo)
	case domain.RoleUser:
		return false
	default:
		return false
	}
}

// CanUpdateTicketPriority determines if user can change priority
func CanUpdateTicketPriority(auth AuthContext, ticket *domain.Ticket) bool {
	return auth.Role == domain.RoleAdmin
}

// CanAssignTicket determines if user can assign tickets
func CanAssignTicket(auth AuthContext, ticket *domain.Ticket) bool {
	return auth.Role == domain.RoleAdmin
}

// CanDeleteTicket determines if user can delete ticket
func CanDeleteTicket(auth AuthContext, ticket *domain.Ticket) bool {
	return auth.Role == domain.RoleAdmin
}

// CanCommentOnTicket determines if user can comment on ticket
func CanCommentOnTicket(auth AuthContext, ticket *domain.Ticket) bool {
	switch auth.Role {
	case domain.RoleAdmin:
		return true
	case domain.RoleAgent:
		return isUserInList(auth.UserID, ticket.AssignedTo)
	case domain.RoleUser:
		return ticket.CreatedBy == auth.UserID
	default:
		return false
	}
}

// CanManageUsers determines if user can manage users
func CanManageUsers(auth AuthContext) bool {
	return auth.Role == domain.RoleAdmin
}

// Helper function to check if UUID is in list
func isUserInList(userID uuid.UUID, list []uuid.UUID) bool {
	for _, id := range list {
		if id == userID {
			return true
		}
	}
	return false
}
