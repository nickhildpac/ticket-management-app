package apperrors

import "errors"

var (
	ErrNotFound       = errors.New("resource not found")
	ErrDuplicateEmail = errors.New("email already exists")
	ErrBadRequest     = errors.New("bad request")
	ErrUnauthorized   = errors.New("unauthorized")
	// ErrDuplicateIdentity means a local row is already linked to the Keycloak
	// subject being provisioned — normally the losing side of a race between two
	// concurrent first requests for the same subject.
	ErrDuplicateIdentity = errors.New("keycloak identity already linked")
)
