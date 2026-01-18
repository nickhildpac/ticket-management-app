package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
)

type ticketRepoStub struct {
	getFn    func(ctx context.Context, id uuid.UUID) (*domain.Ticket, error)
	updateFn func(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error)
}

func (s *ticketRepoStub) ListAll(ctx context.Context, limit, offset int32) ([]domain.Ticket, error) {
	return nil, nil
}
func (s *ticketRepoStub) ListByCreator(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error) {
	return nil, nil
}
func (s *ticketRepoStub) ListByAssignee(ctx context.Context, id uuid.UUID, limit, offset int32) ([]domain.Ticket, error) {
	return nil, nil
}
func (s *ticketRepoStub) Get(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	return s.getFn(ctx, id)
}
func (s *ticketRepoStub) GetByNumber(ctx context.Context, ticketNumber int64) (*domain.Ticket, error) {
	return nil, nil
}
func (s *ticketRepoStub) Create(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
	return nil, nil
}
func (s *ticketRepoStub) Update(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
	return s.updateFn(ctx, ticket)
}
func (s *ticketRepoStub) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func adminCtx() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, configs.UserIDKey, uuid.New().String())
	ctx = context.WithValue(ctx, configs.UserRoleKey, string(domain.RoleAdmin))
	return ctx
}

func TestUpdateTicket_AutoAssignValidatesTransitions(t *testing.T) {
	now := time.Now()
	ticketID := uuid.New()
	assignee := uuid.New()
	stub := &ticketRepoStub{}
	stub.getFn = func(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
		return &domain.Ticket{ID: ticketID, State: domain.TicketStateOpen, CreatedAt: now, UpdatedAt: now}, nil
	}
	stub.updateFn = func(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
		return &ticket, nil
	}

	svc := NewTicketService(stub)
	updated, err := svc.UpdateTicket(adminCtx(), domain.Ticket{
		ID:          ticketID,
		State:       domain.TicketStateOpen,
		AssignedTo:  []uuid.UUID{assignee},
		Title:       "t",
		Description: "d",
	}, []string{"assigned_to"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.State != domain.TicketStatePending {
		t.Fatalf("expected state pending after assignment, got %v", updated.State)
	}
}

func TestUpdateTicket_AutoAssignBlocksInvalidTransition(t *testing.T) {
	now := time.Now()
	ticketID := uuid.New()
	stub := &ticketRepoStub{}
	stub.getFn = func(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
		return &domain.Ticket{ID: ticketID, State: domain.TicketStateClosed, CreatedAt: now, UpdatedAt: now}, nil
	}
	stub.updateFn = func(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
		t.Fatalf("update should not be called on invalid transition")
		return nil, nil
	}

	svc := NewTicketService(stub)
	_, err := svc.UpdateTicket(adminCtx(), domain.Ticket{
		ID:          ticketID,
		State:       domain.TicketStateClosed,
		AssignedTo:  []uuid.UUID{uuid.New()},
		Title:       "t",
		Description: "d",
	}, []string{"assigned_to"})
	if err == nil {
		t.Fatalf("expected error for invalid transition, got nil")
	}
	if err != domain.ErrInvalidStatusTransition {
		t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
	}
}
