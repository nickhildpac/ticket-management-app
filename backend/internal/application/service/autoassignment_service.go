package service

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/internal/ports"
)

type AutoAssignmentService struct {
	userRepo   ports.UserRepository
	ticketRepo ports.TicketRepository
}

func NewAutoAssignmentService(userRepo ports.UserRepository, ticketRepo ports.TicketRepository) *AutoAssignmentService {
	return &AutoAssignmentService{
		userRepo:   userRepo,
		ticketRepo: ticketRepo,
	}
}

func (s *AutoAssignmentService) FindBestAgentForTicket(ctx context.Context, ticket *domain.Ticket) (*domain.User, error) {
	// Get all agents
	agents, err := s.userRepo.GetAllAgents(ctx)
	if err != nil {
		log.Printf("Error getting agents: %v", err)
		return nil, err
	}

	if len(agents) == 0 {
		log.Printf("No agents available")
		return nil, nil
	}

	// Convert to pointers for domain function
	agentPtrs := make([]*domain.User, len(agents))
	for i := range agents {
		agentPtrs[i] = &agents[i]
	}

	// Get active tickets for workload calculation
	activeTickets, err := s.ticketRepo.GetActiveTickets(ctx)
	if err != nil {
		log.Printf("Error getting active tickets: %v", err)
		return nil, err
	}

	// Calculate workload: count Open/Pending tickets per agent
	ticketCounts := make(map[uuid.UUID]int)
	for _, ticket := range activeTickets {
		for _, assignee := range ticket.AssignedTo {
			ticketCounts[assignee]++
		}
	}

	// Find best agent using domain logic
	bestAgent := domain.FindBestAgent(ticket, agentPtrs, ticketCounts)

	if bestAgent != nil {
		log.Printf("Auto-assigned to agent %s %s", bestAgent.FirstName, bestAgent.LastName)
	} else {
		log.Printf("No qualified agent found for skills: %v", ticket.Skills.ToSlice())
	}

	return bestAgent, nil
}
