package handlers

import (
	"bytes"
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

func TestGetMe_ResponseOmitsHashedPassword(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	handler := newHandlerWithMocks(nil, &userServiceMock{
		getUserByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return &domain.User{
				ID:             id,
				FirstName:      "Jane",
				LastName:       "Doe",
				Email:          "jane@example.com",
				Role:           domain.RoleAdmin,
				HashedPassword: "secret-hash",
				Skills:         domain.NewSkillsFromSlice([]string{"incident-management"}),
				CreatedAt:      now,
				UpdatedAt:      now,
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
	if _, ok := body["hashed_password"]; ok {
		t.Fatalf("expected hashed_password to be omitted from response")
	}
}

func TestCreateUser_DuplicateEmailReturns409(t *testing.T) {
	handler := newHandlerWithMocks(nil, &userServiceMock{
		createUserFn: func(ctx context.Context, user domain.User) (*domain.User, error) {
			return nil, apperrors.ErrDuplicateEmail
		},
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBufferString(`{
		"first_name":"Jane",
		"last_name":"Doe",
		"email":"jane@example.com",
		"password":"password123"
	}`))

	rr := httptest.NewRecorder()
	handler.CreateUser(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rr.Code)
	}
}

func TestCreateUser_ResponseOmitsHashedPassword(t *testing.T) {
	now := time.Now()
	handler := newHandlerWithMocks(nil, &userServiceMock{
		createUserFn: func(ctx context.Context, user domain.User) (*domain.User, error) {
			user.ID = uuid.New()
			user.Role = domain.RoleUser
			user.HashedPassword = "secret-hash"
			user.CreatedAt = now
			user.UpdatedAt = now
			return &user, nil
		},
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBufferString(`{
		"first_name":"Jane",
		"last_name":"Doe",
		"email":"jane@example.com",
		"password":"password123"
	}`))

	rr := httptest.NewRecorder()
	handler.CreateUser(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if _, ok := body["hashed_password"]; ok {
		t.Fatalf("expected hashed_password to be omitted from response")
	}
}
