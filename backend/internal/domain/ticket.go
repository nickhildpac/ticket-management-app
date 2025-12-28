package domain

import (
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
		return "cancel"
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
	ID          int64          `json:"id" db:"id"`
	CreatedBy   uuid.UUID      `json:"created_by" db:"created_by"`
	AssignedTo  uuid.UUID      `json:"assigned_to" db:"assigned_to"`
	Title       string         `json:"title" db:"title"`
	Description string         `json:"description" db:"description"`
	State       TicketState    `json:"state" db:"state"`
	Priority    TicketPriority `json:"priority" db:"priority"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}
