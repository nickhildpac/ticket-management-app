package handlers

import (
	"log"

	"github.com/nickhildpac/ticket-management-app/configs"
	db "github.com/nickhildpac/ticket-management-app/internal/db/sqlc"
)

type Repository struct {
	Config *configs.Config
	Store  db.Store
}

var Repo *Repository

func NewRepo(config *configs.Config, store db.Store) *Repository {
	log.Println(config)
	return &Repository{
		Config: config,
		Store:  store,
	}
}

func NewHandlers(r *Repository) {
	Repo = r
}
