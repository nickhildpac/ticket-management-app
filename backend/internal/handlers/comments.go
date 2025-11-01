package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nickhildpac/ticket-management-app/internal/config"
	db "github.com/nickhildpac/ticket-management-app/internal/db/sqlc"
)

type Comment struct {
	TicketID    int64  `json:"ticket_id"`
	Description string `json:"description"`
	CreatedBy   string `json:"created_by"`
}

func (repo *Repository) GetCommentsHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
	}
	arg := db.ListCommentParams{
		TicketID: tid,
		Offset:   0,
		Limit:    10,
	}
	comments, err := repo.Store.ListComment(r.Context(), arg)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
	}
	writeJson(w, http.StatusOK, comments)
}

func (repo *Repository) GetCommentHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
	}
	comment, err := repo.Store.GetComment(r.Context(), tid)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}
	writeJson(w, http.StatusOK, comment)
}

func (repo *Repository) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	var payload Comment
	username := r.Context().Value(config.UsernameKey).(string)
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}
	arg := db.CreateCommentParams{
		TicketID:    payload.TicketID,
		Description: payload.Description,
		CreatedBy:   username,
	}
	comment, err := repo.Store.CreateComment(r.Context(), arg)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}
	writeJson(w, http.StatusAccepted, comment)
}
