package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/nickhildpac/ticket-management-ai-service/internal/keycloak"
)

// TokenVerifier validates an inbound access token. Implemented by
// keycloak.Verifier.
type TokenVerifier interface {
	Verify(ctx context.Context, rawToken string) (*keycloak.Claims, error)
}

// requireAuth rejects requests without a valid Keycloak access token. This
// prevents anyone who can reach the port from triggering paid model calls.
//
// Note it does not check role: the admin-only enforcement for document ingest
// lives in the ticket-service proxy handler, which mounts it behind
// AdminRequired and forwards the caller's token verbatim. That token is issued
// by the same realm and carries the same audience, so it validates here too.
func requireAuth(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized", "invalid token", nil)
				return
			}
			if _, err := verifier.Verify(r.Context(), token); err != nil {
				writeError(w, http.StatusUnauthorized, "unauthorized", "invalid token", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	scheme, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		return "", false
	}
	return strings.TrimSpace(token), true
}
