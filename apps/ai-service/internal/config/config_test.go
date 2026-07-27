package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultSettingsWorkWithoutEnv(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("JWT_SECRET", "")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "development", cfg.AppEnv)
	assert.Equal(t, "local-dev-only-secret", cfg.JWTSecret)
	assert.Equal(t, 0.75, cfg.AutoAnswerConfidenceThreshold)
	assert.Equal(t, 6, cfg.TriageMaxIterations)
	assert.Equal(t, 384, cfg.EmbeddingDim)
	assert.Equal(t, 60, cfg.RRFK)
}

func TestLocalEnvGetsDevelopmentSecretFallback(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	t.Setenv("JWT_SECRET", "")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "local-dev-only-secret", cfg.JWTSecret)
}

func TestProductionRejectsMissingJWTSecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

func TestProductionRejectsKnownWeakJWTSecret(t *testing.T) {
	for _, weak := range []string{"secret", "changeme", "change-me", "local-dev-only-secret"} {
		t.Run(weak, func(t *testing.T) {
			t.Setenv("APP_ENV", "production")
			t.Setenv("JWT_SECRET", weak)

			_, err := Load()
			require.Error(t, err)
		})
	}
}

func TestProductionAcceptsStrongJWTSecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "a-real-production-secret")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "a-real-production-secret", cfg.JWTSecret)
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
