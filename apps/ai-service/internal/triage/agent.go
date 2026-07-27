package triage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/nickhildpac/ticket-management-ai-service/internal/rag"
)

// maxTokens caps the model's budget per triage turn — thinking, a few
// retrieval round-trips and one terminal decision fit comfortably here.
const maxTokens = 4096

// messageClient is the slice of the Anthropic SDK the agent uses; it is
// satisfied by *anthropic.MessageService. Tests inject a fake rather than
// depending on the SDK's wire format.
type messageClient interface {
	New(ctx context.Context, params anthropic.MessageNewParams,
		opts ...option.RequestOption) (*anthropic.Message, error)
}

// Agent runs the RAG-backed triage loop for a single ticket at a time.
type Agent struct {
	client   messageClient
	store    rag.SearchStore
	reranker rag.Reranker

	model               string
	confidenceThreshold float64
	ragTopK             int
	candidateK          int
	rerankPool          int
	rrfK                int
	maxIterations       int
}

// Options configures an Agent. Zero values fall back to the same defaults the
// Python agent used.
type Options struct {
	Model               string
	ConfidenceThreshold float64
	RAGTopK             int
	CandidateK          int
	RerankPool          int
	RRFK                int
	MaxIterations       int
}

// NewAgent builds a triage agent over an Anthropic client, a knowledge-base
// store and a cross-encoder re-ranker.
func NewAgent(client *anthropic.Client, store rag.SearchStore, reranker rag.Reranker, opts Options) *Agent {
	return newAgent(&client.Messages, store, reranker, opts)
}

// newAgent is the injectable constructor the tests use.
func newAgent(client messageClient, store rag.SearchStore, reranker rag.Reranker, opts Options) *Agent {
	a := &Agent{
		client:              client,
		store:               store,
		reranker:            reranker,
		model:               opts.Model,
		confidenceThreshold: opts.ConfidenceThreshold,
		ragTopK:             orDefault(opts.RAGTopK, 5),
		candidateK:          orDefault(opts.CandidateK, 20),
		rerankPool:          orDefault(opts.RerankPool, 10),
		rrfK:                orDefault(opts.RRFK, 60),
		maxIterations:       orDefault(opts.MaxIterations, 6),
	}
	return a
}

func orDefault(v, fallback int) int {
	if v <= 0 {
		return fallback
	}
	return v
}

// session is the per-triage state shared by the tool handlers.
type session struct {
	ticketID string
	// candidates are the KB passages surfaced by search_docs, keyed by chunk
	// id, so rerank_results can reorder them without the model re-sending
	// passage text. order preserves first-seen sequence.
	candidates map[int]rag.KBChunk
	order      []int
	// decision is the model's proposal, recorded by a terminal tool. The
	// deterministic safety gate runs on this after the loop — the tools never
	// touch the ticket.
	decision *TriageDecision
}

// Triage runs the agentic loop for one ticket and returns the gated decision.
// Every failure path escalates: a refusal, a transport error, a loop that runs
// out of iterations, and a loop that ends without a terminal tool call.
func (a *Agent) Triage(ctx context.Context, ticket TicketContext) TriageResult {
	sess := &session{ticketID: ticket.TicketID, candidates: map[int]rag.KBChunk{}}

	slog.Info("triaging ticket",
		"ticket_id", ticket.TicketID, "priority", ticket.Priority, "state", ticket.State)

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(a.model),
		MaxTokens: maxTokens,
		// Adaptive thinking: the model decides how much to reason per ticket.
		Thinking: anthropic.ThinkingConfigParamUnion{
			OfAdaptive: &anthropic.ThinkingConfigAdaptiveParam{},
		},
		System: []anthropic.TextBlockParam{{
			Text:         SystemPrompt,
			CacheControl: anthropic.NewCacheControlEphemeralParam(),
		}},
		Tools: toolDefinitions(),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userContent(ticket))),
		},
	}

	// A manual tool loop rather than the SDK tool runner: the runner defers
	// tool execution to the start of the next turn, so a terminal tool
	// (draft_reply / escalate_ticket) would cost one more model round-trip
	// before we could see the decision and stop.
	exhausted := true
	for range a.maxIterations {
		resp, err := a.client.New(ctx, params)
		if err != nil {
			slog.Error("triage model call failed; escalating",
				"ticket_id", ticket.TicketID, "error", err)
			return escalate(ticket, "triage model call failed", []string{"model_error"})
		}

		// Fail safe on a safety refusal — a refused request is not a green light.
		if resp.StopReason == anthropic.StopReasonRefusal {
			slog.Warn("model refused triage; escalating", "ticket_id", ticket.TicketID)
			return escalate(ticket, "model declined to process the request", []string{"model_refusal"})
		}

		params.Messages = append(params.Messages, resp.ToParam())

		results := a.runToolCalls(ctx, sess, resp)
		// A terminal tool recorded a decision; stop before paying for another
		// model round-trip.
		if sess.decision != nil {
			exhausted = false
			break
		}
		if len(results) == 0 {
			// The model answered without calling a tool — no decision to gate.
			exhausted = false
			break
		}
		params.Messages = append(params.Messages, anthropic.NewUserMessage(results...))
	}

	if exhausted {
		slog.Warn("triage hit max_iterations with no decision; escalating",
			"ticket_id", ticket.TicketID, "max_iterations", a.maxIterations)
	}
	if sess.decision == nil {
		return escalate(ticket, "model made no triage decision", []string{"no_decision"})
	}
	return a.ApplySafetyGate(ticket, *sess.decision)
}

// runToolCalls executes every tool_use block in a response and returns the
// matching tool_result blocks. Tool errors come back as error results so the
// model can recover, rather than aborting the loop.
func (a *Agent) runToolCalls(ctx context.Context, sess *session, resp *anthropic.Message) []anthropic.ContentBlockParamUnion {
	var results []anthropic.ContentBlockParamUnion
	for _, block := range resp.Content {
		use, ok := block.AsAny().(anthropic.ToolUseBlock)
		if !ok {
			continue
		}
		out, err := a.dispatch(ctx, sess, use.Name, use.Input)
		if err != nil {
			slog.Error("triage tool failed",
				"ticket_id", sess.ticketID, "tool", use.Name, "error", err)
			results = append(results,
				anthropic.NewToolResultBlock(use.ID, "tool failed: "+err.Error(), true))
			continue
		}
		results = append(results, anthropic.NewToolResultBlock(use.ID, out, false))
	}
	return results
}

func (a *Agent) dispatch(ctx context.Context, sess *session, name string, input json.RawMessage) (string, error) {
	switch name {
	case toolSearchDocs:
		var in searchDocsInput
		if err := json.Unmarshal(input, &in); err != nil {
			return "", err
		}
		k := a.ragTopK
		if in.K != nil && *in.K > 0 {
			k = *in.K
		}
		return a.searchDocs(ctx, sess, in.Query, k)
	case toolRerankResults:
		var in rerankResultsInput
		if err := json.Unmarshal(input, &in); err != nil {
			return "", err
		}
		return a.rerankResults(ctx, sess, in.Query)
	case toolDraftReply:
		var in draftReplyInput
		if err := json.Unmarshal(input, &in); err != nil {
			return "", err
		}
		return draftReply(sess, in), nil
	case toolEscalateTicket:
		var in escalateTicketInput
		if err := json.Unmarshal(input, &in); err != nil {
			return "", err
		}
		return escalateTicket(sess, in), nil
	default:
		return "", fmt.Errorf("unknown tool %q", name)
	}
}

// searchDocs runs the hybrid retrieval lanes, fuses them and caches the
// candidates so rerank_results can reorder the same set.
func (a *Agent) searchDocs(ctx context.Context, sess *session, query string, k int) (string, error) {
	semantic, err := a.store.SearchSemantic(ctx, query, a.candidateK)
	if err != nil {
		return "", err
	}
	keyword, err := a.store.SearchKeyword(ctx, query, a.candidateK)
	if err != nil {
		return "", err
	}
	fused := rag.ReciprocalRankFusion([][]rag.KBChunk{semantic, keyword}, a.rrfK)
	if len(fused) > k {
		fused = fused[:k]
	}
	for _, chunk := range fused {
		if chunk.ID == 0 {
			continue
		}
		if _, seen := sess.candidates[chunk.ID]; !seen {
			sess.order = append(sess.order, chunk.ID)
		}
		sess.candidates[chunk.ID] = chunk
	}

	slog.Info("search_docs", "ticket_id", sess.ticketID, "query", query, "passages", len(fused))
	for rank, chunk := range fused {
		slog.Info("retrieved chunk",
			"ticket_id", sess.ticketID, "rank", rank+1, "source", chunk.Source,
			"distance", chunk.Distance, "content", chunk.Content)
	}
	return rag.FormatCandidates(fused), nil
}

// rerankResults reorders the cached candidates with the cross-encoder.
func (a *Agent) rerankResults(ctx context.Context, sess *session, query string) (string, error) {
	poolIDs := sess.order
	if len(poolIDs) > a.rerankPool {
		poolIDs = poolIDs[:a.rerankPool]
	}
	if len(poolIDs) == 0 {
		return "No candidates to re-rank yet — call search_docs first.", nil
	}

	pool := make([]rag.KBChunk, len(poolIDs))
	passages := make([]string, len(poolIDs))
	for i, id := range poolIDs {
		pool[i] = sess.candidates[id]
		passages[i] = pool[i].Content
	}

	scores, err := a.reranker.Score(ctx, query, passages)
	if err != nil {
		return "", err
	}
	if len(scores) != len(pool) {
		return "", fmt.Errorf("reranker returned %d scores for %d passages", len(scores), len(pool))
	}

	reranked := sortByScoreDesc(pool, scores)
	slog.Info("rerank_results", "ticket_id", sess.ticketID, "passages", len(reranked))
	return rag.FormatCandidates(reranked), nil
}

// draftReply records the model's auto-answer proposal.
func draftReply(sess *session, in draftReplyInput) string {
	// Escalation is sticky: never downgrade a recorded human handoff to an
	// auto-answer.
	if sess.decision != nil && sess.decision.Action == ActionEscalate {
		return "An escalation was already recorded; keeping the human handoff."
	}
	sess.decision = &TriageDecision{
		Action:      ActionAutoAnswer,
		Confidence:  clamp01(in.Confidence),
		DraftReply:  in.Reply,
		SafetyFlags: nonNil(in.SafetyFlags),
	}
	return "Draft reply recorded for safety review."
}

// escalateTicket records a human handoff.
func escalateTicket(sess *session, in escalateTicketInput) string {
	sess.decision = &TriageDecision{
		Action:           ActionEscalate,
		Confidence:       0,
		EscalationReason: in.Reason,
		SafetyFlags:      nonNil(in.SafetyFlags),
	}
	return "Escalation recorded."
}

// ApplySafetyGate is the deterministic final gate. Auto-answer only when the
// model chose to, is confident enough, raised no safety flags, and actually
// drafted a reply.
func (a *Agent) ApplySafetyGate(ticket TicketContext, decision TriageDecision) TriageResult {
	autoOK := decision.Action == ActionAutoAnswer &&
		decision.Confidence >= a.confidenceThreshold &&
		len(decision.SafetyFlags) == 0 &&
		strings.TrimSpace(decision.DraftReply) != ""

	if autoOK {
		return TriageResult{
			TicketID:    ticket.TicketID,
			Action:      ActionAutoAnswer,
			Confidence:  decision.Confidence,
			DraftReply:  decision.DraftReply,
			SafetyFlags: []string{},
		}
	}

	reason := decision.EscalationReason
	if reason == "" {
		reason = gateReason(decision, a.confidenceThreshold)
	}
	return TriageResult{
		TicketID:         ticket.TicketID,
		Action:           ActionEscalate,
		Confidence:       decision.Confidence,
		EscalationReason: reason,
		SafetyFlags:      nonNil(decision.SafetyFlags),
	}
}

func userContent(ticket TicketContext) string {
	number := ""
	if ticket.TicketNumber != nil {
		number = strconv.Itoa(*ticket.TicketNumber)
	}
	return fmt.Sprintf(
		"Ticket #%s (priority=%s, state=%s)\nTitle: %s\nDescription: %s\n\n"+
			"Use search_docs (and optionally rerank_results) to find relevant knowledge-base "+
			"passages, then make exactly one terminal decision by calling draft_reply or "+
			"escalate_ticket.",
		number, ticket.Priority, ticket.State, ticket.Title, ticket.Description)
}

func escalate(ticket TicketContext, reason string, flags []string) TriageResult {
	return TriageResult{
		TicketID:         ticket.TicketID,
		Action:           ActionEscalate,
		Confidence:       0,
		EscalationReason: reason,
		SafetyFlags:      flags,
	}
}

// gateReason is the customer-safe escalation message for gate overrides. It
// deliberately omits internal telemetry — confidence scores, thresholds and raw
// safety-flag slugs — which the worker already records in the logs, not in the
// ticket comment the end user can see.
func gateReason(decision TriageDecision, threshold float64) string {
	switch {
	case len(decision.SafetyFlags) > 0:
		return "This request needs review by our team before we can respond."
	case decision.Confidence < threshold:
		return "We couldn't answer this automatically with enough certainty, so a teammate will follow up."
	default:
		return "A teammate will follow up on this ticket."
	}
}

func clamp01(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}

// nonNil normalises a nil slice to an empty one so JSON encodes `[]`, matching
// the Python service's response shape.
func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// sortByScoreDesc returns pool ordered best-first, rewriting Distance from the
// re-ranker score.
func sortByScoreDesc(pool []rag.KBChunk, scores []float64) []rag.KBChunk {
	idx := make([]int, len(pool))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool { return scores[idx[a]] > scores[idx[b]] })

	out := make([]rag.KBChunk, len(pool))
	for rank, i := range idx {
		out[rank] = rag.KBChunk{
			ID:       pool[i].ID,
			Content:  pool[i].Content,
			Source:   pool[i].Source,
			Distance: rag.ScoreToDistance(scores[i]),
		}
	}
	return out
}
