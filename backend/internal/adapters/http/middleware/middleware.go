package middlewares

import (
	"context"
	"net/http"

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

func AuthRequired(conf *configs.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := util.GetTokenFromHeaderAndVerify(conf, w, r)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), configs.UsernameKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminRequired(conf *configs.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := util.GetTokenFromHeaderAndVerify(conf, w, r)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if claims.Role != "admin" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			ctx := context.WithValue(r.Context(), configs.UsernameKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
