package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
	"github.com/nickhildpac/ticket-management-app/pkg/util"
)

type CommentPayload struct {
	TicketID    string `json:"ticket_id"`
	Description string `json:"description"`
}

// @Summary		Get ticket comments
// @Description	Retrieve all comments for a specific ticket
// @Tags			Comments
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		string				true	"Ticket UUID"	format(uuid)
// @Success		200		{array}		CommentResponse
// @Failure		400		{object}	util.ErrorBody
// @Failure		401		{object}	util.ErrorBody
// @Failure		403		{object}	util.ErrorBody
// @Failure		500		{object}	util.ErrorBody
// @Router			/tickets/{id}/comments [get]
func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	tid, err := uuid.Parse(idParam)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	comments, err := h.commentService.ListByTicket(r.Context(), tid, 10, 0)
	if err != nil {
		writeHandlerError(w, r, err)
		return
	}

	// Fetch creator details for each comment
	response := make([]CommentResponse, len(comments))
	for i, comment := range comments {
		creator, err := h.userService.GetUserByID(r.Context(), comment.CreatedBy)
		if err != nil {
			writeHandlerError(w, r, err)
			return
		}

		response[i] = CommentResponse{
			ID:          comment.ID,
			TicketID:    comment.TicketID,
			CreatedBy:   comment.CreatedBy,
			Creator:     newUserInfo(creator),
			Description: comment.Description,
			CreatedAt:   comment.CreatedAt,
		}
	}

	util.WriteResponse(w, http.StatusOK, response)
}

// @Summary		Get comment by ID
// @Description	Retrieve a specific comment
// @Tags			Comments
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		string				true	"Comment UUID"	format(uuid)
// @Success		200		{object}	CommentResponse
// @Failure		400		{object}	util.ErrorBody
// @Failure		401		{object}	util.ErrorBody
// @Failure		404		{object}	util.ErrorBody
// @Router			/comments/{id} [get]
func (h *Handler) GetComment(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	tid, err := uuid.Parse(idParam)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	comment, err := h.commentService.GetComment(r.Context(), tid)
	if err != nil {
		writeHandlerError(w, r, err)
		return
	}
	_, err = h.ticketService.GetTicket(r.Context(), comment.TicketID)
	if err != nil {
		writeHandlerError(w, r, err)
		return
	}
	util.WriteResponse(w, http.StatusOK, comment)
}

// @Summary		Create comment
// @Description	Add a new comment to a ticket
// @Tags			Comments
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			request	body		object{ticket_id=string,description=string}	true	"Comment details"
// @Success		202		{object}	CommentResponse
// @Failure		400		{object}	util.ErrorBody
// @Failure		401		{object}	util.ErrorBody
// @Failure		403		{object}	util.ErrorBody
// @Failure		500		{object}	util.ErrorBody
// @Router			/comments [post]
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	var payload CommentPayload
	userIDStr := r.Context().Value(configs.UserIDKey).(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	ticketID, err := uuid.Parse(payload.TicketID)
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	comment, err := h.commentService.CreateComment(r.Context(), domain.Comment{
		TicketID:    ticketID,
		Description: payload.Description,
		CreatedBy:   userID,
	})
	if err != nil {
		writeHandlerError(w, r, err)
		return
	}
	util.WriteResponse(w, http.StatusAccepted, comment)
}
