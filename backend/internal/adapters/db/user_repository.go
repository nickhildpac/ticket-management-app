package db

import (
	"context"
	"database/sql"
	"log"

	"github.com/google/uuid"
	sqlc "github.com/nickhildpac/ticket-management-app/internal/adapters/db/sqlc"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

type UserRepository struct {
	store sqlc.Store
}

func NewUserRepository(store sqlc.Store) *UserRepository {
	return &UserRepository{store: store}
}

func (r *UserRepository) GetUser(ctx context.Context, email string) (*domain.User, error) {
	user, err := r.store.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return mapUser(user), nil
}
func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := r.store.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapUser(user), nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user domain.User) (*domain.User, error) {
	created, err := r.store.CreateUser(ctx, sqlc.CreateUserParams{
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
		UpdatedAt:      user.UpdatedAt,
	})
	if err != nil {
		log.Println("Error creating userrepo:", err)
		return nil, err
	}
	return mapUser(created), nil
}

func (r *UserRepository) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	users, err := r.store.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.User, len(users))
	for i, user := range users {
		result[i] = domain.User{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		}
	}
	return result, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	updated, err := r.store.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      sql.NullString{String: string(user.Role), Valid: user.Role != ""},
		UpdatedAt: user.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}
	return mapUser(updated), nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return r.store.DeleteUser(ctx, id)
}
