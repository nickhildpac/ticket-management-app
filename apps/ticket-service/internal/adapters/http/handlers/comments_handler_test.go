package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/application/apperrors"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

func TestGetComment_NotFoundReturns404(t *testing.T) {
	commentID := uuid.New()
	handler := newHandlerWithMocks(nil, nil, &commentServiceMock{
		getCommentFn: func(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
			return nil, apperrors.ErrNotFound
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/comments/"+commentID.String(), nil)
	req = addRouteParam(req, "id", commentID.String())

	rr := httptest.NewRecorder()
	handler.GetComment(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestGetComments_BatchesCreatorLookup(t *testing.T) {
	ticketID := uuid.New()
	creatorOne := uuid.New()
	creatorTwo := uuid.New()
	now := time.Now()

	var gotIDs []uuid.UUID
	handler := newHandlerWithMocks(nil, &userServiceMock{
		getUserByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			t.Fatalf("GetUserByID should not be called when batched lookup is available")
			return nil, nil
		},
		getUsersByIDsFn: func(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.User, error) {
			gotIDs = append([]uuid.UUID(nil), ids...)
			return map[uuid.UUID]*domain.User{
				creatorOne: {
					ID:        creatorOne,
					FirstName: "Ada",
					LastName:  "Lovelace",
					Email:     "ada@example.com",
				},
				creatorTwo: {
					ID:        creatorTwo,
					FirstName: "Grace",
					LastName:  "Hopper",
					Email:     "grace@example.com",
				},
			}, nil
		},
	}, &commentServiceMock{
		listByTicketFn: func(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Comment, error) {
			if id != ticketID {
				t.Fatalf("expected ticket id %s, got %s", ticketID, id)
			}
			return []domain.Comment{
				{ID: uuid.New(), TicketID: ticketID, CreatedBy: creatorOne, Description: "first", CreatedAt: now},
				{ID: uuid.New(), TicketID: ticketID, CreatedBy: creatorOne, Description: "second", CreatedAt: now},
				{ID: uuid.New(), TicketID: ticketID, CreatedBy: creatorTwo, Description: "third", CreatedAt: now},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/tickets/"+ticketID.String()+"/comments", nil)
	req = addRouteParam(req, "id", ticketID.String())
	rr := httptest.NewRecorder()

	handler.GetComments(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
	if len(gotIDs) != 2 {
		t.Fatalf("expected 2 unique creator ids, got %d: %v", len(gotIDs), gotIDs)
	}
	gotSet := map[uuid.UUID]bool{}
	for _, id := range gotIDs {
		gotSet[id] = true
	}
	if !gotSet[creatorOne] || !gotSet[creatorTwo] {
		t.Fatalf("expected creator ids %s and %s, got %v", creatorOne, creatorTwo, gotIDs)
	}

	var body []CommentResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body) != 3 {
		t.Fatalf("expected 3 comments, got %d", len(body))
	}
	if body[0].Creator.Email != "ada@example.com" || body[2].Creator.Email != "grace@example.com" {
		t.Fatalf("expected creator details from batched lookup, got %+v", body)
	}
}
