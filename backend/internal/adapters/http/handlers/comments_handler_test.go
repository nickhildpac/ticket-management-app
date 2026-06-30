package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
