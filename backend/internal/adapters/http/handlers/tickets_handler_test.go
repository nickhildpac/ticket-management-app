package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/application/apperrors"
	"github.com/nickhildpac/ticket-management-app/internal/application/authorization"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

type ticketServiceMock struct {
	listAllFn        func(ctx context.Context, limit, offset int32) ([]domain.Ticket, error)
	listByCreatorFn  func(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error)
	listByAssigneeFn func(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error)
	getTicketFn      func(ctx context.Context, id uuid.UUID) (*domain.Ticket, error)
	getByNumberFn    func(ctx context.Context, ticketNumber int64) (*domain.Ticket, error)
	createFn         func(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error)
	updateFn         func(ctx context.Context, ticket domain.Ticket, updatedFields []string) (*domain.Ticket, error)
	deleteFn         func(ctx context.Context, id uuid.UUID) error
}

func (m *ticketServiceMock) ListAll(ctx context.Context, limit, offset int32) ([]domain.Ticket, error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx, limit, offset)
	}
	return nil, nil
}

func (m *ticketServiceMock) ListByCreator(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error) {
	if m.listByCreatorFn != nil {
		return m.listByCreatorFn(ctx, id, limit, offset)
	}
	return nil, nil
}

func (m *ticketServiceMock) ListByAssignee(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error) {
	if m.listByAssigneeFn != nil {
		return m.listByAssigneeFn(ctx, id, limit, offset)
	}
	return nil, nil
}

func (m *ticketServiceMock) GetTicket(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	if m.getTicketFn != nil {
		return m.getTicketFn(ctx, id)
	}
	return nil, nil
}

func (m *ticketServiceMock) GetTicketByNumber(ctx context.Context, ticketNumber int64) (*domain.Ticket, error) {
	if m.getByNumberFn != nil {
		return m.getByNumberFn(ctx, ticketNumber)
	}
	return nil, nil
}

func (m *ticketServiceMock) CreateTicket(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
	if m.createFn != nil {
		return m.createFn(ctx, ticket)
	}
	return &ticket, nil
}

func (m *ticketServiceMock) UpdateTicket(ctx context.Context, ticket domain.Ticket, updatedFields []string) (*domain.Ticket, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, ticket, updatedFields)
	}
	return &ticket, nil
}

func (m *ticketServiceMock) DeleteTicket(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type userServiceMock struct {
	getUserFn               func(ctx context.Context, email string) (*domain.User, error)
	getUserByIDFn           func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	createUserFn            func(ctx context.Context, user domain.User) (*domain.User, error)
	getAllUsersFn           func(ctx context.Context) ([]domain.User, error)
	getAllUsersAssignmentFn func(ctx context.Context) ([]domain.User, error)
	updateUserRoleFn        func(ctx context.Context, id uuid.UUID, role domain.UserRole) (*domain.User, error)
	deleteUserFn            func(ctx context.Context, id uuid.UUID) error
}

func (m *userServiceMock) GetUser(ctx context.Context, email string) (*domain.User, error) {
	if m.getUserFn != nil {
		return m.getUserFn(ctx, email)
	}
	return nil, nil
}

func (m *userServiceMock) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.getUserByIDFn != nil {
		return m.getUserByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *userServiceMock) CreateUser(ctx context.Context, user domain.User) (*domain.User, error) {
	if m.createUserFn != nil {
		return m.createUserFn(ctx, user)
	}
	return &user, nil
}

func (m *userServiceMock) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	if m.getAllUsersFn != nil {
		return m.getAllUsersFn(ctx)
	}
	return nil, nil
}

func (m *userServiceMock) GetAllUsersForAssignment(ctx context.Context) ([]domain.User, error) {
	if m.getAllUsersAssignmentFn != nil {
		return m.getAllUsersAssignmentFn(ctx)
	}
	return nil, nil
}

func (m *userServiceMock) UpdateUserRole(ctx context.Context, id uuid.UUID, role domain.UserRole) (*domain.User, error) {
	if m.updateUserRoleFn != nil {
		return m.updateUserRoleFn(ctx, id, role)
	}
	return nil, nil
}

func (m *userServiceMock) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if m.deleteUserFn != nil {
		return m.deleteUserFn(ctx, id)
	}
	return nil
}

type commentServiceMock struct {
	listByTicketFn  func(ctx context.Context, ticketID uuid.UUID, limit, offset int32) ([]domain.Comment, error)
	getCommentFn    func(ctx context.Context, id uuid.UUID) (*domain.Comment, error)
	createCommentFn func(ctx context.Context, comment domain.Comment) (*domain.Comment, error)
}

func (m *commentServiceMock) ListByTicket(ctx context.Context, ticketID uuid.UUID, limit, offset int32) ([]domain.Comment, error) {
	if m.listByTicketFn != nil {
		return m.listByTicketFn(ctx, ticketID, limit, offset)
	}
	return nil, nil
}

func (m *commentServiceMock) GetComment(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	if m.getCommentFn != nil {
		return m.getCommentFn(ctx, id)
	}
	return nil, nil
}

func (m *commentServiceMock) CreateComment(ctx context.Context, comment domain.Comment) (*domain.Comment, error) {
	if m.createCommentFn != nil {
		return m.createCommentFn(ctx, comment)
	}
	return &comment, nil
}

func newHandlerWithMocks(ticketSvc *ticketServiceMock, userSvc *userServiceMock, commentSvc *commentServiceMock) *Handler {
	if ticketSvc == nil {
		ticketSvc = &ticketServiceMock{}
	}
	if userSvc == nil {
		userSvc = &userServiceMock{}
	}
	if commentSvc == nil {
		commentSvc = &commentServiceMock{}
	}
	return NewHandler(nil, userSvc, ticketSvc, commentSvc)
}

func addRouteParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestUpdateTicket_InvalidPriority(t *testing.T) {
	ticketID := uuid.New()
	handler := newHandlerWithMocks(&ticketServiceMock{
		getTicketFn: func(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
			return &domain.Ticket{ID: id}, nil
		},
	}, nil, nil)

	body := bytes.NewBufferString(`{"priority":"invalid"}`)
	req := httptest.NewRequest(http.MethodPatch, "/ticket/"+ticketID.String(), body)
	req = addRouteParam(req, "id", ticketID.String())

	rr := httptest.NewRecorder()
	handler.UpdateTicket(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestUpdateTicket_InvalidTransitionReturnsBadRequest(t *testing.T) {
	ticketID := uuid.New()
	handler := newHandlerWithMocks(&ticketServiceMock{
		getTicketFn: func(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
			return &domain.Ticket{ID: id, Title: "t", Description: "d"}, nil
		},
		updateFn: func(ctx context.Context, ticket domain.Ticket, updatedFields []string) (*domain.Ticket, error) {
			return nil, domain.ErrInvalidStatusTransition
		},
	}, nil, nil)

	req := httptest.NewRequest(http.MethodPatch, "/ticket/"+ticketID.String(), bytes.NewBufferString(`{"state":"closed"}`))
	req = addRouteParam(req, "id", ticketID.String())

	rr := httptest.NewRecorder()
	handler.UpdateTicket(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestGetTicket_NotFoundReturns404(t *testing.T) {
	ticketID := uuid.New()
	handler := newHandlerWithMocks(&ticketServiceMock{
		getTicketFn: func(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
			return nil, apperrors.ErrNotFound
		},
	}, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/ticket/"+ticketID.String(), nil)
	req = addRouteParam(req, "id", ticketID.String())

	rr := httptest.NewRecorder()
	handler.GetTicket(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestGetTicket_AccessDeniedReturns403(t *testing.T) {
	ticketID := uuid.New()
	handler := newHandlerWithMocks(&ticketServiceMock{
		getTicketFn: func(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
			return nil, authorization.ErrAccessDenied
		},
	}, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/ticket/"+ticketID.String(), nil)
	req = addRouteParam(req, "id", ticketID.String())

	rr := httptest.NewRecorder()
	handler.GetTicket(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}
