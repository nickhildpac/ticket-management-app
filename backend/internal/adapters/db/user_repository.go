package db

import (
	"context"

	sqlc "github.com/nickhildpac/ticket-management-app/internal/adapters/db/sqlc"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

type UserRepository struct {
	store sqlc.Store
}

func NewUserRepository(store sqlc.Store) *UserRepository {
	return &UserRepository{store: store}
}

func (r *UserRepository) GetUser(ctx context.Context, username string) (*domain.User, error) {
	user, err := r.store.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}
	return mapUser(user), nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user domain.User) (*domain.User, error) {
	created, err := r.store.CreateUser(ctx, sqlc.CreateUserParams{
		Username:       user.Username,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
	})
	if err != nil {
		return nil, err
	}
	return mapUser(created), nil
}
