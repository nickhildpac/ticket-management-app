// Package middlewares contains shared HTTP middleware for the API server.
package middlewares

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/nickhildpac/ticket-management-app/internal/adapters/auth"
	"github.com/nickhildpac/ticket-management-app/internal/application/service"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
	"github.com/nickhildpac/ticket-management-app/pkg/util"
)

func EnableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", configs.GetString("FRONTEND_URL", "http://localhost:5173"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type,X-CSRF-Token,Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// TokenVerifier verifies a raw bearer token. Implemented by
// internal/adapters/auth.Verifier; an interface so tests can substitute a stub.
type TokenVerifier interface {
	Verify(ctx context.Context, rawToken string) (*auth.Claims, error)
}

// IdentityResolver maps verified claims to the local user row that owns
// tickets and comments. Implemented by service.IdentityService.
type IdentityResolver interface {
	Resolve(ctx context.Context, claims *auth.Claims) (*domain.User, error)
}

// Authenticator turns a Keycloak access token into request context.
//
// This is the only place tokens are inspected. Downstream handlers and services
// read the local user id and role from the context keys it sets, so the
// authorization package is unchanged by the move to Keycloak.
type Authenticator struct {
	verifier TokenVerifier
	identity IdentityResolver
}

func NewAuthenticator(verifier TokenVerifier, identity IdentityResolver) *Authenticator {
	return &Authenticator{verifier: verifier, identity: identity}
}

// claimsContextKey carries the full token claims for handlers that need more
// than id and role (e.g. the realm role list on /me).
type claimsContextKey struct{}

// ClaimsFromContext returns the verified token claims, if the request went
// through AuthRequired.
func ClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey{}).(*auth.Claims)
	return claims, ok
}

// AuthRequired rejects any request without a valid Keycloak access token.
func (a *Authenticator) AuthRequired() func(http.Handler) http.Handler {
	return a.RequireAnyRole()
}

// AdminRequired admits only tokens carrying the `admin` realm role.
func (a *Authenticator) AdminRequired() func(http.Handler) http.Handler {
	return a.RequireAnyRole(domain.RoleAdmin)
}

// RequireAnyRole authenticates the request and, when roles are given, requires
// the token's effective role to be one of them. With no roles it only
// authenticates.
//
// Role checks here are coarse route gating. Per-record rules (may this agent
// see this particular ticket?) stay in internal/application/authorization,
// because they depend on rows in Postgres rather than on the token.
func (a *Authenticator) RequireAnyRole(roles ...domain.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Vary", "Authorization")

			rawToken, err := bearerToken(r)
			if err != nil {
				util.ErrorResponse(w, http.StatusUnauthorized, errors.New("unauthorized"))
				return
			}

			claims, err := a.verifier.Verify(r.Context(), rawToken)
			if err != nil {
				// Detail stays in the log; the client gets a bare 401 so the
				// response can't be used to probe why a token was rejected.
				log.Printf("auth: token rejected for %s %s: %v", r.Method, r.URL.Path, err)
				util.ErrorResponse(w, http.StatusUnauthorized, errors.New("unauthorized"))
				return
			}

			if !hasRole(claims.Role, roles) {
				util.ErrorResponse(w, http.StatusForbidden, errors.New("access denied"))
				return
			}

			// Resolve *after* the role check so an unauthorized caller can never
			// cause a user row to be provisioned.
			user, err := a.identity.Resolve(r.Context(), claims)
			if err != nil {
				if errors.Is(err, service.ErrIdentityConflict) {
					util.ErrorResponse(w, http.StatusConflict,
						errors.New("this email is already linked to a different account"))
					return
				}
				log.Printf("auth: resolving identity for subject %s failed: %v", claims.Subject, err)
				util.ErrorResponse(w, http.StatusInternalServerError, errors.New("could not resolve user"))
				return
			}

			ctx := context.WithValue(r.Context(), configs.UserIDKey, user.ID.String())
			ctx = context.WithValue(ctx, configs.UserRoleKey, string(user.Role))
			ctx = context.WithValue(ctx, claimsContextKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// hasRole reports whether role satisfies the requirement. An empty requirement
// means "authenticated is enough".
func hasRole(role domain.UserRole, required []domain.UserRole) bool {
	if len(required) == 0 {
		return true
	}
	for _, want := range required {
		if role == want {
			return true
		}
	}
	return false
}

// bearerToken extracts the credential from the Authorization header.
func bearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no auth header")
	}
	scheme, token, found := strings.Cut(authHeader, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		return "", errors.New("invalid auth header")
	}
	return strings.TrimSpace(token), nil
}
