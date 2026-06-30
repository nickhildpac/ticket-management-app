package util

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorResponseUsesStructuredEnvelope(t *testing.T) {
	rr := httptest.NewRecorder()

	ErrorResponse(rr, http.StatusNotFound, errors.New("ticket not found"))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}

	var body ErrorBody
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Code != "not_found" {
		t.Fatalf("expected code not_found, got %q", body.Code)
	}
	if body.Message != "ticket not found" {
		t.Fatalf("expected message ticket not found, got %q", body.Message)
	}
	if len(body.Details) != 0 {
		t.Fatalf("expected no details, got %+v", body.Details)
	}
}

func TestErrorResponseHidesInternalErrorMessage(t *testing.T) {
	rr := httptest.NewRecorder()

	ErrorResponse(rr, http.StatusInternalServerError, errors.New("pq: password authentication failed for user postgres"))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	var body ErrorBody
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Code != "internal_server_error" {
		t.Fatalf("expected code internal_server_error, got %q", body.Code)
	}
	if body.Message != "internal server error" {
		t.Fatalf("expected generic internal message, got %q", body.Message)
	}
}
