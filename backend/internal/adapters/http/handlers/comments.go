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

type CommentPayload struct {
	TicketID    int64  `json:"ticket_id"`
	Description string `json:"description"`
}

func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	tid, err := strconv.ParseInt(idParam, 10, 64)
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
	tid, err := strconv.ParseInt(idParam, 10, 64)
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
	username := r.Context().Value(configs.UsernameKey).(string)
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	comment, err := h.commentService.CreateComment(r.Context(), domain.Comment{
		TicketID:    payload.TicketID,
		Description: payload.Description,
		CreatedBy:   username,
	})
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	util.WriteResponse(w, http.StatusAccepted, comment)
}
