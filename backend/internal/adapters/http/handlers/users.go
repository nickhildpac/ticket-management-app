package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/nickhildpac/ticket-management-app/configs"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/internal/usecase"
	"github.com/nickhildpac/ticket-management-app/pkg/util"
)

type UserHandler struct {
	service *usecase.UserService
	config  *configs.Config
}

type User struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func NewUserHandler(service *usecase.UserService, config *configs.Config) *UserHandler {
	return &UserHandler{
		service: service,
		config:  config,
	}
}

func (h *UserHandler) HandleLoginUser(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestPayload)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	user, err := h.service.GetUser(r.Context(), requestPayload.Username)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	if err = util.CheckPassword(user.HashedPassword, requestPayload.Password); err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	u := util.JWTUser{
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      user.Role,
	}
	tokens, err := util.GenerateTokenPair(h.config, &u)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	log.Println(tokens.RefreshToken)
	refreshCookie := util.GetRefreshCookie(h.config, tokens.RefreshToken)
	rsp := struct {
		AccessToken string       `json:"access_token"`
		User        util.JWTUser `json:"user"`
	}{
		AccessToken: tokens.Token,
		User:        u,
	}
	http.SetCookie(w, refreshCookie)
	util.WriteJSON(w, http.StatusAccepted, rsp)
}

func (h *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
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
	arg := domain.User{
		Username:       payload.Username,
		FirstName:      payload.FirstName,
		LastName:       payload.LastName,
		Email:          payload.Email,
		HashedPassword: hashedPassword,
	}
	user, err := h.service.CreateUser(r.Context(), arg)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	util.WriteJSON(w, http.StatusAccepted, user)
}
