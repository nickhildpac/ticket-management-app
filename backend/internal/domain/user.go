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
	ID             uuid.UUID `json:"id"`
	HashedPassword string    `json:"hashed_password,omitempty"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Email          string    `json:"email"`
	Role           UserRole  `json:"role"`
	UpdatedAt      time.Time `json:"updated_at"`
	CreatedAt      time.Time `json:"created_at"`
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
