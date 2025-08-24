package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	db "github.com/nickhildpac/ticket-management-app/internal/db/sqlc"
)

type Comment struct {
	TicketID    int64  `json:"ticket_id"`
	Description string `json:"description"`
	CreatedBy   string `json:"created_by"`
}

func (app *application) getCommentsHandler(w http.ResponseWriter, r *http.Request) {
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
	comments, err := app.Store.ListComment(r.Context(), arg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = json.NewEncoder(w).Encode(comments)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *application) getCommentHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	comment, err := app.Store.GetComment(r.Context(), tid)
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

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	var payload Comment
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	arg := db.CreateCommentParams{
		TicketID:    payload.TicketID,
		Description: payload.Description,
		CreatedBy:   payload.CreatedBy,
	}
	comment, err := app.Store.CreateComment(r.Context(), arg)
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
