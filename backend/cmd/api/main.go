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
		log.Fatal("failed to load config ", err)
	}
	conn, err := sql.Open("postgres", conf.DSN)
	if err != nil {
		log.Fatal("failed to connect to db ", err)
	}
	log.Println("DB connected successfully")

	store := sqldb.NewStore(conn)
	userRepo := adapterdb.NewUserRepository(store)
	ticketRepo := adapterdb.NewTicketRepository(store)
	commentRepo := adapterdb.NewCommentRepository(store)

	userSvc := service.NewUserService(userRepo)
	ticketSvc := service.NewTicketService(ticketRepo)
	commentSvc := service.NewCommentService(commentRepo, ticketRepo)

	handler := httphandlers.NewHandler(conf, userSvc, ticketSvc, commentSvc)

	log.Printf("server is listening on port %d ", conf.ADDR)
	err = http.ListenAndServe(fmt.Sprintf(":%d", conf.ADDR), httpadapter.Router(conf, handler))
	if err != nil {
		log.Fatal(err)
	}
}
