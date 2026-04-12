package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
	Priority    string   `json:"priority"`
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
	counts, err := h.ticketService.GetTicketStats(r.Context())
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	util.WriteResponse(w, http.StatusOK, TicketStats{
		Total: counts.Total, Open: counts.Open, Pending: counts.Pending,
		Resolved: counts.Resolved, Mine: counts.Mine,
	})
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
	tickets, err := h.ticketService.ListAll(r.Context(), 20, 0, domain.TicketListSortCreatedDesc)
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	util.WriteResponse(w, http.StatusOK, h.ticketSummariesWithCreators(r.Context(), tickets))
}

// @Summary		Get all tickets
// @Description	Retrieve a list of tickets. Without filter query parameters, listing follows role rules (admin: all, agent: assigned, user: created). With any of state, priority, created_by, assignee, assigned_to, or ticket_number, filters are applied with the same role scoping (non-admins cannot widen scope via query params).
// @Tags			Tickets
// @Produce		json
// @Security		BearerAuth
// @Param			limit			query		int		false	"Page size (default 20, max 100)"
// @Param			offset			query		int		false	"Offset (default 0)"
// @Param			state			query		string	false	"Filter by state (e.g. open, pending)"
// @Param			priority		query		string	false	"Filter by priority (critical, high, medium, low)"
// @Param			created_by		query		string	false	"Filter by creator user UUID (admin only; standard users are scoped to self)"
// @Param			assignee		query		string	false	"Filter by assignee user UUID (admin or matching agent)"
// @Param			assigned_to		query		string	false	"Alias for assignee"
// @Param			ticket_number	query		int		false	"Filter by ticket number"
// @Param			sort			query		string	false	"Sort field: ticket_number or created_at (default created_at)"
// @Param			order			query		string	false	"asc or desc (default desc)"
// @Success		200	{array}		TicketSummaryResponse
// @Failure		400	{object}	map[string]string
// @Failure		401	{object}	map[string]string
// @Failure		403	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Router			/ticket/all [get]
func (h *Handler) GetTickets(w http.ResponseWriter, r *http.Request) {
	filterParams, useFilters, err := parseListTicketsQuery(r.URL.Query())
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	var tickets []domain.Ticket
	if useFilters {
		tickets, err = h.ticketService.ListTicketsWithFilters(r.Context(), filterParams)
	} else {
		tickets, err = h.ticketService.ListAll(r.Context(), filterParams.LimitVal, filterParams.OffsetVal, filterParams.SortVal)
	}
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	util.WriteResponse(w, http.StatusOK, h.ticketSummariesWithCreators(r.Context(), tickets))
}

const (
	defaultTicketListLimit int32 = 20
	maxTicketListLimit     int32 = 100
)

func parseListTicketsQuery(q url.Values) (params domain.ListAllTicketsByStatePriorityParams, useFilters bool, err error) {
	params.LimitVal = defaultTicketListLimit
	params.OffsetVal = 0

	if v := strings.TrimSpace(q.Get("limit")); v != "" {
		l, convErr := strconv.ParseInt(v, 10, 32)
		if convErr != nil || l < 1 || l > int64(maxTicketListLimit) {
			return params, false, fmt.Errorf("invalid limit")
		}
		params.LimitVal = int32(l)
	}
	if v := strings.TrimSpace(q.Get("offset")); v != "" {
		o, convErr := strconv.ParseInt(v, 10, 32)
		if convErr != nil || o < 0 {
			return params, false, fmt.Errorf("invalid offset")
		}
		params.OffsetVal = int32(o)
	}

	if v := strings.TrimSpace(q.Get("state")); v != "" {
		st, stErr := domain.GetTicketState(v)
		if stErr != nil {
			return params, false, stErr
		}
		params.FilterState = sql.NullInt32{Int32: int32(st), Valid: true}
		useFilters = true
	}
	if v := strings.TrimSpace(q.Get("priority")); v != "" {
		pr, prErr := domain.ParseTicketPriority(v)
		if prErr != nil {
			return params, false, prErr
		}
		params.FilterPriority = sql.NullInt32{Int32: int32(pr), Valid: true}
		useFilters = true
	}
	if v := strings.TrimSpace(q.Get("created_by")); v != "" {
		id, idErr := uuid.Parse(v)
		if idErr != nil {
			return params, false, fmt.Errorf("invalid created_by: %w", idErr)
		}
		params.FilterCreatedBy = uuid.NullUUID{UUID: id, Valid: true}
		useFilters = true
	}
	assignee, assigneeSet, assigneeErr := assigneeFromQuery(q)
	if assigneeErr != nil {
		return params, false, assigneeErr
	}
	if assigneeSet {
		params.FilterAssignee = assignee
		useFilters = true
	}
	if v := strings.TrimSpace(q.Get("ticket_number")); v != "" {
		n, nErr := strconv.ParseInt(v, 10, 64)
		if nErr != nil {
			return params, false, fmt.Errorf("invalid ticket_number")
		}
		params.FilterTicketNumber = sql.NullInt64{Int64: n, Valid: true}
		useFilters = true
	}

	sortVal, sortErr := domain.TicketListSortFromQuery(q.Get("sort"), q.Get("order"))
	if sortErr != nil {
		return params, false, sortErr
	}
	params.SortVal = sortVal

	return params, useFilters, nil
}

func assigneeFromQuery(q url.Values) (uuid.NullUUID, bool, error) {
	a := strings.TrimSpace(q.Get("assignee"))
	b := strings.TrimSpace(q.Get("assigned_to"))
	if a == "" && b == "" {
		return uuid.NullUUID{}, false, nil
	}
	if a != "" && b != "" && a != b {
		return uuid.NullUUID{}, false, errors.New("assignee and assigned_to must match when both are set")
	}
	raw := a
	if raw == "" {
		raw = b
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.NullUUID{}, true, fmt.Errorf("invalid assignee: %w", err)
	}
	return uuid.NullUUID{UUID: id, Valid: true}, true, nil
}

// @Summary		Get assigned tickets
// @Description	Retrieve tickets assigned to the current user. Supports the same filter query parameters as GET /ticket/all (limit, offset, state, priority, ticket_number); scope is always limited to tickets assigned to the caller.
// @Tags			Tickets
// @Produce		json
// @Security		BearerAuth
// @Param			limit			query		int		false	"Page size (default 20, max 100)"
// @Param			offset			query		int		false	"Offset (default 0)"
// @Param			state			query		string	false	"Filter by state (e.g. open, pending)"
// @Param			priority		query		string	false	"Filter by priority (critical, high, medium, low)"
// @Param			ticket_number	query		int		false	"Filter by ticket number"
// @Param			sort			query		string	false	"Sort field: ticket_number or created_at (default created_at)"
// @Param			order			query		string	false	"asc or desc (default desc)"
// @Success		200	{array}		TicketSummaryResponse
// @Failure		400	{object}	map[string]string
// @Failure		401	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Router			/ticket/assigned [get]
func (h *Handler) GetAssignedTickets(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(configs.UserIDKey).(string)
	_, err := uuid.Parse(userIDStr)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	filterParams, _, err := parseListTicketsQuery(r.URL.Query())
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	filterParams.FilterCreatedBy = uuid.NullUUID{}
	// Assignee is forced to JWT user inside the service; ignore assignee/created_by from query for this route.

	tickets, err := h.ticketService.ListAssignedToCurrentUser(r.Context(), filterParams)
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

	var priority domain.TicketPriority
	if strings.TrimSpace(payload.Priority) == "" {
		priority = domain.TicketPriorityLow
	} else {
		priority, err = domain.ParseTicketPriority(payload.Priority)
		if err != nil {
			util.ErrorResponse(w, http.StatusBadRequest, err)
			return
		}
	}

	ticket, err := h.ticketService.CreateTicket(r.Context(), domain.Ticket{
		Title:       payload.Title,
		Description: payload.Description,
		CreatedBy:   userID,
		Skills:      *skills,
		Priority:    priority,
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
		writeHandlerError(w, err)
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
		priority, err := domain.ParseTicketPriority(*payload.Priority)
		if err != nil {
			util.ErrorResponse(w, http.StatusBadRequest, err)
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
	ids := make([]uuid.UUID, 0, len(uniqueCreators))
	for id := range uniqueCreators {
		ids = append(ids, id)
	}
	usersByID, err := h.userService.GetUsersByIDs(ctx, ids)
	if err != nil || usersByID == nil {
		usersByID = make(map[uuid.UUID]*domain.User)
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
