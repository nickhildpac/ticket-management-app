package db

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestPostgresConnection(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	require.NoError(t, db.PingContext(ctx))

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, "SELECT 1")
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())
}
