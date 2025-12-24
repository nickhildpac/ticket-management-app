package service

import (
	"context"

	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/internal/ports"
)

type UserService struct {
	repo ports.UserRepository
}

func NewUserService(r ports.UserRepository) *UserService {
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

func (s *UserService) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	return s.repo.GetAllUsers(ctx)
}
