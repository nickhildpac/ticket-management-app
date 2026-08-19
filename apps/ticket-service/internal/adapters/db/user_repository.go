package db

import (
	"context"
	"database/sql"
	"log"
	"math"

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
		return nil, normalizeDBError(err)
	}
	return mapUser(user), nil
}
func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := r.store.GetUser(ctx, id)
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return mapUser(user), nil
}

func (r *UserRepository) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.User, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]*domain.User{}, nil
	}
	rows, err := r.store.GetUsersByIDs(ctx, ids)
	if err != nil {
		return nil, normalizeDBError(err)
	}
	out := make(map[uuid.UUID]*domain.User, len(rows))
	for _, row := range rows {
		u := mapUser(row)
		out[u.ID] = u
	}
	return out, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user domain.User) (*domain.User, error) {
	created, err := r.store.CreateUser(ctx, sqlc.CreateUserParams{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		// Credentials live in Keycloak; this is a placeholder so the NOT NULL
		// column is satisfied and the value can never match a bcrypt hash.
		HashedPassword: externalPasswordPlaceholder,
		UpdatedAt:      user.UpdatedAt,
		Skills:         user.Skills.ToSlice(),
	})
	if err != nil {
		log.Println("Error creating userrepo:", err)
		return nil, normalizeDBError(err)
	}
	return mapUser(created), nil
}

// externalPasswordPlaceholder is stored in users.hashed_password now that
// authentication is delegated to Keycloak. It is not a valid bcrypt hash.
const externalPasswordPlaceholder = "!external"

// GetUserByKeycloakID looks up the local row linked to a Keycloak subject.
func (r *UserRepository) GetUserByKeycloakID(ctx context.Context, keycloakID uuid.UUID) (*domain.User, error) {
	user, err := r.store.GetUserByKeycloakID(ctx, uuid.NullUUID{UUID: keycloakID, Valid: true})
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return mapUser(user), nil
}

// CreateUserFromKeycloak provisions a new local row for a Keycloak subject.
func (r *UserRepository) CreateUserFromKeycloak(ctx context.Context, keycloakID uuid.UUID, user domain.User) (*domain.User, error) {
	created, err := r.store.CreateUserFromKeycloak(ctx, sqlc.CreateUserFromKeycloakParams{
		KeycloakID: uuid.NullUUID{UUID: keycloakID, Valid: true},
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Email:      user.Email,
		Role:       sql.NullString{String: string(user.Role), Valid: user.Role != ""},
	})
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return mapUser(created), nil
}

// LinkUserToKeycloak claims an existing unlinked row for a Keycloak subject.
// It returns ErrNotFound if the row was already linked to another subject.
func (r *UserRepository) LinkUserToKeycloak(ctx context.Context, localID, keycloakID uuid.UUID, user domain.User) (*domain.User, error) {
	linked, err := r.store.LinkUserToKeycloak(ctx, sqlc.LinkUserToKeycloakParams{
		ID:         localID,
		KeycloakID: uuid.NullUUID{UUID: keycloakID, Valid: true},
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Role:       sql.NullString{String: string(user.Role), Valid: user.Role != ""},
	})
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return mapUser(linked), nil
}

// SyncUserFromKeycloak refreshes a linked row from the token's current profile.
func (r *UserRepository) SyncUserFromKeycloak(ctx context.Context, keycloakID uuid.UUID, user domain.User) (*domain.User, error) {
	synced, err := r.store.SyncUserFromKeycloak(ctx, sqlc.SyncUserFromKeycloakParams{
		KeycloakID: uuid.NullUUID{UUID: keycloakID, Valid: true},
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Role:       sql.NullString{String: string(user.Role), Valid: user.Role != ""},
	})
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return mapUser(synced), nil
}

func (r *UserRepository) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	users, err := r.store.ListUsers(ctx, sqlc.ListUsersParams{
		Limit:  math.MaxInt32,
		Offset: 0,
	})
	if err != nil {
		return nil, normalizeDBError(err)
	}

	result := make([]domain.User, len(users))
	for i, user := range users {
		result[i] = *mapUser(user)
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
		Skills:    user.Skills.ToSlice(),
	})
	if err != nil {
		return nil, normalizeDBError(err)
	}
	return mapUser(updated), nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return normalizeDBError(r.store.DeleteUser(ctx, id))
}

func (r *UserRepository) GetAllAgents(ctx context.Context) ([]domain.User, error) {
	agents, err := r.store.GetAllAgents(ctx)
	if err != nil {
		return nil, normalizeDBError(err)
	}

	result := make([]domain.User, len(agents))
	for i, agent := range agents {
		mapped := mapUser(agent)
		result[i] = *mapped
	}
	return result, nil
}

func (r *UserRepository) GetAutoAssignmentCandidates(ctx context.Context, requiredSkills []string, activeStates []domain.TicketState) ([]domain.AutoAssignmentCandidate, error) {
	if len(requiredSkills) == 0 {
		return []domain.AutoAssignmentCandidate{}, nil
	}

	stateIDs := make([]int32, len(activeStates))
	for i, state := range activeStates {
		stateIDs[i] = int32(state)
	}

	rows, err := r.store.GetAutoAssignmentCandidates(ctx, sqlc.GetAutoAssignmentCandidatesParams{
		RequiredSkills: requiredSkills,
		ActiveStates:   stateIDs,
	})
	if err != nil {
		return nil, normalizeDBError(err)
	}

	candidates := make([]domain.AutoAssignmentCandidate, len(rows))
	for i, row := range rows {
		agent := mapUser(sqlc.User{
			ID:             row.ID,
			HashedPassword: row.HashedPassword,
			FirstName:      row.FirstName,
			LastName:       row.LastName,
			Email:          row.Email,
			Role:           row.Role,
			UpdatedAt:      row.UpdatedAt,
			CreatedAt:      row.CreatedAt,
			Skills:         row.Skills,
		})
		candidates[i] = domain.AutoAssignmentCandidate{
			Agent:             *agent,
			ActiveTicketCount: int(row.ActiveTicketCount),
		}
	}

	return candidates, nil
}
