package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

type ticketServiceMock struct {
	ticket       domain.Ticket
	updateCalled bool
}

func (m *ticketServiceMock) ListAll(ctx context.Context, limit, offset int32) ([]domain.Ticket, error) {
	return nil, nil
}
func (m *ticketServiceMock) ListByCreator(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error) {
	return nil, nil
}
func (m *ticketServiceMock) ListByAssignee(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error) {
	return nil, nil
}
func (m *ticketServiceMock) GetTicket(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	return &m.ticket, nil
}
func (m *ticketServiceMock) CreateTicket(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
	return &ticket, nil
}
func (m *ticketServiceMock) UpdateTicket(ctx context.Context, ticket domain.Ticket, updatedFields []string) (*domain.Ticket, error) {
	m.updateCalled = true
	return &ticket, nil
}
func (m *ticketServiceMock) DeleteTicket(ctx context.Context, id uuid.UUID) error {
	return nil
}

func newHandlerWithTicketMock(t *testing.T, ticket domain.Ticket) (*Handler, *ticketServiceMock) {
	t.Helper()
	ticketSvc := &ticketServiceMock{ticket: ticket}
	userSvc := &noopUserService{}
	commentSvc := &noopCommentService{}
	return NewHandler(nil, userSvc, ticketSvc, commentSvc), ticketSvc
}

type noopUserService struct{}

type noopCommentService struct{}

func (s *noopUserService) GetUser(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
func (s *noopUserService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return nil, nil
}
func (s *noopUserService) CreateUser(ctx context.Context, user domain.User) (*domain.User, error) {
	return nil, nil
}
func (s *noopUserService) GetAllUsers(ctx context.Context) ([]domain.User, error) { return nil, nil }
func (s *noopUserService) GetAllUsersForAssignment(ctx context.Context) ([]domain.User, error) {
	return nil, nil
}
func (s *noopUserService) UpdateUserRole(ctx context.Context, id uuid.UUID, role domain.UserRole) (*domain.User, error) {
	return nil, nil
}
func (s *noopUserService) DeleteUser(ctx context.Context, id uuid.UUID) error { return nil }

func (s *noopCommentService) ListByTicket(ctx context.Context, ticketID uuid.UUID, limit, offset int32) ([]domain.Comment, error) {
	return nil, nil
}
func (s *noopCommentService) GetComment(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	return nil, nil
}
func (s *noopCommentService) CreateComment(ctx context.Context, comment domain.Comment) (*domain.Comment, error) {
	return nil, nil
}

func TestUpdateTicket_InvalidPriority(t *testing.T) {
	ticketID := uuid.New()
	handler, ticketSvc := newHandlerWithTicketMock(t, domain.Ticket{ID: ticketID})

	body := bytes.NewBufferString(`{"priority":"invalid"}`)
	req := httptest.NewRequest(http.MethodPatch, "/ticket/"+ticketID.String(), body)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", ticketID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.UpdateTicket(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, status)
	}
	if ticketSvc.updateCalled {
		t.Fatalf("update should not be called for invalid priority")
	}
}
