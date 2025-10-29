package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	db "github.com/nickhildpac/ticket-management-app/internal/db/sqlc"
	"github.com/nickhildpac/ticket-management-app/pkg/util"
)

type User struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func (repo *Repository) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func (repo *Repository) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookieNotFound := true
	for _, cookie := range r.Cookies() {
		if cookie.Name == repo.Config.CookieName {
			cookieNotFound = false
			claims := &util.Claims{}
			refreshToken := cookie.Value
			// parse token to get the claims
			_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(repo.Config.JWTSecret), nil
			})
			if err != nil {
				errorResponse(w, http.StatusUnauthorized, err)
				return
			}
			// get user id from token claim
			username := claims.Subject
			user, err := repo.Store.GetUser(r.Context(), username)
			if err != nil {
				errorResponse(w, http.StatusInternalServerError, err)
				return
			}
			u := util.JWTUser{
				Username:  username,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Email:     user.Email,
				Role:      user.Role.String,
			}
			tokenPairs, err := util.GenerateTokenPair(repo.Config, &u)
			if err != nil {
				errorResponse(w, http.StatusInternalServerError, err)
				return
			}
			http.SetCookie(w, util.GetRefreshCookie(repo.Config, tokenPairs.RefreshToken))
			writeJson(w, http.StatusAccepted, struct {
				AccessToken string       `json:"access_token"`
				User        util.JWTUser `json:"user"`
			}{
				AccessToken: tokenPairs.Token,
				User:        u,
			})
		}
	}
	if cookieNotFound {
		errorResponse(w, http.StatusUnauthorized, errors.New("error generating token"))
	}
}

func (repo *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, util.GetExpiredRefreshCookie(repo.Config))
	w.WriteHeader(http.StatusAccepted)
}

func (repo *Repository) Login(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestPayload)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}
	user, err := repo.Store.GetUser(r.Context(), requestPayload.Username)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}
	if err = util.CheckPassword(user.HashedPassword, requestPayload.Password); err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}
	u := util.JWTUser{
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      user.Role.String,
	}
	tokens, err := util.GenerateTokenPair(repo.Config, &u)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}
	refreshCookie := util.GetRefreshCookie(repo.Config, tokens.RefreshToken)
	rsp := struct {
		AccessToken string       `json:"access_token"`
		User        util.JWTUser `json:"user"`
	}{
		AccessToken: tokens.Token,
		User:        u,
	}
	http.SetCookie(w, refreshCookie)
	writeJson(w, http.StatusAccepted, rsp)
}

func (repo *Repository) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	user, err := repo.Store.GetUser(r.Context(), username)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}
	writeJson(w, http.StatusOK, user)
}

func (repo *Repository) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload User
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	hashedPassword, err := util.HashPassword(payload.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	arg := db.CreateUserParams{
		Username:       payload.Username,
		FirstName:      payload.FirstName,
		LastName:       payload.LastName,
		Email:          payload.Email,
		HashedPassword: hashedPassword,
	}
	user, err := repo.Store.CreateUser(r.Context(), arg)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}
	writeJson(w, http.StatusAccepted, user)
}

func writeJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(status)
}

func errorResponse(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err = json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
