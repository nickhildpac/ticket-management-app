// Package util contains helper fn used for jwt-tokens
package util

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
)

type Claims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

type RefreshClaims struct {
	jwt.RegisteredClaims
}

type JWTUser struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
}
type TokenPairs struct {
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func GenerateTokenPair(conf *configs.Config, user *JWTUser) (TokenPairs, error) {
	// create a token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			Issuer:    conf.JWTIssuer,
			Audience:  jwt.ClaimStrings{conf.JWTAudience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(conf.TokenExpiry)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Role: user.Role,
	})

	// create a signed token
	signedAccessToken, err := accessToken.SignedString([]byte(conf.JWTSecret))
	if err != nil {
		return TokenPairs{}, err
	}

	// create refresh token and set claims
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(conf.RefreshExpiry)),
		},
	})
	// create signed refresh token
	signedRefreshToken, err := refreshToken.SignedString([]byte(conf.JWTSecret))
	if err != nil {
		return TokenPairs{}, err
	}
	// create token pair and populate signed token
	tokenPairs := TokenPairs{
		Token:        signedAccessToken,
		RefreshToken: signedRefreshToken,
	}
	// return toke pair
	return tokenPairs, nil
}

func GetExpiredRefreshCookie(conf *configs.Config) *http.Cookie {
	return &http.Cookie{
		Name:     conf.CookieName,
		Path:     conf.CookiePath,
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		SameSite: http.SameSiteStrictMode,
		Domain:   conf.CookieDomain,
		HttpOnly: true,
		Secure:   true,
	}
}

func GetRefreshCookie(conf *configs.Config, refreshToken string) *http.Cookie {
	return &http.Cookie{
		Name:     conf.CookieName,
		Path:     conf.CookiePath,
		Value:    refreshToken,
		Expires:  time.Now().Add(conf.RefreshExpiry),
		MaxAge:   int(conf.RefreshExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   conf.CookieDomain,
		HttpOnly: true,
		Secure:   true,
	}
}

func GetTokenFromHeaderAndVerify(conf *configs.Config, w http.ResponseWriter, r *http.Request) (*Claims, error) {
	w.Header().Add("Vary", "Authorization")
	// get auth header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("no auth header")
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 {
		return nil, errors.New("invalid auth header")
	}
	if headerParts[0] != "Bearer" {
		return nil, errors.New("invalid auth header")
	}
	token := headerParts[1]
	// declare empty claims
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		return []byte(conf.JWTSecret), nil
	})
	if err != nil {
		if strings.HasPrefix(err.Error(), "token is expired by") {
			return nil, errors.New("expired token")
		}
		return nil, err
	}

	if claims.Issuer != conf.JWTIssuer {
		return nil, errors.New("invalid issuer")
	}
	return claims, nil
}
