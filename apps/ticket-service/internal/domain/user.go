package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAgent UserRole = "agent"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID uuid.UUID `json:"id"`
	// KeycloakID links this row to the Keycloak subject (`sub`) that
	// authenticates as it. Nil for rows that predate Keycloak and have not been
	// claimed by a sign-in yet. The local ID stays the identity used by
	// tickets/comments foreign keys; see docs/adr/0003-keycloak-authentication.md.
	KeycloakID *uuid.UUID `json:"-"`
	FirstName  string     `json:"first_name"`
	LastName   string     `json:"last_name"`
	Email      string     `json:"email"`
	Role       UserRole   `json:"role"`
	Skills     Skills     `json:"skills"`
	UpdatedAt  time.Time  `json:"updated_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

func GetRole(s string) (UserRole, error) {
	switch strings.ToLower(s) {
	case "user":
		return RoleUser, nil
	case "agent":
		return RoleAgent, nil
	case "admin":
		return RoleAdmin, nil
	default:
		return "", fmt.Errorf("invalid role: %s", s)
	}
}
