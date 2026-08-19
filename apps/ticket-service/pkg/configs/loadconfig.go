// Package configs loads environment variables and application configuration values.
package configs

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"
	"time"
)

type userContextKey string

const (
	UserIDKey   userContextKey = "user_id"
	UserRoleKey userContextKey = "user_role"
)

type Config struct {
	ADDR   int
	DSN    string
	AppEnv string
	Domain string

	// ---- Keycloak (OIDC) ----------------------------------------------
	// This service is a resource server: it validates tokens and holds no
	// signing key. See docs/adr/0003-keycloak-authentication.md.

	// KeycloakIssuerURL is the realm URL as it appears in a token's `iss`,
	// e.g. http://localhost:8090/realms/ticket-management.
	KeycloakIssuerURL string
	// KeycloakInternalIssuerURL, when set, is where discovery and JWKS are
	// fetched from. Needed in Docker, where the browser and this service reach
	// Keycloak on different hostnames.
	KeycloakInternalIssuerURL string
	// KeycloakAudience must appear in the token's `aud`.
	KeycloakAudience string
	// KeycloakWebClientID is the public SPA client, served to the browser by
	// GET /api/v1/auth/config so the frontend needs no per-environment rebuild.
	KeycloakWebClientID string
	// KeycloakAdminClientID/Secret authenticate the Admin API calls that write
	// realm role changes back to Keycloak. Optional: without them the service
	// still runs, and admin role edits are rejected rather than silently
	// applied locally and then overwritten on the user's next login.
	KeycloakAdminClientID     string
	KeycloakAdminClientSecret string

	// IdentityCacheTTL bounds how long a resolved Keycloak subject → local user
	// mapping is reused before the database is read again.
	IdentityCacheTTL time.Duration

	AIServiceURL string
}

func LoadConfig() (*Config, error) {
	var config Config
	config.ADDR = GetInt("PORT", 8080)
	config.DSN = GetString("DB_ADDR", "postgres://postgres:postgres@localhost/ticket_management?sslmode=disable")
	config.AppEnv = strings.ToLower(GetString("APP_ENV", GetString("ENV", "local")))

	defaultIssuer := ""
	if isLocalOrTest(config.AppEnv) {
		defaultIssuer = "http://localhost:8090/realms/ticket-management"
	}

	flag.StringVar(&config.KeycloakIssuerURL, "keycloak-issuer", GetString("KEYCLOAK_ISSUER_URL", defaultIssuer), "Keycloak realm issuer URL")
	flag.StringVar(&config.KeycloakInternalIssuerURL, "keycloak-internal-issuer", GetString("KEYCLOAK_INTERNAL_ISSUER_URL", ""), "Keycloak realm URL for in-network discovery")
	flag.StringVar(&config.KeycloakAudience, "keycloak-audience", GetString("KEYCLOAK_AUDIENCE", "ticket-service"), "required token audience")
	flag.StringVar(&config.Domain, "domain", GetString("Domain", "example.com"), "domain")
	flag.Parse()

	config.KeycloakWebClientID = GetString("KEYCLOAK_WEB_CLIENT_ID", "ticket-web")
	config.KeycloakAdminClientID = GetString("KEYCLOAK_ADMIN_CLIENT_ID", "")
	config.KeycloakAdminClientSecret = GetString("KEYCLOAK_ADMIN_CLIENT_SECRET", "")
	config.IdentityCacheTTL = time.Second * time.Duration(GetInt("IDENTITY_CACHE_TTL_SECONDS", 60))

	if strings.TrimSpace(config.KeycloakIssuerURL) == "" {
		return nil, errors.New("KEYCLOAK_ISSUER_URL is required")
	}
	// An audience-less verifier accepts any token the realm issues to any
	// client, which is not acceptable outside local development.
	if !isLocalOrTest(config.AppEnv) && strings.TrimSpace(config.KeycloakAudience) == "" {
		return nil, errors.New("KEYCLOAK_AUDIENCE is required outside local/test")
	}
	// http issuers are fine on localhost, never in a deployed environment:
	// tokens would cross the network in the clear.
	if !isLocalOrTest(config.AppEnv) && strings.HasPrefix(config.KeycloakIssuerURL, "http://") {
		return nil, errors.New("KEYCLOAK_ISSUER_URL must use https outside local/test")
	}
	// The dev secret is checked into the realm export, so it must never be
	// usable anywhere real.
	if !isLocalOrTest(config.AppEnv) && isWeakSecret(config.KeycloakAdminClientSecret) {
		return nil, errors.New("KEYCLOAK_ADMIN_CLIENT_SECRET must not use a known weak value outside local/test")
	}

	// Base URL of the ai-service, which the admin document-upload endpoint proxies
	// to. Service-name DNS in compose; localhost for direct local runs.
	config.AIServiceURL = GetString("AI_SERVICE_URL", "http://localhost:8081")
	return &config, nil
}

// ReachableKeycloakURL is the realm URL this process can actually connect to,
// which differs from the issuer inside Docker (the issuer is browser-facing).
// Use it for any outbound call to Keycloak; use KeycloakIssuerURL only for
// validating a token's `iss`.
func (c *Config) ReachableKeycloakURL() string {
	if s := strings.TrimSpace(c.KeycloakInternalIssuerURL); s != "" {
		return s
	}
	return c.KeycloakIssuerURL
}

// isWeakSecret reports whether a secret is a placeholder that ships in the repo.
// An empty secret is not weak here: the admin client is optional, and leaving it
// unset simply disables role writes.
func isWeakSecret(secret string) bool {
	switch strings.TrimSpace(secret) {
	case "secret", "changeme", "change-me", "local-dev-only-secret",
		"ticket-service-dev-secret", "ai-service-dev-secret":
		return true
	default:
		return false
	}
}

func isLocalOrTest(env string) bool {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "", "local", "test":
		return true
	default:
		return false
	}
}

func GetString(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return val
}

func GetBool(key string, fallback bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(val))
	if err != nil {
		return fallback
	}
	return parsed
}

func GetInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return valAsInt
}
