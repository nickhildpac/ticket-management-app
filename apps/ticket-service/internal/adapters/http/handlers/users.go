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
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
	"github.com/nickhildpac/ticket-management-app/pkg/util"
)

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	util.WriteResponse(w, http.StatusOK, struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	})
}

// @Summary		Login user
// @Description	Authenticate user with email and password
// @Tags			Authentication
// @Accept			json
// @Produce		json
// @Param			request	body		object{email=string,password=string}	true	"Login credentials"
// @Success		200		{object}	object{access_token=string,user=object{id=string,first_name=string,last_name=string,email=string,role=string}}
// @Failure		401		{object}	util.ErrorBody
// @Router			/auth/login [post]
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
		writeInternalError(w, r, err)
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

// @Summary		Logout user
// @Description	Clear refresh token cookie
// @Tags			Authentication
// @Produce		json
// @Success		202
// @Router			/auth/logout [get]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, util.GetExpiredRefreshCookie(h.config))
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	user, err := h.userService.GetUser(r.Context(), username)
	if err != nil {
		writeHandlerError(w, r, err)
		return
	}
	util.WriteResponse(w, http.StatusOK, newUserResponse(user))
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
		writeHandlerError(w, r, err)
		return
	}
	util.WriteResponse(w, http.StatusOK, newUserResponse(user))
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers(r.Context())
	if err != nil {
		writeHandlerError(w, r, err)
		return
	}

	response := make([]UserResponse, 0, len(users))
	for i := range users {
		response = append(response, newUserResponse(&users[i]))
	}

	util.WriteResponse(w, http.StatusOK, response)
}

//	@Summary		Get users for assignment
//	@Description	Get list of users with basic info for ticket assignment
//	@Tags			Users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		AssignmentUserResponse
//	@Failure		401	{object}	util.ErrorBody
//	@Router			/users [get]
//
// GetBasicUsers returns a list of users with basic info (id, name, email) for ticket assignment
// This endpoint is accessible to all authenticated users, not just admins
func (h *Handler) GetBasicUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsersForAssignment(r.Context())
	if err != nil {
		writeHandlerError(w, r, err)
		return
	}

	response := make([]AssignmentUserResponse, 0, len(users))
	for i := range users {
		response = append(response, newAssignmentUserResponse(&users[i]))
	}

	util.WriteResponse(w, http.StatusOK, response)
}

// @Summary		Get current user
// @Description	Get profile of authenticated user
// @Tags			Users
// @Produce		json
// @Security		BearerAuth
// @Success		200	{object}	UserResponse
// @Failure		401	{object}	util.ErrorBody
// @Router			/me [get]
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value(configs.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		writeHandlerError(w, r, err)
		return
	}
	util.WriteResponse(w, http.StatusOK, newUserResponse(user))
}

// @Summary		Update current user skills
// @Description	Update the authenticated user's skills (partial profile update)
// @Tags			Users
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			request	body		object{skills=[]string}	true	"Skills list (validated against allowed values)"
// @Success		200		{object}	UserResponse
// @Failure		400		{object}	util.ErrorBody
// @Failure		401		{object}	util.ErrorBody
// @Router			/me [patch]
func (h *Handler) PatchMe(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Skills []string `json:"skills"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	user, err := h.userService.UpdateMySkills(r.Context(), body.Skills)
	if err != nil {
		writeHandlerError(w, r, err)
		return
	}
	util.WriteResponse(w, http.StatusOK, newUserResponse(user))
}

// @Summary		Create new user
// @Description	Register a new user account
// @Tags			Authentication
// @Accept			json
// @Produce		json
// @Param			request	body		object{first_name=string,last_name=string,email=string,password=string,skills=[]string}	true	"User registration details"
// @Success		201		{object}	UserResponse
// @Failure		400		{object}	util.ErrorBody
// @Failure		500		{object}	util.ErrorBody
// @Router			/users [post]
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		FirstName string   `json:"first_name"`
		LastName  string   `json:"last_name"`
		Email     string   `json:"email"`
		Password  string   `json:"password"`
		Skills    []string `json:"skills"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	hashedPassword, err := util.HashPassword(requestPayload.Password)
	if err != nil {
		writeInternalError(w, r, err)
		return
	}

	// Validate skills
	skills, err := domain.NewSkills(requestPayload.Skills)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	user, err := h.userService.CreateUser(r.Context(), domain.User{
		FirstName:      requestPayload.FirstName,
		LastName:       requestPayload.LastName,
		Email:          requestPayload.Email,
		HashedPassword: hashedPassword,
		Skills:         *skills,
	})
	if err != nil {
		log.Println("Error creating user:", err)
		writeHandlerError(w, r, err)
		return
	}
	util.WriteResponse(w, http.StatusCreated, newUserResponse(user))
}

// @Summary		Refresh access token
// @Description	Get new access token using refresh token from cookie
// @Tags			Authentication
// @Accept			json
// @Produce		json
// @Success		200	{object}	object{access_token=string,user=object{id=string,first_name=string,last_name=string,email=string,role=string}}
// @Failure		401	{object}	util.ErrorBody
// @Router			/auth/refresh [get]
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
	}, jwt.WithIssuer(h.config.JWTIssuer))
	if err != nil {
		util.ErrorResponse(w, http.StatusUnauthorized, err)
		return
	}
	// Only tokens minted as refresh tokens may mint new pairs; an access token
	// presented here (same secret, also parseable) must be rejected.
	if claims.TokenType != util.RefreshTokenType {
		util.ErrorResponse(w, http.StatusUnauthorized, errors.New("not a refresh token"))
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		util.ErrorResponse(w, http.StatusUnauthorized, err)
		return
	}
	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		writeInternalError(w, r, err)
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
		writeInternalError(w, r, err)
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

// @Summary		Update user role
// @Description	Update a user's role (admin only)
// @Tags			Admin
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		string									true	"User UUID"	format(uuid)
// @Param			request	body		object{role=string}						true	"New role (user, agent, admin)"
// @Success		200		{object}	UserResponse
// @Failure		400		{object}	util.ErrorBody
// @Failure		401		{object}	util.ErrorBody
// @Failure		403		{object}	util.ErrorBody
// @Failure		500		{object}	util.ErrorBody
// @Router			/admin/users/{id}/role [put]
func (h *Handler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	var payload struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	role, err := domain.GetRole(payload.Role)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	user, err := h.userService.UpdateUserRole(r.Context(), userID, role)
	if err != nil {
		writeHandlerError(w, r, err)
		return
	}

	util.WriteResponse(w, http.StatusOK, newUserResponse(user))
}

// @Summary		Delete user
// @Description	Delete a user account (admin only)
// @Tags			Admin
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		string				true	"User UUID"	format(uuid)
// @Success		204
// @Failure		400		{object}	util.ErrorBody
// @Failure		401		{object}	util.ErrorBody
// @Failure		403		{object}	util.ErrorBody
// @Failure		500		{object}	util.ErrorBody
// @Router			/admin/users/{id} [delete]
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	err = h.userService.DeleteUser(r.Context(), userID)
	if err != nil {
		writeHandlerError(w, r, err)
		return
	}

	util.WriteResponse(w, http.StatusNoContent, nil)
}
