package config

import (
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/nickhildpac/ticket-management-app/internal/env"
)

type userContextKey string

const UsernameKey userContextKey = "username"

type Config struct {
	ADDR          int
	DSN           string
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
	config.ADDR = env.GetInt("PORT", 8081)
	config.DSN = env.GetString("DB_ADDR", "postgres://postgres:postgres@localhost/ticket_management?sslmode=disable")
	flag.StringVar(&config.JWTSecret, "jwt-secret", env.GetString("JWTSecret", "secret"), "signing secret")
	flag.StringVar(&config.JWTIssuer, "jwt-issuer", env.GetString("JWTIssuer", "example.com"), "signing issuer")
	flag.StringVar(&config.JWTAudience, "jwt-audience", env.GetString("JWTAudience", "example.com"), "signing audience")
	flag.StringVar(&config.CookieDomain, "cookie-domain", env.GetString("CookieDomain", "localhost"), "cookie domain")
	flag.StringVar(&config.Domain, "domain", env.GetString("Domain", "example.com"), "domain")
	flag.Parse()
	config.CookieName = env.GetString("RefreshCookieName", "refresh_token")
	config.CookiePath = env.GetString("CookiePath", "/")
	config.TokenExpiry = time.Minute * time.Duration(env.GetInt("TokenExpiry", 15))
	config.RefreshExpiry = time.Hour * time.Duration(env.GetInt("RefreshTokenExpiry", 24))
	return &config, nil
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
