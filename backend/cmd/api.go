package main

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func (app *application) mount() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(app.enableCORS)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)
		r.Post("/login", app.login)
		r.Post("/user", app.createUserHandler)
		r.Get("/user/{username}", app.getUserHandler)
		r.Get("/user/{username}/tickets", app.getTicketsHandler)
		r.Get("/tickets", app.getTicketsHandler)
		r.Post("/ticket", app.createTicketHandler)
		r.Get("/ticket/{id}", app.getTicketHandler)
		r.Get("/ticket/{id}/comments", app.getCommentsHandler)
		r.Post("/comment", app.createCommentHandler)
		r.Get("/comment/{id}", app.getCommentHandler)
	})
	return r
}
