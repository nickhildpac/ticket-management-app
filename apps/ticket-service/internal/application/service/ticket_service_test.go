package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/application/apperrors"
	"github.com/nickhildpac/ticket-management-app/internal/application/authorization"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/pkg/configs"
)

type ticketRepoStub struct {
	getFn              func(ctx context.Context, id uuid.UUID) (*domain.Ticket, error)
	getActiveTicketsFn func(ctx context.Context) ([]domain.Ticket, error)
	createFn           func(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error)
	updateFn           func(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error)
	listAllFilteredFn  func(ctx context.Context, params domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error)
}

func (s *ticketRepoStub) ListAll(ctx context.Context, limit, offset, sortVal int32) ([]domain.Ticket, error) {
	return nil, nil
}
func (s *ticketRepoStub) ListAllFiltered(ctx context.Context, params domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error) {
	if s.listAllFilteredFn != nil {
		return s.listAllFilteredFn(ctx, params)
	}
	return nil, nil
}
func (s *ticketRepoStub) ListByCreator(ctx context.Context, id uuid.UUID, limit, offset, sortVal int32) ([]domain.Ticket, error) {
	return nil, nil
}
func (s *ticketRepoStub) ListByAssignee(ctx context.Context, id uuid.UUID, limit, offset, sortVal int32) ([]domain.Ticket, error) {
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

// The WithEvent variants delegate to the plain stubs; unit tests don't assert
// on outbox rows (that's covered by the repository integration tests).
func (s *ticketRepoStub) CreateWithEvent(ctx context.Context, ticket domain.Ticket, eventType string) (*domain.Ticket, error) {
	return s.Create(ctx, ticket)
}
func (s *ticketRepoStub) UpdateWithEvent(ctx context.Context, ticket domain.Ticket, eventType string) (*domain.Ticket, error) {
	return s.Update(ctx, ticket)
}
func (s *ticketRepoStub) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (s *ticketRepoStub) CountTicketStatsAll(ctx context.Context, currentUserID uuid.UUID) (domain.TicketListStats, error) {
	return domain.TicketListStats{}, nil
}

func (s *ticketRepoStub) CountTicketStatsByCreator(ctx context.Context, creatorID, currentUserID uuid.UUID) (domain.TicketListStats, error) {
	return domain.TicketListStats{}, nil
}

func (s *ticketRepoStub) CountTicketStatsByAssignee(ctx context.Context, assigneeID, currentUserID uuid.UUID) (domain.TicketListStats, error) {
	return domain.TicketListStats{}, nil
}

func adminCtx() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, configs.UserIDKey, uuid.New().String())
	ctx = context.WithValue(ctx, configs.UserRoleKey, string(domain.RoleAdmin))
	return ctx
}

func ctxWithRole(role domain.UserRole, userID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, configs.UserIDKey, userID.String())
	ctx = context.WithValue(ctx, configs.UserRoleKey, string(role))
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
	assignedTo := []uuid.UUID{assignee}
	updated, err := svc.UpdateTicket(adminCtx(), ticketID, domain.TicketPatch{AssignedTo: &assignedTo})
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
	assignedTo := []uuid.UUID{uuid.New()}
	_, err := svc.UpdateTicket(adminCtx(), ticketID, domain.TicketPatch{AssignedTo: &assignedTo})
	if err == nil {
		t.Fatalf("expected error for invalid transition, got nil")
	}
	if err != domain.ErrInvalidStatusTransition {
		t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
	}
}

func TestUpdateTicket_EmptyPatchReturnsBadRequest(t *testing.T) {
	stub := &ticketRepoStub{
		getFn: func(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
			t.Fatalf("repo get should not be called for an empty patch")
			return nil, nil
		},
	}

	svc := NewTicketService(stub, nil)
	_, err := svc.UpdateTicket(adminCtx(), uuid.New(), domain.TicketPatch{})
	if !errors.Is(err, apperrors.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
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
		// Auto-assignment resolves before the insert, so the create call must
		// already carry the assignment and pending state (one atomic write).
		createFn: func(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
			if len(ticket.AssignedTo) != 1 || ticket.AssignedTo[0] != agentID {
				t.Fatalf("expected assigned agent %s at create time, got %v", agentID, ticket.AssignedTo)
			}
			if ticket.State != domain.TicketStatePending {
				t.Fatalf("expected pending state at create time, got %v", ticket.State)
			}
			created := ticket
			created.ID = createdID
			created.CreatedAt = now
			return &created, nil
		},
		updateFn: func(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
			t.Fatal("no post-create update expected; assignment persists with the insert")
			return nil, nil
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

func TestCreateTicket_ProceedsUnassignedWhenAutoAssignLookupFails(t *testing.T) {
	now := time.Now()
	createdID := uuid.New()
	stub := &ticketRepoStub{
		getActiveTicketsFn: func(ctx context.Context) ([]domain.Ticket, error) {
			return nil, nil
		},
		createFn: func(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
			created := ticket
			created.ID = createdID
			created.CreatedAt = now
			return &created, nil
		},
	}

	// Candidate lookup errors must not fail creation: the ticket is created
	// unassigned in the open state.
	autoAssignmentSvc := NewAutoAssignmentService(&mockUserRepository{
		err: errors.New("database error"),
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
		t.Fatalf("expected open state, got %v", created.State)
	}
	if len(created.AssignedTo) != 0 {
		t.Fatalf("expected ticket to remain unassigned, got %v", created.AssignedTo)
	}
}

func TestListTicketsWithFilters_Admin_PassesParamsToRepo(t *testing.T) {
	creator := uuid.New()
	var got domain.ListAllTicketsByStatePriorityParams
	stub := &ticketRepoStub{
		listAllFilteredFn: func(ctx context.Context, p domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error) {
			got = p
			return nil, nil
		},
	}
	svc := NewTicketService(stub, nil)
	in := domain.ListAllTicketsByStatePriorityParams{LimitVal: 10, OffsetVal: 1}
	in.FilterCreatedBy = uuid.NullUUID{UUID: creator, Valid: true}
	_, err := svc.ListTicketsWithFilters(adminCtx(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.FilterCreatedBy.Valid || got.FilterCreatedBy.UUID != creator {
		t.Fatalf("expected created_by filter preserved, got %+v", got.FilterCreatedBy)
	}
	if got.LimitVal != 10 || got.OffsetVal != 1 {
		t.Fatalf("expected limit/offset preserved, got limit=%d offset=%d", got.LimitVal, got.OffsetVal)
	}
}

func TestListTicketsWithFilters_Agent_ForcesAssignee(t *testing.T) {
	agentID := uuid.New()
	var got domain.ListAllTicketsByStatePriorityParams
	stub := &ticketRepoStub{
		listAllFilteredFn: func(ctx context.Context, p domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error) {
			got = p
			return nil, nil
		},
	}
	svc := NewTicketService(stub, nil)
	in := domain.ListAllTicketsByStatePriorityParams{LimitVal: 20, OffsetVal: 0}
	_, err := svc.ListTicketsWithFilters(ctxWithRole(domain.RoleAgent, agentID), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.FilterAssignee.Valid || got.FilterAssignee.UUID != agentID {
		t.Fatalf("expected assignee %s, got %+v", agentID, got.FilterAssignee)
	}
}

func TestListTicketsWithFilters_Agent_RejectsCreatedBy(t *testing.T) {
	agentID := uuid.New()
	other := uuid.New()
	stub := &ticketRepoStub{
		listAllFilteredFn: func(ctx context.Context, p domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error) {
			t.Fatal("ListAllFiltered should not be called")
			return nil, nil
		},
	}
	svc := NewTicketService(stub, nil)
	in := domain.ListAllTicketsByStatePriorityParams{LimitVal: 20, OffsetVal: 0}
	in.FilterCreatedBy = uuid.NullUUID{UUID: other, Valid: true}
	_, err := svc.ListTicketsWithFilters(ctxWithRole(domain.RoleAgent, agentID), in)
	if err != authorization.ErrAccessDenied {
		t.Fatalf("expected ErrAccessDenied, got %v", err)
	}
}

func TestListTicketsWithFilters_Agent_RejectsOtherAssignee(t *testing.T) {
	agentID := uuid.New()
	other := uuid.New()
	stub := &ticketRepoStub{}
	svc := NewTicketService(stub, nil)
	in := domain.ListAllTicketsByStatePriorityParams{}
	in.FilterAssignee = uuid.NullUUID{UUID: other, Valid: true}
	_, err := svc.ListTicketsWithFilters(ctxWithRole(domain.RoleAgent, agentID), in)
	if err != authorization.ErrAccessDenied {
		t.Fatalf("expected ErrAccessDenied, got %v", err)
	}
}

func TestListTicketsWithFilters_User_ForcesCreatedBy(t *testing.T) {
	userID := uuid.New()
	var got domain.ListAllTicketsByStatePriorityParams
	stub := &ticketRepoStub{
		listAllFilteredFn: func(ctx context.Context, p domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error) {
			got = p
			return nil, nil
		},
	}
	svc := NewTicketService(stub, nil)
	in := domain.ListAllTicketsByStatePriorityParams{LimitVal: 20, OffsetVal: 0}
	_, err := svc.ListTicketsWithFilters(ctxWithRole(domain.RoleUser, userID), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.FilterCreatedBy.Valid || got.FilterCreatedBy.UUID != userID {
		t.Fatalf("expected created_by %s, got %+v", userID, got.FilterCreatedBy)
	}
}

func TestListTicketsWithFilters_User_RejectsAssignee(t *testing.T) {
	userID := uuid.New()
	stub := &ticketRepoStub{}
	svc := NewTicketService(stub, nil)
	in := domain.ListAllTicketsByStatePriorityParams{}
	in.FilterAssignee = uuid.NullUUID{UUID: uuid.New(), Valid: true}
	_, err := svc.ListTicketsWithFilters(ctxWithRole(domain.RoleUser, userID), in)
	if err != authorization.ErrAccessDenied {
		t.Fatalf("expected ErrAccessDenied, got %v", err)
	}
}

func TestListTicketsWithFilters_User_RejectsOtherCreator(t *testing.T) {
	userID := uuid.New()
	other := uuid.New()
	stub := &ticketRepoStub{}
	svc := NewTicketService(stub, nil)
	in := domain.ListAllTicketsByStatePriorityParams{}
	in.FilterCreatedBy = uuid.NullUUID{UUID: other, Valid: true}
	_, err := svc.ListTicketsWithFilters(ctxWithRole(domain.RoleUser, userID), in)
	if err != authorization.ErrAccessDenied {
		t.Fatalf("expected ErrAccessDenied, got %v", err)
	}
}
