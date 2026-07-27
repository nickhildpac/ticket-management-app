// Package app wires the ai-service's collaborators together. Both entrypoints
// (the HTTP API and the triage worker) build the same agent, so the
// construction lives here rather than being duplicated per binary.
package app

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	_ "github.com/lib/pq" // postgres driver

	"github.com/nickhildpac/ticket-management-ai-service/internal/config"
	"github.com/nickhildpac/ticket-management-ai-service/internal/dbmigrate"
	"github.com/nickhildpac/ticket-management-ai-service/internal/rag"
	"github.com/nickhildpac/ticket-management-ai-service/internal/triage"
)

// OpenDB connects to Postgres and verifies the connection.
func OpenDB(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

// Migrate applies the ai-service's own migrations (kb_chunks and its indexes).
// The path is resolved relative to the working directory; override it with
// MIGRATIONS_PATH when running from somewhere else.
func Migrate(db *sql.DB) error {
	path := os.Getenv("MIGRATIONS_PATH")
	if path == "" {
		path = "migrations"
	}
	return dbmigrate.Run(db, path)
}

// NewStore builds the pgvector-backed knowledge-base store with the configured
// embedder.
func NewStore(db *sql.DB, cfg *config.Config) *rag.VectorStore {
	embedder := rag.BuildEmbedder(rag.EmbedderSettings{
		OpenAIAPIKey:   cfg.OpenAIAPIKey,
		EmbeddingModel: cfg.EmbeddingModel,
		EmbeddingDim:   cfg.EmbeddingDim,
	})
	if cfg.OpenAIAPIKey == "" {
		slog.Warn("OPENAI_API_KEY unset — falling back to the offline hashing embedder; " +
			"retrieval quality will be poor")
	}
	return rag.NewVectorStore(db, embedder)
}

// NewAgent builds the triage agent over the store, the cross-encoder re-ranker
// and the Anthropic client.
func NewAgent(store *rag.VectorStore, cfg *config.Config) *triage.Agent {
	reranker := rag.NewOpenRouterReranker(cfg.OpenRouterAPIKey, cfg.CrossEncoderModel)
	client := anthropic.NewClient(option.WithAPIKey(cfg.AnthropicAPIKey))
	if strings.TrimSpace(cfg.AnthropicAPIKey) == "" {
		slog.Warn("ANTHROPIC_API_KEY unset — every triage will fail and escalate")
	}
	return triage.NewAgent(&client, store, reranker, triage.Options{
		Model:               cfg.TriageModel,
		ConfidenceThreshold: cfg.AutoAnswerConfidenceThreshold,
		RAGTopK:             cfg.RAGTopK,
		CandidateK:          cfg.RAGCandidateK,
		RerankPool:          cfg.RAGRerankPool,
		RRFK:                cfg.RRFK,
		MaxIterations:       cfg.TriageMaxIterations,
	})
}

// SetupLogging installs a JSON slog handler as the default logger.
func SetupLogging(cfg *config.Config) {
	level := slog.LevelInfo
	if strings.EqualFold(cfg.AppEnv, "development") || strings.EqualFold(cfg.AppEnv, "local") {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
}
