// Package keycloak provides the ai-service's two halves of Keycloak
// integration: verifying inbound access tokens on its own HTTP surface, and
// obtaining outbound service-account tokens for callbacks into ticket-service.
//
// Before this, the worker signed its own HS256 tokens with a secret shared with
// ticket-service — which meant it could mint a token for *any* user id, and a
// leak of that one secret forged arbitrary identities. Now it asks Keycloak for
// a token bound to its own service account, and holds no signing key at all.
package keycloak

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// ErrInvalidToken is returned for any inbound verification failure.
var ErrInvalidToken = errors.New("invalid token")

// VerifierConfig configures inbound token verification.
type VerifierConfig struct {
	// IssuerURL is what tokens must carry in `iss`.
	IssuerURL string
	// DiscoveryURL, when set and different, is where discovery/JWKS are fetched
	// from (Docker split-horizon; see infra/keycloak/README.md).
	DiscoveryURL string
	// Audience must appear in `aud`; empty disables the check.
	Audience string
}

// Verifier validates inbound Keycloak access tokens.
type Verifier struct {
	verifier *oidc.IDTokenVerifier
}

// NewVerifier runs OIDC discovery, retrying while Keycloak boots.
func NewVerifier(ctx context.Context, cfg VerifierConfig) (*Verifier, error) {
	if strings.TrimSpace(cfg.IssuerURL) == "" {
		return nil, errors.New("keycloak issuer URL is required")
	}
	issuer := strings.TrimRight(cfg.IssuerURL, "/")
	discoveryURL := strings.TrimRight(strings.TrimSpace(cfg.DiscoveryURL), "/")
	split := discoveryURL != "" && discoveryURL != issuer

	discoveryCtx := ctx
	fetchFrom := issuer
	if split {
		discoveryCtx = oidc.InsecureIssuerURLContext(ctx, issuer)
		fetchFrom = discoveryURL
	}

	provider, err := discoverWithRetry(discoveryCtx, fetchFrom)
	if err != nil {
		return nil, err
	}

	// Keycloak advertises jwks_uri on its public hostname, which this process
	// can't reach inside Docker. Re-home it onto the in-network origin, or
	// discovery succeeds and every subsequent verification fails fetching keys.
	jwksURL, err := jwksURLFrom(provider, fetchFrom, split)
	if err != nil {
		return nil, err
	}

	oidcCfg := &oidc.Config{
		ClientID:             cfg.Audience,
		SupportedSigningAlgs: []string{oidc.RS256},
	}
	if cfg.Audience == "" {
		log.Println("WARNING: KEYCLOAK_AUDIENCE is unset; accepting tokens issued to any client in the realm")
		oidcCfg.SkipClientIDCheck = true
	}
	// Explicit key set rather than provider.Verifier, so the URL above is the
	// one actually used. `iss` is still required to equal the public issuer.
	return &Verifier{
		verifier: oidc.NewVerifier(issuer, oidc.NewRemoteKeySet(ctx, jwksURL), oidcCfg),
	}, nil
}

// jwksURLFrom reads jwks_uri from the discovery document, re-homing it onto the
// discovery origin when discovery and issuer differ.
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

func discoverWithRetry(ctx context.Context, issuer string) (*oidc.Provider, error) {
	const attempts = 10
	var lastErr error
	for i := range attempts {
		provider, err := oidc.NewProvider(ctx, issuer)
		if err == nil {
			return provider, nil
		}
		lastErr = err
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

// Claims is the subset of an inbound token this service reads.
type Claims struct {
	Subject  string
	Username string
	Roles    []string
}

// Verify checks signature, issuer, audience and expiry.
func (v *Verifier) Verify(ctx context.Context, rawToken string) (*Claims, error) {
	token, err := v.verifier.Verify(ctx, rawToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	var raw struct {
		Subject     string `json:"sub"`
		Username    string `json:"preferred_username"`
		RealmAccess struct {
			Roles []string `json:"roles"`
		} `json:"realm_access"`
	}
	if err := token.Claims(&raw); err != nil {
		return nil, fmt.Errorf("%w: malformed claims: %v", ErrInvalidToken, err)
	}
	return &Claims{Subject: raw.Subject, Username: raw.Username, Roles: raw.RealmAccess.Roles}, nil
}

// ---- outbound service-account tokens ------------------------------------

// TokenSourceConfig configures the client_credentials grant.
type TokenSourceConfig struct {
	// TokenURL is the realm's token endpoint. Derived from IssuerURL when empty.
	TokenURL string
	// IssuerURL is used to derive TokenURL and is otherwise unused.
	IssuerURL    string
	ClientID     string
	ClientSecret string
	// Scopes are optional; the realm's default client scopes already carry the
	// audience mapper this service needs.
	Scopes []string
}

// TokenSource hands out service-account access tokens, refreshing them before
// expiry. It is safe for concurrent use.
type TokenSource struct {
	source oauth2.TokenSource
}

// NewTokenSource builds a client_credentials token source.
func NewTokenSource(cfg TokenSourceConfig) (*TokenSource, error) {
	if strings.TrimSpace(cfg.ClientID) == "" || strings.TrimSpace(cfg.ClientSecret) == "" {
		return nil, errors.New("KEYCLOAK_CLIENT_ID and KEYCLOAK_CLIENT_SECRET are required for ticket-service callbacks")
	}

	tokenURL := strings.TrimSpace(cfg.TokenURL)
	if tokenURL == "" {
		derived, err := TokenURLFromIssuer(cfg.IssuerURL)
		if err != nil {
			return nil, err
		}
		tokenURL = derived
	}

	conf := &clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     tokenURL,
		Scopes:       cfg.Scopes,
		AuthStyle:    oauth2.AuthStyleInParams,
	}
	// oauth2's ReuseTokenSource caches until ~10s before expiry, so the token
	// endpoint is hit once per token lifetime rather than once per event.
	return &TokenSource{source: oauth2.ReuseTokenSource(nil, conf.TokenSource(context.Background()))}, nil
}

// TokenURLFromIssuer derives the OIDC token endpoint from a realm URL.
func TokenURLFromIssuer(issuer string) (string, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(issuer), "/")
	if trimmed == "" {
		return "", errors.New("keycloak issuer URL is required to derive the token endpoint")
	}
	if _, err := url.Parse(trimmed); err != nil {
		return "", fmt.Errorf("parse issuer URL: %w", err)
	}
	if !strings.Contains(trimmed, "/realms/") {
		return "", fmt.Errorf("issuer URL %q is not a /realms/<realm> URL", issuer)
	}
	return trimmed + "/protocol/openid-connect/token", nil
}

// Token returns a valid access token, fetching or refreshing as needed.
func (t *TokenSource) Token(ctx context.Context) (string, error) {
	// oauth2.TokenSource has no context parameter; the HTTP client used for the
	// exchange is carried on the context the source was built with. Honour
	// cancellation explicitly so a shutting-down worker doesn't block here.
	if err := ctx.Err(); err != nil {
		return "", err
	}
	token, err := t.source.Token()
	if err != nil {
		return "", fmt.Errorf("keycloak client_credentials: %w", err)
	}
	return token.AccessToken, nil
}

// staticTokenSource lets tests inject a fixed token.
type staticTokenSource struct {
	token string
	err   error
}

func (s staticTokenSource) Token() (*oauth2.Token, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &oauth2.Token{AccessToken: s.token, TokenType: "Bearer", Expiry: time.Now().Add(time.Hour)}, nil
}

// NewStaticTokenSource returns a TokenSource that always yields token. For tests.
func NewStaticTokenSource(token string) *TokenSource {
	return &TokenSource{source: staticTokenSource{token: token}}
}
