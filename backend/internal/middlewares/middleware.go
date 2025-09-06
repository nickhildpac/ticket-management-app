package middlewares

import (
	"context"
	"net/http"

	"github.com/nickhildpac/ticket-management-app/internal/config"
	"github.com/nickhildpac/ticket-management-app/internal/env"
	"github.com/nickhildpac/ticket-management-app/pkg/util"
)

func EnableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", env.GetString("FRONTEND_URL", "http://localhost:5173"))
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type,X-CSRF-Token,Authorization")
			return
		} else {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type,X-CSRF-Token,Authorization")
			h.ServeHTTP(w, r)
		}
	})
}

func AuthRequired(conf *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, _, err := util.GetTokenFromHeaderAndVerify(conf, w, r)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), config.UsernameKey, username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
