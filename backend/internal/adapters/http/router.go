// Package http sets up the HTTP router and middleware wiring.
package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/nickhildpac/ticket-management-app/docs" // Import generated docs
	"github.com/nickhildpac/ticket-management-app/internal/adapters/http/handlers"
	middlewares "github.com/nickhildpac/ticket-management-app/internal/adapters/http/middleware"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

func Router(conf *configs.Config, h *handlers.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middlewares.EnableCORS)

	// Health check endpoint
	r.Get("/health", h.HealthCheck)

	// Prometheus metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// Swagger documentation endpoint
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/api/v1", func(r chi.Router) {
		// Public endpoints
		r.Get("/health", h.HealthCheck)
		r.Post("/login", h.Login)
		r.Get("/logout", h.Logout)
		r.Post("/user", h.CreateUser)
		r.Get("/refresh", h.RefreshToken)
		r.With(middlewares.AuthRequired(conf)).Get("/me", h.GetMe)

		// Ticket routes (authenticated)
		r.Route("/ticket", func(mux chi.Router) {
			mux.Use(middlewares.AuthRequired(conf))
			mux.Get("/stats", h.GetTicketStats)
			mux.Get("/all", h.GetTickets)
			mux.Get("/assigned", h.GetAssignedTickets)
			mux.Post("/", h.CreateTicket)
			mux.Get("/number/{number}", h.GetTicketByNumber)
			mux.Get("/{id}", h.GetTicket)
			mux.Patch("/{id}", h.UpdateTicket)
			mux.Delete("/{id}", h.DeleteTicket)
			mux.Get("/{id}/comments", h.GetComments)
		})

		// Comment routes (authenticated)
		r.With(middlewares.AuthRequired(conf)).Post("/comment", h.CreateComment)
		r.With(middlewares.AuthRequired(conf)).Get("/comment/{id}", h.GetComment)

		// User routes (authenticated) - for getting user list for assignments
		r.With(middlewares.AuthRequired(conf)).Get("/users", h.GetBasicUsers)
		r.With(middlewares.AuthRequired(conf)).Get("/me", h.GetMe)
		r.With(middlewares.AuthRequired(conf)).Patch("/me", h.PatchMe)

		// Admin-only user management routes
		r.Route("/admin/users", func(mux chi.Router) {
			mux.Use(middlewares.AdminRequired(conf))
			mux.Get("/", h.GetAllUsers)
			mux.Put("/{id}/role", h.UpdateUserRole)
			mux.Delete("/{id}", h.DeleteUser)
		})

		// Legacy admin endpoint (can be deprecated)
		r.With(middlewares.AdminRequired(conf)).Get("/admin/tickets", h.GetAllTickets)
	})
	return r
}
