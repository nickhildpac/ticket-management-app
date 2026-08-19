package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/adapters/auth"
	"github.com/nickhildpac/ticket-management-app/internal/application/service"
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

// AuthConfigResponse tells the SPA where to authenticate.
type AuthConfigResponse struct {
	Issuer   string `json:"issuer" example:"http://localhost:8090/realms/ticket-management"`
	ClientID string `json:"client_id" example:"ticket-web"`
}

// @Summary		Get authentication configuration
// @Description	Public OIDC parameters the SPA needs to start an Authorization Code + PKCE flow
// @Tags			Authentication
// @Produce		json
// @Success		200	{object}	AuthConfigResponse
// @Router			/auth/config [get]
//
// AuthConfig is served unauthenticated and contains no secrets: the issuer URL
// and the public client id are both visible in the browser's redirect anyway.
// Serving them rather than baking them into the bundle means the SPA does not
// need rebuilding per environment.
func (h *Handler) AuthConfig(w http.ResponseWriter, r *http.Request) {
	util.WriteResponse(w, http.StatusOK, AuthConfigResponse{
		Issuer:   h.config.KeycloakIssuerURL,
		ClientID: h.config.KeycloakWebClientID,
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
// @Failure		501		{object}	util.ErrorBody
// @Router			/admin/users/{id}/role [put]
//
// Roles live in Keycloak, so this writes there first and mirrors the result
// locally. Writing only locally would look like it worked and then silently
// revert the next time the user presented a token.
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
		switch {
		case errors.Is(err, auth.ErrAdminNotConfigured):
			util.ErrorResponse(w, http.StatusNotImplemented, errors.New(
				"role management is unavailable: set KEYCLOAK_ADMIN_CLIENT_ID/SECRET, or change the role in the Keycloak console"))
		case errors.Is(err, service.ErrUserNotLinked):
			util.ErrorResponse(w, http.StatusConflict, errors.New(
				"this user has never signed in via Keycloak, so their role cannot be changed here"))
		default:
			writeHandlerError(w, r, err)
		}
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
