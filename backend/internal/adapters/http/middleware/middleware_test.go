package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
	"github.com/nickhildpac/ticket-management-app/pkg/util"
)

func TestAuthRequiredUnauthorizedUsesErrorEnvelope(t *testing.T) {
	conf := &configs.Config{JWTSecret: "test-secret", JWTIssuer: "test-issuer"}
	handler := AuthRequired(conf)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/protected", nil))

	assertErrorEnvelope(t, rr, http.StatusUnauthorized, "unauthorized")
}

func TestAdminRequiredForbiddenUsesErrorEnvelope(t *testing.T) {
	conf := &configs.Config{JWTSecret: "test-secret", JWTIssuer: "test-issuer"}
	handler := AdminRequired(conf)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+signedAccessToken(t, conf, "user"))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assertErrorEnvelope(t, rr, http.StatusForbidden, "forbidden")
}

func assertErrorEnvelope(t *testing.T, rr *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()

	if rr.Code != wantStatus {
		t.Fatalf("expected status %d, got %d", wantStatus, rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected JSON content type, got %q", got)
	}

	var body util.ErrorBody
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != wantCode {
		t.Fatalf("expected code %q, got %q", wantCode, body.Code)
	}
	if body.Message == "" {
		t.Fatalf("expected non-empty error message")
	}
}

func signedAccessToken(t *testing.T, conf *configs.Config, role string) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &util.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "00000000-0000-0000-0000-000000000001",
			Issuer:  conf.JWTIssuer,
		},
		Role: role,
	})
	signed, err := token.SignedString([]byte(conf.JWTSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}
