package triage

import "github.com/anthropics/anthropic-sdk-go"

// Tool names the model may call. The first two retrieve; the last two are
// terminal decisions.
const (
	toolSearchDocs     = "search_docs"
	toolRerankResults  = "rerank_results"
	toolDraftReply     = "draft_reply"
	toolEscalateTicket = "escalate_ticket"
)

// searchDocsInput is the model's argument shape for search_docs.
type searchDocsInput struct {
	Query string `json:"query"`
	K     *int   `json:"k,omitempty"`
}

// rerankResultsInput is the model's argument shape for rerank_results.
type rerankResultsInput struct {
	Query string `json:"query"`
}

// draftReplyInput is the model's argument shape for the draft_reply terminal
// decision.
type draftReplyInput struct {
	Reply       string   `json:"reply"`
	Confidence  float64  `json:"confidence"`
	SafetyFlags []string `json:"safety_flags,omitempty"`
}

// escalateTicketInput is the model's argument shape for the escalate_ticket
// terminal decision.
type escalateTicketInput struct {
	Reason      string   `json:"reason"`
	SafetyFlags []string `json:"safety_flags,omitempty"`
}

// safetyFlagsSchema is the shared JSON schema fragment for a safety_flags
// argument. The taxonomy itself is spelled out in the system prompt.
var safetyFlagsSchema = map[string]any{
	"type":  "array",
	"items": map[string]any{"type": "string"},
	"description": "Leave empty for a clean self-service answer; any flag forces escalation. " +
		"Use only values from the safety_flags taxonomy in the system prompt.",
}

// toolDefinitions returns the four triage tools. The schemas and descriptions
// are the model's only contract for these tools, so they carry the same
// guidance the Python docstrings did.
func toolDefinitions() []anthropic.ToolUnionParam {
	tools := []anthropic.ToolParam{
		{
			Name: toolSearchDocs,
			Description: anthropic.String(
				"Search the knowledge base for passages relevant to a query.\n\n" +
					"Runs semantic + keyword retrieval and fuses the results. Returns passages " +
					"labelled by a numeric [id] you can cite in a draft reply. Call again with a " +
					"refined query if the results are thin."),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Natural-language search query describing what you need.",
					},
					"k": map[string]any{
						"type":        "integer",
						"description": "Maximum number of fused passages to return.",
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name: toolRerankResults,
			Description: anthropic.String(
				"Re-rank the passages found by search_docs using a cross-encoder.\n\n" +
					"Call after search_docs to reorder the cached candidates by how well they " +
					"answer the query. Returns passages best-first, labelled by [id]."),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The query to score passages against (usually the ticket's need).",
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name: toolDraftReply,
			Description: anthropic.String(
				"Propose a grounded, cited suggested reply (terminal decision).\n\n" +
					"Call this exactly once when the KB fully supports a correct, complete, " +
					"self-service answer. The reply is posted as a suggested comment for a human " +
					"to review, and is still subject to a safety gate."),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"reply": map[string]any{
						"type": "string",
						"description": "The suggested reply. Every step/claim must cite an [id] " +
							"from search results.",
					},
					"confidence": map[string]any{
						"type":    "number",
						"minimum": 0,
						"maximum": 1,
						"description": "0..1 confidence the reply is correct and fully grounded " +
							"in the knowledge base.",
					},
					"safety_flags": safetyFlagsSchema,
				},
				Required: []string{"reply", "confidence"},
			},
		},
		{
			Name: toolEscalateTicket,
			Description: anthropic.String(
				"Hand the ticket to a human (terminal decision).\n\n" +
					"Call this exactly once when the KB can't fully support a safe answer, or the " +
					"ticket needs a sensitive action/decision or a human/Anthropic-side remedy."),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"reason": map[string]any{
						"type": "string",
						"description": "Brief, customer-safe explanation for the handoff. " +
							"No internal metrics or flag slugs.",
					},
					"safety_flags": safetyFlagsSchema,
				},
				Required: []string{"reason"},
			},
		},
	}

	out := make([]anthropic.ToolUnionParam, len(tools))
	for i := range tools {
		out[i] = anthropic.ToolUnionParam{OfTool: &tools[i]}
	}
	return out
}
