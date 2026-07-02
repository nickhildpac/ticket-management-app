// Package main boots the ticket management API server.
//
//	@title			Ticket Management System API
//	@version		1.0
//	@description	Complete API documentation for the ticket management system
//	@host			localhost:8080
//	@BasePath		/api/v1
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name							Authorization
//	@description					Type "Bearer" followed by a space and JWT token.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	adapterdb "github.com/nickhildpac/ticket-management-app/internal/adapters/db"
	sqldb "github.com/nickhildpac/ticket-management-app/internal/adapters/db/sqlc"
	"github.com/nickhildpac/ticket-management-app/internal/adapters/events"
	httpadapter "github.com/nickhildpac/ticket-management-app/internal/adapters/http"
	httphandlers "github.com/nickhildpac/ticket-management-app/internal/adapters/http/handlers"
	"github.com/nickhildpac/ticket-management-app/internal/application/service"
	"github.com/nickhildpac/ticket-management-app/internal/ports"
	"github.com/redis/go-redis/v9"

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

	// Domain-event publisher (transactional outbox). If a broker is configured,
	// start the relay that drains the outbox to Redis for the AI/RAG service.
	var publisher ports.EventPublisher = events.NewOutboxPublisher(conn)
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
		relay := events.NewRelay(conn, rdb)
		go relay.Run(context.Background())
		log.Printf("outbox relay started, publishing to redis at %s", redisAddr)
	}

	userSvc := service.NewUserService(userRepo)
	autoAssignmentSvc := service.NewAutoAssignmentService(userRepo, ticketRepo)
	ticketSvc := service.NewTicketService(ticketRepo, autoAssignmentSvc, publisher)
	commentSvc := service.NewCommentService(commentRepo, ticketRepo)

	handler := httphandlers.NewHandler(conf, userSvc, ticketSvc, commentSvc)

	addr := fmt.Sprintf(":%d", conf.ADDR)
	log.Printf("starting HTTP server on port %d", conf.ADDR)

	if err := http.ListenAndServe(addr, httpadapter.Router(conf, handler)); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}
