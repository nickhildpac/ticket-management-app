package keycloak

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRealm serves the OIDC discovery document, a JWKS and a token endpoint,
// standing in for a Keycloak realm.
type fakeRealm struct {
	server *httptest.Server
	key    *rsa.PrivateKey

	tokenRequests atomic.Int32
	lastForm      url.Values
	tokenTTL      int
}

func newFakeRealm(t *testing.T) *fakeRealm {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	fr := &fakeRealm{key: key, tokenTTL: 300}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                fr.issuer(),
			"token_endpoint":                        fr.issuer() + "/protocol/openid-connect/token",
			"jwks_uri":                              fr.issuer() + "/protocol/openid-connect/certs",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	mux.HandleFunc("/protocol/openid-connect/certs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{{
				Key: key.Public(), KeyID: "kid-1", Algorithm: string(jose.RS256), Use: "sig",
			}},
		})
	})
	mux.HandleFunc("/protocol/openid-connect/token", func(w http.ResponseWriter, r *http.Request) {
		fr.tokenRequests.Add(1)
		_ = r.ParseForm()
		fr.lastForm = r.PostForm

		if r.PostForm.Get("client_secret") != "correct-secret" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid_client"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "issued-token",
			"token_type":   "Bearer",
			"expires_in":   fr.tokenTTL,
		})
	})

	fr.server = httptest.NewServer(mux)
	t.Cleanup(fr.server.Close)
	return fr
}

func (f *fakeRealm) issuer() string { return f.server.URL }

func (f *fakeRealm) tokenURL() string { return f.server.URL + "/protocol/openid-connect/token" }

type claimSet struct {
	Issuer      string   `json:"iss"`
	Subject     string   `json:"sub"`
	Audience    []string `json:"aud"`
	Expiry      int64    `json:"exp"`
	IssuedAt    int64    `json:"iat"`
	Username    string   `json:"preferred_username,omitempty"`
	RealmAccess struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
}

func (f *fakeRealm) sign(t *testing.T, c claimSet) string {
	t.Helper()
	if c.Issuer == "" {
		c.Issuer = f.issuer()
	}
	if c.Expiry == 0 {
		c.Expiry = time.Now().Add(5 * time.Minute).Unix()
	}
	if c.IssuedAt == 0 {
		c.IssuedAt = time.Now().Unix()
	}
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: f.key},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader(jose.HeaderKey("kid"), "kid-1"),
	)
	require.NoError(t, err)
	raw, err := jwt.Signed(signer).Claims(c).Serialize()
	require.NoError(t, err)
	return raw
}

func serviceAccountClaims() claimSet {
	c := claimSet{
		Subject:  "00000000-0000-4000-8000-0000000000a1",
		Audience: []string{"ticket-service"},
		Username: "service-account-ai-service",
	}
	c.RealmAccess.Roles = []string{"admin"}
	return c
}

func TestVerifierAcceptsRealmToken(t *testing.T) {
	realm := newFakeRealm(t)
	v, err := NewVerifier(context.Background(), VerifierConfig{
		IssuerURL: realm.issuer(), Audience: "ticket-service",
	})
	require.NoError(t, err)

	claims, err := v.Verify(context.Background(), realm.sign(t, serviceAccountClaims()))
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-4000-8000-0000000000a1", claims.Subject)
	assert.Equal(t, "service-account-ai-service", claims.Username)
	assert.Contains(t, claims.Roles, "admin")
}

// Paid model calls sit behind this endpoint, so a token from a realm we don't
// trust must not get through.
func TestVerifierRejectsForeignKeyAndIssuerAndAudience(t *testing.T) {
	realm := newFakeRealm(t)
	v, err := NewVerifier(context.Background(), VerifierConfig{
		IssuerURL: realm.issuer(), Audience: "ticket-service",
	})
	require.NoError(t, err)

	t.Run("forged signature", func(t *testing.T) {
		attackerKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)
		signer, err := jose.NewSigner(
			jose.SigningKey{Algorithm: jose.RS256, Key: attackerKey},
			(&jose.SignerOptions{}).WithType("JWT").WithHeader(jose.HeaderKey("kid"), "kid-1"),
		)
		require.NoError(t, err)
		forged, err := jwt.Signed(signer).Claims(serviceAccountClaims()).Serialize()
		require.NoError(t, err)

		_, err = v.Verify(context.Background(), forged)
		require.Error(t, err)
	})

	t.Run("wrong audience", func(t *testing.T) {
		c := serviceAccountClaims()
		c.Audience = []string{"another-service"}
		_, err := v.Verify(context.Background(), realm.sign(t, c))
		require.Error(t, err)
	})

	t.Run("wrong issuer", func(t *testing.T) {
		c := serviceAccountClaims()
		c.Issuer = "https://evil.test/realms/x"
		_, err := v.Verify(context.Background(), realm.sign(t, c))
		require.Error(t, err)
	})

	t.Run("expired", func(t *testing.T) {
		c := serviceAccountClaims()
		c.IssuedAt = time.Now().Add(-2 * time.Hour).Unix()
		c.Expiry = time.Now().Add(-time.Hour).Unix()
		_, err := v.Verify(context.Background(), realm.sign(t, c))
		require.Error(t, err)
	})

	t.Run("garbage", func(t *testing.T) {
		_, err := v.Verify(context.Background(), "not-a-jwt")
		require.Error(t, err)
	})
}

// Inside Docker, Keycloak advertises both `iss` and `jwks_uri` on its public
// hostname, which this process cannot reach. The verifier must validate against
// the public issuer while fetching keys over the in-network URL; trusting
// jwks_uri verbatim makes every token fail with a connection error.
func TestVerifierSeparateDiscoveryURL(t *testing.T) {
	realm := newFakeRealm(t)
	externalIssuer := "http://localhost:8090/realms/ticket-management"

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                externalIssuer,
			"jwks_uri":                              externalIssuer + "/protocol/openid-connect/certs",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	// Same path as the public URL advertises; only scheme and host differ.
	mux.HandleFunc("/realms/ticket-management/protocol/openid-connect/certs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{{
				Key: realm.key.Public(), KeyID: "kid-1", Algorithm: string(jose.RS256), Use: "sig",
			}},
		})
	})
	internal := httptest.NewServer(mux)
	defer internal.Close()

	v, err := NewVerifier(context.Background(), VerifierConfig{
		IssuerURL:    externalIssuer,
		DiscoveryURL: internal.URL,
		Audience:     "ticket-service",
	})
	require.NoError(t, err)

	c := serviceAccountClaims()
	c.Issuer = externalIssuer
	_, err = v.Verify(context.Background(), realm.sign(t, c))
	require.NoError(t, err)

	// The in-network URL must not be accepted as an issuer.
	c.Issuer = internal.URL
	_, err = v.Verify(context.Background(), realm.sign(t, c))
	require.Error(t, err)
}

func TestTokenSourceFetchesAndCaches(t *testing.T) {
	realm := newFakeRealm(t)
	ts, err := NewTokenSource(TokenSourceConfig{
		TokenURL: realm.tokenURL(), ClientID: "ai-service", ClientSecret: "correct-secret",
	})
	require.NoError(t, err)

	for range 5 {
		token, err := ts.Token(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "issued-token", token)
	}

	// One exchange for five calls: the worker would otherwise hit the token
	// endpoint on every consumed event.
	assert.Equal(t, int32(1), realm.tokenRequests.Load())
	assert.Equal(t, "client_credentials", realm.lastForm.Get("grant_type"))
	assert.Equal(t, "ai-service", realm.lastForm.Get("client_id"))
}

func TestTokenSourceSurfacesBadCredentials(t *testing.T) {
	realm := newFakeRealm(t)
	ts, err := NewTokenSource(TokenSourceConfig{
		TokenURL: realm.tokenURL(), ClientID: "ai-service", ClientSecret: "wrong-secret",
	})
	require.NoError(t, err)

	_, err = ts.Token(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client_credentials")
}

func TestNewTokenSourceRequiresCredentials(t *testing.T) {
	_, err := NewTokenSource(TokenSourceConfig{IssuerURL: "http://x/realms/y", ClientID: "ai-service"})
	require.Error(t, err)

	_, err = NewTokenSource(TokenSourceConfig{IssuerURL: "http://x/realms/y", ClientSecret: "s"})
	require.Error(t, err)
}

func TestTokenURLFromIssuer(t *testing.T) {
	got, err := TokenURLFromIssuer("http://keycloak:8080/realms/ticket-management/")
	require.NoError(t, err)
	assert.Equal(t, "http://keycloak:8080/realms/ticket-management/protocol/openid-connect/token", got)

	_, err = TokenURLFromIssuer("http://keycloak:8080/not-a-realm")
	require.Error(t, err)

	_, err = TokenURLFromIssuer("")
	require.Error(t, err)
}

func TestTokenRespectsCancelledContext(t *testing.T) {
	realm := newFakeRealm(t)
	ts, err := NewTokenSource(TokenSourceConfig{
		TokenURL: realm.tokenURL(), ClientID: "ai-service", ClientSecret: "correct-secret",
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = ts.Token(ctx)
	require.ErrorIs(t, err, context.Canceled)
}
