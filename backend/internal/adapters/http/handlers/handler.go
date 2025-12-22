// Package handlers
package handlers

import (
	"github.com/nickhildpac/ticket-management-app/internal/ports"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
)

type Handler struct {
	config         *configs.Config
	userService    ports.UserService
	ticketService  ports.TicketService
	commentService ports.CommentService
}

func NewHandler(cfg *configs.Config, u ports.UserService, t ports.TicketService, c ports.CommentService) *Handler {
	return &Handler{
		config:         cfg,
		userService:    u,
		ticketService:  t,
		commentService: c,
	}
}
