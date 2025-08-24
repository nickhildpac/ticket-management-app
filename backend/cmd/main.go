package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	db "github.com/nickhildpac/ticket-management-app/internal/db/sqlc"
	"github.com/nickhildpac/ticket-management-app/internal/env"
)

type application struct {
	ADDR         int
	Store        db.Store
	Auth         Auth
	Domain       string
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	CookieDomain string
}

func main() {
	var app application
	app.ADDR = env.GetInt("PORT", 8081)
	flag.StringVar(&app.JWTSecret, "jwt-secret", env.GetString("JWTSecret", ""), "signing secret")
	flag.StringVar(&app.JWTIssuer, "jwt-issuer", env.GetString("JWTIssuer", ""), "signing issuer")
	flag.StringVar(&app.JWTAudience, "jwt-audience", env.GetString("JWTAudience", ""), "signing audience")
	flag.StringVar(&app.CookieDomain, "cookie-domain", env.GetString("CookieDomain", ""), "cookie domain")
	flag.StringVar(&app.Domain, "domain", env.GetString("Domain", ""), "domain")
	flag.Parse()
	app.Auth = Auth{
		Issuer:        app.JWTIssuer,
		Audience:      app.JWTAudience,
		Secret:        app.JWTSecret,
		TokenExpiry:   time.Minute * time.Duration(env.GetInt("TokenExpiry", 15)),
		RefreshExpiry: time.Hour * time.Duration(env.GetInt("RefreshTokenExpiry", 24)),
		CookiePath:    env.GetString("CookiePath", ""),
		CookieName:    env.GetString("RefreshCookieName", "app-refresh_token"),
		CookieDomain:  app.CookieDomain,
	}

	conn, err := sql.Open("postgres", env.GetString("DB_ADDR", ""))
	if err != nil {
		log.Fatal("failed to connect to db ", err)
	}
	log.Println("DB connected successfully")
	app.Store = db.NewStore(conn)
	log.Printf("server is listening on port %d ", app.ADDR)
	err = http.ListenAndServe(fmt.Sprintf(":%d", app.ADDR), app.mount())
	if err != nil {
		log.Fatal(err)
	}
}
