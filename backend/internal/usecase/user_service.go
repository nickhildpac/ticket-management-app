package usecase

import (
	"context"

	db "github.com/nickhildpac/ticket-management-app/internal/adapters/db/sqlc"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

type UserService struct {
	repo db.Store
}

func NewUserService(r db.Store) *UserService {
	return &UserService{
		repo: r,
	}
}

func (s *UserService) GetUser(ctx context.Context, username string) (*domain.User, error) {
	user, err := s.repo.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}
	return &domain.User{
		Username:       user.Username,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
	}, nil
}

func (s *UserService) CreateUser(ctx context.Context, u domain.User) (*domain.User, error) {
	arg := db.CreateUserParams{
		Username:       u.Username,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		Email:          u.Email,
		HashedPassword: u.HashedPassword,
	}
	user, err := s.repo.CreateUser(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &domain.User{
		Username:       user.Username,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
	}, nil
}
