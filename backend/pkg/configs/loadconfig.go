// Package config loads env variables
package configs

import (
	"flag"
	"os"
	"strconv"
	"time"
)

type userContextKey string

const UserIDKey userContextKey = "user_id"
const UserRoleKey userContextKey = "user_role"

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
	config.ADDR = GetInt("PORT", 8081)
	config.DSN = GetString("DB_ADDR", "postgres://postgres:postgres@localhost/ticket_management?sslmode=disable")
	jwtSecret := GetString("JWT_SECRET", GetString("JWTSecret", "secret"))
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
