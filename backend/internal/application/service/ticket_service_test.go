package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
)

type ticketRepoStub struct {
	getFn              func(ctx context.Context, id uuid.UUID) (*domain.Ticket, error)
	getActiveTicketsFn func(ctx context.Context) ([]domain.Ticket, error)
	createFn           func(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error)
	updateFn           func(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error)
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
func (s *ticketRepoStub) GetActiveTickets(ctx context.Context) ([]domain.Ticket, error) {
	if s.getActiveTicketsFn != nil {
		return s.getActiveTicketsFn(ctx)
	}
	return nil, nil
}
func (s *ticketRepoStub) Create(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
	if s.createFn != nil {
		return s.createFn(ctx, ticket)
	}
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

	svc := NewTicketService(stub, nil)
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

	svc := NewTicketService(stub, nil)
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

func TestCreateTicket_ReturnsPersistedAssignedTicketOnAutoAssignSuccess(t *testing.T) {
	now := time.Now()
	agentID := uuid.New()
	createdID := uuid.New()
	stub := &ticketRepoStub{
		getActiveTicketsFn: func(ctx context.Context) ([]domain.Ticket, error) {
			return nil, nil
		},
		createFn: func(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
			return &domain.Ticket{
				ID:          createdID,
				CreatedBy:   ticket.CreatedBy,
				Title:       ticket.Title,
				Description: ticket.Description,
				Skills:      ticket.Skills,
				State:       domain.TicketStateOpen,
				Priority:    domain.TicketPriorityLow,
				CreatedAt:   now,
				UpdatedAt:   ticket.UpdatedAt,
			}, nil
		},
		updateFn: func(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
			if len(ticket.AssignedTo) != 1 || ticket.AssignedTo[0] != agentID {
				t.Fatalf("expected assigned agent %s, got %v", agentID, ticket.AssignedTo)
			}
			if ticket.State != domain.TicketStatePending {
				t.Fatalf("expected pending state, got %v", ticket.State)
			}
			return &ticket, nil
		},
	}

	autoAssignmentSvc := NewAutoAssignmentService(&mockUserRepository{
		agents: []domain.User{
			{
				ID:     agentID,
				Role:   domain.RoleAgent,
				Skills: domain.NewSkillsFromSlice([]string{"incident-management"}),
			},
		},
	}, stub)

	svc := NewTicketService(stub, autoAssignmentSvc)
	created, err := svc.CreateTicket(context.Background(), domain.Ticket{
		CreatedBy:   uuid.New(),
		Title:       "Test",
		Description: "Description",
		Skills:      domain.NewSkillsFromSlice([]string{"incident-management"}),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created.State != domain.TicketStatePending {
		t.Fatalf("expected pending state, got %v", created.State)
	}
	if len(created.AssignedTo) != 1 || created.AssignedTo[0] != agentID {
		t.Fatalf("expected persisted assignment to agent %s, got %v", agentID, created.AssignedTo)
	}
}

func TestCreateTicket_ReturnsOriginalTicketWhenAutoAssignUpdateFails(t *testing.T) {
	now := time.Now()
	createdID := uuid.New()
	stub := &ticketRepoStub{
		getActiveTicketsFn: func(ctx context.Context) ([]domain.Ticket, error) {
			return nil, nil
		},
		createFn: func(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
			return &domain.Ticket{
				ID:          createdID,
				CreatedBy:   ticket.CreatedBy,
				Title:       ticket.Title,
				Description: ticket.Description,
				Skills:      ticket.Skills,
				State:       domain.TicketStateOpen,
				Priority:    domain.TicketPriorityLow,
				CreatedAt:   now,
				UpdatedAt:   ticket.UpdatedAt,
			}, nil
		},
		updateFn: func(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
			return nil, errors.New("update failed")
		},
	}

	autoAssignmentSvc := NewAutoAssignmentService(&mockUserRepository{
		agents: []domain.User{
			{
				ID:     uuid.New(),
				Role:   domain.RoleAgent,
				Skills: domain.NewSkillsFromSlice([]string{"incident-management"}),
			},
		},
	}, stub)

	svc := NewTicketService(stub, autoAssignmentSvc)
	created, err := svc.CreateTicket(context.Background(), domain.Ticket{
		CreatedBy:   uuid.New(),
		Title:       "Test",
		Description: "Description",
		Skills:      domain.NewSkillsFromSlice([]string{"incident-management"}),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created.State != domain.TicketStateOpen {
		t.Fatalf("expected original open state, got %v", created.State)
	}
	if len(created.AssignedTo) != 0 {
		t.Fatalf("expected original ticket to remain unassigned, got %v", created.AssignedTo)
	}
}
