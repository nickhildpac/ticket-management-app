package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/nickhildpac/ticket-management-ai-service/internal/triage"
)

// triageHandler runs on-demand triage for a single ticket.
//
// The primary path is the async worker consuming ticket events; this endpoint
// exists for manual re-runs and testing. It returns the decision without
// applying it back to the ticket service.
func triageHandler(agent Triager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ticket triage.TicketContext
		if err := json.NewDecoder(r.Body).Decode(&ticket); err != nil {
			writeError(w, http.StatusBadRequest, "validation_error",
				"request validation failed", err.Error())
			return
		}
		if missing := missingFields(ticket); len(missing) > 0 {
			writeError(w, http.StatusBadRequest, "validation_error",
				"request validation failed", missing)
			return
		}
		if ticket.Priority == "" {
			ticket.Priority = "low"
		}
		writeJSON(w, http.StatusOK, agent.Triage(r.Context(), ticket))
	}
}

// missingFields reports the required TicketContext fields the payload omitted,
// in the same {field, message} shape FastAPI/pydantic produced.
func missingFields(t triage.TicketContext) []map[string]string {
	required := []struct {
		name  string
		value string
	}{
		{"ticket_id", t.TicketID},
		{"title", t.Title},
		{"description", t.Description},
		{"state", t.State},
	}
	var out []map[string]string
	for _, f := range required {
		if strings.TrimSpace(f.value) == "" {
			out = append(out, map[string]string{
				"field":   f.name,
				"message": "field required",
				"type":    "missing",
			})
		}
	}
	return out
}
