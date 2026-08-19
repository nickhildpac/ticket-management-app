// Package auth verifies Keycloak-issued OIDC access tokens and maps their
// realm roles onto the domain role model.
//
// The ticket-service is a pure OAuth2 resource server: it never sees a
// password, never mints a token, and holds no signing secret. It trusts
// RS256-signed tokens from the configured issuer, checked against the issuer's
// published JWKS (fetched lazily and re-fetched on key rotation by go-oidc).
package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

// ErrInvalidToken is returned for every verification failure. Callers turn it
// into a 401 without echoing the underlying reason to the client — the detail
// (bad signature vs. expired vs. wrong audience) is useful in logs but is a
// probing oracle in a response body.
var ErrInvalidToken = errors.New("invalid token")

// Claims is the subset of a Keycloak access token the application cares about.
type Claims struct {
	// Subject is the Keycloak user id (`sub`) — the stable external identity.
	// It is *not* the local users.id; see the identity service for the mapping.
	Subject   string
	Email     string
	FirstName string
	LastName  string
	Username  string

	// Roles is the raw realm_access.roles list, and Role is that list reduced
	// to the single effective domain role.
	Roles []string
	Role  domain.UserRole

	// ClientID is `azp` (authorized party) — which client the token was issued
	// to. ServiceAccount is true for client_credentials tokens (no human).
	ClientID       string
	ServiceAccount bool

	// ExpiresAt is used only for logging/diagnostics; expiry is already
	// enforced during verification.
	ExpiresAt time.Time
}

// Config describes the issuer this service accepts tokens from.
type Config struct {
	// IssuerURL is the value tokens must carry in `iss`, e.g.
	// http://localhost:8090/realms/ticket-management.
	IssuerURL string

	// DiscoveryURL, when set and different from IssuerURL, is where discovery
	// and JWKS are actually fetched. Needed inside Docker, where the browser
	// reaches Keycloak at localhost:8090 but this container reaches it at
	// keycloak:8080 — the token's `iss` is still the browser-facing URL.
	DiscoveryURL string

	// Audience must appear in the token's `aud`. Empty disables the check,
	// which is only safe in a single-client realm.
	Audience string
}

// Verifier validates access tokens against a Keycloak realm.
type Verifier struct {
	verifier *oidc.IDTokenVerifier
	issuer   string
}

// NewVerifier performs OIDC discovery against the issuer and returns a
// verifier. Discovery is a network call, so it retries: in Compose the API can
// win the startup race against Keycloak, and failing fast there would crash-loop
// the service for no reason.
func NewVerifier(ctx context.Context, cfg Config) (*Verifier, error) {
	if strings.TrimSpace(cfg.IssuerURL) == "" {
		return nil, errors.New("keycloak issuer URL is required")
	}
	issuer := strings.TrimRight(cfg.IssuerURL, "/")
	discoveryURL := strings.TrimRight(strings.TrimSpace(cfg.DiscoveryURL), "/")
	split := discoveryURL != "" && discoveryURL != issuer

	discoveryCtx := ctx
	fetchFrom := issuer
	if split {
		// Fetch from the in-network URL but keep requiring the browser-facing
		// issuer in `iss`. This is the documented escape hatch for exactly the
		// split-horizon DNS case, not a relaxation of issuer checking.
		discoveryCtx = oidc.InsecureIssuerURLContext(ctx, issuer)
		fetchFrom = discoveryURL
	}

	provider, err := discoverWithRetry(discoveryCtx, fetchFrom)
	if err != nil {
		return nil, err
	}

	// Keycloak advertises `jwks_uri` using its public hostname, which in Docker
	// this process cannot reach. Rewriting it onto the in-network origin is the
	// other half of the split — without it, discovery succeeds and then every
	// verification fails trying to fetch keys from the browser-facing URL.
	jwksURL, err := jwksURLFrom(provider, fetchFrom, split)
	if err != nil {
		return nil, err
	}

	oidcCfg := &oidc.Config{
		ClientID:             cfg.Audience,
		SupportedSigningAlgs: []string{oidc.RS256},
	}
	if cfg.Audience == "" {
		// Keep this loud: without an audience check any token the realm issues
		// to any client is accepted here.
		log.Println("WARNING: KEYCLOAK_AUDIENCE is unset; accepting tokens issued to any client in the realm")
		oidcCfg.SkipClientIDCheck = true
	}

	// Build the verifier from an explicit key set rather than provider.Verifier,
	// so the JWKS URL above is actually the one used. Issuer checking is
	// unchanged: NewVerifier still requires `iss` to equal the public issuer.
	return &Verifier{
		verifier: oidc.NewVerifier(issuer, oidc.NewRemoteKeySet(ctx, jwksURL), oidcCfg),
		issuer:   issuer,
	}, nil
}

// jwksURLFrom reads `jwks_uri` out of the discovery document, re-homing it onto
// the discovery origin when discovery and issuer differ.
func jwksURLFrom(provider *oidc.Provider, discoveryURL string, split bool) (string, error) {
	var doc struct {
		JWKSURL string `json:"jwks_uri"`
	}
	if err := provider.Claims(&doc); err != nil {
		return "", fmt.Errorf("read jwks_uri from discovery document: %w", err)
	}
	if doc.JWKSURL == "" {
		return "", errors.New("discovery document contained no jwks_uri")
	}
	if !split {
		return doc.JWKSURL, nil
	}

	advertised, err := url.Parse(doc.JWKSURL)
	if err != nil {
		return "", fmt.Errorf("parse jwks_uri %q: %w", doc.JWKSURL, err)
	}
	internal, err := url.Parse(discoveryURL)
	if err != nil {
		return "", fmt.Errorf("parse discovery URL %q: %w", discoveryURL, err)
	}
	advertised.Scheme, advertised.Host = internal.Scheme, internal.Host
	return advertised.String(), nil
}

// discoverWithRetry retries OIDC discovery with a linear backoff, giving
// Keycloak time to finish booting.
func discoverWithRetry(ctx context.Context, issuer string) (*oidc.Provider, error) {
	const attempts = 10
	var lastErr error
	for i := range attempts {
		provider, err := oidc.NewProvider(ctx, issuer)
		if err == nil {
			return provider, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			break
		}
		wait := time.Duration(i+1) * time.Second
		log.Printf("keycloak discovery failed (attempt %d/%d), retrying in %s: %v", i+1, attempts, wait, err)
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("keycloak discovery cancelled: %w", ctx.Err())
		case <-time.After(wait):
		}
	}
	return nil, fmt.Errorf("keycloak discovery at %s failed: %w", issuer, lastErr)
}

// Issuer returns the issuer URL tokens are validated against.
func (v *Verifier) Issuer() string { return v.issuer }

// rawClaims mirrors the Keycloak access token body.
type rawClaims struct {
	Subject     string `json:"sub"`
	Email       string `json:"email"`
	GivenName   string `json:"given_name"`
	FamilyName  string `json:"family_name"`
	Name        string `json:"name"`
	Username    string `json:"preferred_username"`
	AZP         string `json:"azp"`
	RealmAccess struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	Expiry int64 `json:"exp"`
}

// Verify checks signature, issuer, audience and expiry, then extracts claims.
func (v *Verifier) Verify(ctx context.Context, rawToken string) (*Claims, error) {
	token, err := v.verifier.Verify(ctx, rawToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	var raw rawClaims
	if err := token.Claims(&raw); err != nil {
		return nil, fmt.Errorf("%w: malformed claims: %v", ErrInvalidToken, err)
	}
	if raw.Subject == "" {
		return nil, fmt.Errorf("%w: token has no subject", ErrInvalidToken)
	}

	first, last := raw.GivenName, raw.FamilyName
	if first == "" && last == "" {
		first, last = splitName(raw.Name, raw.Username, raw.Email)
	}

	return &Claims{
		Subject:   raw.Subject,
		Email:     raw.Email,
		FirstName: first,
		LastName:  last,
		Username:  raw.Username,
		Roles:     raw.RealmAccess.Roles,
		Role:      RoleFromRealmRoles(raw.RealmAccess.Roles),
		ClientID:  raw.AZP,
		// Keycloak names service-account users after their client; there is no
		// dedicated claim distinguishing them from human logins.
		ServiceAccount: strings.HasPrefix(raw.Username, "service-account-"),
		ExpiresAt:      token.Expiry,
	}, nil
}

// RoleFromRealmRoles reduces a token's realm roles to one effective domain
// role, most-privileged first.
//
// The fallback is the least-privileged role, never the most: an unrecognised or
// missing role must not be able to escalate. A realm where the roles mapper is
// misconfigured degrades to end-user access rather than granting admin.
func RoleFromRealmRoles(roles []string) domain.UserRole {
	var agent bool
	for _, r := range roles {
		switch strings.ToLower(strings.TrimSpace(r)) {
		case string(domain.RoleAdmin):
			return domain.RoleAdmin
		case string(domain.RoleAgent):
			agent = true
		}
	}
	if agent {
		return domain.RoleAgent
	}
	return domain.RoleUser
}

// splitName derives a first/last name when the token carries no given_name or
// family_name (service accounts, or a realm without the profile scope).
func splitName(fullName, username, email string) (first, last string) {
	candidate := strings.TrimSpace(fullName)
	if candidate == "" {
		candidate = strings.TrimSpace(username)
	}
	if candidate == "" {
		candidate, _, _ = strings.Cut(email, "@")
	}
	if candidate == "" {
		return "Unknown", "User"
	}
	if name, rest, found := strings.Cut(candidate, " "); found {
		return name, rest
	}
	return candidate, ""
}
