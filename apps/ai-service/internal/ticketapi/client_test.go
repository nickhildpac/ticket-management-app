package ticketapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSecret   = "test-secret"
	testIssuer   = "example.com"
	testAudience = "example.com"
	testAccount  = "00000000-0000-4000-8000-0000000000a1"
)

// capture records the request the client made so the test can assert on the
// method, path, body and minted token.
type capture struct {
	method string
	path   string
	query  string
	auth   string
	body   map[string]any
}

// newStubTicketService returns a client pointed at a stub API plus the capture
// it fills in.
func newStubTicketService(t *testing.T, status int) (*Client, *capture) {
	t.Helper()
	got := &capture{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.path = r.URL.Path
		got.query = r.URL.RawQuery
		got.auth = r.Header.Get("Authorization")
		if raw, err := io.ReadAll(r.Body); err == nil && len(raw) > 0 {
			_ = json.Unmarshal(raw, &got.body)
		}
		w.WriteHeader(status)
		if status >= 400 {
			_, _ = w.Write([]byte(`{"error":"forbidden"}`))
		}
	}))
	t.Cleanup(srv.Close)

	return New(Settings{
		TicketServiceURL:   srv.URL,
		JWTSecret:          testSecret,
		JWTIssuer:          testIssuer,
		JWTAudience:        testAudience,
		ServiceAccountID:   testAccount,
		ServiceAccountRole: "admin",
	}), got
}

// parseServiceToken verifies the bearer token the same way the ticket-service
// middleware does, and returns its claims.
func parseServiceToken(t *testing.T, header string) *claims {
	t.Helper()
	require.True(t, len(header) > len("Bearer "), "missing bearer token")
	out := &claims{}
	_, err := jwt.ParseWithClaims(header[len("Bearer "):], out, func(*jwt.Token) (any, error) {
		return []byte(testSecret), nil
	}, jwt.WithAudience(testAudience), jwt.WithIssuer(testIssuer))
	require.NoError(t, err)
	return out
}

func TestAddCommentPostsAsTheServiceAccount(t *testing.T) {
	client, got := newStubTicketService(t, http.StatusCreated)

	err := client.AddComment(context.Background(), "t-1", "[AI-suggested reply]\n\nUse the link.")

	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, got.method)
	assert.Equal(t, "/api/v1/comments", got.path)
	assert.Equal(t, "t-1", got.body["ticket_id"])
	assert.Equal(t, "[AI-suggested reply]\n\nUse the link.", got.body["description"])

	// The minted token must satisfy the ticket-service middleware: an existing
	// admin user id as `sub`, with matching issuer and audience.
	c := parseServiceToken(t, got.auth)
	assert.Equal(t, testAccount, c.Subject)
	assert.Equal(t, "admin", c.Role)
	assert.Equal(t, testIssuer, c.Issuer)
	assert.Contains(t, c.Audience, testAudience)
	assert.WithinDuration(t, c.IssuedAt.Time.Add(tokenTTL), c.ExpiresAt.Time, 0)
}

func TestVerifyAccessProbesTheTicketList(t *testing.T) {
	client, got := newStubTicketService(t, http.StatusOK)

	require.NoError(t, client.VerifyAccess(context.Background()))

	assert.Equal(t, http.MethodGet, got.method)
	assert.Equal(t, "/api/v1/tickets", got.path)
	assert.Equal(t, "limit=1&offset=0", got.query)
}

func TestVerifyAccessSurfacesAuthFailures(t *testing.T) {
	client, _ := newStubTicketService(t, http.StatusForbidden)

	err := client.VerifyAccess(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestAddCommentSurfacesNonSuccessStatus(t *testing.T) {
	client, _ := newStubTicketService(t, http.StatusUnauthorized)

	err := client.AddComment(context.Background(), "t-1", "hi")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestSetStatePatchesTheTicket(t *testing.T) {
	client, got := newStubTicketService(t, http.StatusOK)

	require.NoError(t, client.SetState(context.Background(), "t-1", "pending"))

	assert.Equal(t, http.MethodPatch, got.method)
	assert.Equal(t, "/api/v1/tickets/t-1", got.path)
	assert.Equal(t, "pending", got.body["state"])
}

func TestBaseURLTolerAtesATrailingSlash(t *testing.T) {
	client := New(Settings{TicketServiceURL: "http://ticket-service:8080/"})

	assert.Equal(t, "http://ticket-service:8080/api/v1", client.baseURL)
}
