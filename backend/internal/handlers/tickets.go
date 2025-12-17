// Package handlers consists of user and ticket controllers
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	config "github.com/nickhildpac/ticket-management-app/configs"
	db "github.com/nickhildpac/ticket-management-app/internal/db/sqlc"
)

type Ticket struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	ID          int64     `json:"id"`
	AssignedTo  string    `json:"assigned_to"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (repo *Repository) GetAllTicketsHandler(w http.ResponseWriter, r *http.Request) {
	arg := db.ListAllTicketsParams{
		Limit:  20,
		Offset: 0,
	}
	tickets, err := repo.Store.ListAllTickets(r.Context(), arg)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, tickets)
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
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, tickets)
}

func (repo *Repository) GetTicketHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}
	ticket, err := repo.Store.GetTicket(r.Context(), tid)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}
	resp := Ticket{
		Title:       ticket.Title,
		Description: ticket.Description,
		CreatedBy:   ticket.CreatedBy,
		ID:          ticket.ID,
		AssignedTo:  ticket.AssignedTo.String,
		CreatedAt:   ticket.CreatedAt,
		UpdatedAt:   ticket.UpdatedAt.Time,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (repo *Repository) CreateTicketHandler(w http.ResponseWriter, r *http.Request) {
	var payload Ticket
	username := r.Context().Value(config.UsernameKey).(string)
	log.Println(username)
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}
	arg := db.CreateTicketParams{
		Title:       payload.Title,
		Description: payload.Description,
		CreatedBy:   username,
	}
	ticket, err := repo.Store.CreateTicket(r.Context(), arg)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusAccepted, ticket)
}
