// Package ticketapi calls back into the authoritative Go ticket API to apply
// triage results.
//
// Auth: it mints a short-lived HS256 JWT for the AI service account. The claim
// shape (RegisteredClaims + `role`) mirrors the ticket-service auth middleware;
// the `sub` must be an existing agent/admin user id.
package ticketapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// tokenTTL keeps the minted service token short-lived; it is re-minted per
// request.
const tokenTTL = 5 * time.Minute

// Settings is the slice of configuration the client needs.
type Settings struct {
	TicketServiceURL   string
	JWTSecret          string
	JWTIssuer          string
	JWTAudience        string
	ServiceAccountID   string
	ServiceAccountRole string
}

// Client is an HTTP client for the ticket-service API.
type Client struct {
	baseURL  string
	settings Settings
	http     *http.Client
}

// Option customises a Client (tests point it at a stub server).
type Option func(*Client)

// WithHTTPClient overrides the underlying HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(t *Client) { t.http = c }
}

// New builds a ticket-service client from settings.
func New(s Settings, opts ...Option) *Client {
	c := &Client{
		baseURL:  strings.TrimRight(s.TicketServiceURL, "/") + "/api/v1",
		settings: s,
		http:     &http.Client{Timeout: 10 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// claims mirrors the ticket-service's util.Claims.
type claims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

func (c *Client) token() (string, error) {
	now := time.Now()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   c.settings.ServiceAccountID,
			Issuer:    c.settings.JWTIssuer,
			Audience:  jwt.ClaimStrings{c.settings.JWTAudience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
		Role: c.settings.ServiceAccountRole,
	})
	return tok.SignedString([]byte(c.settings.JWTSecret))
}

// AddComment posts a comment on a ticket as the AI service account.
func (c *Client) AddComment(ctx context.Context, ticketID, description string) error {
	body, err := json.Marshal(map[string]string{
		"ticket_id":   ticketID,
		"description": description,
	})
	if err != nil {
		return err
	}
	return c.do(ctx, http.MethodPost, c.baseURL+"/comments", bytes.NewReader(body), true)
}

// SetState transitions a ticket. The worker deliberately does not call this —
// it only comments, so it can't violate the ticket state machine — but the
// endpoint is kept available for manual tooling.
func (c *Client) SetState(ctx context.Context, ticketID, state string) error {
	body, err := json.Marshal(map[string]string{"state": state})
	if err != nil {
		return err
	}
	return c.do(ctx, http.MethodPatch,
		fmt.Sprintf("%s/tickets/%s", c.baseURL, url.PathEscape(ticketID)),
		bytes.NewReader(body), true)
}

// VerifyAccess probes the ticket API with the service token; it returns an
// error on failure.
//
// This surfaces misconfiguration (unset/invalid AI_SERVICE_ACCOUNT_ID, wrong
// role, iss/aud mismatch, or an unreachable API) loudly at worker startup
// instead of silently 403-ing on every consumed event.
func (c *Client) VerifyAccess(ctx context.Context) error {
	return c.do(ctx, http.MethodGet, c.baseURL+"/tickets?limit=1&offset=0", nil, false)
}

func (c *Client) do(ctx context.Context, method, target string, body io.Reader, jsonBody bool) error {
	token, err := c.token()
	if err != nil {
		return fmt.Errorf("mint service token: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if jsonBody {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, target, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Bound the echoed body so a large error page can't blow up the log line.
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("%s %s: unexpected status %d: %s",
			method, target, resp.StatusCode, strings.TrimSpace(string(detail)))
	}
	// Drain so the connection can be reused.
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}
