package db

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/nickhildpac/ticket-management-app/internal/application/apperrors"
)

func normalizeDBError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return apperrors.ErrNotFound
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		switch pqErr.Constraint {
		case "users_email_key":
			return apperrors.ErrDuplicateEmail
		case "idx_users_keycloak_id":
			return apperrors.ErrDuplicateIdentity
		}
	}

	return err
}
