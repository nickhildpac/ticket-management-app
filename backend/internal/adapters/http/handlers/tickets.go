package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Skills      []string `json:"skills"`
}

// @Summary		Get ticket statistics
// @Description	Get ticket statistics based on user role (admin/agent see all, users see own)
// @Tags			Tickets
// @Produce		json
// @Security		BearerAuth
// @Success		200	{object}	TicketStats
// @Failure		401	{object}	map[string]string
// @Router			/ticket/stats [get]
func (h *Handler) GetTicketStats(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(configs.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	// Role-scoped listing matches TicketService.ListAll (admin: all, agent: assigned, user: created)
	tickets, err := h.ticketService.ListAll(r.Context(), 1000, 0)
	if err != nil {
		writeHandlerError(w, err)
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
	Skills      *[]string    `json:"skills"`
}

func (h *Handler) GetAllTickets(w http.ResponseWriter, r *http.Request) {
	tickets, err := h.ticketService.ListAll(r.Context(), 20, 0)
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	util.WriteResponse(w, http.StatusOK, h.ticketSummariesWithCreators(r.Context(), tickets))
}

// @Summary		Get all tickets
// @Description	Retrieve a list of all tickets
// @Tags			Tickets
// @Produce		json
// @Security		BearerAuth
// @Success		200	{array}		TicketSummaryResponse
// @Failure		401	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Router			/ticket/all [get]
func (h *Handler) GetTickets(w http.ResponseWriter, r *http.Request) {
	tickets, err := h.ticketService.ListAll(r.Context(), 20, 0)
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	util.WriteResponse(w, http.StatusOK, h.ticketSummariesWithCreators(r.Context(), tickets))
}

// @Summary		Get assigned tickets
// @Description	Retrieve tickets assigned to the current user
// @Tags			Tickets
// @Produce		json
// @Security		BearerAuth
// @Success		200	{array}		TicketSummaryResponse
// @Failure		401	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Router			/ticket/assigned [get]
func (h *Handler) GetAssignedTickets(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(configs.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	tickets, err := h.ticketService.ListByAssignee(r.Context(), userID, 20, 0)
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	util.WriteResponse(w, http.StatusOK, h.ticketSummariesWithCreators(r.Context(), tickets))
}

// @Summary		Get ticket by ID
// @Description	Retrieve full details of a specific ticket
// @Tags			Tickets
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		string				true	"Ticket UUID"	format(uuid)
// @Success		200		{object}	TicketResponse
// @Failure		400		{object}	map[string]string
// @Failure		401		{object}	map[string]string
// @Failure		403		{object}	map[string]string
// @Failure		404		{object}	map[string]string
// @Router			/ticket/{id} [get]
func (h *Handler) GetTicket(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	tid, err := uuid.Parse(idParam)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	ticket, err := h.ticketService.GetTicket(r.Context(), tid)
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	// Fetch creator details
	creator, err := h.userService.GetUserByID(r.Context(), ticket.CreatedBy)
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	resp := TicketResponse{
		TicketID:     ticket.ID,
		TicketNumber: ticket.TicketNumber,
		Title:        ticket.Title,
		Description:  ticket.Description,
		CreatedBy:    ticket.CreatedBy,
		Creator:      newUserInfo(creator),
		CreatedAt:    ticket.CreatedAt,
		UpdatedAt:    ticket.UpdatedAt,
		State:        ticket.State.String(),
		Priority:     ticket.Priority.String(),
		AssignedTo:   ticket.AssignedTo,
		Skills:       ticket.Skills.ToSlice(),
	}
	util.WriteResponse(w, http.StatusOK, resp)
}

// @Summary		Get ticket by number
// @Description	Retrieve ticket by its ticket number
// @Tags			Tickets
// @Produce		json
// @Security		BearerAuth
// @Param			number	path		int				true	"Ticket Number"
// @Success		200		{object}	TicketResponse
// @Failure		400		{object}	map[string]string
// @Failure		401		{object}	map[string]string
// @Failure		403		{object}	map[string]string
// @Failure		404		{object}	map[string]string
// @Router			/ticket/number/{number} [get]
func (h *Handler) GetTicketByNumber(w http.ResponseWriter, r *http.Request) {
	numberParam := chi.URLParam(r, "number")
	var ticketNumber int64
	_, err := fmt.Sscanf(numberParam, "%d", &ticketNumber)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, errors.New("invalid ticket number"))
		return
	}

	ticket, err := h.ticketService.GetTicketByNumber(r.Context(), ticketNumber)
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	// Fetch creator details
	creator, err := h.userService.GetUserByID(r.Context(), ticket.CreatedBy)
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	resp := TicketResponse{
		TicketID:     ticket.ID,
		TicketNumber: ticket.TicketNumber,
		Title:        ticket.Title,
		Description:  ticket.Description,
		CreatedBy:    ticket.CreatedBy,
		Creator:      newUserInfo(creator),
		CreatedAt:    ticket.CreatedAt,
		UpdatedAt:    ticket.UpdatedAt,
		State:        ticket.State.String(),
		Priority:     ticket.Priority.String(),
		AssignedTo:   ticket.AssignedTo,
		Skills:       ticket.Skills.ToSlice(),
	}
	util.WriteResponse(w, http.StatusOK, resp)
}

// @Summary		Create new ticket
// @Description	Create a new ticket (accessible to all authenticated users)
// @Tags			Tickets
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			request	body		object{title=string,description=string,skills=[]string}	true	"Ticket creation details"
// @Success		202		{object}	TicketSummaryResponse
// @Failure		400		{object}	map[string]string
// @Failure		401		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router			/ticket [post]
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

	// Validate skills
	skills, err := domain.NewSkills(payload.Skills)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	ticket, err := h.ticketService.CreateTicket(r.Context(), domain.Ticket{
		Title:       payload.Title,
		Description: payload.Description,
		CreatedBy:   userID,
		Skills:      *skills,
	})
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// Return summary response
	summary := TicketSummaryResponse{
		ID:           ticket.ID,
		TicketNumber: ticket.TicketNumber,
		CreatedBy:    ticket.CreatedBy,
		Title:        ticket.Title,
		Description:  ticket.Description,
		State:        ticket.State.String(),
		Priority:     ticket.Priority.String(),
		CreatedAt:    ticket.CreatedAt,
		UpdatedAt:    ticket.UpdatedAt,
	}
	if creator, err := h.userService.GetUserByID(r.Context(), ticket.CreatedBy); err == nil && creator != nil {
		ui := newUserInfo(creator)
		summary.Creator = &ui
	}
	util.WriteResponse(w, http.StatusAccepted, summary)
}

// @Summary		Update ticket
// @Description	Partially update ticket fields. Valid state transitions: open→pending/cancelled, pending→open/resolved/cancelled, resolved→open/pending/closed/cancelled
// @Tags			Tickets
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		string									true	"Ticket UUID"	format(uuid)
// @Param			request	body		object{title=string,description=string,state=string,priority=string,assigned_to=[]string,skills=[]string}	false	"Fields to update"
// @Success		200		{object}	TicketSummaryResponse
// @Failure		400		{object}	map[string]string	"Invalid state transition or invalid priority"
// @Failure		401		{object}	map[string]string
// @Failure		403		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router			/ticket/{id} [patch]
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
	if payload.Skills != nil {
		skills, err := domain.NewSkills(*payload.Skills)
		if err != nil {
			util.ErrorResponse(w, http.StatusBadRequest, err)
			return
		}
		ticket.Skills = *skills
		changed = true
		updatedFields = append(updatedFields, "skills")
	}

	if !changed {
		util.ErrorResponse(w, http.StatusBadRequest, errors.New("no fields provided to update"))
		return
	}

	updated, err := h.ticketService.UpdateTicket(r.Context(), *ticket, updatedFields)
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	// Return summary response
	summary := TicketSummaryResponse{
		ID:           updated.ID,
		TicketNumber: updated.TicketNumber,
		CreatedBy:    updated.CreatedBy,
		Title:        updated.Title,
		Description:  updated.Description,
		State:        updated.State.String(),
		Priority:     updated.Priority.String(),
		CreatedAt:    updated.CreatedAt,
		UpdatedAt:    updated.UpdatedAt,
	}
	if creator, err := h.userService.GetUserByID(r.Context(), updated.CreatedBy); err == nil && creator != nil {
		ui := newUserInfo(creator)
		summary.Creator = &ui
	}
	util.WriteResponse(w, http.StatusOK, summary)
}

// @Summary		Delete ticket
// @Description	Permanently delete a ticket (admin/agent only for tickets not created by them)
// @Tags			Tickets
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		string				true	"Ticket UUID"	format(uuid)
// @Success		204
// @Failure		400		{object}	map[string]string
// @Failure		401		{object}	map[string]string
// @Failure		403		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router			/ticket/{id} [delete]
func (h *Handler) DeleteTicket(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	tid, err := uuid.Parse(idParam)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	err = h.ticketService.DeleteTicket(r.Context(), tid)
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	util.WriteResponse(w, http.StatusNoContent, nil)
}

func (h *Handler) ticketSummariesWithCreators(ctx context.Context, tickets []domain.Ticket) []TicketSummaryResponse {
	uniqueCreators := make(map[uuid.UUID]struct{})
	for _, t := range tickets {
		if t.CreatedBy != uuid.Nil {
			uniqueCreators[t.CreatedBy] = struct{}{}
		}
	}
	usersByID := make(map[uuid.UUID]*domain.User, len(uniqueCreators))
	for id := range uniqueCreators {
		u, err := h.userService.GetUserByID(ctx, id)
		if err == nil && u != nil {
			usersByID[id] = u
		}
	}

	summaries := make([]TicketSummaryResponse, 0, len(tickets))
	for _, t := range tickets {
		s := TicketSummaryResponse{
			ID:           t.ID,
			TicketNumber: t.TicketNumber,
			CreatedBy:    t.CreatedBy,
			Title:        t.Title,
			Description:  t.Description,
			State:        t.State.String(),
			Priority:     t.Priority.String(),
			CreatedAt:    t.CreatedAt,
			UpdatedAt:    t.UpdatedAt,
		}
		if u, ok := usersByID[t.CreatedBy]; ok {
			ui := newUserInfo(u)
			s.Creator = &ui
		} else {
			// Always emit creator so clients get a stable shape; names/email empty if user missing.
			s.Creator = &UserInfo{ID: t.CreatedBy}
		}
		summaries = append(summaries, s)
	}
	return summaries
}
