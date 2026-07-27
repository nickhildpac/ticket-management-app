// Package httpapi exposes the ai-service's HTTP surface: on-demand triage and
// knowledge-base ingestion.
//
// The ticket lifecycle lives in the Go ticket-service (ADR 0002); this service
// exposes only AI/RAG endpoints.
package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// errorBody is the error envelope the Python service returned, kept byte-for-byte
// so existing callers (including the ticket-service ingest proxy, which streams
// this response through unchanged) keep working.
type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("failed to write response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string, details any) {
	writeJSON(w, status, errorBody{Code: code, Message: message, Details: details})
}
