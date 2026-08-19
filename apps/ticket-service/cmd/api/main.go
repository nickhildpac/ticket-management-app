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
	"github.com/nickhildpac/ticket-management-app/internal/adapters/auth"
	adapterdb "github.com/nickhildpac/ticket-management-app/internal/adapters/db"
	sqldb "github.com/nickhildpac/ticket-management-app/internal/adapters/db/sqlc"
	"github.com/nickhildpac/ticket-management-app/internal/adapters/events"
	httpadapter "github.com/nickhildpac/ticket-management-app/internal/adapters/http"
	httphandlers "github.com/nickhildpac/ticket-management-app/internal/adapters/http/handlers"
	middlewares "github.com/nickhildpac/ticket-management-app/internal/adapters/http/middleware"
	"github.com/nickhildpac/ticket-management-app/internal/application/service"
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

	// Go owns the ticket schema (ADR 0002); apply migrations on startup unless
	// disabled (e.g. when a DB is managed externally). Set AUTO_MIGRATE=false to skip.
	if os.Getenv("AUTO_MIGRATE") != "false" {
		migrationsPath := configs.GetString("MIGRATIONS_PATH", "migrations")
		if err := adapterdb.RunMigrations(conn, migrationsPath); err != nil {
			log.Fatalf("failed to run migrations: %v", err)
		}
		log.Println("migrations applied")
	}

	store := sqldb.NewStore(conn)
	userRepo := adapterdb.NewUserRepository(store)
	ticketRepo := adapterdb.NewTicketRepository(store)
	commentRepo := adapterdb.NewCommentRepository(store)

	// Domain events are written to the event_outbox table inside the ticket
	// repository's transactions (transactional outbox). If a broker is
	// configured, start the relay that drains the outbox to Redis for the
	// AI/RAG service.
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
		relay := events.NewRelay(conn, rdb)
		go relay.Run(context.Background())
		log.Printf("outbox relay started, publishing to redis at %s", redisAddr)
	}

	// Authentication is delegated to Keycloak: this service verifies tokens
	// against the realm's JWKS and holds no signing key of its own.
	// See docs/adr/0003-keycloak-authentication.md.
	verifier, err := auth.NewVerifier(context.Background(), auth.Config{
		IssuerURL:    conf.KeycloakIssuerURL,
		DiscoveryURL: conf.KeycloakInternalIssuerURL,
		Audience:     conf.KeycloakAudience,
	})
	if err != nil {
		log.Fatalf("failed to initialise keycloak verifier: %v", err)
	}
	log.Printf("authenticating against keycloak issuer %s", verifier.Issuer())

	// Optional: only needed to write role changes back to the realm. Without it
	// the service runs normally and the admin role endpoint reports itself
	// unavailable rather than making a change Keycloak would later undo.
	adminClient, err := auth.NewAdminClient(auth.AdminConfig{
		// Outbound call, so this must be the reachable URL, not the
		// browser-facing issuer.
		IssuerURL: conf.ReachableKeycloakURL(),
		ClientID:  conf.KeycloakAdminClientID,
		Secret:    conf.KeycloakAdminClientSecret,
	})
	if err != nil {
		log.Fatalf("failed to initialise keycloak admin client: %v", err)
	}
	if adminClient == nil {
		log.Println("keycloak admin credentials not set; role changes must be made in the Keycloak console")
	}

	identitySvc := service.NewIdentityService(userRepo, conf.IdentityCacheTTL)
	userSvc := service.NewUserService(userRepo,
		service.WithRealmRoleWriter(adminClient),
		service.WithIdentityCache(identitySvc),
	)
	autoAssignmentSvc := service.NewAutoAssignmentService(userRepo, ticketRepo)
	ticketSvc := service.NewTicketService(ticketRepo, autoAssignmentSvc)
	commentSvc := service.NewCommentService(commentRepo, ticketRepo)

	handler := httphandlers.NewHandler(conf, userSvc, ticketSvc, commentSvc)
	authenticator := middlewares.NewAuthenticator(verifier, identitySvc)

	addr := fmt.Sprintf(":%d", conf.ADDR)
	log.Printf("starting HTTP server on port %d", conf.ADDR)

	if err := http.ListenAndServe(addr, httpadapter.Router(authenticator, handler)); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}
