// Package dbmigrate applies the ai-service's own SQL migrations.
package dbmigrate

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // file:// migration source
)

// MigrationsTable is deliberately *not* golang-migrate's default
// ("schema_migrations"): the Go ticket-service applies its own migrations into
// that table in the same database, so the two services would overwrite each
// other's version rows. Each service tracks its own history.
const MigrationsTable = "ai_schema_migrations"

// Run applies the SQL migrations in path (golang-migrate format) against the
// given connection. Idempotent: ErrNoChange is treated as success.
//
// The migrations are written with IF NOT EXISTS throughout, so they also apply
// cleanly to a database whose kb_chunks table was created by the previous
// Alembic-managed revisions.
func Run(conn *sql.DB, path string) error {
	driver, err := migratepg.WithInstance(conn, &migratepg.Config{
		MigrationsTable: MigrationsTable,
	})
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
