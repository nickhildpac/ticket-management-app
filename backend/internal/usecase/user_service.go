package usecase

import (
	"context"

	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/internal/usecase/port"
)

type UserService struct {
	repo port.UserRepository
}

func NewUserService(r port.UserRepository) *UserService {
	return &UserService{
		repo: r,
	}
}

func (s *UserService) GetUser(ctx context.Context, username string) (*domain.User, error) {
	return s.repo.GetUser(ctx, username)
}

func (s *UserService) CreateUser(ctx context.Context, u domain.User) (*domain.User, error) {
	return s.repo.CreateUser(ctx, u)
}

var _ port.UserService = (*UserService)(nil)
