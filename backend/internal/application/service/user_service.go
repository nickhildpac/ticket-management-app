package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
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

func (s *UserService) GetUser(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.GetUser(ctx, email)
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *UserService) CreateUser(ctx context.Context, u domain.User) (*domain.User, error) {
	u.UpdatedAt = time.Now()
	user, err := s.repo.CreateUser(ctx, u)
	if err != nil {
		log.Println("Error creating userserver:", err.Error())
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	return s.repo.GetAllUsers(ctx)
}
