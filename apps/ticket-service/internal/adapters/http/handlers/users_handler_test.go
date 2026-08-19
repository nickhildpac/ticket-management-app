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
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
)

func TestGetMe_NotFoundReturns404(t *testing.T) {
	userID := uuid.New()
	handler := newHandlerWithMocks(nil, &userServiceMock{
		getUserByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return nil, apperrors.ErrNotFound
		},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	ctx := context.WithValue(req.Context(), configs.UserIDKey, userID.String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.GetMe(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestGetMe_ResponseExposesNoCredentialFields(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	handler := newHandlerWithMocks(nil, &userServiceMock{
		getUserByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return &domain.User{
				ID:        id,
				FirstName: "Jane",
				LastName:  "Doe",
				Email:     "jane@example.com",
				Role:      domain.RoleAdmin,
				Skills:    domain.NewSkillsFromSlice([]string{"incident-management"}),
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	ctx := context.WithValue(req.Context(), configs.UserIDKey, userID.String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.GetMe(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	for _, field := range []string{"hashed_password", "password", "keycloak_id"} {
		if _, ok := body[field]; ok {
			t.Fatalf("expected %q to be absent from the response", field)
		}
	}
}
