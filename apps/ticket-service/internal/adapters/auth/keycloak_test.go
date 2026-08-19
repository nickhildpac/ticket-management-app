package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

// fakeIssuer is a minimal stand-in for a Keycloak realm: the discovery
// document plus a JWKS, signing with a key the test controls. It lets the
// tests exercise the real go-oidc verification path (signature, issuer,
// audience, expiry) without a Keycloak container.
type fakeIssuer struct {
	server *httptest.Server
	key    *rsa.PrivateKey
	keyID  string
}

func newFakeIssuer(t *testing.T) *fakeIssuer {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	fi := &fakeIssuer{key: key, keyID: "test-key-1"}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                fi.issuer(),
			"authorization_endpoint":                fi.issuer() + "/protocol/openid-connect/auth",
			"token_endpoint":                        fi.issuer() + "/protocol/openid-connect/token",
			"jwks_uri":                              fi.issuer() + "/protocol/openid-connect/certs",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	mux.HandleFunc("/protocol/openid-connect/certs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{{
				Key:       key.Public(),
				KeyID:     fi.keyID,
				Algorithm: string(jose.RS256),
				Use:       "sig",
			}},
		})
	})

	fi.server = httptest.NewServer(mux)
	t.Cleanup(fi.server.Close)
	return fi
}

func (f *fakeIssuer) issuer() string { return f.server.URL }

// tokenClaims is the shape the fake issuer signs; it mirrors a Keycloak access token.
type tokenClaims struct {
	Issuer      string   `json:"iss"`
	Subject     string   `json:"sub"`
	Audience    []string `json:"aud"`
	Expiry      int64    `json:"exp"`
	IssuedAt    int64    `json:"iat"`
	Email       string   `json:"email,omitempty"`
	GivenName   string   `json:"given_name,omitempty"`
	FamilyName  string   `json:"family_name,omitempty"`
	Name        string   `json:"name,omitempty"`
	Username    string   `json:"preferred_username,omitempty"`
	AZP         string   `json:"azp,omitempty"`
	RealmAccess struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
}

func (f *fakeIssuer) sign(t *testing.T, c tokenClaims, signWith *rsa.PrivateKey) string {
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
	if signWith == nil {
		signWith = f.key
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: signWith},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader(jose.HeaderKey("kid"), f.keyID),
	)
	if err != nil {
		t.Fatalf("new signer: %v", err)
	}
	raw, err := jwt.Signed(signer).Claims(c).Serialize()
	if err != nil {
		t.Fatalf("sign claims: %v", err)
	}
	return raw
}

func (f *fakeIssuer) verifier(t *testing.T, audience string) *Verifier {
	t.Helper()
	v, err := NewVerifier(context.Background(), Config{IssuerURL: f.issuer(), Audience: audience})
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}
	return v
}

func baseClaims() tokenClaims {
	c := tokenClaims{
		Subject:    "a1111111-1111-4111-8111-111111111111",
		Audience:   []string{"ticket-service"},
		Email:      "alice@admin.com",
		GivenName:  "Alice",
		FamilyName: "Admin",
		Username:   "alice@admin.com",
		AZP:        "ticket-web",
	}
	c.RealmAccess.Roles = []string{"admin", "default-roles-ticket-management"}
	return c
}

func TestVerifyValidToken(t *testing.T) {
	fi := newFakeIssuer(t)
	v := fi.verifier(t, "ticket-service")

	claims, err := v.Verify(context.Background(), fi.sign(t, baseClaims(), nil))
	if err != nil {
		t.Fatalf("expected token to verify, got %v", err)
	}

	if claims.Subject != "a1111111-1111-4111-8111-111111111111" {
		t.Errorf("subject = %q", claims.Subject)
	}
	if claims.Role != domain.RoleAdmin {
		t.Errorf("role = %q, want admin", claims.Role)
	}
	if claims.Email != "alice@admin.com" {
		t.Errorf("email = %q", claims.Email)
	}
	if claims.FirstName != "Alice" || claims.LastName != "Admin" {
		t.Errorf("name = %q %q", claims.FirstName, claims.LastName)
	}
	if claims.ServiceAccount {
		t.Error("human login should not be flagged as a service account")
	}
}

// A token signed by a key the issuer never published must be rejected — this is
// the check that stops a forged token from being accepted.
func TestVerifyRejectsForeignSigningKey(t *testing.T) {
	fi := newFakeIssuer(t)
	v := fi.verifier(t, "ticket-service")

	attacker, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate attacker key: %v", err)
	}

	forged := fi.sign(t, baseClaims(), attacker)
	if _, err := v.Verify(context.Background(), forged); err == nil {
		t.Fatal("expected a token signed with an unknown key to be rejected")
	} else if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestVerifyRejectsWrongAudience(t *testing.T) {
	fi := newFakeIssuer(t)
	v := fi.verifier(t, "ticket-service")

	c := baseClaims()
	c.Audience = []string{"some-other-service"}

	if _, err := v.Verify(context.Background(), fi.sign(t, c, nil)); err == nil {
		t.Fatal("expected a token for another audience to be rejected")
	}
}

func TestVerifyRejectsWrongIssuer(t *testing.T) {
	fi := newFakeIssuer(t)
	v := fi.verifier(t, "ticket-service")

	c := baseClaims()
	c.Issuer = "https://evil.example.com/realms/ticket-management"

	if _, err := v.Verify(context.Background(), fi.sign(t, c, nil)); err == nil {
		t.Fatal("expected a token from another issuer to be rejected")
	}
}

func TestVerifyRejectsExpiredToken(t *testing.T) {
	fi := newFakeIssuer(t)
	v := fi.verifier(t, "ticket-service")

	c := baseClaims()
	c.IssuedAt = time.Now().Add(-2 * time.Hour).Unix()
	c.Expiry = time.Now().Add(-time.Hour).Unix()

	if _, err := v.Verify(context.Background(), fi.sign(t, c, nil)); err == nil {
		t.Fatal("expected an expired token to be rejected")
	}
}

func TestVerifyRejectsGarbage(t *testing.T) {
	fi := newFakeIssuer(t)
	v := fi.verifier(t, "ticket-service")

	for _, raw := range []string{"", "not-a-jwt", "a.b.c"} {
		if _, err := v.Verify(context.Background(), raw); err == nil {
			t.Errorf("expected %q to be rejected", raw)
		}
	}
}

func TestVerifyServiceAccountToken(t *testing.T) {
	fi := newFakeIssuer(t)
	v := fi.verifier(t, "ticket-service")

	c := tokenClaims{
		Subject:  "00000000-0000-4000-8000-0000000000a1",
		Audience: []string{"ticket-service"},
		Username: "service-account-ai-service",
		AZP:      "ai-service",
		Email:    "ai-triage@service.local",
	}
	c.RealmAccess.Roles = []string{"admin"}

	claims, err := v.Verify(context.Background(), fi.sign(t, c, nil))
	if err != nil {
		t.Fatalf("expected service account token to verify, got %v", err)
	}
	if !claims.ServiceAccount {
		t.Error("expected ServiceAccount to be true")
	}
	if claims.ClientID != "ai-service" {
		t.Errorf("clientID = %q", claims.ClientID)
	}
	if claims.Role != domain.RoleAdmin {
		t.Errorf("role = %q, want admin", claims.Role)
	}
	// No given_name/family_name on service accounts; the fallback must still
	// produce something usable for the local user row.
	if claims.FirstName == "" {
		t.Error("expected a derived first name")
	}
}

// The realm may hand out an audience-less token if the ticket-audience scope is
// dropped; with the check disabled that must still verify, and the constructor
// is expected to have warned.
func TestVerifyWithoutAudienceCheck(t *testing.T) {
	fi := newFakeIssuer(t)
	v := fi.verifier(t, "")

	c := baseClaims()
	c.Audience = []string{"account"}

	if _, err := v.Verify(context.Background(), fi.sign(t, c, nil)); err != nil {
		t.Fatalf("expected token to verify with audience check disabled, got %v", err)
	}
}

// Docker split-horizon: `iss` is the browser-facing URL, discovery happens over
// the in-network URL.
func TestVerifierSeparateDiscoveryURL(t *testing.T) {
	fi := newFakeIssuer(t)
	externalIssuer := "http://localhost:8090/realms/ticket-management"

	// Re-serve discovery the way Keycloak does with KC_HOSTNAME set: both the
	// issuer *and* the advertised jwks_uri use the public hostname, which this
	// process cannot reach. The verifier has to re-home the JWKS URL onto the
	// in-network origin; if it just trusts jwks_uri, key fetching hangs or
	// connection-refuses and every token is rejected.
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                externalIssuer,
			"jwks_uri":                              externalIssuer + "/protocol/openid-connect/certs",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	// Served at the same path the external URL advertises: only scheme and host
	// differ between the two, which is exactly the rewrite the verifier performs.
	mux.HandleFunc("/realms/ticket-management/protocol/openid-connect/certs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{{
				Key: fi.key.Public(), KeyID: fi.keyID, Algorithm: string(jose.RS256), Use: "sig",
			}},
		})
	})
	internal := httptest.NewServer(mux)
	defer internal.Close()

	v, err := NewVerifier(context.Background(), Config{
		IssuerURL:    externalIssuer,
		DiscoveryURL: internal.URL,
		Audience:     "ticket-service",
	})
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}

	c := baseClaims()
	c.Issuer = externalIssuer
	if _, err := v.Verify(context.Background(), fi.sign(t, c, nil)); err != nil {
		t.Fatalf("expected token with external issuer to verify, got %v", err)
	}

	// The in-network URL must not be accepted as `iss`.
	c.Issuer = internal.URL
	if _, err := v.Verify(context.Background(), fi.sign(t, c, nil)); err == nil {
		t.Fatal("expected the internal discovery URL to be rejected as an issuer")
	}
}

func TestRoleFromRealmRoles(t *testing.T) {
	tests := []struct {
		name  string
		roles []string
		want  domain.UserRole
	}{
		{"admin wins over agent", []string{"agent", "admin"}, domain.RoleAdmin},
		{"admin wins over user", []string{"user", "admin"}, domain.RoleAdmin},
		{"agent wins over user", []string{"user", "agent"}, domain.RoleAgent},
		{"plain user", []string{"user"}, domain.RoleUser},
		{"case insensitive", []string{"ADMIN"}, domain.RoleAdmin},
		{"whitespace tolerated", []string{" agent "}, domain.RoleAgent},
		{"no roles falls back to user", nil, domain.RoleUser},
		{"unknown roles fall back to user", []string{"offline_access", "uma_authorization"}, domain.RoleUser},
		{"unknown role never escalates", []string{"superadmin", "root"}, domain.RoleUser},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := RoleFromRealmRoles(tc.roles); got != tc.want {
				t.Errorf("RoleFromRealmRoles(%v) = %q, want %q", tc.roles, got, tc.want)
			}
		})
	}
}

func TestSplitName(t *testing.T) {
	tests := []struct {
		full, username, email string
		wantFirst, wantLast   string
	}{
		{"Alice Admin", "", "", "Alice", "Admin"},
		{"Mary Jane Watson", "", "", "Mary", "Jane Watson"},
		{"", "service-account-ai-service", "", "service-account-ai-service", ""},
		{"", "", "charlie@user.com", "charlie", ""},
		{"", "", "", "Unknown", "User"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s|%s|%s", tc.full, tc.username, tc.email), func(t *testing.T) {
			first, last := splitName(tc.full, tc.username, tc.email)
			if first != tc.wantFirst || last != tc.wantLast {
				t.Errorf("splitName(%q,%q,%q) = %q,%q; want %q,%q",
					tc.full, tc.username, tc.email, first, last, tc.wantFirst, tc.wantLast)
			}
		})
	}
}
