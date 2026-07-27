package httpapi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// requireAuth rejects requests without a valid ticket-service JWT (shared HS256
// secret). This prevents anyone who can reach the port from triggering paid
// model calls.
//
// Note it does not check role: the admin-only enforcement for document ingest
// lives in the ticket-service proxy handler, which mounts it behind
// AdminRequired and forwards the caller's token verbatim.
func requireAuth(secret, issuer, audience string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized", "invalid token", nil)
				return
			}
			if _, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
				}
				return []byte(secret), nil
			}, jwt.WithAudience(audience), jwt.WithIssuer(issuer)); err != nil {
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
	if !found || !strings.EqualFold(scheme, "Bearer") || token == "" {
		return "", false
	}
	return token, true
}
