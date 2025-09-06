package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nickhildpac/ticket-management-app/internal/config"
	db "github.com/nickhildpac/ticket-management-app/internal/db/sqlc"
)

type Ticket struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedBy   string `json:"created_by"`
}

func (repo *Repository) GetTicketsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(config.UsernameKey).(string)
	arg := db.ListTicketsParams{
		CreatedBy: username,
		Limit:     20,
		Offset:    0,
	}
	tickets, err := repo.Store.ListTickets(r.Context(), arg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(tickets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (repo *Repository) GetTicketHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	ticket, err := repo.Store.GetTicket(r.Context(), tid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(ticket)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (repo *Repository) CreateTicketHandler(w http.ResponseWriter, r *http.Request) {
	var payload Ticket
	username := r.Context().Value(config.UsernameKey).(string)
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	arg := db.CreateTicketParams{
		Title:       payload.Title,
		Description: payload.Description,
		CreatedBy:   username,
	}
	ticket, err := repo.Store.CreateTicket(r.Context(), arg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(ticket)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
