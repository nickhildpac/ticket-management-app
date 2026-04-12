package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
)

// mockUserRepository is a mock implementation of UserRepository
type mockUserRepository struct {
	agents []domain.User
	err    error
}

func (m *mockUserRepository) GetUser(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepository) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.User, error) {
	return map[uuid.UUID]*domain.User{}, nil
}

func (m *mockUserRepository) CreateUser(ctx context.Context, user domain.User) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepository) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	return nil, nil
}

func (m *mockUserRepository) GetAllAgents(ctx context.Context) ([]domain.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.agents, nil
}

func (m *mockUserRepository) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return nil
}

// mockTicketRepository is a mock implementation of TicketRepository
type mockTicketRepository struct {
	activeTickets []domain.Ticket
	err           error
}

func (m *mockTicketRepository) ListAll(ctx context.Context, limit, offset, sortVal int32) ([]domain.Ticket, error) {
	return nil, nil
}

func (m *mockTicketRepository) ListAllFiltered(ctx context.Context, params domain.ListAllTicketsByStatePriorityParams) ([]domain.Ticket, error) {
	return nil, nil
}

func (m *mockTicketRepository) ListByCreator(ctx context.Context, id uuid.UUID, limit, offset, sortVal int32) ([]domain.Ticket, error) {
	return nil, nil
}

func (m *mockTicketRepository) ListByAssignee(ctx context.Context, id uuid.UUID, limit, offset, sortVal int32) ([]domain.Ticket, error) {
	return nil, nil
}

func (m *mockTicketRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	return nil, nil
}

func (m *mockTicketRepository) GetByNumber(ctx context.Context, ticketNumber int64) (*domain.Ticket, error) {
	return nil, nil
}

func (m *mockTicketRepository) GetActiveTickets(ctx context.Context) ([]domain.Ticket, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.activeTickets, nil
}

func (m *mockTicketRepository) Create(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
	return nil, nil
}

func (m *mockTicketRepository) Update(ctx context.Context, ticket domain.Ticket) (*domain.Ticket, error) {
	return nil, nil
}

func (m *mockTicketRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockTicketRepository) CountTicketStatsAll(ctx context.Context, currentUserID uuid.UUID) (domain.TicketListStats, error) {
	return domain.TicketListStats{}, nil
}

func (m *mockTicketRepository) CountTicketStatsByCreator(ctx context.Context, creatorID, currentUserID uuid.UUID) (domain.TicketListStats, error) {
	return domain.TicketListStats{}, nil
}

func (m *mockTicketRepository) CountTicketStatsByAssignee(ctx context.Context, assigneeID, currentUserID uuid.UUID) (domain.TicketListStats, error) {
	return domain.TicketListStats{}, nil
}

func TestFindBestAgentForTicket_Success(t *testing.T) {
	ctx := context.Background()

	agentID1 := uuid.New()
	agentID2 := uuid.New()
	ticketID := uuid.New()

	userRepo := &mockUserRepository{
		agents: []domain.User{
			{
				ID:        agentID1,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Role:      domain.RoleAgent,
				Skills:    domain.NewSkillsFromSlice([]string{"incident-management", "log-analysis"}),
			},
			{
				ID:        agentID2,
				FirstName: "Jane",
				LastName:  "Smith",
				Email:     "jane@example.com",
				Role:      domain.RoleAgent,
				Skills:    domain.NewSkillsFromSlice([]string{"incident-management"}),
			},
		},
	}

	// Create an active ticket assigned to agentID2 to increase workload
	activeTicketID := uuid.New()
	ticketRepo := &mockTicketRepository{
		activeTickets: []domain.Ticket{
			{
				ID:         activeTicketID,
				AssignedTo: []uuid.UUID{agentID2},
				State:      domain.TicketStateOpen,
			},
		},
	}

	ticket := &domain.Ticket{
		ID:          ticketID,
		Title:       "Test Ticket",
		Description: "Test Description",
		Skills:      domain.NewSkillsFromSlice([]string{"incident-management", "log-analysis"}),
	}

	svc := NewAutoAssignmentService(userRepo, ticketRepo)
	bestAgent, err := svc.FindBestAgentForTicket(ctx, ticket)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if bestAgent == nil {
		t.Fatalf("expected best agent, got nil")
	}
	if bestAgent.ID != agentID1 {
		t.Fatalf("expected agent1 (perfect match, no workload), got %v", bestAgent.ID)
	}
}

func TestFindBestAgentForTicket_NoAgents(t *testing.T) {
	ctx := context.Background()

	ticketID := uuid.New()

	userRepo := &mockUserRepository{
		agents: []domain.User{},
	}

	ticketRepo := &mockTicketRepository{
		activeTickets: []domain.Ticket{},
	}

	ticket := &domain.Ticket{
		ID:          ticketID,
		Title:       "Test Ticket",
		Description: "Test Description",
		Skills:      domain.NewSkillsFromSlice([]string{"incident-management"}),
	}

	svc := NewAutoAssignmentService(userRepo, ticketRepo)
	bestAgent, err := svc.FindBestAgentForTicket(ctx, ticket)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if bestAgent != nil {
		t.Fatalf("expected nil (no agents), got %v", bestAgent)
	}
}

func TestFindBestAgentForTicket_NoSkillMatches(t *testing.T) {
	ctx := context.Background()

	agentID := uuid.New()
	ticketID := uuid.New()

	userRepo := &mockUserRepository{
		agents: []domain.User{
			{
				ID:        agentID,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Role:      domain.RoleAgent,
				Skills:    domain.NewSkillsFromSlice([]string{"production-support"}),
			},
		},
	}

	ticketRepo := &mockTicketRepository{
		activeTickets: []domain.Ticket{},
	}

	ticket := &domain.Ticket{
		ID:          ticketID,
		Title:       "Test Ticket",
		Description: "Test Description",
		Skills:      domain.NewSkillsFromSlice([]string{"incident-management", "log-analysis"}),
	}

	svc := NewAutoAssignmentService(userRepo, ticketRepo)
	bestAgent, err := svc.FindBestAgentForTicket(ctx, ticket)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if bestAgent != nil {
		t.Fatalf("expected nil (no skill matches), got %v", bestAgent)
	}
}

func TestFindBestAgentForTicket_TicketNoSkills(t *testing.T) {
	ctx := context.Background()

	agentID := uuid.New()
	ticketID := uuid.New()

	userRepo := &mockUserRepository{
		agents: []domain.User{
			{
				ID:        agentID,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Role:      domain.RoleAgent,
				Skills:    domain.NewSkillsFromSlice([]string{"incident-management"}),
			},
		},
	}

	ticketRepo := &mockTicketRepository{
		activeTickets: []domain.Ticket{},
	}

	ticket := &domain.Ticket{
		ID:          ticketID,
		Title:       "Test Ticket",
		Description: "Test Description",
		Skills:      domain.NewSkillsFromSlice([]string{}),
	}

	svc := NewAutoAssignmentService(userRepo, ticketRepo)
	bestAgent, err := svc.FindBestAgentForTicket(ctx, ticket)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if bestAgent != nil {
		t.Fatalf("expected nil (ticket has no skills), got %v", bestAgent)
	}
}

func TestFindBestAgentForTicket_RepositoryError(t *testing.T) {
	ctx := context.Background()

	ticketID := uuid.New()

	userRepo := &mockUserRepository{
		err: fmt.Errorf("database error"),
	}

	ticketRepo := &mockTicketRepository{}

	ticket := &domain.Ticket{
		ID:          ticketID,
		Title:       "Test Ticket",
		Description: "Test Description",
		Skills:      domain.NewSkillsFromSlice([]string{"incident-management"}),
	}

	svc := NewAutoAssignmentService(userRepo, ticketRepo)
	bestAgent, err := svc.FindBestAgentForTicket(ctx, ticket)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if bestAgent != nil {
		t.Fatalf("expected nil agent on error, got %v", bestAgent)
	}
}
