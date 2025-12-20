package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/nickhildpac/ticket-management-app/configs"
	"github.com/nickhildpac/ticket-management-app/internal/usecase/port"
)

type Handler struct {
	config         *configs.Config
	userService    port.UserService
	ticketService  port.TicketService
	commentService port.CommentService
}

func NewHandler(cfg *configs.Config, u port.UserService, t port.TicketService, c port.CommentService) *Handler {
	return &Handler{
		config:         cfg,
		userService:    u,
		ticketService:  t,
		commentService: c,
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func errorResponse(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
}
