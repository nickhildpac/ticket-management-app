package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultSettingsWorkWithoutEnv(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("KEYCLOAK_CLIENT_SECRET", "")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "development", cfg.AppEnv)
	assert.Equal(t, "http://localhost:8090/realms/ticket-management", cfg.KeycloakIssuerURL)
	assert.Equal(t, "ticket-service", cfg.KeycloakAudience)
	assert.Equal(t, "ai-service", cfg.KeycloakClientID)
	assert.Equal(t, "ai-service-dev-secret", cfg.KeycloakClientSecret)
	assert.Equal(t, 0.75, cfg.AutoAnswerConfidenceThreshold)
	assert.Equal(t, 6, cfg.TriageMaxIterations)
	assert.Equal(t, 384, cfg.EmbeddingDim)
	assert.Equal(t, 60, cfg.RRFK)
}

func TestLocalEnvGetsDevelopmentSecretFallback(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	t.Setenv("KEYCLOAK_CLIENT_SECRET", "")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "ai-service-dev-secret", cfg.KeycloakClientSecret)
}

func TestProductionRejectsMissingClientSecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("KEYCLOAK_ISSUER_URL", "https://sso.example.com/realms/ticket-management")
	t.Setenv("KEYCLOAK_CLIENT_SECRET", "")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "KEYCLOAK_CLIENT_SECRET")
}

// The dev secret is checked in to the realm export, so it must never be usable
// outside development.
func TestProductionRejectsKnownWeakClientSecret(t *testing.T) {
	for _, weak := range []string{"secret", "changeme", "change-me", "local-dev-only-secret", "ai-service-dev-secret"} {
		t.Run(weak, func(t *testing.T) {
			t.Setenv("APP_ENV", "production")
			t.Setenv("KEYCLOAK_ISSUER_URL", "https://sso.example.com/realms/ticket-management")
			t.Setenv("KEYCLOAK_CLIENT_SECRET", weak)

			_, err := Load()
			require.Error(t, err)
		})
	}
}

func TestProductionAcceptsStrongClientSecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("KEYCLOAK_ISSUER_URL", "https://sso.example.com/realms/ticket-management")
	t.Setenv("KEYCLOAK_CLIENT_SECRET", "a-real-production-secret")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "a-real-production-secret", cfg.KeycloakClientSecret)
}

// Tokens would otherwise cross the network in the clear.
func TestProductionRejectsPlaintextIssuer(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("KEYCLOAK_ISSUER_URL", "http://sso.example.com/realms/ticket-management")
	t.Setenv("KEYCLOAK_CLIENT_SECRET", "a-real-production-secret")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "https")
}

// Inside Docker the browser-facing issuer is not reachable, so discovery uses
// the in-network URL while `iss` stays the public one.
func TestDiscoveryIssuerURLPrefersInternal(t *testing.T) {
	t.Setenv("KEYCLOAK_ISSUER_URL", "http://localhost:8090/realms/ticket-management")
	t.Setenv("KEYCLOAK_INTERNAL_ISSUER_URL", "http://keycloak:8080/realms/ticket-management")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "http://keycloak:8080/realms/ticket-management", cfg.DiscoveryIssuerURL())

	t.Setenv("KEYCLOAK_INTERNAL_ISSUER_URL", "")
	cfg, err = Load()
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8090/realms/ticket-management", cfg.DiscoveryIssuerURL())
}

func TestInvalidNumericEnvIsRejected(t *testing.T) {
	t.Setenv("RAG_TOP_K", "not-a-number")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "RAG_TOP_K")
}

// normalizeDSN keeps a SQLAlchemy-style DATABASE_URL carried over from the
// Python service working, and defaults sslmode so lib/pq doesn't demand TLS
// from the local/compose Postgres.
func TestNormalizeDSN(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "strips the SQLAlchemy driver suffix",
			in:   "postgresql+psycopg://postgres:postgres@localhost:5432/ticket_management",
			want: "postgresql://postgres:postgres@localhost:5432/ticket_management?sslmode=disable",
		},
		{
			name: "adds sslmode=disable when absent",
			in:   "postgres://u:p@host:5432/db",
			want: "postgres://u:p@host:5432/db?sslmode=disable",
		},
		{
			name: "preserves an explicit sslmode",
			in:   "postgres://u:p@host:5432/db?sslmode=require",
			want: "postgres://u:p@host:5432/db?sslmode=require",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, normalizeDSN(tc.in))
		})
	}
}

func TestCORSOriginsSplitAndTrim(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "http://a.test , http://b.test ,, ")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"http://a.test", "http://b.test"}, cfg.CORSOrigins)
}
