package util

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nickhildpac/ticket-management-app/internal/config"
)

type Claims struct {
	jwt.RegisteredClaims
}

type JWTUser struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}
type TokenPairs struct {
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func GenerateTokenPair(conf *config.Config, user *JWTUser) (TokenPairs, error) {
	// create a token
	token := jwt.New(jwt.SigningMethodHS256)
	// set the clains
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	claims["sub"] = fmt.Sprint(user.Username)
	claims["aud"] = conf.JWTAudience
	claims["iss"] = conf.JWTIssuer
	claims["iat"] = time.Now().UTC().Unix()
	claims["typ"] = "JWT"

	// set expiry for JWT
	claims["exp"] = time.Now().UTC().Add(conf.TokenExpiry).Unix()

	// create a signed token
	signedAccessToken, err := token.SignedString([]byte(conf.JWTSecret))
	if err != nil {
		return TokenPairs{}, err
	}

	// create refresh token and set claims
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshTokenClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshTokenClaims["sub"] = fmt.Sprint(user.Username)
	refreshTokenClaims["iat"] = time.Now().UTC().Unix()

	refreshTokenClaims["exp"] = time.Now().UTC().Add(conf.RefreshExpiry).Unix()
	// create signed refresh token
	signedRefreshToken, err := refreshToken.SignedString([]byte(conf.JWTSecret))

	if err != nil {
		return TokenPairs{}, err
	}
	// create token pair and populate signed token
	var tokenPairs = TokenPairs{
		Token:        signedAccessToken,
		RefreshToken: signedRefreshToken,
	}
	// return toke pair
	return tokenPairs, nil

}

func GetExpiredRefreshCookie(conf *config.Config) *http.Cookie {
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
func GetRefreshCookie(conf *config.Config, refreshToken string) *http.Cookie {
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

func GetTokenFromHeaderAndVerify(conf *config.Config, w http.ResponseWriter, r *http.Request) (string, *Claims, error) {
	w.Header().Add("Vary", "Authorization")
	// get auth header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", nil, errors.New("no auth header")
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 {
		return "", nil, errors.New("invalid auth header")
	}
	if headerParts[0] != "Bearer" {
		return "", nil, errors.New("invalid auth header")
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
			return "", nil, errors.New("expired token")
		}
		return "", nil, err
	}
	username := claims.Subject

	if claims.Issuer != conf.JWTIssuer {
		return "", nil, errors.New("invalid issuer")
	}
	return username, claims, nil
}
