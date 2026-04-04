package handlers

import (
	"errors"
	"net/http"

	"github.com/nickhildpac/ticket-management-app/internal/application/apperrors"
	"github.com/nickhildpac/ticket-management-app/internal/application/authorization"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/pkg/util"
)

func writeHandlerError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authorization.ErrAccessDenied):
		util.ErrorResponse(w, http.StatusForbidden, err)
	case errors.Is(err, apperrors.ErrNotFound):
		util.ErrorResponse(w, http.StatusNotFound, err)
	case errors.Is(err, apperrors.ErrDuplicateEmail):
		util.ErrorResponse(w, http.StatusConflict, err)
	case errors.Is(err, domain.ErrInvalidStatusTransition), errors.Is(err, apperrors.ErrBadRequest):
		util.ErrorResponse(w, http.StatusBadRequest, err)
	default:
		util.ErrorResponse(w, http.StatusInternalServerError, err)
	}
}
