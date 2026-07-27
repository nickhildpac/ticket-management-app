// Package config loads the ai-service settings from the environment.
//
// Every knob keeps the env var name the Python service used, so an existing
// .env keeps working unchanged. Defaults match app/core/config.py.
package config

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config is the ai-service's runtime configuration.
type Config struct {
	AppName     string
	AppEnv      string
	APIV1Prefix string
	Port        string
	DatabaseURL string
	CORSOrigins []string

	// ---- auth: shared secret with the Go ticket-service ----------------
	JWTSecret   string
	JWTIssuer   string
	JWTAudience string

	// ---- AI / RAG / triage --------------------------------------------
	RedisURL         string
	TicketServiceURL string
	AnthropicAPIKey  string
	TriageModel      string
	// AutoAnswerConfidenceThreshold is the deterministic safety gate's bar:
	// auto-answer only when the model is at least this confident.
	AutoAnswerConfidenceThreshold float64
	// TriageMaxIterations caps model turns in the agentic loop before we fail
	// safe and escalate.
	TriageMaxIterations int

	EmbeddingDim int
	// OpenAIAPIKey selects the semantic embedder; empty falls back to the
	// offline hashing embedder (see rag.BuildEmbedder).
	OpenAIAPIKey   string
	EmbeddingModel string

	// RAGTopK is the number of passages returned after fusion/re-ranking.
	RAGTopK int
	// RAGCandidateK is fetched from each retrieval lane (semantic/keyword).
	RAGCandidateK int
	// RAGRerankPool is how many fused candidates the cross-encoder scores.
	RAGRerankPool int
	// RRFK is the Reciprocal Rank Fusion constant (standard value is 60).
	RRFK int

	OpenRouterAPIKey  string
	CrossEncoderModel string

	// Redis Streams consumer group + this replica's consumer name (set per
	// replica when scaling horizontally so pending-entry recovery can tell
	// them apart).
	ConsumerGroup string
	ConsumerName  string

	// Service account the worker uses to call back into the Go ticket API.
	// The minted JWT must satisfy the ticket-service auth middleware: an
	// existing agent/admin user id as `sub`, matching issuer/audience, signed
	// with the shared JWTSecret. See docs/adr/0002-service-topology.md.
	// Defaults match the seeded AI service account (migration 000013). It is
	// an admin because CanCommentOnTicket only lets admins comment on tickets
	// they aren't assigned to, and the worker comments on arbitrary tickets.
	ServiceAccountID   string
	ServiceAccountRole string
}

// weakSecrets are rejected outside local/development/test so a placeholder
// never ships to a real environment.
var weakSecrets = map[string]struct{}{
	"":                      {},
	"secret":                {},
	"changeme":              {},
	"change-me":             {},
	"local-dev-only-secret": {},
}

// Load reads the configuration from the environment and validates it.
//
// A .env file in the working directory is loaded first, matching the Python
// service's pydantic-settings behaviour so `go run ./cmd/worker` picks up local
// API keys. It never overrides an already-set variable, so compose's explicit
// `environment:` entries still win. A missing file is not an error.
func Load() (*Config, error) {
	_ = godotenv.Load()

	c := &Config{
		AppName:     env("APP_NAME", "Ticket Management AI Service"),
		AppEnv:      env("APP_ENV", "development"),
		APIV1Prefix: env("API_V1_PREFIX", "/api/v1"),
		Port:        env("PORT", "8081"),
		DatabaseURL: normalizeDSN(env("DATABASE_URL",
			"postgres://postgres:postgres@localhost:5432/ticket_management")),
		CORSOrigins: splitCSV(env("CORS_ORIGINS", "http://localhost:5173")),

		JWTSecret:   os.Getenv("JWT_SECRET"),
		JWTIssuer:   env("JWT_ISSUER", "example.com"),
		JWTAudience: env("JWT_AUDIENCE", "example.com"),

		RedisURL:         env("REDIS_URL", "redis://localhost:6379/0"),
		TicketServiceURL: env("TICKET_SERVICE_URL", "http://localhost:8080"),
		AnthropicAPIKey:  os.Getenv("ANTHROPIC_API_KEY"),
		TriageModel:      env("TRIAGE_MODEL", "claude-opus-4-8"),

		OpenAIAPIKey:   os.Getenv("OPENAI_API_KEY"),
		EmbeddingModel: env("EMBEDDING_MODEL", "text-embedding-3-small"),

		OpenRouterAPIKey:  os.Getenv("OPENROUTER_API_KEY"),
		CrossEncoderModel: env("CROSS_ENCODER_MODEL", "cohere/rerank-v3.5"),

		ConsumerGroup: env("CONSUMER_GROUP", "ai-triage"),
		ConsumerName:  env("CONSUMER_NAME", "worker-1"),

		ServiceAccountID:   env("AI_SERVICE_ACCOUNT_ID", "00000000-0000-4000-8000-0000000000a1"),
		ServiceAccountRole: env("AI_SERVICE_ACCOUNT_ROLE", "admin"),
	}

	var err error
	if c.AutoAnswerConfidenceThreshold, err = envFloat("AUTO_ANSWER_CONFIDENCE_THRESHOLD", 0.75); err != nil {
		return nil, err
	}
	if c.TriageMaxIterations, err = envInt("TRIAGE_MAX_ITERATIONS", 6); err != nil {
		return nil, err
	}
	if c.EmbeddingDim, err = envInt("EMBEDDING_DIM", 384); err != nil {
		return nil, err
	}
	if c.RAGTopK, err = envInt("RAG_TOP_K", 5); err != nil {
		return nil, err
	}
	if c.RAGCandidateK, err = envInt("RAG_CANDIDATE_K", 20); err != nil {
		return nil, err
	}
	if c.RAGRerankPool, err = envInt("RAG_RERANK_POOL", 10); err != nil {
		return nil, err
	}
	if c.RRFK, err = envInt("RRF_K", 60); err != nil {
		return nil, err
	}

	if err := c.validateJWTSecret(); err != nil {
		return nil, err
	}
	return c, nil
}

// validateJWTSecret mirrors app/core/config.py: local-ish environments get a
// dev default, everything else must supply a real secret.
func (c *Config) validateJWTSecret() error {
	switch strings.ToLower(strings.TrimSpace(c.AppEnv)) {
	case "", "local", "development", "test":
		if c.JWTSecret == "" {
			c.JWTSecret = "local-dev-only-secret"
		}
		return nil
	}
	if _, weak := weakSecrets[strings.TrimSpace(c.JWTSecret)]; weak {
		return fmt.Errorf(
			"JWT_SECRET is required and must not use a known weak value outside local/development/test")
	}
	return nil
}

// sqlAlchemyDriver matches the "+driver" suffix SQLAlchemy DSNs carry
// (postgresql+psycopg://...), which Go's postgres drivers reject.
var sqlAlchemyDriver = regexp.MustCompile(`^(postgres(?:ql)?)\+[a-z0-9]+://`)

// normalizeDSN turns a SQLAlchemy-style DATABASE_URL into one lib/pq accepts,
// so an .env carried over from the Python service keeps working. It also
// defaults sslmode=disable: lib/pq defaults URL DSNs to sslmode=require, and
// the local/compose Postgres does not serve TLS.
func normalizeDSN(dsn string) string {
	dsn = sqlAlchemyDriver.ReplaceAllString(dsn, "$1://")
	u, err := url.Parse(dsn)
	if err != nil || u.Scheme == "" {
		return dsn
	}
	q := u.Query()
	if q.Get("sslmode") == "" {
		q.Set("sslmode", "disable")
		u.RawQuery = q.Encode()
	}
	return u.String()
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) (int, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return v, nil
}

func envFloat(key string, fallback float64) (float64, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback, nil
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return v, nil
}

func splitCSV(raw string) []string {
	var out []string
	for item := range strings.SplitSeq(raw, ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}
