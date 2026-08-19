package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/adapters/auth"
	"github.com/nickhildpac/ticket-management-app/internal/application/apperrors"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

// identityRepo is an in-memory stand-in for the users table, with just enough
// behaviour to exercise the link/provision/sync paths including their unique
// constraints.
type identityRepo struct {
	mockUserRepository

	mu       sync.Mutex
	byID     map[uuid.UUID]*domain.User
	creates  int
	links    int
	syncs    int
	failSync bool
}

func newIdentityRepo(seed ...*domain.User) *identityRepo {
	r := &identityRepo{byID: map[uuid.UUID]*domain.User{}}
	for _, u := range seed {
		r.byID[u.ID] = u
	}
	return r
}

func (r *identityRepo) GetUser(_ context.Context, email string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.byID {
		if u.Email == email {
			copied := *u
			return &copied, nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func (r *identityRepo) GetUserByKeycloakID(_ context.Context, keycloakID uuid.UUID) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.byID {
		if u.KeycloakID != nil && *u.KeycloakID == keycloakID {
			copied := *u
			return &copied, nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func (r *identityRepo) CreateUserFromKeycloak(_ context.Context, keycloakID uuid.UUID, user domain.User) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.creates++
	for _, u := range r.byID {
		if u.KeycloakID != nil && *u.KeycloakID == keycloakID {
			return nil, apperrors.ErrDuplicateIdentity
		}
		if u.Email == user.Email {
			return nil, apperrors.ErrDuplicateEmail
		}
	}
	created := user
	created.ID = uuid.New()
	created.KeycloakID = &keycloakID
	r.byID[created.ID] = &created
	copied := created
	return &copied, nil
}

func (r *identityRepo) LinkUserToKeycloak(_ context.Context, localID, keycloakID uuid.UUID, user domain.User) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.links++
	existing, ok := r.byID[localID]
	// Mirrors the `WHERE id = $1 AND keycloak_id IS NULL` guard in the query.
	if !ok || existing.KeycloakID != nil {
		return nil, apperrors.ErrNotFound
	}
	existing.KeycloakID = &keycloakID
	existing.FirstName = user.FirstName
	existing.LastName = user.LastName
	existing.Role = user.Role
	copied := *existing
	return &copied, nil
}

func (r *identityRepo) SyncUserFromKeycloak(_ context.Context, keycloakID uuid.UUID, user domain.User) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.syncs++
	if r.failSync {
		return nil, errors.New("database is unavailable")
	}
	for _, u := range r.byID {
		if u.KeycloakID != nil && *u.KeycloakID == keycloakID {
			u.Email = user.Email
			u.FirstName = user.FirstName
			u.LastName = user.LastName
			u.Role = user.Role
			copied := *u
			return &copied, nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func claimsFor(sub uuid.UUID, email string, role domain.UserRole) *auth.Claims {
	return &auth.Claims{
		Subject:   sub.String(),
		Email:     email,
		FirstName: "Test",
		LastName:  "User",
		Role:      role,
	}
}

func TestResolveProvisionsNewUser(t *testing.T) {
	repo := newIdentityRepo()
	svc := NewIdentityService(repo, time.Minute)
	sub := uuid.New()

	user, err := svc.Resolve(context.Background(), claimsFor(sub, "new@example.com", domain.RoleUser))
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if user.KeycloakID == nil || *user.KeycloakID != sub {
		t.Errorf("expected the new row to be linked to subject %s", sub)
	}
	if user.Email != "new@example.com" || user.Role != domain.RoleUser {
		t.Errorf("unexpected provisioned user %+v", user)
	}
	// The local id must be freshly generated, not the Keycloak subject.
	if user.ID == sub {
		t.Error("local user id should not be the keycloak subject")
	}
}

// The migration path that matters most: an account that existed before Keycloak
// must keep its id, because tickets and comments point at it.
func TestResolveLinksExistingUserByEmailPreservingLocalID(t *testing.T) {
	existingID := uuid.New()
	repo := newIdentityRepo(&domain.User{
		ID:        existingID,
		Email:     "charlie@user.com",
		FirstName: "Charlie",
		LastName:  "User",
		Role:      domain.RoleUser,
	})
	svc := NewIdentityService(repo, time.Minute)
	sub := uuid.New()

	user, err := svc.Resolve(context.Background(), claimsFor(sub, "charlie@user.com", domain.RoleUser))
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if user.ID != existingID {
		t.Fatalf("expected the pre-existing local id %s to be kept, got %s", existingID, user.ID)
	}
	if user.KeycloakID == nil || *user.KeycloakID != sub {
		t.Error("expected the existing row to be linked to the keycloak subject")
	}
	if repo.creates != 0 {
		t.Errorf("expected no new row to be created, got %d creates", repo.creates)
	}
}

// A second Keycloak account sharing an email must not inherit the first one's
// tickets.
func TestResolveRefusesToStealAlreadyLinkedEmail(t *testing.T) {
	firstSubject := uuid.New()
	repo := newIdentityRepo(&domain.User{
		ID:         uuid.New(),
		KeycloakID: &firstSubject,
		Email:      "shared@example.com",
		Role:       domain.RoleAdmin,
	})
	svc := NewIdentityService(repo, time.Minute)

	_, err := svc.Resolve(context.Background(), claimsFor(uuid.New(), "shared@example.com", domain.RoleUser))
	if !errors.Is(err, ErrIdentityConflict) {
		t.Fatalf("expected ErrIdentityConflict, got %v", err)
	}
}

// Keycloak is authoritative for roles: a promotion there must reach the local
// row rather than being ignored because the row already exists.
func TestResolveSyncsRoleChangeFromToken(t *testing.T) {
	sub := uuid.New()
	repo := newIdentityRepo(&domain.User{
		ID:         uuid.New(),
		KeycloakID: &sub,
		Email:      "bob@agent.com",
		FirstName:  "Test",
		LastName:   "User",
		Role:       domain.RoleUser,
	})
	svc := NewIdentityService(repo, time.Minute)

	user, err := svc.Resolve(context.Background(), claimsFor(sub, "bob@agent.com", domain.RoleAgent))
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if user.Role != domain.RoleAgent {
		t.Fatalf("expected role to be synced to agent, got %q", user.Role)
	}
	if repo.syncs != 1 {
		t.Errorf("expected exactly one sync write, got %d", repo.syncs)
	}
}

// A stale row must not lock the user out when the refresh write fails.
func TestResolveFallsBackToStoredRowWhenSyncFails(t *testing.T) {
	sub := uuid.New()
	repo := newIdentityRepo(&domain.User{
		ID:         uuid.New(),
		KeycloakID: &sub,
		Email:      "bob@agent.com",
		FirstName:  "Test",
		LastName:   "User",
		Role:       domain.RoleUser,
	})
	repo.failSync = true
	svc := NewIdentityService(repo, time.Minute)

	user, err := svc.Resolve(context.Background(), claimsFor(sub, "bob@agent.com", domain.RoleAgent))
	if err != nil {
		t.Fatalf("expected the stored row to be used, got error %v", err)
	}
	if user.Role != domain.RoleUser {
		t.Errorf("expected the stored role, got %q", user.Role)
	}
}

func TestResolveIsCachedAcrossRequests(t *testing.T) {
	sub := uuid.New()
	repo := newIdentityRepo(&domain.User{
		ID:         uuid.New(),
		KeycloakID: &sub,
		Email:      "bob@agent.com",
		FirstName:  "Test",
		LastName:   "User",
		Role:       domain.RoleAgent,
	})
	svc := NewIdentityService(repo, time.Minute)
	claims := claimsFor(sub, "bob@agent.com", domain.RoleAgent)

	for range 5 {
		if _, err := svc.Resolve(context.Background(), claims); err != nil {
			t.Fatalf("resolve: %v", err)
		}
	}
	if repo.creates != 0 || repo.links != 0 || repo.syncs != 0 {
		t.Errorf("expected no writes on the cached path, got creates=%d links=%d syncs=%d",
			repo.creates, repo.links, repo.syncs)
	}
}

// A privilege change must not wait out the cache TTL: a token carrying a
// different role has to bypass the cached entry.
func TestCacheIsBypassedWhenTokenRoleChanges(t *testing.T) {
	sub := uuid.New()
	repo := newIdentityRepo(&domain.User{
		ID:         uuid.New(),
		KeycloakID: &sub,
		Email:      "bob@agent.com",
		FirstName:  "Test",
		LastName:   "User",
		Role:       domain.RoleAgent,
	})
	svc := NewIdentityService(repo, time.Hour)

	if _, err := svc.Resolve(context.Background(), claimsFor(sub, "bob@agent.com", domain.RoleAgent)); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	// Same subject, demoted in Keycloak.
	user, err := svc.Resolve(context.Background(), claimsFor(sub, "bob@agent.com", domain.RoleUser))
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if user.Role != domain.RoleUser {
		t.Fatalf("expected the demotion to take effect immediately, got %q", user.Role)
	}
}

func TestInvalidateCacheForcesReread(t *testing.T) {
	sub := uuid.New()
	local := &domain.User{
		ID:         uuid.New(),
		KeycloakID: &sub,
		Email:      "bob@agent.com",
		FirstName:  "Test",
		LastName:   "User",
		Role:       domain.RoleAgent,
	}
	repo := newIdentityRepo(local)
	svc := NewIdentityService(repo, time.Hour)
	claims := claimsFor(sub, "bob@agent.com", domain.RoleAgent)

	if _, err := svc.Resolve(context.Background(), claims); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	// Change a locally-owned field. Name and role would be no good here: the
	// token owns those, so a resolve would sync them straight back.
	repo.mu.Lock()
	repo.byID[local.ID].Skills = domain.NewSkillsFromSlice([]string{"log-analysis"})
	repo.mu.Unlock()

	// Still cached, so the stale copy is served.
	cached, err := svc.Resolve(context.Background(), claims)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(cached.Skills.ToSlice()) != 0 {
		t.Fatalf("expected the cached row, got skills %v", cached.Skills.ToSlice())
	}

	svc.InvalidateCache(local.ID)

	user, err := svc.Resolve(context.Background(), claims)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got := user.Skills.ToSlice(); len(got) != 1 || got[0] != "log-analysis" {
		t.Errorf("expected a fresh read after invalidation, got skills %v", got)
	}
}

func TestResolveRejectsNonUUIDSubject(t *testing.T) {
	svc := NewIdentityService(newIdentityRepo(), time.Minute)

	_, err := svc.Resolve(context.Background(), &auth.Claims{Subject: "not-a-uuid", Role: domain.RoleUser})
	if !errors.Is(err, apperrors.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

// A service account token carries no email; the row still needs a unique one.
func TestResolveSynthesisesEmailWhenTokenHasNone(t *testing.T) {
	repo := newIdentityRepo()
	svc := NewIdentityService(repo, time.Minute)

	user, err := svc.Resolve(context.Background(), &auth.Claims{
		Subject:   uuid.NewString(),
		Username:  "service-account-ai-service",
		FirstName: "AI",
		Role:      domain.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if user.Email != "service-account-ai-service@keycloak.local" {
		t.Errorf("unexpected synthesised email %q", user.Email)
	}
}

// Two concurrent first requests for the same new subject must converge on one
// row, not create two.
func TestConcurrentResolveProvisionsOnce(t *testing.T) {
	repo := newIdentityRepo()
	svc := NewIdentityService(repo, time.Minute)
	sub := uuid.New()
	claims := claimsFor(sub, "racer@example.com", domain.RoleUser)

	const goroutines = 8
	var wg sync.WaitGroup
	ids := make([]uuid.UUID, goroutines)
	errs := make([]error, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			user, err := svc.Resolve(context.Background(), claims)
			errs[i] = err
			if user != nil {
				ids[i] = user.ID
			}
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: %v", i, err)
		}
	}
	for i, id := range ids {
		if id != ids[0] {
			t.Fatalf("goroutine %d resolved to %s, want %s — provisioning raced", i, id, ids[0])
		}
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if len(repo.byID) != 1 {
		t.Fatalf("expected exactly one user row, got %d", len(repo.byID))
	}
}
