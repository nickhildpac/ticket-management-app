package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
	"github.com/nickhildpac/ticket-management-app/pkg/util"
)

type CommentPayload struct {
	TicketID    string `json:"ticket_id"`
	Description string `json:"description"`
}

func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	tid, err := uuid.Parse(idParam)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	comments, err := h.commentService.ListByTicket(r.Context(), tid, 10, 0)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	util.WriteResponse(w, http.StatusOK, comments)
}

func (h *Handler) GetComment(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	tid, err := uuid.Parse(idParam)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	comment, err := h.commentService.GetComment(r.Context(), tid)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	util.WriteResponse(w, http.StatusOK, comment)
}

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	var payload CommentPayload
	userIDStr := r.Context().Value(configs.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	ticketID, err := uuid.Parse(payload.TicketID)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	comment, err := h.commentService.CreateComment(r.Context(), domain.Comment{
		TicketID:    ticketID,
		Description: payload.Description,
		CreatedBy:   user.ID,
	})
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	util.WriteResponse(w, http.StatusAccepted, comment)
}
