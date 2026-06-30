package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/nickhildpac/ticket-management-app/internal/application/apperrors"
	"github.com/nickhildpac/ticket-management-app/internal/application/authorization"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/pkg/util"
)

func writeHandlerError(w http.ResponseWriter, r *http.Request, err error) {
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
		writeInternalError(w, r, err)
	}
}

func writeInternalError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error request_id=%s error=%v", middleware.GetReqID(r.Context()), err)
	util.ErrorResponse(w, http.StatusInternalServerError, err)
}
