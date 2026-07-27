// Command worker consumes the ticket-events Redis stream, triages each ticket
// and applies the decision by commenting back via the Go ticket API.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"

	"github.com/nickhildpac/ticket-management-ai-service/internal/app"
	"github.com/nickhildpac/ticket-management-ai-service/internal/config"
	"github.com/nickhildpac/ticket-management-ai-service/internal/ticketapi"
	"github.com/nickhildpac/ticket-management-ai-service/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	app.SetupLogging(cfg)

	db, err := app.OpenDB(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Error("invalid REDIS_URL", "error", err)
		os.Exit(1)
	}
	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()

	agent := app.NewAgent(app.NewStore(db, cfg), cfg)
	tickets := ticketapi.New(ticketapi.Settings{
		TicketServiceURL:   cfg.TicketServiceURL,
		JWTSecret:          cfg.JWTSecret,
		JWTIssuer:          cfg.JWTIssuer,
		JWTAudience:        cfg.JWTAudience,
		ServiceAccountID:   cfg.ServiceAccountID,
		ServiceAccountRole: cfg.ServiceAccountRole,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Fail loudly at startup if the callback path is misconfigured, rather
	// than 403-ing on every event with the failure buried in the logs. This is
	// deliberately not fatal: the worker still drains the stream, and the
	// operator sees one obvious error instead of a silent backlog.
	if err := tickets.VerifyAccess(ctx); err != nil {
		slog.Error("ticket API callback check FAILED — triage results won't be applied. "+
			"Check AI_SERVICE_ACCOUNT_ID/ROLE, JWT_SECRET, and TICKET_SERVICE_URL.",
			"error", err)
	} else {
		slog.Info("ticket API callback verified")
	}

	w := worker.New(rdb, agent, tickets, cfg.ConsumerGroup, cfg.ConsumerName)
	if err := w.Run(ctx); err != nil {
		slog.Error("worker stopped", "error", err)
		os.Exit(1)
	}
	slog.Info("worker shut down")
}
