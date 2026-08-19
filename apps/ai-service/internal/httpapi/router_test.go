package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nickhildpac/ticket-management-ai-service/internal/keycloak"
	"github.com/nickhildpac/ticket-management-ai-service/internal/triage"
)

// validToken is the one token stubVerifier accepts. Real token validation
// (signature, issuer, audience, expiry) is covered in internal/keycloak; these
// tests only care that the router refuses anything the verifier rejects.
const validToken = "valid-keycloak-token"

type stubVerifier struct{}

func (stubVerifier) Verify(_ context.Context, raw string) (*keycloak.Claims, error) {
	if raw != validToken {
		return nil, keycloak.ErrInvalidToken
	}
	return &keycloak.Claims{
		Subject:  "00000000-0000-4000-8000-0000000000a1",
		Username: "service-account-ai-service",
		Roles:    []string{"admin"},
	}, nil
}

type stubAgent struct {
	result  triage.TriageResult
	triaged []triage.TicketContext
}

func (s *stubAgent) Triage(_ context.Context, ticket triage.TicketContext) triage.TriageResult {
	s.triaged = append(s.triaged, ticket)
	return s.result
}

// newTestRouter wires the router without a database, so only the endpoints that
// don't touch the store are exercised here.
func newTestRouter(agent Triager) http.Handler {
	return NewRouter(Deps{
		Agent:       agent,
		APIV1Prefix: "/api/v1",
		Verifier:    stubVerifier{},
		CORSOrigins: []string{"http://localhost:5173"},
	})
}

func do(t *testing.T, h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body
}

func TestHealthEndpointsAreUnauthenticated(t *testing.T) {
	h := newTestRouter(&stubAgent{})

	for _, path := range []string{"/health", "/api/v1/health"} {
		rec := do(t, h, httptest.NewRequest(http.MethodGet, path, nil))

		assert.Equal(t, http.StatusOK, rec.Code, path)
		assert.Equal(t, map[string]any{"status": "ok"}, decode(t, rec))
	}
}

func TestTriageRequiresAValidToken(t *testing.T) {
	h := newTestRouter(&stubAgent{})

	tests := []struct {
		name   string
		header string
	}{
		{"no header", ""},
		{"wrong scheme", "Basic " + validToken},
		{"empty credential", "Bearer "},
		{"garbage token", "Bearer not-a-jwt"},
		{"token the verifier rejects", "Bearer some-other-realms-token"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/triage", bytes.NewReader([]byte(`{}`)))
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}

			rec := do(t, h, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.Equal(t, "unauthorized", decode(t, rec)["code"])
		})
	}
}

func TestTriageReturnsTheGatedDecision(t *testing.T) {
	agent := &stubAgent{result: triage.TriageResult{
		TicketID:    "t-1",
		Action:      triage.ActionAutoAnswer,
		Confidence:  0.9,
		DraftReply:  "Use the reset link [1].",
		SafetyFlags: []string{},
	}}
	h := newTestRouter(agent)

	body := `{"ticket_id":"t-1","title":"Reset","description":"help","state":"open"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/triage", bytes.NewReader([]byte(body)))
	req.Header.Set("Authorization", "Bearer "+validToken)

	rec := do(t, h, req)

	require.Equal(t, http.StatusOK, rec.Code)
	out := decode(t, rec)
	assert.Equal(t, "auto_answer", out["action"])
	assert.Equal(t, "Use the reset link [1].", out["draft_reply"])
	// An unset priority defaults to "low", matching the worker's behaviour.
	require.Len(t, agent.triaged, 1)
	assert.Equal(t, "low", agent.triaged[0].Priority)
}

func TestTriageValidationErrorUsesStructuredEnvelope(t *testing.T) {
	h := newTestRouter(&stubAgent{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/triage", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Authorization", "Bearer "+validToken)

	rec := do(t, h, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	body := decode(t, rec)
	assert.Equal(t, "validation_error", body["code"])
	assert.Equal(t, "request validation failed", body["message"])

	details, ok := body["details"].([]any)
	require.True(t, ok, "details should list the missing fields")
	first, ok := details[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "ticket_id", first["field"])
	assert.NotEmpty(t, first["message"])
}

func TestTriageRejectsMalformedJSON(t *testing.T) {
	h := newTestRouter(&stubAgent{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/triage", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Authorization", "Bearer "+validToken)

	rec := do(t, h, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "validation_error", decode(t, rec)["code"])
}

// A panic in a handler must not leak internals to the caller.
func TestPanicIsRecoveredWithoutLeakingInternals(t *testing.T) {
	h := NewRouter(Deps{
		Agent:       panickingAgent{},
		APIV1Prefix: "/api/v1",
		Verifier:    stubVerifier{},
	})

	body := `{"ticket_id":"t-1","title":"t","description":"d","state":"open"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/triage", bytes.NewReader([]byte(body)))
	req.Header.Set("Authorization", "Bearer "+validToken)

	rec := do(t, h, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "database password leaked")
	assert.Equal(t, map[string]any{
		"code":    "internal_server_error",
		"message": "internal server error",
		"details": nil,
	}, decode(t, rec))
}

type panickingAgent struct{}

func (panickingAgent) Triage(context.Context, triage.TicketContext) triage.TriageResult {
	panic("database password leaked")
}

func TestIngestRequiresMultipartBody(t *testing.T) {
	h := newTestRouter(&stubAgent{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ingest", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Authorization", "Bearer "+validToken)
	req.Header.Set("Content-Type", "application/json")

	rec := do(t, h, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "validation_error", decode(t, rec)["code"])
}

func TestIngestRejectsAnUploadWithoutFilesField(t *testing.T) {
	h := newTestRouter(&stubAgent{})

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	require.NoError(t, mw.WriteField("notes", "no files here"))
	require.NoError(t, mw.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ingest", &buf)
	req.Header.Set("Authorization", "Bearer "+validToken)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	rec := do(t, h, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "validation_error", decode(t, rec)["code"])
}

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	h := newTestRouter(&stubAgent{})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/triage", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Headers", "authorization")

	rec := do(t, h, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "http://localhost:5173", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "authorization", rec.Header().Get("Access-Control-Allow-Headers"))
}

func TestCORSIgnoresUnknownOrigin(t *testing.T) {
	h := newTestRouter(&stubAgent{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "http://evil.test")

	rec := do(t, h, req)

	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}
