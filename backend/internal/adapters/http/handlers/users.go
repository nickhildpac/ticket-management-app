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
		util.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	if err = util.CheckPassword(user.HashedPassword, requestPayload.Password); err != nil {
		util.ErrorResponse(w, http.StatusUnauthorized, err)
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

	loginResponse := struct {
		AccessToken string       `json:"access_token"`
		User        util.JWTUser `json:"user"`
	}{
		AccessToken: tokenPairs.Token,
		User:        util.JWTUser{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Role:      user.Role,
		},
	}
	util.WriteResponse(w, http.StatusOK, loginResponse)
}