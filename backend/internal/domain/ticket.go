package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type TicketState int
type TicketPriority int

const (
	// States
	TicketStateOpen      TicketState = iota + 1 // 1
	TicketStatePending                          // 2
	TicketStateResolved                         // 3
	TicketStateClosed                           // 4
	TicketStateCancelled                        // 5
)

const (
	// Priorities
	TicketPriorityCritical TicketPriority = iota + 1 // 1
	TicketPriorityHigh                               // 2
	TicketPriorityMedium                             // 3
	TicketPriorityLow                                // 4
)

func (s TicketState) String() string {
	switch s {
	case TicketStateOpen:
		return "open"
	case TicketStatePending:
		return "pending"
	case TicketStateResolved:
		return "resolved"
	case TicketStateClosed:
		return "closed"
	case TicketStateCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

func GetTicketPriority(s string) TicketPriority {
	switch strings.ToLower(s) {
	case "critical":
		return TicketPriorityCritical
	case "high":
		return TicketPriorityHigh
	case "medium":
		return TicketPriorityMedium
	case "low":
		return TicketPriorityLow
	default:
		return -1
	}
}

func (p TicketPriority) String() string {
	switch p {
	case TicketPriorityLow:
		return "low"
	case TicketPriorityMedium:
		return "medium"
	case TicketPriorityHigh:
		return "high"
	case TicketPriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

type Ticket struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	CreatedBy   uuid.UUID      `json:"created_by" db:"created_by"`
	AssignedTo  uuid.UUID      `json:"assigned_to" db:"assigned_to"`
	Title       string         `json:"title" db:"title"`
	Description string         `json:"description" db:"description"`
	State       TicketState    `json:"state" db:"state"`
	Priority    TicketPriority `json:"priority" db:"priority"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

var allowedTransitions = map[TicketState]map[TicketState]struct{}{
	// Open tickets can move to Pending, be Cancelled, or stay Open
	TicketStateOpen: {
		TicketStatePending:   {},
		TicketStateCancelled: {},
	},
	// Pending tickets can move back to Open, be Resolved, or be Cancelled
	TicketStatePending: {
		TicketStateOpen:      {},
		TicketStateResolved:  {},
		TicketStateCancelled: {},
	},
	// Resolved tickets can move back to Open/Pending (reopened), be Closed, or be Cancelled
	TicketStateResolved: {
		TicketStateOpen:      {},
		TicketStatePending:   {},
		TicketStateClosed:    {},
		TicketStateCancelled: {},
	},
	// Closed tickets are final - no transitions allowed
	TicketStateClosed: {},
	// Cancelled tickets are final - no transitions allowed
	TicketStateCancelled: {},
}

func CanTransition(from TicketState, to TicketState) bool {
	if from == to {
		return true
	}
	next, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	_, ok = next[to]
	return ok
}

// GetValidTransitions returns all valid states that can be transitioned to from the given state
func GetValidTransitions(from TicketState) []TicketState {
	var validStates []TicketState

	// Always include current state (no change)
	validStates = append(validStates, from)

	// Add allowed transition states
	if transitions, ok := allowedTransitions[from]; ok {
		for state := range transitions {
			validStates = append(validStates, state)
		}
	}

	return validStates
}

var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

// GetTransitionError returns a more descriptive error for invalid transitions
func GetTransitionError(from TicketState, to TicketState) error {
	return fmt.Errorf("cannot transition ticket from %s to %s: %w", from.String(), to.String(), ErrInvalidStatusTransition)
}
