package service

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/adapters/auth"
	"github.com/nickhildpac/ticket-management-app/internal/application/apperrors"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/internal/ports"
)

// IdentityService resolves a verified Keycloak token into the local user row
// that owns tickets and comments.
//
// Keycloak is authoritative for *who you are* and *what role you hold*; the
// local users table stays authoritative for *what you own*, because
// tickets.created_by, tickets.assigned_to[] and comments.created_by are foreign
// keys into it. Resolution order for a subject:
//
//  1. already linked (users.keycloak_id = sub) — refresh the row if the token's
//     profile or role has drifted;
//  2. an unlinked row with the same email — claim it, so accounts that existed
//     before Keycloak keep their ticket history;
//  3. otherwise provision a new row (JIT).
type IdentityService struct {
	repo  ports.UserRepository
	cache *identityCache
}

// NewIdentityService builds the service. ttl bounds how long a token's
// resolution is reused before the database is consulted again; zero uses the
// default.
func NewIdentityService(repo ports.UserRepository, ttl time.Duration) *IdentityService {
	if ttl <= 0 {
		ttl = defaultIdentityCacheTTL
	}
	return &IdentityService{repo: repo, cache: newIdentityCache(ttl)}
}

// ErrIdentityConflict means the token's email belongs to a local row already
// linked to a different Keycloak subject. Auto-resolving that would let a new
// Keycloak account inherit another user's tickets, so it is refused.
var ErrIdentityConflict = errors.New("email already linked to a different identity")

// Resolve maps verified token claims to the local user.
func (s *IdentityService) Resolve(ctx context.Context, claims *auth.Claims) (*domain.User, error) {
	keycloakID, err := uuid.Parse(claims.Subject)
	if err != nil {
		// Keycloak subjects are UUIDs; anything else means a realm this service
		// was not built for.
		return nil, apperrors.ErrUnauthorized
	}

	if user, ok := s.cache.get(keycloakID, claims); ok {
		return user, nil
	}

	user, err := s.resolveFromStore(ctx, keycloakID, claims)
	if err != nil {
		return nil, err
	}

	s.cache.put(keycloakID, claims, user)
	return user, nil
}

// InvalidateCache drops any cached resolution for a local user. Call it after
// changing a user's role locally so the next request re-reads the row.
func (s *IdentityService) InvalidateCache(localID uuid.UUID) {
	s.cache.invalidateByLocalID(localID)
}

func (s *IdentityService) resolveFromStore(ctx context.Context, keycloakID uuid.UUID, claims *auth.Claims) (*domain.User, error) {
	existing, err := s.repo.GetUserByKeycloakID(ctx, keycloakID)
	switch {
	case err == nil:
		// The token is the source of truth for role and name; write back only
		// when something actually changed, to keep this a read on the hot path.
		if identityDrifted(existing, claims) {
			synced, syncErr := s.repo.SyncUserFromKeycloak(ctx, keycloakID, userFromClaims(claims))
			if syncErr != nil {
				// A failed refresh must not lock the user out — the row is
				// still valid, just stale.
				log.Printf("identity: refreshing user %s from token failed, using stored row: %v", existing.ID, syncErr)
				return existing, nil
			}
			return synced, nil
		}
		return existing, nil

	case errors.Is(err, apperrors.ErrNotFound):
		// fall through to link-or-provision

	default:
		return nil, err
	}

	return s.linkOrProvision(ctx, keycloakID, claims)
}

func (s *IdentityService) linkOrProvision(ctx context.Context, keycloakID uuid.UUID, claims *auth.Claims) (*domain.User, error) {
	if claims.Email != "" {
		byEmail, err := s.repo.GetUser(ctx, claims.Email)
		switch {
		case err == nil && byEmail.KeycloakID == nil:
			linked, linkErr := s.repo.LinkUserToKeycloak(ctx, byEmail.ID, keycloakID, userFromClaims(claims))
			if linkErr == nil {
				log.Printf("identity: linked existing user %s (%s) to keycloak subject %s",
					byEmail.ID, byEmail.Email, keycloakID)
				return linked, nil
			}
			if !errors.Is(linkErr, apperrors.ErrNotFound) {
				return nil, linkErr
			}
			// The row was claimed between the read and the update; re-resolve
			// so we see whoever won.
			return s.resolveAfterRace(ctx, keycloakID, claims)

		case err == nil && byEmail.KeycloakID != nil && *byEmail.KeycloakID != keycloakID:
			log.Printf("identity: refusing to link keycloak subject %s to user %s — already linked to %s",
				keycloakID, byEmail.ID, *byEmail.KeycloakID)
			return nil, ErrIdentityConflict

		case err != nil && !errors.Is(err, apperrors.ErrNotFound):
			return nil, err
		}
	}

	created, err := s.repo.CreateUserFromKeycloak(ctx, keycloakID, userFromClaims(claims))
	if err != nil {
		// Two concurrent first requests for the same new subject both reach
		// here; the loser hits a unique index and just reads the winner's row.
		if errors.Is(err, apperrors.ErrDuplicateIdentity) || errors.Is(err, apperrors.ErrDuplicateEmail) {
			return s.resolveAfterRace(ctx, keycloakID, claims)
		}
		return nil, err
	}
	log.Printf("identity: provisioned local user %s for keycloak subject %s (role %s)",
		created.ID, keycloakID, created.Role)
	return created, nil
}

// resolveAfterRace re-reads by subject after losing a provisioning race.
func (s *IdentityService) resolveAfterRace(ctx context.Context, keycloakID uuid.UUID, claims *auth.Claims) (*domain.User, error) {
	user, err := s.repo.GetUserByKeycloakID(ctx, keycloakID)
	if err != nil {
		return nil, err
	}
	if identityDrifted(user, claims) {
		if synced, syncErr := s.repo.SyncUserFromKeycloak(ctx, keycloakID, userFromClaims(claims)); syncErr == nil {
			return synced, nil
		}
	}
	return user, nil
}

// identityDrifted reports whether the stored row disagrees with the token.
func identityDrifted(user *domain.User, claims *auth.Claims) bool {
	if user.Role != claims.Role {
		return true
	}
	if claims.Email != "" && user.Email != claims.Email {
		return true
	}
	return user.FirstName != claims.FirstName || user.LastName != claims.LastName
}

func userFromClaims(claims *auth.Claims) domain.User {
	email := claims.Email
	if email == "" {
		// The row needs a unique non-empty email; Keycloak guarantees a unique
		// username, so synthesise from it when the token carries no email.
		email = claims.Username + "@keycloak.local"
	}
	return domain.User{
		Email:     email,
		FirstName: claims.FirstName,
		LastName:  claims.LastName,
		Role:      claims.Role,
	}
}

// ---- cache --------------------------------------------------------------

// defaultIdentityCacheTTL bounds staleness after an out-of-band change (a role
// edited directly in Keycloak's console, say). It is deliberately shorter than
// the realm's 5-minute access token lifespan, so a re-login is never the slower
// path to picking up a change.
const defaultIdentityCacheTTL = time.Minute

// identityCache avoids a database round-trip on every authenticated request.
//
// Entries are keyed by Keycloak subject and additionally validated against the
// token's role and profile: if a *new* token carries a different role, the
// cached entry is ignored rather than served. That means a privilege change in
// Keycloak takes effect as soon as the user presents a token carrying it, not
// after the TTL.
type identityCache struct {
	mu      sync.RWMutex
	ttl     time.Duration
	entries map[uuid.UUID]identityCacheEntry
}

type identityCacheEntry struct {
	user      *domain.User
	role      domain.UserRole
	email     string
	firstName string
	lastName  string
	expiresAt time.Time
}

func newIdentityCache(ttl time.Duration) *identityCache {
	return &identityCache{ttl: ttl, entries: make(map[uuid.UUID]identityCacheEntry)}
}

func (c *identityCache) get(keycloakID uuid.UUID, claims *auth.Claims) (*domain.User, bool) {
	c.mu.RLock()
	entry, ok := c.entries[keycloakID]
	c.mu.RUnlock()

	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	if entry.role != claims.Role || entry.firstName != claims.FirstName || entry.lastName != claims.LastName {
		return nil, false
	}
	if claims.Email != "" && entry.email != claims.Email {
		return nil, false
	}
	return entry.user, true
}

func (c *identityCache) put(keycloakID uuid.UUID, claims *auth.Claims, user *domain.User) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Bounded so a burst of distinct subjects can't grow this without limit.
	if len(c.entries) >= maxIdentityCacheEntries {
		c.evictExpiredLocked()
		if len(c.entries) >= maxIdentityCacheEntries {
			clear(c.entries)
		}
	}

	c.entries[keycloakID] = identityCacheEntry{
		user:      user,
		role:      claims.Role,
		email:     claims.Email,
		firstName: claims.FirstName,
		lastName:  claims.LastName,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *identityCache) invalidateByLocalID(localID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, entry := range c.entries {
		if entry.user != nil && entry.user.ID == localID {
			delete(c.entries, k)
		}
	}
}

func (c *identityCache) evictExpiredLocked() {
	now := time.Now()
	for k, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, k)
		}
	}
}

const maxIdentityCacheEntries = 10000
