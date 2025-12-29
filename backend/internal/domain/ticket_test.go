package domain

import (
	"testing"
)

func TestCanTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     TicketState
		to       TicketState
		expected bool
	}{
		// Open ticket transitions
		{"Open to Pending", TicketStateOpen, TicketStatePending, true},
		{"Open to Cancelled", TicketStateOpen, TicketStateCancelled, true},
		{"Open to Closed", TicketStateOpen, TicketStateClosed, false},
		{"Open to Resolved", TicketStateOpen, TicketStateResolved, false},

		// Pending ticket transitions
		{"Pending to Open", TicketStatePending, TicketStateOpen, true},
		{"Pending to Resolved", TicketStatePending, TicketStateResolved, true},
		{"Pending to Cancelled", TicketStatePending, TicketStateCancelled, true},
		{"Pending to Closed", TicketStatePending, TicketStateClosed, false},

		// Resolved ticket transitions
		{"Resolved to Open", TicketStateResolved, TicketStateOpen, true},
		{"Resolved to Pending", TicketStateResolved, TicketStatePending, true},
		{"Resolved to Closed", TicketStateResolved, TicketStateClosed, true},
		{"Resolved to Cancelled", TicketStateResolved, TicketStateCancelled, true},

		// Closed ticket transitions (final state)
		{"Closed to Open", TicketStateClosed, TicketStateOpen, false},
		{"Closed to Closed", TicketStateClosed, TicketStateClosed, true},

		// Cancelled ticket transitions (final state)
		{"Cancelled to Open", TicketStateCancelled, TicketStateOpen, false},
		{"Cancelled to Cancelled", TicketStateCancelled, TicketStateCancelled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanTransition(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("CanTransition(%v, %v) = %v; want %v", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestGetValidTransitions(t *testing.T) {
	tests := []struct {
		name           string
		from           TicketState
		expectedCount  int
		expectedStates []TicketState
	}{
		{
			name:           "Open states",
			from:           TicketStateOpen,
			expectedCount:  3, // Open + Pending + Cancelled
			expectedStates: []TicketState{TicketStateOpen, TicketStatePending, TicketStateCancelled},
		},
		{
			name:           "Pending states",
			from:           TicketStatePending,
			expectedCount:  4, // Pending + Open + Resolved + Cancelled
			expectedStates: []TicketState{TicketStatePending, TicketStateOpen, TicketStateResolved, TicketStateCancelled},
		},
		{
			name:           "Closed states",
			from:           TicketStateClosed,
			expectedCount:  1, // Only Closed
			expectedStates: []TicketState{TicketStateClosed},
		},
		{
			name:           "Cancelled states",
			from:           TicketStateCancelled,
			expectedCount:  1, // Only Cancelled
			expectedStates: []TicketState{TicketStateCancelled},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetValidTransitions(tt.from)
			if len(result) != tt.expectedCount {
				t.Errorf("GetValidTransitions(%v) returned %d states; want %d", tt.from, len(result), tt.expectedCount)
			}

			for _, expectedState := range tt.expectedStates {
				found := false
				for _, state := range result {
					if state == expectedState {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetValidTransitions(%v) missing expected state %v", tt.from, expectedState)
				}
			}
		})
	}
}

func TestTicketStateString(t *testing.T) {
	tests := []struct {
		name     string
		state    TicketState
		expected string
	}{
		{"Open", TicketStateOpen, "open"},
		{"Pending", TicketStatePending, "pending"},
		{"Resolved", TicketStateResolved, "resolved"},
		{"Closed", TicketStateClosed, "closed"},
		{"Cancelled", TicketStateCancelled, "cancelled"},
		{"Unknown", TicketState(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.state.String()
			if result != tt.expected {
				t.Errorf("%v.String() = %q; want %q", tt.state, result, tt.expected)
			}
		})
	}
}
