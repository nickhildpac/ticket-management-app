package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/application/apperrors"
	"github.com/nickhildpac/ticket-management-app/internal/application/authorization"
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

func (s *UserService) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.User, error) {
	return s.repo.GetUsersByIDs(ctx, ids)
}

// UpdateMySkills updates the authenticated user's skills (self-service profile).
func (s *UserService) UpdateMySkills(ctx context.Context, skills []string) (*domain.User, error) {
	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}
	sk, err := domain.NewSkills(skills)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperrors.ErrBadRequest, err)
	}
	user, err := s.repo.GetUserByID(ctx, auth.UserID)
	if err != nil {
		return nil, err
	}
	user.Skills = *sk
	user.UpdatedAt = time.Now()
	return s.repo.UpdateUser(ctx, user)
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
	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	if !authorization.CanManageUsers(auth) {
		return nil, authorization.ErrAccessDenied
	}

	return s.repo.GetAllUsers(ctx)
}

// GetAllUsersForAssignment returns all users for ticket assignment purposes
// This method is accessible to all authenticated users, not just admins
func (s *UserService) GetAllUsersForAssignment(ctx context.Context) ([]domain.User, error) {
	return s.repo.GetAllUsers(ctx)
}

func (s *UserService) UpdateUserRole(ctx context.Context, id uuid.UUID, role domain.UserRole) (*domain.User, error) {
	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	if !authorization.CanManageUsers(auth) {
		return nil, authorization.ErrAccessDenied
	}

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Role = role
	user.UpdatedAt = time.Now()

	return s.repo.UpdateUser(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	auth, err := authorization.GetAuthContext(ctx)
	if err != nil {
		return err
	}

	if !authorization.CanManageUsers(auth) {
		return authorization.ErrAccessDenied
	}

	// Prevent self-deletion
	if auth.UserID == id {
		return fmt.Errorf("%w: cannot delete your own account", apperrors.ErrBadRequest)
	}

	return s.repo.DeleteUser(ctx, id)
}
