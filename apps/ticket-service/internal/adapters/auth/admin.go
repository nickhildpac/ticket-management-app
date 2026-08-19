package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

// AdminClient writes realm role assignments back to Keycloak.
//
// Keycloak owns roles, so an admin changing someone's role has to change it
// *there* — updating only the local row would be silently reverted the next
// time that user presented a token. This client is what keeps the existing
// admin endpoint honest.
//
// It authenticates with the ticket-service confidential client's service
// account, which needs the `realm-management` client roles `view-users` and
// `manage-users`.
type AdminClient struct {
	baseURL  string
	realm    string
	clientID string
	secret   string
	http     *http.Client

	token *cachedToken
}

// AdminConfig configures the Admin API client.
type AdminConfig struct {
	// IssuerURL is the realm URL; the admin base URL and realm name are derived
	// from it. This must be a URL *this process* can reach — inside Docker that
	// is the in-network URL, not the browser-facing issuer the verifier checks
	// tokens against.
	IssuerURL string
	ClientID  string
	Secret    string
}

// ErrAdminNotConfigured is returned when role writes are attempted without
// admin credentials. Callers surface it rather than falling back to a local-only
// write that Keycloak would later overwrite.
var ErrAdminNotConfigured = errors.New("keycloak admin credentials are not configured")

// NewAdminClient builds an Admin API client. It returns (nil, nil) when no
// credentials are configured — role management is then unavailable, which the
// caller reports as such.
func NewAdminClient(cfg AdminConfig) (*AdminClient, error) {
	if strings.TrimSpace(cfg.ClientID) == "" || strings.TrimSpace(cfg.Secret) == "" {
		return nil, nil
	}
	base, realm, err := splitRealmURL(cfg.IssuerURL)
	if err != nil {
		return nil, err
	}
	return &AdminClient{
		baseURL:  base,
		realm:    realm,
		clientID: cfg.ClientID,
		secret:   cfg.Secret,
		http:     &http.Client{Timeout: 10 * time.Second},
		token:    &cachedToken{},
	}, nil
}

// splitRealmURL turns https://host/realms/<realm> into (https://host, <realm>).
func splitRealmURL(issuer string) (baseURL, realm string, err error) {
	trimmed := strings.TrimRight(strings.TrimSpace(issuer), "/")
	u, err := url.Parse(trimmed)
	if err != nil {
		return "", "", fmt.Errorf("parse issuer URL: %w", err)
	}
	base, realmName, found := strings.Cut(u.Path, "/realms/")
	if !found || realmName == "" {
		return "", "", fmt.Errorf("issuer URL %q is not a /realms/<realm> URL", issuer)
	}
	return strings.TrimRight(u.Scheme+"://"+u.Host+base, "/"), realmName, nil
}

// SetRealmRole makes `role` the user's only application realm role, removing
// the other two. Roles are mutually exclusive in this domain (domain.UserRole
// is a single value), so assigning without removing would leave a user holding
// both `agent` and `admin`, and RoleFromRealmRoles would keep resolving them to
// the more privileged one.
func (c *AdminClient) SetRealmRole(ctx context.Context, keycloakID uuid.UUID, role domain.UserRole) error {
	if c == nil {
		return ErrAdminNotConfigured
	}

	available, err := c.realmRoles(ctx)
	if err != nil {
		return err
	}

	var toAdd []realmRole
	var toRemove []realmRole
	for _, name := range []domain.UserRole{domain.RoleAdmin, domain.RoleAgent, domain.RoleUser} {
		r, ok := available[string(name)]
		if !ok {
			return fmt.Errorf("realm role %q does not exist in realm %q", name, c.realm)
		}
		if name == role {
			toAdd = append(toAdd, r)
		} else {
			toRemove = append(toRemove, r)
		}
	}

	path := fmt.Sprintf("/admin/realms/%s/users/%s/role-mappings/realm", c.realm, keycloakID)

	// Add before removing: if the second call fails, the user is left with an
	// extra role rather than none at all.
	if err := c.do(ctx, http.MethodPost, path, toAdd, nil); err != nil {
		return fmt.Errorf("assign realm role %q: %w", role, err)
	}
	if err := c.do(ctx, http.MethodDelete, path, toRemove, nil); err != nil {
		return fmt.Errorf("remove superseded realm roles: %w", err)
	}
	return nil
}

// realmRole is Keycloak's RoleRepresentation, trimmed to what role-mapping
// calls require.
type realmRole struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *AdminClient) realmRoles(ctx context.Context) (map[string]realmRole, error) {
	var roles []realmRole
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/admin/realms/%s/roles", c.realm), nil, &roles); err != nil {
		return nil, fmt.Errorf("list realm roles: %w", err)
	}
	out := make(map[string]realmRole, len(roles))
	for _, r := range roles {
		out[r.Name] = r
	}
	return out, nil
}

func (c *AdminClient) do(ctx context.Context, method, path string, body, out any) error {
	token, err := c.accessToken(ctx)
	if err != nil {
		return err
	}

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("keycloak admin API %s %s: %s: %s", method, path, resp.Status, strings.TrimSpace(string(detail)))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

// cachedToken holds the service account token between calls.
type cachedToken struct {
	value     string
	expiresAt time.Time
}

func (c *AdminClient) accessToken(ctx context.Context) (string, error) {
	if c.token.value != "" && time.Now().Before(c.token.expiresAt) {
		return c.token.value, nil
	}

	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {c.clientID},
		"client_secret": {c.secret},
	}
	endpoint := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", c.baseURL, c.realm)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("keycloak token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("keycloak token request: %s: %s", resp.Status, strings.TrimSpace(string(detail)))
	}

	var payload struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.AccessToken == "" {
		return "", errors.New("keycloak token response contained no access_token")
	}

	// Renew early so a token can't expire mid-flight.
	c.token.value = payload.AccessToken
	c.token.expiresAt = time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second).Add(-30 * time.Second)
	return payload.AccessToken, nil
}
