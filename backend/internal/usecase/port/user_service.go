package port

import (
	"context"

	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

type UserService interface {
	GetUser(ctx context.Context, username string) (*domain.User, error)
	CreateUser(ctx context.Context, user domain.User) (*domain.User, error)
}
