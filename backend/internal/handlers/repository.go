package handlers

import (
	"log"

	"github.com/nickhildpac/ticket-management-app/internal/config"
	db "github.com/nickhildpac/ticket-management-app/internal/db/sqlc"
)

type Repository struct {
	Config *config.Config
	Store  db.Store
}

var Repo *Repository

func NewRepo(config *config.Config, store db.Store) *Repository {
	log.Println(config)
	return &Repository{
		Config: config,
		Store:  store,
	}
}

func NewHandlers(r *Repository) {
	Repo = r
}
