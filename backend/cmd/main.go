package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/nickhildpac/ticket-management-app/internal/config"
	db "github.com/nickhildpac/ticket-management-app/internal/db/sqlc"
	"github.com/nickhildpac/ticket-management-app/internal/handlers"
)

func main() {
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatal("failed to load config ", err)
	}
	conn, err := sql.Open("postgres", conf.DSN)
	if err != nil {
		log.Fatal("failed to connect to db ", err)
	}
	log.Println("DB connected successfully")
	store := db.NewStore(conn)
	repo := handlers.NewRepo(conf, store)
	handlers.NewHandlers(repo)
	log.Printf("server is listening on port %d ", conf.ADDR)
	err = http.ListenAndServe(fmt.Sprintf(":%d", conf.ADDR), mount(conf))
	if err != nil {
		log.Fatal(err)
	}
}
