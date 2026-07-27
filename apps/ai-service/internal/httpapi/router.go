package httpapi

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/nickhildpac/ticket-management-ai-service/internal/rag"
	"github.com/nickhildpac/ticket-management-ai-service/internal/triage"
)

// Triager produces a gated triage decision for a ticket.
type Triager interface {
	Triage(ctx context.Context, ticket triage.TicketContext) triage.TriageResult
}

// Deps are the collaborators the HTTP handlers need. The agent and store are
// process singletons: building them opens a database connection pool and an
// Anthropic client, so they must not be rebuilt per request.
type Deps struct {
	Agent       Triager
	Store       *rag.VectorStore
	APIV1Prefix string
	JWTSecret   string
	JWTIssuer   string
	JWTAudience string
	CORSOrigins []string
}

// NewRouter wires the ai-service HTTP surface.
func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, recoverer)
	r.Use(corsMiddleware(d.CORSOrigins))

	r.Get("/health", healthHandler)

	prefix := d.APIV1Prefix
	if prefix == "" {
		prefix = "/api/v1"
	}
	r.Route(prefix, func(api chi.Router) {
		api.Get("/health", healthHandler)

		api.Group(func(secured chi.Router) {
			secured.Use(requireAuth(d.JWTSecret, d.JWTIssuer, d.JWTAudience))
			secured.Post("/triage", triageHandler(d.Agent))
			secured.Post("/ingest", ingestHandler(d.Store))
		})
	})
	return r
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
