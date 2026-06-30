// Package domain defines core entities and business rules for tickets.
package domain

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type (
	TicketState    int
	TicketPriority int
)

const (
	// Priorities
	TicketPriorityCritical TicketPriority = iota + 1 // 1
	TicketPriorityHigh                               // 2
	TicketPriorityMedium                             // 3
	TicketPriorityLow                                // 4
)

func (s TicketState) String() string {
	if wire, ok := ticketStateWireValues[s]; ok {
		return wire
	}
	return "unknown"
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
		return 4
	}
}

// ParseTicketPriority parses a priority string for strict validation (unknown values are an error).
func ParseTicketPriority(s string) (TicketPriority, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "critical":
		return TicketPriorityCritical, nil
	case "high":
		return TicketPriorityHigh, nil
	case "medium":
		return TicketPriorityMedium, nil
	case "low":
		return TicketPriorityLow, nil
	default:
		return 0, fmt.Errorf("invalid ticket priority: %s", s)
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
	ID           uuid.UUID      `json:"id" db:"id"`
	TicketNumber int64          `json:"ticket_number" db:"ticket_number"`
	CreatedBy    uuid.UUID      `json:"created_by" db:"created_by"`
	AssignedTo   []uuid.UUID    `json:"assigned_to" db:"assigned_to"`
	Skills       Skills         `json:"skills" db:"skills"`
	Title        string         `json:"title" db:"title"`
	Description  string         `json:"description" db:"description"`
	State        TicketState    `json:"state" db:"state"`
	Priority     TicketPriority `json:"priority" db:"priority"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}

// TicketPatch represents a partial ticket update after transport-level decoding.
// Nil fields are left unchanged; non-nil fields are intentionally updated.
type TicketPatch struct {
	Title       *string
	Description *string
	State       *TicketState
	Priority    *TicketPriority
	AssignedTo  *[]uuid.UUID
	Skills      *Skills
}

func (p TicketPatch) IsEmpty() bool {
	return p.Title == nil &&
		p.Description == nil &&
		p.State == nil &&
		p.Priority == nil &&
		p.AssignedTo == nil &&
		p.Skills == nil
}

// TicketListStats holds role-scoped ticket counts (dashboards / GET /tickets/stats).
type TicketListStats struct {
	Total    int32
	Open     int32
	Pending  int32
	Resolved int32
	Mine     int32
}

// Ticket list sort modes (passed to SQL; validated in HTTP layer).
const (
	TicketListSortIDAsc            int32 = 0
	TicketListSortTicketNumberAsc  int32 = 1
	TicketListSortTicketNumberDesc int32 = 2
	TicketListSortCreatedAsc       int32 = 3
	TicketListSortCreatedDesc      int32 = 4
)

// TicketListSortFromQuery maps ?sort=&order= to SortVal. Empty sort defaults to newest-first by created time.
func TicketListSortFromQuery(sortField, order string) (int32, error) {
	f := strings.ToLower(strings.TrimSpace(sortField))
	o := strings.ToLower(strings.TrimSpace(order))
	if f == "" {
		return TicketListSortCreatedDesc, nil
	}
	if o == "" {
		o = "desc"
	}
	if o != "asc" && o != "desc" {
		return 0, fmt.Errorf("invalid order: use asc or desc")
	}
	switch f {
	case "ticket_number":
		if o == "asc" {
			return TicketListSortTicketNumberAsc, nil
		}
		return TicketListSortTicketNumberDesc, nil
	case "created_at":
		if o == "asc" {
			return TicketListSortCreatedAsc, nil
		}
		return TicketListSortCreatedDesc, nil
	default:
		return 0, fmt.Errorf("invalid sort: use ticket_number or created_at")
	}
}

// ListAllTicketsByStatePriorityParams holds optional filters for listing tickets (unset/null fields are ignored).
type ListAllTicketsByStatePriorityParams struct {
	FilterState        sql.NullInt32
	FilterPriority     sql.NullInt32
	FilterCreatedBy    uuid.NullUUID
	FilterAssignee     uuid.NullUUID
	FilterTicketNumber sql.NullInt64
	OffsetVal          int32
	LimitVal           int32
	SortVal            int32
}

func GetTicketState(s string) (TicketState, error) {
	normalized := strings.ToLower(strings.TrimSpace(s))
	if state, ok := ticketStateByInput[normalized]; ok {
		return state, nil
	}
	return 0, fmt.Errorf("invalid ticket state: %s", s)
}

func CanTransition(from TicketState, to TicketState) bool {
	if from == to {
		return true
	}
	next, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	for _, state := range next {
		if state == to {
			return true
		}
	}
	return false
}

// GetValidTransitions returns all valid states that can be transitioned to from the given state
func GetValidTransitions(from TicketState) []TicketState {
	var validStates []TicketState

	// Always include current state (no change)
	validStates = append(validStates, from)

	transitions, ok := allowedTransitions[from]
	if !ok {
		return validStates
	}
	validStates = append(validStates, transitions...)
	return validStates
}

var ErrInvalidStatusTransition = errors.New("invalid status transition")

// GetTransitionError returns a more descriptive error for invalid transitions
func GetTransitionError(from TicketState, to TicketState) error {
	return fmt.Errorf("cannot transition ticket from %s to %s: %w", from.String(), to.String(), ErrInvalidStatusTransition)
}
