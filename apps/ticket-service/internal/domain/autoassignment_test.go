package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestCalculateAgentScores_PerfectMatchLowWorkload(t *testing.T) {
	// Agent with all required skills, 0 tickets
	agentID1 := uuid.New()
	agentID2 := uuid.New()
	ticketID := uuid.New()

	agents := []*User{
		{
			ID:        agentID1,
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Role:      RoleAgent,
			Skills:    MustNewSkills([]string{"incident-management", "log-analysis"}),
		},
		{
			ID:        agentID2,
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane@example.com",
			Role:      RoleAgent,
			Skills:    MustNewSkills([]string{"incident-management"}),
		},
	}

	ticket := &Ticket{
		ID:          ticketID,
		Title:       "Test Ticket",
		Description: "Test Description",
		Skills:      MustNewSkills([]string{"incident-management", "log-analysis"}),
	}

	ticketCounts := map[uuid.UUID]int{
		agentID1: 0,
		agentID2: 0,
	}

	scores := CalculateAgentScores(ticket, agents, ticketCounts)

	if len(scores) != 2 {
		t.Fatalf("expected 2 scores, got %d", len(scores))
	}

	// Agent 1 should have perfect score (70 skills + 30 workload = 100)
	if scores[0].User.ID != agentID1 {
		t.Fatalf("expected agent1 to be first, got %v", scores[0].User.ID)
	}
	if scores[0].Skills != 70.0 {
		t.Fatalf("expected skills score of 70, got %f", scores[0].Skills)
	}
	if scores[0].Workload != 30.0 {
		t.Fatalf("expected workload score of 30, got %f", scores[0].Workload)
	}
	if scores[0].Total != 100.0 {
		t.Fatalf("expected total score of 100, got %f", scores[0].Total)
	}

	// Agent 2 should have 50% match (35 skills + 30 workload = 65)
	if scores[1].User.ID != agentID2 {
		t.Fatalf("expected agent2 to be second, got %v", scores[1].User.ID)
	}
	if scores[1].Skills != 35.0 {
		t.Fatalf("expected skills score of 35, got %f", scores[1].Skills)
	}
	if scores[1].Total != 65.0 {
		t.Fatalf("expected total score of 65, got %f", scores[1].Total)
	}
}

func TestCalculateAgentScores_PartialMatchHighWorkload(t *testing.T) {
	// Agent with 50% skills, 5 tickets
	agentID := uuid.New()
	ticketID := uuid.New()

	agents := []*User{
		{
			ID:        agentID,
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Role:      RoleAgent,
			Skills:    MustNewSkills([]string{"incident-management"}),
		},
	}

	ticket := &Ticket{
		ID:          ticketID,
		Title:       "Test Ticket",
		Description: "Test Description",
		Skills:      MustNewSkills([]string{"incident-management", "log-analysis"}),
	}

	ticketCounts := map[uuid.UUID]int{
		agentID: 5,
	}

	scores := CalculateAgentScores(ticket, agents, ticketCounts)

	if len(scores) != 1 {
		t.Fatalf("expected 1 score, got %d", len(scores))
	}

	// Skills: (1/2) * 70 = 35
	// Workload: (1/6) * 30 = 5
	// Total: 35 + 5 = 40
	if scores[0].Skills != 35.0 {
		t.Fatalf("expected skills score of 35, got %f", scores[0].Skills)
	}
	if scores[0].Workload != 5.0 {
		t.Fatalf("expected workload score of 5, got %f", scores[0].Workload)
	}
	if scores[0].Total != 40.0 {
		t.Fatalf("expected total score of 40, got %f", scores[0].Total)
	}
}

func TestCalculateAgentScores_NoSkillMatch(t *testing.T) {
	// Agent with 0 matching skills
	agentID := uuid.New()
	ticketID := uuid.New()

	agents := []*User{
		{
			ID:        agentID,
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Role:      RoleAgent,
			Skills:    MustNewSkills([]string{"production-support"}),
		},
	}

	ticket := &Ticket{
		ID:          ticketID,
		Title:       "Test Ticket",
		Description: "Test Description",
		Skills:      MustNewSkills([]string{"incident-management", "log-analysis"}),
	}

	ticketCounts := map[uuid.UUID]int{
		agentID: 0,
	}

	scores := CalculateAgentScores(ticket, agents, ticketCounts)

	if len(scores) != 0 {
		t.Fatalf("expected 0 scores (no skill matches), got %d", len(scores))
	}
}

func TestCalculateAgentScores_NonAgent(t *testing.T) {
	// User with role != 'agent'
	userID := uuid.New()
	ticketID := uuid.New()

	agents := []*User{
		{
			ID:        userID,
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Role:      RoleUser,
			Skills:    MustNewSkills([]string{"incident-management", "log-analysis"}),
		},
	}

	ticket := &Ticket{
		ID:          ticketID,
		Title:       "Test Ticket",
		Description: "Test Description",
		Skills:      MustNewSkills([]string{"incident-management", "log-analysis"}),
	}

	ticketCounts := map[uuid.UUID]int{
		userID: 0,
	}

	scores := CalculateAgentScores(ticket, agents, ticketCounts)

	if len(scores) != 0 {
		t.Fatalf("expected 0 scores (non-agent), got %d", len(scores))
	}
}

func TestCalculateAgentScores_NoRequiredSkills(t *testing.T) {
	// Ticket with no skills required
	agentID := uuid.New()
	ticketID := uuid.New()

	agents := []*User{
		{
			ID:        agentID,
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Role:      RoleAgent,
			Skills:    MustNewSkills([]string{"incident-management"}),
		},
	}

	ticket := &Ticket{
		ID:          ticketID,
		Title:       "Test Ticket",
		Description: "Test Description",
		Skills:      MustNewSkills([]string{}),
	}

	ticketCounts := map[uuid.UUID]int{
		agentID: 0,
	}

	scores := CalculateAgentScores(ticket, agents, ticketCounts)

	if len(scores) != 0 {
		t.Fatalf("expected 0 scores (no required skills), got %d", len(scores))
	}
}

func TestFindBestAgent(t *testing.T) {
	agentID1 := uuid.New()
	agentID2 := uuid.New()
	ticketID := uuid.New()

	agents := []*User{
		{
			ID:        agentID1,
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Role:      RoleAgent,
			Skills:    MustNewSkills([]string{"incident-management", "log-analysis"}),
		},
		{
			ID:        agentID2,
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane@example.com",
			Role:      RoleAgent,
			Skills:    MustNewSkills([]string{"incident-management"}),
		},
	}

	ticket := &Ticket{
		ID:          ticketID,
		Title:       "Test Ticket",
		Description: "Test Description",
		Skills:      MustNewSkills([]string{"incident-management", "log-analysis"}),
	}

	ticketCounts := map[uuid.UUID]int{
		agentID1: 0,
		agentID2: 0,
	}

	bestAgent := FindBestAgent(ticket, agents, ticketCounts)

	if bestAgent == nil {
		t.Fatalf("expected best agent, got nil")
	}
	if bestAgent.ID != agentID1 {
		t.Fatalf("expected agent1, got %v", bestAgent.ID)
	}
}

func TestFindBestAgent_NoQualifiedAgents(t *testing.T) {
	// No agents with matching skills
	agentID := uuid.New()
	ticketID := uuid.New()

	agents := []*User{
		{
			ID:        agentID,
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Role:      RoleAgent,
			Skills:    MustNewSkills([]string{"production-support"}),
		},
	}

	ticket := &Ticket{
		ID:          ticketID,
		Title:       "Test Ticket",
		Description: "Test Description",
		Skills:      MustNewSkills([]string{"incident-management", "log-analysis"}),
	}

	ticketCounts := map[uuid.UUID]int{
		agentID: 0,
	}

	bestAgent := FindBestAgent(ticket, agents, ticketCounts)

	if bestAgent != nil {
		t.Fatalf("expected nil (no qualified agents), got %v", bestAgent)
	}
}

func TestFindBestAgent_EmptyAgentList(t *testing.T) {
	ticketID := uuid.New()

	ticket := &Ticket{
		ID:          ticketID,
		Title:       "Test Ticket",
		Description: "Test Description",
		Skills:      MustNewSkills([]string{"incident-management"}),
	}

	ticketCounts := map[uuid.UUID]int{}

	bestAgent := FindBestAgent(ticket, []*User{}, ticketCounts)

	if bestAgent != nil {
		t.Fatalf("expected nil (empty agent list), got %v", bestAgent)
	}
}
