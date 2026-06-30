package domain

import (
	"sort"

	"github.com/google/uuid"
)

// AgentScore represents an agent's calculated score for assignment
type AgentScore struct {
	User     *User
	Skills   float64 // 0-70 points
	Workload float64 // 0-30 points
	Total    float64 // Skills + Workload
}

type AutoAssignmentCandidate struct {
	Agent             User
	ActiveTicketCount int
}

// CalculateAgentScores calculates scores for all agents
func CalculateAgentScores(ticket *Ticket, agents []*User, ticketCounts map[uuid.UUID]int) []*AgentScore {
	var scores []*AgentScore

	for _, agent := range agents {
		if agent.Role != RoleAgent {
			continue
		}

		// Check skill matches
		matchingSkills := 0
		requiredSkills := ticket.Skills.ToSlice()

		if len(requiredSkills) == 0 {
			continue // No skills required, skip auto-assignment
		}

		for _, skill := range requiredSkills {
			if agent.Skills.Contains(skill) {
				matchingSkills++
			}
		}

		// Safety: At least one skill match required
		if matchingSkills == 0 {
			continue
		}

		// Calculate scores
		skillsScore := (float64(matchingSkills) / float64(len(requiredSkills))) * 70
		workload := float64(ticketCounts[agent.ID])
		workloadScore := (1 / (workload + 1)) * 30
		totalScore := skillsScore + workloadScore

		scores = append(scores, &AgentScore{
			User:     agent,
			Skills:   skillsScore,
			Workload: workloadScore,
			Total:    totalScore,
		})
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Total > scores[j].Total
	})

	return scores
}

// FindBestAgent returns the best agent for a ticket (or nil if none qualified)
func FindBestAgent(ticket *Ticket, agents []*User, ticketCounts map[uuid.UUID]int) *User {
	scores := CalculateAgentScores(ticket, agents, ticketCounts)

	if len(scores) == 0 {
		return nil // No qualified agents
	}

	return scores[0].User
}
