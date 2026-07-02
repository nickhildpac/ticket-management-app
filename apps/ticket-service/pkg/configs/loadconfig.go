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
	ADDR          int
	DSN           string
	AppEnv        string
	JWTIssuer     string
	JWTAudience   string
	JWTSecret     string
	Domain        string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
	CookieDomain  string
	CookiePath    string
	CookieName    string
}

func LoadConfig() (*Config, error) {
	var config Config
	config.ADDR = GetInt("PORT", 8080)
	config.DSN = GetString("DB_ADDR", "postgres://postgres:postgres@localhost/ticket_management?sslmode=disable")
	config.AppEnv = strings.ToLower(GetString("APP_ENV", GetString("ENV", "local")))
	jwtSecret := GetString("JWT_SECRET", GetString("JWTSecret", ""))
	if jwtSecret == "" && isLocalOrTest(config.AppEnv) {
		jwtSecret = "local-dev-only-secret"
	}
	if !isLocalOrTest(config.AppEnv) && isWeakJWTSecret(jwtSecret) {
		return nil, errors.New("JWT_SECRET is required and must not use a known weak value outside local/test")
	}
	flag.StringVar(&config.JWTSecret, "jwt-secret", jwtSecret, "signing secret")
	flag.StringVar(&config.JWTIssuer, "jwt-issuer", GetString("JWTIssuer", "example.com"), "signing issuer")
	flag.StringVar(&config.JWTAudience, "jwt-audience", GetString("JWTAudience", "example.com"), "signing audience")
	flag.StringVar(&config.CookieDomain, "cookie-domain", GetString("CookieDomain", "localhost"), "cookie domain")
	flag.StringVar(&config.Domain, "domain", GetString("Domain", "example.com"), "domain")
	flag.Parse()
	config.CookieName = GetString("RefreshCookieName", "tapp-refresh_token")
	config.CookiePath = GetString("CookiePath", "/")
	config.TokenExpiry = time.Minute * time.Duration(GetInt("TokenExpiry", 15))
	config.RefreshExpiry = time.Hour * time.Duration(GetInt("RefreshTokenExpiry", 24))
	if !isLocalOrTest(config.AppEnv) && isWeakJWTSecret(config.JWTSecret) {
		return nil, errors.New("JWT secret flag must not be empty or weak outside local/test")
	}
	return &config, nil
}

func isLocalOrTest(env string) bool {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "", "local", "test":
		return true
	default:
		return false
	}
}

func isWeakJWTSecret(secret string) bool {
	switch strings.TrimSpace(secret) {
	case "", "secret", "changeme", "change-me", "local-dev-only-secret":
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
