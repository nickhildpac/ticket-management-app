package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	Querier
	// ExecTx runs fn inside a database transaction, giving it both the raw
	// *sql.Tx (for statements sqlc doesn't generate, e.g. the outbox insert)
	// and a tx-bound *Queries. Commits on nil error, rolls back otherwise.
	ExecTx(ctx context.Context, fn func(tx *sql.Tx, q *Queries) error) error
}

type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	store := &SQLStore{
		db:      db,
		Queries: New(db),
	}
	return store
}

func (s *SQLStore) ExecTx(ctx context.Context, fn func(tx *sql.Tx, q *Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx, New(tx)); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v; rollback error: %w", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}
