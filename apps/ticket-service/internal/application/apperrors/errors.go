package apperrors

import "errors"

var (
	ErrNotFound       = errors.New("resource not found")
	ErrDuplicateEmail = errors.New("email already exists")
	ErrBadRequest     = errors.New("bad request")
)
