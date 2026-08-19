// Command api serves the ai-service HTTP surface: on-demand triage and
// knowledge-base ingestion.
//
// The primary triage path is the async worker (cmd/worker) consuming ticket
// events; this API exists for manual re-runs, testing, and the admin document
// upload the ticket-service proxies through.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nickhildpac/ticket-management-ai-service/internal/app"
	"github.com/nickhildpac/ticket-management-ai-service/internal/config"
	"github.com/nickhildpac/ticket-management-ai-service/internal/httpapi"
	"github.com/nickhildpac/ticket-management-ai-service/internal/keycloak"
)

// shutdownTimeout bounds how long in-flight requests get to finish. Ingest
// embeds every chunk, so a multi-file upload can legitimately take a while.
const shutdownTimeout = 60 * time.Second

func main() {
	// migrateOnly backs `make migrate`: apply the AI-owned schema out of band
	// without starting the server.
	migrateOnly := flag.Bool("migrate-only", false, "apply migrations and exit")
	flag.Parse()

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

	if err := app.Migrate(db); err != nil {
		slog.Error("failed to apply migrations", "error", err)
		os.Exit(1)
	}
	if *migrateOnly {
		slog.Info("migrations applied")
		return
	}

	store := app.NewStore(db, cfg)
	agent := app.NewAgent(store, cfg)

	// Inbound tokens on /triage and /ingest are verified against the realm's
	// JWKS. Discovery may need to retry while Keycloak finishes booting.
	verifier, err := keycloak.NewVerifier(context.Background(), keycloak.VerifierConfig{
		IssuerURL:    cfg.KeycloakIssuerURL,
		DiscoveryURL: cfg.KeycloakInternalIssuerURL,
		Audience:     cfg.KeycloakAudience,
	})
	if err != nil {
		slog.Error("failed to initialise keycloak verifier", "error", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: httpapi.NewRouter(httpapi.Deps{
			Agent:       agent,
			Store:       store,
			APIV1Prefix: cfg.APIV1Prefix,
			Verifier:    verifier,
			CORSOrigins: cfg.CORSOrigins,
		}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("ai-service listening", "port", cfg.Port, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}
