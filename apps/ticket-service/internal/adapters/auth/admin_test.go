package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

func TestSplitRealmURL(t *testing.T) {
	tests := []struct {
		in        string
		wantBase  string
		wantRealm string
		wantErr   bool
	}{
		{"http://keycloak:8080/realms/ticket-management", "http://keycloak:8080", "ticket-management", false},
		{"https://sso.example.com/realms/prod/", "https://sso.example.com", "prod", false},
		// Keycloak behind a path prefix.
		{"https://example.com/auth/realms/prod", "https://example.com/auth", "prod", false},
		{"https://example.com/not-a-realm", "", "", true},
		{"", "", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			base, realm, err := splitRealmURL(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error for %q", tc.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if base != tc.wantBase || realm != tc.wantRealm {
				t.Errorf("splitRealmURL(%q) = %q,%q; want %q,%q", tc.in, base, realm, tc.wantBase, tc.wantRealm)
			}
		})
	}
}

// Without credentials the client is nil, and role writes must report themselves
// unavailable rather than silently doing nothing.
func TestNewAdminClientWithoutCredentials(t *testing.T) {
	c, err := NewAdminClient(AdminConfig{IssuerURL: "http://keycloak:8080/realms/ticket-management"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c != nil {
		t.Fatal("expected a nil client when no credentials are configured")
	}
	if err := c.SetRealmRole(context.Background(), uuid.New(), domain.RoleAdmin); err != ErrAdminNotConfigured {
		t.Errorf("expected ErrAdminNotConfigured, got %v", err)
	}
}

// Roles are mutually exclusive in this domain, so assigning one must also strip
// the others — otherwise a demoted admin keeps `admin` alongside `agent` and
// RoleFromRealmRoles keeps resolving them to admin.
func TestSetRealmRoleAddsOneAndRemovesTheOthers(t *testing.T) {
	var added, removed []string

	mux := http.NewServeMux()
	mux.HandleFunc("/realms/ticket-management/protocol/openid-connect/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "admin-token", "expires_in": 60})
	})
	mux.HandleFunc("/admin/realms/ticket-management/roles", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]realmRole{
			{ID: "id-admin", Name: "admin"},
			{ID: "id-agent", Name: "agent"},
			{ID: "id-user", Name: "user"},
			{ID: "id-other", Name: "offline_access"},
		})
	})
	mux.HandleFunc("/admin/realms/ticket-management/users/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/role-mappings/realm") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer admin-token" {
			t.Errorf("expected the service-account token, got %q", got)
		}
		var roles []realmRole
		_ = json.NewDecoder(r.Body).Decode(&roles)
		for _, role := range roles {
			switch r.Method {
			case http.MethodPost:
				added = append(added, role.Name)
			case http.MethodDelete:
				removed = append(removed, role.Name)
			}
		}
		w.WriteHeader(http.StatusNoContent)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := NewAdminClient(AdminConfig{
		IssuerURL: srv.URL + "/realms/ticket-management",
		ClientID:  "ticket-service",
		Secret:    "s3cret",
	})
	if err != nil {
		t.Fatalf("new admin client: %v", err)
	}

	if err := client.SetRealmRole(context.Background(), uuid.New(), domain.RoleAgent); err != nil {
		t.Fatalf("SetRealmRole: %v", err)
	}

	if len(added) != 1 || added[0] != "agent" {
		t.Errorf("added = %v, want [agent]", added)
	}
	wantRemoved := map[string]bool{"admin": true, "user": true}
	if len(removed) != 2 {
		t.Fatalf("removed = %v, want admin and user", removed)
	}
	for _, r := range removed {
		if !wantRemoved[r] {
			t.Errorf("unexpectedly removed %q", r)
		}
	}
}

func TestAdminTokenIsCachedBetweenCalls(t *testing.T) {
	var tokenRequests int

	mux := http.NewServeMux()
	mux.HandleFunc("/realms/ticket-management/protocol/openid-connect/token", func(w http.ResponseWriter, _ *http.Request) {
		tokenRequests++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "admin-token", "expires_in": 300})
	})
	mux.HandleFunc("/admin/realms/ticket-management/roles", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]realmRole{
			{ID: "id-admin", Name: "admin"}, {ID: "id-agent", Name: "agent"}, {ID: "id-user", Name: "user"},
		})
	})
	mux.HandleFunc("/admin/realms/ticket-management/users/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := NewAdminClient(AdminConfig{
		IssuerURL: srv.URL + "/realms/ticket-management",
		ClientID:  "ticket-service",
		Secret:    "s3cret",
	})
	if err != nil {
		t.Fatalf("new admin client: %v", err)
	}

	for range 3 {
		if err := client.SetRealmRole(context.Background(), uuid.New(), domain.RoleUser); err != nil {
			t.Fatalf("SetRealmRole: %v", err)
		}
	}
	if tokenRequests != 1 {
		t.Errorf("expected the token to be fetched once, got %d requests", tokenRequests)
	}
}

// A realm missing one of the three app roles is a misconfiguration; failing is
// better than half-applying a role change.
func TestSetRealmRoleFailsWhenARoleIsMissing(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/realms/ticket-management/protocol/openid-connect/token", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "t", "expires_in": 60})
	})
	mux.HandleFunc("/admin/realms/ticket-management/roles", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode([]realmRole{{ID: "id-admin", Name: "admin"}})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, err := NewAdminClient(AdminConfig{
		IssuerURL: srv.URL + "/realms/ticket-management",
		ClientID:  "ticket-service",
		Secret:    "s3cret",
	})
	if err != nil {
		t.Fatalf("new admin client: %v", err)
	}

	if err := client.SetRealmRole(context.Background(), uuid.New(), domain.RoleAdmin); err == nil {
		t.Fatal("expected an error when a realm role is missing")
	}
}
