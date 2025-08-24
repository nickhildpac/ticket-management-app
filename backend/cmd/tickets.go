package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	db "github.com/nickhildpac/ticket-management-app/internal/db/sqlc"
)

type Ticket struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedBy   string `json:"created_by"`
}

func (app *application) getTicketsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	arg := db.ListTicketsParams{
		CreatedBy: username,
		Limit:     20,
		Offset:    0,
	}
	tickets, err := app.Store.ListTickets(r.Context(), arg)
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

func (app *application) getTicketHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	ticket, err := app.Store.GetTicket(r.Context(), tid)
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

func (app *application) createTicketHandler(w http.ResponseWriter, r *http.Request) {
	var payload Ticket
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	arg := db.CreateTicketParams{
		Title:       payload.Title,
		Description: payload.Description,
		CreatedBy:   payload.CreatedBy,
	}
	ticket, err := app.Store.CreateTicket(r.Context(), arg)
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
