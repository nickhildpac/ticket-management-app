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
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
)

type ticketServiceMock struct {
	listAllFn                   func(ctx context.Context, limit, offset, sortVal int32) ([]domain.Ticket, error)
	listTicketsWithFiltersFn    func(ctx context.Context, params domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error)
	listAssignedToCurrentUserFn func(ctx context.Context, params domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error)
	listByCreatorFn             func(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error)
	listByAssigneeFn            func(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error)
	getTicketFn                 func(ctx context.Context, id uuid.UUID) (*domain.Ticket, error)
	getByNumberFn               func(ctx context.Context, ticketNumber int64) (*domain.Ticket, error)
	createFn                    func(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error)
	updateFn                    func(ctx context.Context, id uuid.UUID, patch domain.TicketPatch) (*domain.Ticket, error)
	deleteFn                    func(ctx context.Context, id uuid.UUID) error
	getTicketStatsFn            func(ctx context.Context) (domain.TicketListStats, error)
}

func (m *ticketServiceMock) ListAll(ctx context.Context, limit, offset, sortVal int32) ([]domain.Ticket, error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx, limit, offset, sortVal)
	}
	return nil, nil
}

func (m *ticketServiceMock) ListTicketsWithFilters(ctx context.Context, params domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error) {
	if m.listTicketsWithFiltersFn != nil {
		return m.listTicketsWithFiltersFn(ctx, params)
	}
	return nil, nil
}

func (m *ticketServiceMock) ListAssignedToCurrentUser(ctx context.Context, params domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error) {
	if m.listAssignedToCurrentUserFn != nil {
		return m.listAssignedToCurrentUserFn(ctx, params)
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

func (m *ticketServiceMock) UpdateTicket(ctx context.Context, id uuid.UUID, patch domain.TicketPatch) (*domain.Ticket, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, patch)
	}
	return &domain.Ticket{ID: id}, nil
}

func (m *ticketServiceMock) DeleteTicket(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *ticketServiceMock) GetTicketStats(ctx context.Context) (domain.TicketListStats, error) {
	if m.getTicketStatsFn != nil {
		return m.getTicketStatsFn(ctx)
	}
	return domain.TicketListStats{}, nil
}

type userServiceMock struct {
	getUserFn               func(ctx context.Context, email string) (*domain.User, error)
	getUserByIDFn           func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	getUsersByIDsFn         func(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.User, error)
	getAllUsersFn           func(ctx context.Context) ([]domain.User, error)
	getAllUsersAssignmentFn func(ctx context.Context) ([]domain.User, error)
	updateUserRoleFn        func(ctx context.Context, id uuid.UUID, role domain.UserRole) (*domain.User, error)
	deleteUserFn            func(ctx context.Context, id uuid.UUID) error
	updateMySkillsFn        func(ctx context.Context, skills []string) (*domain.User, error)
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

func (m *userServiceMock) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.User, error) {
	if m.getUsersByIDsFn != nil {
		return m.getUsersByIDsFn(ctx, ids)
	}
	out := make(map[uuid.UUID]*domain.User)
	if m.getUserByIDFn != nil {
		for _, id := range ids {
			if u, err := m.getUserByIDFn(ctx, id); err == nil && u != nil {
				out[id] = u
			}
		}
	}
	return out, nil
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

func (m *userServiceMock) UpdateMySkills(ctx context.Context, skills []string) (*domain.User, error) {
	if m.updateMySkillsFn != nil {
		return m.updateMySkillsFn(ctx, skills)
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

func withAuthContext(req *http.Request, userID uuid.UUID, role domain.UserRole) *http.Request {
	ctx := req.Context()
	ctx = context.WithValue(ctx, configs.UserIDKey, userID.String())
	ctx = context.WithValue(ctx, configs.UserRoleKey, string(role))
	return req.WithContext(ctx)
}

func TestUpdateTicket_InvalidPriority(t *testing.T) {
	ticketID := uuid.New()
	handler := newHandlerWithMocks(&ticketServiceMock{}, nil, nil)

	body := bytes.NewBufferString(`{"priority":"invalid"}`)
	req := httptest.NewRequest(http.MethodPatch, "/tickets/"+ticketID.String(), body)
	req = addRouteParam(req, "id", ticketID.String())

	rr := httptest.NewRecorder()
	handler.UpdateTicket(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestUpdateTicket_NotFoundReturns404(t *testing.T) {
	ticketID := uuid.New()
	handler := newHandlerWithMocks(&ticketServiceMock{
		updateFn: func(ctx context.Context, id uuid.UUID, patch domain.TicketPatch) (*domain.Ticket, error) {
			return nil, apperrors.ErrNotFound
		},
	}, nil, nil)

	req := httptest.NewRequest(http.MethodPatch, "/tickets/"+ticketID.String(), bytes.NewBufferString(`{"title":"x"}`))
	req = addRouteParam(req, "id", ticketID.String())

	rr := httptest.NewRecorder()
	handler.UpdateTicket(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestUpdateTicket_InvalidTransitionReturnsBadRequest(t *testing.T) {
	ticketID := uuid.New()
	handler := newHandlerWithMocks(&ticketServiceMock{
		updateFn: func(ctx context.Context, id uuid.UUID, patch domain.TicketPatch) (*domain.Ticket, error) {
			return nil, domain.ErrInvalidStatusTransition
		},
	}, nil, nil)

	req := httptest.NewRequest(http.MethodPatch, "/tickets/"+ticketID.String(), bytes.NewBufferString(`{"state":"closed"}`))
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

	req := httptest.NewRequest(http.MethodGet, "/tickets/"+ticketID.String(), nil)
	req = addRouteParam(req, "id", ticketID.String())

	rr := httptest.NewRecorder()
	handler.GetTicket(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestGetTickets_InvalidPriorityReturns400(t *testing.T) {
	handler := newHandlerWithMocks(&ticketServiceMock{}, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/tickets?priority=not-a-priority", nil)
	req = withAuthContext(req, uuid.New(), domain.RoleAdmin)

	rr := httptest.NewRecorder()
	handler.GetTickets(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestGetTickets_WithFilterUsesListTicketsWithFilters(t *testing.T) {
	var gotParams domain.ListAllTicketsByStatePriorityParams
	handler := newHandlerWithMocks(&ticketServiceMock{
		listTicketsWithFiltersFn: func(ctx context.Context, params domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error) {
			gotParams = params
			return nil, nil
		},
	}, &userServiceMock{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/tickets?state=open&limit=5&offset=2", nil)
	req = withAuthContext(req, uuid.New(), domain.RoleAdmin)

	rr := httptest.NewRecorder()
	handler.GetTickets(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !gotParams.FilterState.Valid || gotParams.FilterState.Int32 != int32(domain.TicketStateOpen) {
		t.Fatalf("expected open state filter, got %+v", gotParams.FilterState)
	}
	if gotParams.LimitVal != 5 || gotParams.OffsetVal != 2 {
		t.Fatalf("expected limit 5 offset 2, got limit=%d offset=%d", gotParams.LimitVal, gotParams.OffsetVal)
	}
}

func TestGetTickets_NoFilterUsesListAllWithPagination(t *testing.T) {
	var gotLimit, gotOffset, gotSort int32
	handler := newHandlerWithMocks(&ticketServiceMock{
		listAllFn: func(ctx context.Context, limit, offset, sortVal int32) ([]domain.Ticket, error) {
			gotLimit, gotOffset, gotSort = limit, offset, sortVal
			return nil, nil
		},
	}, &userServiceMock{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/tickets?limit=10&offset=3", nil)
	req = withAuthContext(req, uuid.New(), domain.RoleAdmin)

	rr := httptest.NewRecorder()
	handler.GetTickets(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if gotLimit != 10 || gotOffset != 3 {
		t.Fatalf("expected limit 10 offset 3, got limit=%d offset=%d", gotLimit, gotOffset)
	}
	if gotSort != domain.TicketListSortCreatedDesc {
		t.Fatalf("expected default sort created_at desc (%d), got %d", domain.TicketListSortCreatedDesc, gotSort)
	}
}

func TestGetTickets_AssignedToMeUsesAssignedList(t *testing.T) {
	var called bool
	var gotParams domain.ListAllTicketsByStatePriorityParams
	handler := newHandlerWithMocks(&ticketServiceMock{
		listAssignedToCurrentUserFn: func(ctx context.Context, params domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error) {
			called = true
			gotParams = params
			return nil, nil
		},
	}, &userServiceMock{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/tickets?assigned_to=me&state=open&limit=7&offset=1", nil)
	req = withAuthContext(req, uuid.New(), domain.RoleAgent)

	rr := httptest.NewRecorder()
	handler.GetTickets(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !called {
		t.Fatal("expected assigned-ticket list to be called")
	}
	if !gotParams.FilterState.Valid || gotParams.FilterState.Int32 != int32(domain.TicketStateOpen) {
		t.Fatalf("expected open state filter, got %+v", gotParams.FilterState)
	}
	if gotParams.LimitVal != 7 || gotParams.OffsetVal != 1 {
		t.Fatalf("expected limit 7 offset 1, got limit=%d offset=%d", gotParams.LimitVal, gotParams.OffsetVal)
	}
	if gotParams.FilterAssignee.Valid {
		t.Fatalf("expected assigned_to=me to be handled by service scope, got assignee filter %+v", gotParams.FilterAssignee)
	}
}

func TestGetTicket_AccessDeniedReturns403(t *testing.T) {
	ticketID := uuid.New()
	handler := newHandlerWithMocks(&ticketServiceMock{
		getTicketFn: func(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
			return nil, authorization.ErrAccessDenied
		},
	}, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/tickets/"+ticketID.String(), nil)
	req = addRouteParam(req, "id", ticketID.String())

	rr := httptest.NewRecorder()
	handler.GetTicket(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}
