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
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	arg := db.ListCommentParams{
		TicketID: tid,
		Offset:   0,
		Limit:    10,
	}
	comments, err := repo.Store.ListComment(r.Context(), arg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = json.NewEncoder(w).Encode(comments)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (repo *Repository) GetCommentHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	comment, err := repo.Store.GetComment(r.Context(), tid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (repo *Repository) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	var payload Comment
	username := r.Context().Value(config.UsernameKey).(string)
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	arg := db.CreateCommentParams{
		TicketID:    payload.TicketID,
		Description: payload.Description,
		CreatedBy:   username,
	}
	comment, err := repo.Store.CreateComment(r.Context(), arg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)

}
