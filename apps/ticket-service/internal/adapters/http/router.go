// Package http sets up the HTTP router and middleware wiring.
package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/nickhildpac/ticket-management-app/docs" // Import generated docs
	"github.com/nickhildpac/ticket-management-app/internal/adapters/http/handlers"
	middlewares "github.com/nickhildpac/ticket-management-app/internal/adapters/http/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Router wires the API. Authentication is Keycloak-issued bearer tokens checked
// by auth; the role names below are realm roles carried in the token.
func Router(auth *middlewares.Authenticator, h *handlers.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middlewares.EnableCORS)

	authenticated := auth.AuthRequired()
	adminOnly := auth.AdminRequired()

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

		// Where to authenticate. Login, registration, token refresh and logout
		// all happen against Keycloak directly from the browser (Authorization
		// Code + PKCE), so this service exposes no credential endpoints.
		r.Get("/auth/config", h.AuthConfig)

		// Ticket routes (authenticated)
		r.Route("/tickets", func(mux chi.Router) {
			mux.Use(authenticated)
			mux.Get("/stats", h.GetTicketStats)
			mux.Get("/", h.GetTickets)
			mux.Post("/", h.CreateTicket)
			mux.Get("/number/{number}", h.GetTicketByNumber)
			mux.Get("/{id}", h.GetTicket)
			mux.Patch("/{id}", h.UpdateTicket)
			mux.Delete("/{id}", h.DeleteTicket)
			mux.Get("/{id}/comments", h.GetComments)
		})

		// Comment routes (authenticated)
		r.With(authenticated).Post("/comments", h.CreateComment)
		r.With(authenticated).Get("/comments/{id}", h.GetComment)

		// User routes (authenticated) - for getting user list for assignments
		r.With(authenticated).Get("/users", h.GetBasicUsers)
		r.With(authenticated).Get("/me", h.GetMe)
		r.With(authenticated).Patch("/me", h.PatchMe)

		// Admin-only user management routes
		r.Route("/admin/users", func(mux chi.Router) {
			mux.Use(adminOnly)
			mux.Get("/", h.GetAllUsers)
			mux.Put("/{id}/role", h.UpdateUserRole)
			mux.Delete("/{id}", h.DeleteUser)
		})

		// Legacy admin endpoint (can be deprecated)
		r.With(adminOnly).Get("/admin/tickets", h.GetAllTickets)

		// Admin-only KB document ingestion (proxies multipart upload to ai-service).
		r.With(adminOnly).Post("/admin/documents", h.IngestDocuments)
	})
	return r
}
