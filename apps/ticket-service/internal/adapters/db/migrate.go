package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations applies the SQL migrations in `path` (golang-migrate format)
// against the given connection, using the same schema_migrations tracking the
// project's `migrate` CLI uses. Idempotent: ErrNoChange is treated as success.
//
// The Go ticket-service owns the ticket/user/comment schema (ADR 0002), so it
// applies its own migrations on startup rather than relying on an external tool.
func RunMigrations(conn *sql.DB, path string) error {
	driver, err := migratepg.WithInstance(conn, &migratepg.Config{})
	if err != nil {
		return fmt.Errorf("migrate: init driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+path, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate: open source %q: %w", path, err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate: apply: %w", err)
	}
	return nil
}
