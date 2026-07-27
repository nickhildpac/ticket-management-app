// Package triage holds the support-ticket triage agent: the RAG-backed
// tool loop that proposes a decision, and the deterministic safety gate that
// has the final say on whether we auto-answer.
package triage

// Action is a terminal triage outcome.
type Action string

const (
	// ActionAutoAnswer posts a grounded, cited suggested reply as a comment.
	ActionAutoAnswer Action = "auto_answer"
	// ActionEscalate hands the ticket to a human.
	ActionEscalate Action = "escalate"
)

// TicketContext is the set of ticket fields the triage agent reasons over
// (taken from a ticket event).
type TicketContext struct {
	TicketID     string `json:"ticket_id"`
	TicketNumber *int   `json:"ticket_number,omitempty"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	State        string `json:"state"`
	Priority     string `json:"priority"`
}

// TriageDecision is the structured decision the model records via a terminal
// tool.
//
// This is the model's *proposal*; the deterministic safety gate in
// Agent.ApplySafetyGate has the final say on whether we auto-answer.
type TriageDecision struct {
	Action           Action   `json:"action"`
	Confidence       float64  `json:"confidence"`
	DraftReply       string   `json:"draft_reply,omitempty"`
	EscalationReason string   `json:"escalation_reason,omitempty"`
	SafetyFlags      []string `json:"safety_flags"`
}

// TriageResult is the final decision after applying the deterministic safety
// gate.
type TriageResult struct {
	TicketID         string   `json:"ticket_id"`
	Action           Action   `json:"action"`
	Confidence       float64  `json:"confidence"`
	DraftReply       string   `json:"draft_reply,omitempty"`
	EscalationReason string   `json:"escalation_reason,omitempty"`
	SafetyFlags      []string `json:"safety_flags"`
}
