package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
	"github.com/nickhildpac/ticket-management-app/pkg/util"
)

type TicketPayload struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateTicketPayload struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	State       *string `json:"state"`
	Priority    *string `json:"priority"`
	AssignedTo  *string `json:"assigned_to"`
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

func (h *Handler) GetAssignedTickets(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(configs.UsernameKey).(string)
	tickets, err := h.ticketService.ListByAssignee(r.Context(), username, 20, 0)
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
	resp := TicketResponse{
		TicketID:    ticket.ID,
		Title:       ticket.Title,
		Description: ticket.Description,
		CreatedBy:   ticket.CreatedBy,
		CreatedAt:   ticket.CreatedAt,
		State:       ticket.State.String(),
		Priority:    ticket.Priority.String(),
		AssignedTo:  ticket.AssignedTo,
	}
	util.WriteResponse(w, http.StatusOK, resp)
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

func (h *Handler) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	tid, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	var payload UpdateTicketPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	ticket, err := h.ticketService.GetTicket(r.Context(), tid)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	changed := false
	if payload.Title != nil {
		ticket.Title = *payload.Title
		changed = true
	}
	if payload.Description != nil {
		ticket.Description = *payload.Description
		changed = true
	}
	if payload.State != nil {
		state, err := parseTicketState(*payload.State)
		if err != nil {
			util.ErrorResponse(w, http.StatusBadRequest, err)
			return
		}
		ticket.State = state
		changed = true
	}
	if payload.Priority != nil {
		priority, err := parseTicketPriority(*payload.Priority)
		if err != nil {
			util.ErrorResponse(w, http.StatusBadRequest, err)
			return
		}
		ticket.Priority = priority
		changed = true
	}
	if payload.AssignedTo != nil {
		ticket.AssignedTo = *payload.AssignedTo
		changed = true
	}
	ticket.UpdatedAt = time.Now()

	if !changed {
		util.ErrorResponse(w, http.StatusBadRequest, errors.New("no fields provided to update"))
		return
	}

	updated, err := h.ticketService.UpdateTicket(r.Context(), *ticket)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	util.WriteResponse(w, http.StatusOK, updated)
}

func parseTicketState(state string) (domain.TicketState, error) {
	switch strings.ToLower(state) {
	case "open":
		return domain.TicketStateOpen, nil
	case "pending":
		return domain.TicketStatePending, nil
	case "resolved":
		return domain.TicketStateResolved, nil
	case "closed":
		return domain.TicketStateClosed, nil
	case "cancel", "cancelled", "canceled":
		return domain.TicketStateCancelled, nil
	default:
		return 0, fmt.Errorf("invalid ticket state: %s", state)
	}
}

func parseTicketPriority(priority string) (domain.TicketPriority, error) {
	switch strings.ToLower(priority) {
	case "low":
		return domain.TicketPriorityLow, nil
	case "medium":
		return domain.TicketPriorityMedium, nil
	case "high":
		return domain.TicketPriorityHigh, nil
	case "critical":
		return domain.TicketPriorityCritical, nil
	default:
		return 0, fmt.Errorf("invalid ticket priority: %s", priority)
	}
}
