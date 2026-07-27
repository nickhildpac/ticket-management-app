// Command ingest chunks and embeds a knowledge-base directory into the vector
// store.
//
// Usage: ingest [path]   (default: ./knowledge)
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/nickhildpac/ticket-management-ai-service/internal/app"
	"github.com/nickhildpac/ticket-management-ai-service/internal/config"
	"github.com/nickhildpac/ticket-management-ai-service/internal/rag"
)

func main() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		_, _ = out.Write([]byte(
			"Ingest a knowledge base into the vector store.\n\n" +
				"Usage: ingest [path]\n\n" +
				"  path  Directory of text files to ingest, recursively (default: ./knowledge).\n"))
	}
	flag.Parse()

	root := flag.Arg(0)
	if root == "" {
		root = "knowledge"
	}
	if _, err := os.Stat(root); err != nil {
		slog.Error("knowledge path not found", "path", root, "error", err)
		os.Exit(1)
	}

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

	store := app.NewStore(db, cfg)

	n, err := rag.IngestPath(context.Background(), store, root)
	if err != nil {
		slog.Error("ingest failed", "path", root, "ingested_chunks", n, "error", err)
		os.Exit(1)
	}
	slog.Info("ingest complete", "chunks", n, "path", root)
}
