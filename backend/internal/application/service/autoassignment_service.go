package service

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/nickhildpac/ticket-management-app/internal/domain"
	"github.com/nickhildpac/ticket-management-app/internal/ports"
)

type AutoAssignmentService struct {
	userRepo ports.UserRepository
}

func NewAutoAssignmentService(userRepo ports.UserRepository, _ ports.TicketRepository) *AutoAssignmentService {
	return &AutoAssignmentService{
		userRepo: userRepo,
	}
}

func (s *AutoAssignmentService) FindBestAgentForTicket(ctx context.Context, ticket *domain.Ticket) (*domain.User, error) {
	requiredSkills := ticket.Skills.ToSlice()
	if len(requiredSkills) == 0 {
		log.Printf("No qualified agent found for skills: %v", requiredSkills)
		return nil, nil
	}

	candidates, err := s.userRepo.GetAutoAssignmentCandidates(ctx, requiredSkills, []domain.TicketState{
		domain.TicketStateOpen,
		domain.TicketStatePending,
		domain.TicketStateInProgress,
	})
	if err != nil {
		log.Printf("Error getting auto-assignment candidates: %v", err)
		return nil, err
	}

	if len(candidates) == 0 {
		log.Printf("No qualified agent found for skills: %v", requiredSkills)
		return nil, nil
	}

	agentPtrs := make([]*domain.User, len(candidates))
	ticketCounts := make(map[uuid.UUID]int, len(candidates))
	for i := range candidates {
		agentPtrs[i] = &candidates[i].Agent
		ticketCounts[candidates[i].Agent.ID] = candidates[i].ActiveTicketCount
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
