package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	keycloakauth "github.com/nickhildpac/ticket-management-app/internal/adapters/auth"
	"github.com/nickhildpac/ticket-management-app/internal/application/apperrors"
	"github.com/nickhildpac/ticket-management-app/internal/application/authorization"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/internal/ports"
)

// RealmRoleWriter assigns a realm role in the identity provider. Implemented by
// auth.AdminClient; nil when admin credentials are not configured.
type RealmRoleWriter interface {
	SetRealmRole(ctx context.Context, keycloakID uuid.UUID, role domain.UserRole) error
}

// IdentityCacheInvalidator drops a cached Keycloak-subject → user resolution.
type IdentityCacheInvalidator interface {
	InvalidateCache(localID uuid.UUID)
}

// ErrUserNotLinked means the target row has no Keycloak subject yet — nobody has
// signed in as them — so there is no realm identity whose roles we could change.
var ErrUserNotLinked = errors.New("user is not linked to a keycloak identity")

type UserService struct {
	repo ports.UserRepository
	// roles is nil when Keycloak admin credentials are absent; role changes then
	// fail loudly instead of being applied locally and later overwritten.
	roles    RealmRoleWriter
	identity IdentityCacheInvalidator
}

func NewUserService(r ports.UserRepository, opts ...UserServiceOption) *UserService {
	s := &UserService{repo: r}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// UserServiceOption configures optional collaborators.
type UserServiceOption func(*UserService)

// WithRealmRoleWriter enables writing role changes through to Keycloak.
func WithRealmRoleWriter(w RealmRoleWriter) UserServiceOption {
	return func(s *UserService) {
		// A typed-nil interface would pass a `!= nil` check and then panic, so
		// guard here where the concrete value is still visible.
		if w != nil {
			s.roles = w
		}
	}
}

// WithIdentityCache lets role changes take effect without waiting for the
// identity cache TTL.
func WithIdentityCache(c IdentityCacheInvalidator) UserServiceOption {
	return func(s *UserService) { s.identity = c }
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

// UpdateUserRole changes a user's role in Keycloak, then mirrors it locally.
//
// Keycloak is the source of truth for roles (ADR 0003). The write order matters:
// if the local mirror fails after Keycloak succeeded, the next request from that
// user re-syncs the row from their token, so the two converge. Doing it the other
// way round would leave a local role that no token ever backs.
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

	if s.roles == nil {
		return nil, keycloakauth.ErrAdminNotConfigured
	}
	if user.KeycloakID == nil {
		return nil, ErrUserNotLinked
	}
	if err := s.roles.SetRealmRole(ctx, *user.KeycloakID, role); err != nil {
		return nil, fmt.Errorf("updating role in keycloak: %w", err)
	}

	user.Role = role
	user.UpdatedAt = time.Now()

	updated, err := s.repo.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// Without this the old role would be served from cache until the TTL, even
	// though Keycloak already has the new one.
	if s.identity != nil {
		s.identity.InvalidateCache(id)
	}
	return updated, nil
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
