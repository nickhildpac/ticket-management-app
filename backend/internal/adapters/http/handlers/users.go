package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/pkg/util"
)

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	util.WriteResponse(w, http.StatusOK, struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	user, err := h.userService.GetUser(r.Context(), requestPayload.Email)
	if err != nil {
		util.ErrorResponse(w, http.StatusUnauthorized, errors.New("invalid credentials"))
		return
	}

	if err = util.CheckPassword(user.HashedPassword, requestPayload.Password); err != nil {
		util.ErrorResponse(w, http.StatusUnauthorized, errors.New("invalid credentials"))
		return
	}

	tokenPairs, err := util.GenerateTokenPair(h.config, &util.JWTUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      user.Role,
	})
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	http.SetCookie(w, util.GetRefreshCookie(h.config, tokenPairs.RefreshToken))
	util.WriteResponse(w, http.StatusOK, struct {
		AccessToken string       `json:"access_token"`
		User        util.JWTUser `json:"user"`
	}{
		AccessToken: tokenPairs.Token,
		User: util.JWTUser{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Role:      user.Role,
		},
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, util.GetExpiredRefreshCookie(h.config))
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	user, err := h.userService.GetUser(r.Context(), username)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	util.WriteResponse(w, http.StatusOK, user)
}

func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	util.WriteResponse(w, http.StatusOK, user)
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers(r.Context())
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	util.WriteResponse(w, http.StatusOK, users)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	hashedPassword, err := util.HashPassword(requestPayload.Password)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	user, err := h.userService.CreateUser(r.Context(), domain.User{
		FirstName:      requestPayload.FirstName,
		LastName:       requestPayload.LastName,
		Email:          requestPayload.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		log.Println("Error creating user:", err)
		return
	}
	util.WriteResponse(w, http.StatusCreated, user)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshCookie, err := r.Cookie(h.config.CookieName)
	if err != nil {
		util.ErrorResponse(w, http.StatusUnauthorized, errors.New("missing refresh cookie"))
		return
	}

	claims := &util.RefreshClaims{}
	refreshToken := refreshCookie.Value
	_, err = jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(h.config.JWTSecret), nil
	})
	if err != nil {
		util.ErrorResponse(w, http.StatusUnauthorized, err)
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		util.ErrorResponse(w, http.StatusUnauthorized, err)
		return
	}
	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	tokens, err := util.GenerateTokenPair(h.config, &util.JWTUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      user.Role,
	})
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	http.SetCookie(w, util.GetRefreshCookie(h.config, tokens.RefreshToken))
	util.WriteResponse(w, http.StatusOK, struct {
		AccessToken string       `json:"access_token"`
		User        util.JWTUser `json:"user"`
	}{
		AccessToken: tokens.Token,
		User: util.JWTUser{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Role:      user.Role,
		},
	})
}
