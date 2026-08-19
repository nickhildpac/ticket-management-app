package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/adapters/auth"
	"github.com/nickhildpac/ticket-management-app/internal/application/service"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
	"github.com/nickhildpac/ticket-management-app/pkg/util"
)

// stubVerifier returns canned claims for a known token string. Real token
// verification (signatures, issuer, audience, expiry) is covered in
// internal/adapters/auth; these tests are about what the middleware does with
// the result.
type stubVerifier struct {
	claims map[string]*auth.Claims
}

func (s *stubVerifier) Verify(_ context.Context, raw string) (*auth.Claims, error) {
	if c, ok := s.claims[raw]; ok {
		return c, nil
	}
	return nil, auth.ErrInvalidToken
}

type stubIdentity struct {
	user     *domain.User
	err      error
	calls    int
	lastSeen *auth.Claims
}

func (s *stubIdentity) Resolve(_ context.Context, claims *auth.Claims) (*domain.User, error) {
	s.calls++
	s.lastSeen = claims
	if s.err != nil {
		return nil, s.err
	}
	if s.user != nil {
		return s.user, nil
	}
	return &domain.User{ID: uuid.New(), Role: claims.Role}, nil
}

const (
	adminToken = "token-admin"
	agentToken = "token-agent"
	userToken  = "token-user"
)

func newTestAuthenticator(identity IdentityResolver) *Authenticator {
	verifier := &stubVerifier{claims: map[string]*auth.Claims{
		adminToken: {Subject: uuid.NewString(), Role: domain.RoleAdmin, Email: "alice@admin.com"},
		agentToken: {Subject: uuid.NewString(), Role: domain.RoleAgent, Email: "bob@agent.com"},
		userToken:  {Subject: uuid.NewString(), Role: domain.RoleUser, Email: "charlie@user.com"},
	}}
	if identity == nil {
		identity = &stubIdentity{}
	}
	return NewAuthenticator(verifier, identity)
}

func requestWithToken(token string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func TestAuthRequiredUnauthorizedUsesErrorEnvelope(t *testing.T) {
	handler := newTestAuthenticator(nil).AuthRequired()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, requestWithToken(""))

	assertErrorEnvelope(t, rr, http.StatusUnauthorized, "unauthorized")
}

func TestAuthRequiredRejectsInvalidToken(t *testing.T) {
	handler := newTestAuthenticator(nil).AuthRequired()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, requestWithToken("not-a-known-token"))

	assertErrorEnvelope(t, rr, http.StatusUnauthorized, "unauthorized")
}

func TestAuthRequiredRejectsMalformedHeaders(t *testing.T) {
	headers := []string{
		"Bearer",              // no credential
		"Bearer ",             // empty credential
		adminToken,            // no scheme
		"Basic " + adminToken, // wrong scheme
	}

	for _, header := range headers {
		t.Run(header, func(t *testing.T) {
			handler := newTestAuthenticator(nil).AuthRequired()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Fatal("next handler should not be called")
			}))

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", header)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401 for header %q, got %d", header, rr.Code)
			}
		})
	}
}

// The scheme is case-insensitive per RFC 7235; rejecting "bearer" would break
// otherwise-valid clients.
func TestAuthRequiredAcceptsLowercaseScheme(t *testing.T) {
	var called bool
	handler := newTestAuthenticator(nil).AuthRequired()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "bearer "+adminToken)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatalf("expected handler to run, got status %d", rr.Code)
	}
}

func TestAdminRequiredForbiddenUsesErrorEnvelope(t *testing.T) {
	handler := newTestAuthenticator(nil).AdminRequired()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, requestWithToken(userToken))

	assertErrorEnvelope(t, rr, http.StatusForbidden, "forbidden")
}

func TestAdminRequiredAdmitsAdmin(t *testing.T) {
	var called bool
	handler := newTestAuthenticator(nil).AdminRequired()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, requestWithToken(adminToken))

	if !called {
		t.Fatalf("expected admin to be admitted, got status %d", rr.Code)
	}
}

func TestRequireAnyRole(t *testing.T) {
	tests := []struct {
		name     string
		required []domain.UserRole
		token    string
		want     int
	}{
		{"no roles means authenticated only", nil, userToken, http.StatusOK},
		{"agent admitted to staff route", []domain.UserRole{domain.RoleAgent, domain.RoleAdmin}, agentToken, http.StatusOK},
		{"admin admitted to staff route", []domain.UserRole{domain.RoleAgent, domain.RoleAdmin}, adminToken, http.StatusOK},
		{"user refused staff route", []domain.UserRole{domain.RoleAgent, domain.RoleAdmin}, userToken, http.StatusForbidden},
		{"agent refused admin route", []domain.UserRole{domain.RoleAdmin}, agentToken, http.StatusForbidden},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := newTestAuthenticator(nil).RequireAnyRole(tc.required...)(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, requestWithToken(tc.token))

			if rr.Code != tc.want {
				t.Fatalf("expected status %d, got %d", tc.want, rr.Code)
			}
		})
	}
}

// A rejected request must not cause a user row to be provisioned — otherwise
// anyone holding a token for a different audience could seed the users table.
func TestForbiddenRequestDoesNotProvisionIdentity(t *testing.T) {
	identity := &stubIdentity{}
	handler := newTestAuthenticator(identity).AdminRequired()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, requestWithToken(userToken))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
	if identity.calls != 0 {
		t.Fatalf("expected identity resolution to be skipped, got %d calls", identity.calls)
	}
}

// The rest of the application reads the *local* user id from the context, not
// the Keycloak subject; getting this wrong would break every ownership check.
func TestAuthRequiredPutsLocalUserIDInContext(t *testing.T) {
	localID := uuid.New()
	identity := &stubIdentity{user: &domain.User{ID: localID, Role: domain.RoleAgent}}

	var gotID, gotRole string
	handler := newTestAuthenticator(identity).AuthRequired()(
		http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			gotID, _ = r.Context().Value(configs.UserIDKey).(string)
			gotRole, _ = r.Context().Value(configs.UserRoleKey).(string)
		}))

	handler.ServeHTTP(httptest.NewRecorder(), requestWithToken(adminToken))

	if gotID != localID.String() {
		t.Errorf("context user id = %q, want the local id %q", gotID, localID)
	}
	// The role in context is the resolved user's, which the identity service
	// keeps in step with the token.
	if gotRole != string(domain.RoleAgent) {
		t.Errorf("context role = %q, want %q", gotRole, domain.RoleAgent)
	}
	if identity.lastSeen == nil || identity.lastSeen.Role != domain.RoleAdmin {
		t.Error("expected the verified claims to be handed to the identity resolver")
	}
}

func TestClaimsAvailableInContext(t *testing.T) {
	var claims *auth.Claims
	var ok bool
	handler := newTestAuthenticator(nil).AuthRequired()(
		http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			claims, ok = ClaimsFromContext(r.Context())
		}))

	handler.ServeHTTP(httptest.NewRecorder(), requestWithToken(adminToken))

	if !ok || claims == nil {
		t.Fatal("expected token claims in the request context")
	}
	if claims.Email != "alice@admin.com" {
		t.Errorf("claims.Email = %q", claims.Email)
	}
}

func TestIdentityConflictReturns409(t *testing.T) {
	identity := &stubIdentity{err: service.ErrIdentityConflict}
	handler := newTestAuthenticator(identity).AuthRequired()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, requestWithToken(adminToken))

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestIdentityFailureReturns500(t *testing.T) {
	identity := &stubIdentity{err: errors.New("database is down")}
	handler := newTestAuthenticator(identity).AuthRequired()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, requestWithToken(adminToken))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
	// The underlying error must not leak to the client.
	if body := rr.Body.String(); strings.Contains(body, "database is down") {
		t.Errorf("internal error detail leaked to client: %s", body)
	}
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
