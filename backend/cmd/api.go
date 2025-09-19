package main

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/nickhildpac/ticket-management-app/internal/config"
	"github.com/nickhildpac/ticket-management-app/internal/handlers"
	"github.com/nickhildpac/ticket-management-app/internal/middlewares"
)

func mount(conf *config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middlewares.EnableCORS)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", handlers.Repo.HealthCheckHandler)
		r.Post("/login", handlers.Repo.Login)
		r.Get("/logout", handlers.Repo.Logout)
		r.Post("/user", handlers.Repo.CreateUserHandler)
		r.Get("/user/{username}", handlers.Repo.GetUserHandler)
		r.Get("/refresh", handlers.Repo.RefreshToken)
		r.Route("/ticket", func(mux chi.Router) {
			mux.Use(middlewares.AuthRequired(conf))
			mux.Get("/admin/tickets", handlers.Repo.GetTicketsHandler)
			mux.Get("/all", handlers.Repo.GetTicketsHandler)
			mux.Post("/", handlers.Repo.CreateTicketHandler)
			mux.Get("/{id}", handlers.Repo.GetTicketHandler)
			mux.Get("/{id}/comments", handlers.Repo.GetCommentsHandler)
		})
		r.With(middlewares.AuthRequired(conf)).Post("/comment", handlers.Repo.CreateCommentHandler)
		r.Get("/comment/{id}", handlers.Repo.GetCommentHandler)
	})
	return r
}
