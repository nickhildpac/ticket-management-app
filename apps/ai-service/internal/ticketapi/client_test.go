package ticketapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nickhildpac/ticket-management-ai-service/internal/keycloak"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testServiceToken stands in for the access token Keycloak issues to the
// ai-service service account.
const testServiceToken = "keycloak-service-account-token"

// capture records the request the client made so the test can assert on the
// method, path, body and presented token.
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
		TicketServiceURL: srv.URL,
		Tokens:           keycloak.NewStaticTokenSource(testServiceToken),
	}), got
}

func TestAddCommentPostsAsTheServiceAccount(t *testing.T) {
	client, got := newStubTicketService(t, http.StatusCreated)

	err := client.AddComment(context.Background(), "t-1", "[AI-suggested reply]\n\nUse the link.")

	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, got.method)
	assert.Equal(t, "/api/v1/comments", got.path)
	assert.Equal(t, "t-1", got.body["ticket_id"])
	assert.Equal(t, "[AI-suggested reply]\n\nUse the link.", got.body["description"])

	// The service-account token from Keycloak is presented as-is; this client
	// no longer signs anything itself.
	assert.Equal(t, "Bearer "+testServiceToken, got.auth)
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

// Without a token source there is nothing to authenticate with; the call must
// fail rather than go out unauthenticated.
func TestCallsFailWithoutATokenSource(t *testing.T) {
	client := New(Settings{TicketServiceURL: "http://ticket-service:8080"})

	err := client.AddComment(context.Background(), "t-1", "hi")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no token source")
}
