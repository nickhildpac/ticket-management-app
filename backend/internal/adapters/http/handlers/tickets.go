package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
	"github.com/nickhildpac/ticket-management-app/pkg/util"
)

type TicketPayload struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *Handler) GetAllTickets(w http.ResponseWriter, r *http.Request) {
	tickets, err := h.ticketService.ListAll(r.Context(), 20, 0)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	util.WriteResponse(w, http.StatusOK, tickets)
}

func (h *Handler) GetTickets(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(configs.UsernameKey).(string)
	tickets, err := h.ticketService.ListByCreator(r.Context(), username, 20, 0)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	util.WriteResponse(w, http.StatusOK, tickets)
}

func (h *Handler) GetTicket(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	tid, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	ticket, err := h.ticketService.GetTicket(r.Context(), tid)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	util.WriteResponse(w, http.StatusOK, ticket)
}

func (h *Handler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	var payload TicketPayload
	username := r.Context().Value(configs.UsernameKey).(string)
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	ticket, err := h.ticketService.CreateTicket(r.Context(), domain.Ticket{
		Title:       payload.Title,
		Description: payload.Description,
		CreatedBy:   username,
	})
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	util.WriteResponse(w, http.StatusAccepted, ticket)
}
