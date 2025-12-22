// Package http
package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nickhildpac/ticket-management-app/internal/adapters/http/handlers"
	middlewares "github.com/nickhildpac/ticket-management-app/internal/adapters/http/middleware"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
)

func Router(conf *configs.Config, h *handlers.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middlewares.EnableCORS)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", h.HealthCheck)
		r.Post("/login", h.Login)
		r.Get("/logout", h.Logout)
		r.Post("/user", h.CreateUser)
		r.Get("/user/{username}", h.GetUser)
		r.Get("/refresh", h.RefreshToken)

		r.Route("/ticket", func(mux chi.Router) {
			mux.Use(middlewares.AuthRequired(conf))
			mux.Get("/all", h.GetTickets)
			mux.Post("/", h.CreateTicket)
			mux.Get("/{id}", h.GetTicket)
			mux.Get("/{id}/comments", h.GetComments)
		})

		r.With(middlewares.AdminRequired(conf)).Get("/admin/tickets", h.GetAllTickets)
		r.With(middlewares.AuthRequired(conf)).Post("/comment", h.CreateComment)
		r.Get("/comment/{id}", h.GetComment)
	})
	return r
}
