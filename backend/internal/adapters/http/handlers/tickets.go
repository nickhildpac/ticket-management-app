package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/application/authorization"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
	"github.com/nickhildpac/ticket-management-app/pkg/util"
)

type TicketStats struct {
	Total    int32 `json:"total"`
	Open     int32 `json:"open"`
	Pending  int32 `json:"pending"`
	Resolved int32 `json:"resolved"`
	Mine     int32 `json:"mine"`
}

type TicketPayload struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *Handler) GetTicketStats(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(configs.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	// Get tickets based on user role
	var tickets []domain.Ticket
	auth, err := authorization.GetAuthContext(r.Context())
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	if auth.Role == domain.RoleAdmin {
		tickets, err = h.ticketService.ListAll(r.Context(), 1000, 0)
	} else if auth.Role == domain.RoleAgent {
		tickets, err = h.ticketService.ListByAssignee(r.Context(), userID, 1000, 0)
	} else {
		tickets, err = h.ticketService.ListByCreator(r.Context(), userID, 1000, 0)
	}

	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// Calculate stats
	stats := TicketStats{}
	stats.Total = int32(len(tickets))

	for _, ticket := range tickets {
		switch ticket.State {
		case domain.TicketStateOpen:
			stats.Open++
		case domain.TicketStatePending:
			stats.Pending++
		case domain.TicketStateResolved:
			stats.Resolved++
		}
	}

	// Count tickets assigned to current user
	for _, ticket := range tickets {
		for _, assignee := range ticket.AssignedTo {
			if assignee == userID {
				stats.Mine++
				break
			}
		}
	}

	util.WriteResponse(w, http.StatusOK, stats)
}

type UpdateTicketPayload struct {
	Title       *string      `json:"title"`
	Description *string      `json:"description"`
	State       *string      `json:"state"`
	Priority    *string      `json:"priority"`
	AssignedTo  *[]uuid.UUID `json:"assigned_to"`
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
	tickets, err := h.ticketService.ListAll(r.Context(), 20, 0)
	if err != nil {
		if err == authorization.ErrAccessDenied {
			util.ErrorResponse(w, http.StatusForbidden, err)
			return
		}
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	util.WriteResponse(w, http.StatusOK, tickets)
}

func (h *Handler) GetAssignedTickets(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(configs.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	tickets, err := h.ticketService.ListByAssignee(r.Context(), userID, 20, 0)
	if err != nil {
		if err == authorization.ErrAccessDenied {
			util.ErrorResponse(w, http.StatusForbidden, err)
			return
		}
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	util.WriteResponse(w, http.StatusOK, tickets)
}

func (h *Handler) GetTicket(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	tid, err := uuid.Parse(idParam)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	ticket, err := h.ticketService.GetTicket(r.Context(), tid)
	if err != nil {
		if err == authorization.ErrAccessDenied {
			util.ErrorResponse(w, http.StatusForbidden, err)
			return
		}
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// Fetch creator details
	creator, err := h.userService.GetUserByID(r.Context(), ticket.CreatedBy)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	resp := TicketResponse{
		TicketID:    ticket.ID,
		Title:       ticket.Title,
		Description: ticket.Description,
		CreatedBy:   ticket.CreatedBy,
		Creator: UserInfo{
			ID:        creator.ID,
			FirstName: creator.FirstName,
			LastName:  creator.LastName,
			Email:     creator.Email,
		},
		CreatedAt:  ticket.CreatedAt,
		State:      ticket.State.String(),
		Priority:   ticket.Priority.String(),
		AssignedTo: ticket.AssignedTo,
	}
	util.WriteResponse(w, http.StatusOK, resp)
}

func (h *Handler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	var payload TicketPayload
	userIDStr := r.Context().Value(configs.UserIDKey).(string)
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	ticket, err := h.ticketService.CreateTicket(r.Context(), domain.Ticket{
		Title:       payload.Title,
		Description: payload.Description,
		CreatedBy:   userID,
	})
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	util.WriteResponse(w, http.StatusAccepted, ticket)
}

func (h *Handler) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	tid, err := uuid.Parse(idParam)
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
		if err == authorization.ErrAccessDenied {
			util.ErrorResponse(w, http.StatusForbidden, err)
			return
		}
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	changed := false
	updatedFields := []string{}
	if payload.Title != nil {
		ticket.Title = *payload.Title
		changed = true
		updatedFields = append(updatedFields, "title")
	}
	if payload.Description != nil {
		ticket.Description = *payload.Description
		changed = true
		updatedFields = append(updatedFields, "description")
	}
	if payload.State != nil {
		state, err := domain.GetTicketState(*payload.State)
		if err != nil {
			util.ErrorResponse(w, http.StatusBadRequest, err)
			return
		}
		ticket.State = state
		changed = true
		updatedFields = append(updatedFields, "state")
	}
	if payload.Priority != nil {
		priority := domain.GetTicketPriority(*payload.Priority)
		if priority == -1 {
			util.ErrorResponse(w, http.StatusBadRequest, errors.New("invalid priority"))
			return
		}
		ticket.Priority = priority
		changed = true
		updatedFields = append(updatedFields, "priority")
	}
	if payload.AssignedTo != nil {
		ticket.AssignedTo = *payload.AssignedTo
		changed = true
		updatedFields = append(updatedFields, "assigned_to")
	}

	if !changed {
		util.ErrorResponse(w, http.StatusBadRequest, errors.New("no fields provided to update"))
		return
	}

	updated, err := h.ticketService.UpdateTicket(r.Context(), *ticket, updatedFields)
	if err != nil {
		if err == authorization.ErrAccessDenied {
			util.ErrorResponse(w, http.StatusForbidden, err)
			return
		}
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	util.WriteResponse(w, http.StatusOK, updated)
}

func (h *Handler) DeleteTicket(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	tid, err := uuid.Parse(idParam)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	err = h.ticketService.DeleteTicket(r.Context(), tid)
	if err != nil {
		if err == authorization.ErrAccessDenied {
			util.ErrorResponse(w, http.StatusForbidden, err)
			return
		}
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	util.WriteResponse(w, http.StatusNoContent, nil)
}
