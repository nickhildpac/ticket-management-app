// Package util provides shared helper utilities for HTTP handlers and auth.
package util

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type ErrorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details"`
}

func WriteResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
}

func ErrorResponse(w http.ResponseWriter, status int, err error) {
	ErrorResponseWithCode(w, status, errorCodeForStatus(status), err, nil)
}

func ErrorResponseWithCode(w http.ResponseWriter, status int, code string, err error, details []ErrorDetail) {
	if err == nil {
		err = errors.New("request failed")
	}
	message := err.Error()
	if status >= http.StatusInternalServerError {
		log.Printf("internal server error: %v", err)
		code = "internal_server_error"
		message = "internal server error"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encodeErr := json.NewEncoder(w).Encode(ErrorBody{
		Code:    code,
		Message: message,
		Details: details,
	})
	if encodeErr != nil {
		http.Error(w, encodeErr.Error(), http.StatusInternalServerError)
		return
	}
}

func errorCodeForStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	default:
		if status >= 500 {
			return "internal_server_error"
		}
		return "request_failed"
	}
}
