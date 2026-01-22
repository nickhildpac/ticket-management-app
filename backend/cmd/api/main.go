// Package main boots the ticket management API server.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	adapterdb "github.com/nickhildpac/ticket-management-app/internal/adapters/db"
	sqldb "github.com/nickhildpac/ticket-management-app/internal/adapters/db/sqlc"
	httpadapter "github.com/nickhildpac/ticket-management-app/internal/adapters/http"
	httphandlers "github.com/nickhildpac/ticket-management-app/internal/adapters/http/handlers"
	"github.com/nickhildpac/ticket-management-app/internal/application/service"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
)

func main() {
	conf, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	conn, err := sql.Open("postgres", conf.DSN)
	if err != nil {
		log.Fatalf("failed to open database connection: %v", err)
	}

	if err = conn.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("DB connected successfully")

	store := sqldb.NewStore(conn)
	userRepo := adapterdb.NewUserRepository(store)
	ticketRepo := adapterdb.NewTicketRepository(store)
	commentRepo := adapterdb.NewCommentRepository(store)

	userSvc := service.NewUserService(userRepo)
	autoAssignmentSvc := service.NewAutoAssignmentService(userRepo, ticketRepo)
	ticketSvc := service.NewTicketService(ticketRepo, autoAssignmentSvc)
	commentSvc := service.NewCommentService(commentRepo, ticketRepo)

	handler := httphandlers.NewHandler(conf, userSvc, ticketSvc, commentSvc)

	addr := fmt.Sprintf(":%d", conf.ADDR)
	log.Printf("starting HTTP server on port %d", conf.ADDR)

	if err := http.ListenAndServe(addr, httpadapter.Router(conf, handler)); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}
