from __future__ import annotations

import logging
from collections.abc import Callable
from dataclasses import dataclass, field
from typing import Any, Protocol

from anthropic import beta_tool

from app.ai.rerank import Reranker, score_to_distance
from app.ai.retrieval import format_candidates, reciprocal_rank_fusion
from app.ai.schemas import KBChunk, TicketContext, TriageDecision, TriageResult
from app.ai.vectorstore import VectorStore

logger = logging.getLogger(__name__)

# Cap on the model's tool-use budget per triage — thinking + a few retrieval
# round-trips + one terminal decision fit comfortably here.
_MAX_TOKENS = 4096

# Stable, cacheable system prompt. Keeping it byte-stable lets prompt caching
# reuse it across every triage call (see shared/prompt-caching.md).
SYSTEM_PROMPT = """You are a support-ticket triage agent for a ticketing product.

Your job: decide whether an incoming ticket can be **answered safely** from the
knowledge base (KB), or whether it must be **escalated to a human**.

Workflow — every triage MUST end in exactly one terminal decision:
1. Call `search_docs` with a focused query to retrieve relevant KB passages.
   Search again with a refined query if the first results are thin.
2. Optionally call `rerank_results` to reorder the candidates by relevance.
3. Make exactly ONE terminal call: `draft_reply` (propose a grounded, cited
   suggested reply) or `escalate_ticket` (hand off to a human). Never call both.

What an auto-answer actually is: a grounded, cited *suggested reply* posted as a
comment for a human agent to review. It never changes the account, never sends a
message to the customer directly, and never transitions the ticket. So the bar is a
**correct, complete, KB-supported reply** — not "zero risk".

Call **draft_reply** when BOTH hold:
- The retrieved KB passages clearly and fully support a correct, complete answer, AND
- The ticket only needs information or **self-service steps the user can perform
  themselves** (how-to, configuration, documented troubleshooting). This includes
  auth/security/API/login topics when the documented fix is self-service — e.g. a 403
  caused by a disabled access toggle or a missing API-key scope is a documented,
  user-fixable configuration issue and should be auto-answered.

Call **escalate_ticket** when the ticket involves a sensitive **action, decision, or
risk** — not merely a sensitive-sounding topic. Escalate, and record a matching flag
from the taxonomy below in safety_flags, when:
- The user is asking you to take or authorize something sensitive (change billing/plan/
  seats, issue a refund, cancel, delete data, expose PII, reset security), OR
- Resolution requires a human or Anthropic-side action (account lockout recovery, a
  suspected security incident or leaked credential, a data-pipeline failure, a
  rate-limit increase, legal/compliance, or a CSM/Sales referral), OR
- The user is in distress or mentions self-harm.

safety_flags taxonomy — use ONLY these values, and set a flag only when it genuinely
applies (an empty list means nothing sensitive is at stake):
- account_or_billing_change    — changing plan, seats, billing, or contacts
- refund_or_cancellation
- security_incident_or_compromise  — suspected breach or account takeover
- credential_or_key_leak       — exposed/compromised API key or password
- account_access_lockout       — locked out; needs human account action to regain access
- data_deletion_or_pii         — deleting data or exposing personal data
- legal_or_compliance
- user_distress_or_self_harm
- requires_anthropic_side_action  — remedy is only doable by Anthropic/CSM/Sales

Do NOT escalate just because a ticket mentions auth, security, API, or login when the
documented resolution is self-service configuration the user controls and the KB covers
it. Reserve safety_flags for the cases above.

Grounding and citations:
- `search_docs`/`rerank_results` return passages labelled "[id] source=<path>". Every
  step or claim in a draft_reply must come from these passages; cite the "[id]" you
  relied on.
- Never invent facts, URLs, prices, or policy. If the KB is missing, thin, or only
  partially relevant, call escalate_ticket instead of guessing.
- End the draft by noting that a human will follow up if the steps don't resolve it, and
  direct any genuinely account-side action to support rather than performing it yourself.
- escalate_ticket reasons are shown to the customer: keep them brief and customer-safe —
  never include internal metrics (confidence, thresholds) or raw safety-flag slugs.

Be honest about confidence. Low confidence means escalate. When unsure, escalate."""


class ToolRunnerClient(Protocol):
    """Minimal surface of the Anthropic client used here (eases test stubbing)."""

    beta: Any


# A runner factory takes the raw tool closures + request params and returns an
# object the triage loop can iterate (yielding assistant messages) and drive with
# ``generate_tool_call_response()``. The default wraps the SDK Tool Runner; tests
# inject a fake so they don't depend on the SDK's wire format.
RunnerFactory = Callable[..., Any]


@dataclass
class _TriageSession:
    """Per-triage state shared by the tool closures for a single ticket."""

    ticket_id: str
    # KB candidates surfaced by search_docs, keyed by chunk id, so rerank_results
    # can reorder them without the model re-sending passage text.
    candidates: dict[int, KBChunk] = field(default_factory=dict)
    # The model's proposed decision, recorded by a terminal tool. The deterministic
    # safety gate runs on this after the loop — the tools never touch the ticket.
    decision: TriageDecision | None = None


class TriageAgent:
    def __init__(
        self,
        client: ToolRunnerClient,
        store: VectorStore,
        reranker: Reranker,
        *,
        model: str,
        confidence_threshold: float,
        rag_top_k: int = 5,
        candidate_k: int = 20,
        rerank_pool: int = 10,
        rrf_k: int = 60,
        max_iterations: int = 6,
        runner_factory: RunnerFactory | None = None,
    ) -> None:
        self.client = client
        self.store = store
        self.reranker = reranker
        self.model = model
        self.confidence_threshold = confidence_threshold
        self.rag_top_k = rag_top_k
        self.candidate_k = candidate_k
        self.rerank_pool = rerank_pool
        self.rrf_k = rrf_k
        self.max_iterations = max_iterations
        self._runner_factory = runner_factory or self._default_runner_factory

    def triage(self, ticket: TicketContext) -> TriageResult:
        session = _TriageSession(ticket_id=ticket.ticket_id)
        funcs = self._build_tools(ticket, session)
        system = [
            {"type": "text", "text": SYSTEM_PROMPT, "cache_control": {"type": "ephemeral"}}
        ]
        messages = [{"role": "user", "content": self._user_content(ticket)}]
        logger.info(
            "triaging ticket %s (priority=%s state=%s)",
            ticket.ticket_id,
            ticket.priority,
            ticket.state,
        )

        try:
            runner = self._runner_factory(
                funcs, system=system, messages=messages, max_tokens=_MAX_TOKENS
            )
            for i, message in enumerate(runner):
                # Fail safe on a safety refusal — a refused request is not a green light.
                if getattr(message, "stop_reason", None) == "refusal":
                    logger.warning("model refused triage for ticket %s; escalating", ticket.ticket_id)
                    return _escalate(ticket, "model declined to process the request", ["model_refusal"])
                # Execute this turn's tool calls now (cached; the runner reuses the
                # result). Terminal tools set session.decision here, so we can stop
                # immediately instead of paying for one more model round-trip.
                runner.generate_tool_call_response()
                if session.decision is not None:
                    break
                if i + 1 >= self.max_iterations:
                    logger.warning(
                        "triage for ticket %s hit max_iterations with no decision; escalating",
                        ticket.ticket_id,
                    )
                    break
        except Exception:  # noqa: BLE001 — any model/transport failure must fail safe.
            logger.exception("triage tool loop failed for ticket %s; escalating", ticket.ticket_id)
            return _escalate(ticket, "triage model call failed", ["model_error"])

        if session.decision is None:
            return _escalate(ticket, "model made no triage decision", ["no_decision"])

        return self.apply_safety_gate(ticket, session.decision)

    def _default_runner_factory(
        self,
        funcs: list[Callable[..., str]],
        *,
        system: list[dict[str, Any]],
        messages: list[dict[str, Any]],
        max_tokens: int,
    ) -> Any:
        return self.client.beta.messages.tool_runner(
            model=self.model,
            max_tokens=max_tokens,
            thinking={"type": "adaptive"},
            system=system,
            messages=messages,
            tools=[beta_tool(fn) for fn in funcs],
            max_iterations=self.max_iterations,
        )

    def _user_content(self, ticket: TicketContext) -> str:
        return (
            f"Ticket #{ticket.ticket_number or ''} "
            f"(priority={ticket.priority}, state={ticket.state})\n"
            f"Title: {ticket.title}\n"
            f"Description: {ticket.description}\n\n"
            "Use search_docs (and optionally rerank_results) to find relevant knowledge-base "
            "passages, then make exactly one terminal decision by calling draft_reply or "
            "escalate_ticket."
        )

    def _build_tools(
        self, ticket: TicketContext, session: _TriageSession
    ) -> list[Callable[..., str]]:
        """Build the four triage tools as closures over this ticket's session.

        Returns plain functions (with type hints + docstrings the SDK turns into
        schemas). The default runner wraps them with ``beta_tool``; tests call them
        directly.
        """

        def search_docs(query: str, k: int = 5) -> str:
            """Search the knowledge base for passages relevant to a query.

            Runs semantic + keyword retrieval and fuses the results. Returns passages
            labelled by a numeric [id] you can cite in a draft reply. Call again with a
            refined query if the results are thin.

            Args:
                query: Natural-language search query describing what you need.
                k: Maximum number of fused passages to return.
            """
            semantic = self.store.search_semantic(query, self.candidate_k)
            keyword = self.store.search_keyword(query, self.candidate_k)
            fused = reciprocal_rank_fusion([semantic, keyword], rrf_k=self.rrf_k)[:k]
            for chunk in fused:
                if chunk.id is not None:
                    session.candidates[chunk.id] = chunk
            logger.info(
                "search_docs ticket=%s query=%r returned %d passage(s)",
                session.ticket_id,
                query,
                len(fused),
            )
            for rank, chunk in enumerate(fused, start=1):
                logger.info(
                    "retrieved chunk ticket=%s rank=%d source=%s distance=%.4f content=%r",
                    session.ticket_id,
                    rank,
                    chunk.source,
                    chunk.distance,
                    chunk.content,
                )
            return format_candidates(fused)

        def rerank_results(query: str) -> str:
            """Re-rank the passages found by search_docs using a cross-encoder.

            Call after search_docs to reorder the cached candidates by how well they
            answer the query. Returns passages best-first, labelled by [id].

            Args:
                query: The query to score passages against (usually the ticket's need).
            """
            pool = list(session.candidates.values())[: self.rerank_pool]
            if not pool:
                return "No candidates to re-rank yet — call search_docs first."
            scores = self.reranker.score(query, [c.content for c in pool])
            ranked = sorted(
                zip(pool, scores, strict=True),
                key=lambda item: item[1],
                reverse=True,
            )
            reranked = [
                KBChunk(
                    id=chunk.id,
                    content=chunk.content,
                    source=chunk.source,
                    distance=score_to_distance(score),
                )
                for chunk, score in ranked
            ]
            logger.info(
                "rerank_results ticket=%s reranked %d passage(s)",
                session.ticket_id,
                len(reranked),
            )
            return format_candidates(reranked)

        def draft_reply(reply: str, confidence: float, safety_flags: list[str] | None = None) -> str:
            """Propose a grounded, cited suggested reply (terminal decision).

            Call this exactly once when the KB fully supports a correct, complete,
            self-service answer. The reply is posted as a suggested comment for a human
            to review, and is still subject to a safety gate.

            Args:
                reply: The suggested reply. Every step/claim must cite an [id] from search results.
                confidence: 0..1 confidence the reply is correct and fully grounded in the KB.
                safety_flags: Leave empty for a clean self-service answer; any flag forces escalation.
            """
            # Escalation is sticky: never downgrade a recorded human handoff to an auto-answer.
            if session.decision is not None and session.decision.action == "escalate":
                return "An escalation was already recorded; keeping the human handoff."
            session.decision = TriageDecision(
                action="auto_answer",
                confidence=_clamp01(confidence),
                draft_reply=reply,
                safety_flags=list(safety_flags or []),
            )
            return "Draft reply recorded for safety review."

        def escalate_ticket(reason: str, safety_flags: list[str] | None = None) -> str:
            """Hand the ticket to a human (terminal decision).

            Call this exactly once when the KB can't fully support a safe answer, or the
            ticket needs a sensitive action/decision or a human/Anthropic-side remedy.

            Args:
                reason: Brief, customer-safe explanation for the handoff. No internal metrics or flag slugs.
                safety_flags: Applicable values from the taxonomy; empty if none apply.
            """
            session.decision = TriageDecision(
                action="escalate",
                confidence=0.0,
                escalation_reason=reason,
                safety_flags=list(safety_flags or []),
            )
            return "Escalation recorded."

        return [search_docs, rerank_results, draft_reply, escalate_ticket]

    def apply_safety_gate(self, ticket: TicketContext, decision: TriageDecision) -> TriageResult:
        """Deterministic final gate. Auto-answer only when the model chose to,
        is confident enough, raised no safety flags, and actually drafted a reply."""
        auto_ok = (
            decision.action == "auto_answer"
            and decision.confidence >= self.confidence_threshold
            and not decision.safety_flags
            and bool(decision.draft_reply and decision.draft_reply.strip())
        )
        if auto_ok:
            return TriageResult(
                ticket_id=ticket.ticket_id,
                action="auto_answer",
                confidence=decision.confidence,
                draft_reply=decision.draft_reply,
                safety_flags=[],
            )

        reason = decision.escalation_reason or _gate_reason(decision, self.confidence_threshold)
        return TriageResult(
            ticket_id=ticket.ticket_id,
            action="escalate",
            confidence=decision.confidence,
            escalation_reason=reason,
            safety_flags=decision.safety_flags,
        )


def _clamp01(value: float) -> float:
    return max(0.0, min(1.0, float(value)))


def _escalate(ticket: TicketContext, reason: str, flags: list[str]) -> TriageResult:
    return TriageResult(
        ticket_id=ticket.ticket_id,
        action="escalate",
        confidence=0.0,
        escalation_reason=reason,
        safety_flags=flags,
    )


def _gate_reason(decision: TriageDecision, threshold: float) -> str:
    """Customer-safe escalation message for gate overrides. Deliberately omits internal
    telemetry — confidence scores, thresholds, and raw safety-flag slugs — which the
    worker already records in the logs, not in the ticket comment the end user can see."""
    if decision.safety_flags:
        return "This request needs review by our team before we can respond."
    if decision.confidence < threshold:
        return "We couldn't answer this automatically with enough certainty, so a teammate will follow up."
    return "A teammate will follow up on this ticket."
