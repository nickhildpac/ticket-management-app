package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
	"github.com/nickhildpac/ticket-management-app/pkg/util"
)

const (
	defaultCommentLimit int32 = 50
	maxCommentLimit     int32 = 200
)

type CommentPayload struct {
	TicketID    string `json:"ticket_id"`
	Description string `json:"description"`
}

// parseCommentPagination reads limit/offset from the query with sane defaults
// and bounds, so comment lists aren't silently capped (the AI worker adds
// comments, which previously pushed human comments past a hardcoded limit of 10).
func parseCommentPagination(q url.Values) (limit, offset int32, err error) {
	limit, offset = defaultCommentLimit, 0
	if v := strings.TrimSpace(q.Get("limit")); v != "" {
		l, convErr := strconv.ParseInt(v, 10, 32)
		if convErr != nil || l < 1 || l > int64(maxCommentLimit) {
			return 0, 0, fmt.Errorf("invalid limit")
		}
		limit = int32(l)
	}
	if v := strings.TrimSpace(q.Get("offset")); v != "" {
		o, convErr := strconv.ParseInt(v, 10, 32)
		if convErr != nil || o < 0 {
			return 0, 0, fmt.Errorf("invalid offset")
		}
		offset = int32(o)
	}
	return limit, offset, nil
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

	limit, offset, err := parseCommentPagination(r.URL.Query())
	if err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	comments, err := h.commentService.ListByTicket(r.Context(), tid, limit, offset)
	if err != nil {
		writeHandlerError(w, r, err)
		return
	}

	response, err := h.commentResponsesWithCreators(r.Context(), comments)
	if err != nil {
		writeHandlerError(w, r, err)
		return
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

func (h *Handler) commentResponsesWithCreators(ctx context.Context, comments []domain.Comment) ([]CommentResponse, error) {
	uniqueCreators := make(map[uuid.UUID]struct{})
	for _, comment := range comments {
		if comment.CreatedBy != uuid.Nil {
			uniqueCreators[comment.CreatedBy] = struct{}{}
		}
	}

	ids := make([]uuid.UUID, 0, len(uniqueCreators))
	for id := range uniqueCreators {
		ids = append(ids, id)
	}

	usersByID, err := h.userService.GetUsersByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	if usersByID == nil {
		usersByID = make(map[uuid.UUID]*domain.User)
	}

	response := make([]CommentResponse, 0, len(comments))
	for _, comment := range comments {
		creator := UserInfo{ID: comment.CreatedBy}
		if user, ok := usersByID[comment.CreatedBy]; ok && user != nil {
			creator = newUserInfo(user)
		}
		response = append(response, CommentResponse{
			ID:          comment.ID,
			TicketID:    comment.TicketID,
			CreatedBy:   comment.CreatedBy,
			Creator:     creator,
			Description: comment.Description,
			CreatedAt:   comment.CreatedAt,
		})
	}

	return response, nil
}
