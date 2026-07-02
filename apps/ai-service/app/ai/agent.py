from __future__ import annotations

import logging
from typing import Any, Protocol

from app.ai.retrieval import Retriever, format_context
from app.ai.schemas import TicketContext, TriageDecision, TriageResult

logger = logging.getLogger(__name__)

# Stable, cacheable system prompt. Keeping it byte-stable lets prompt caching
# reuse it across every triage call (see shared/prompt-caching.md).
SYSTEM_PROMPT = """You are a support-ticket triage agent for a ticketing product.

Your job: decide whether an incoming ticket can be **answered safely** from the
provided knowledge-base context, or whether it must be **escalated to a human**.

Rules:
- Only propose an auto-answer when the knowledge-base context clearly and fully
  supports a correct, complete reply. If the context is missing, thin, or only
  partially relevant, escalate.
- Escalate anything that is sensitive or high-impact even if you *could* answer:
  account/billing changes, refunds, cancellations, security or auth issues,
  legal/compliance, data deletion or PII exposure, or user distress/self-harm.
  Record these in safety_flags.
- Never invent facts, URLs, prices, or policy. Ground the reply in the context.
- Be honest about confidence. Low confidence means escalate.

Return your decision in the required structured format."""


class MessagesParseClient(Protocol):
    """Minimal surface of the Anthropic client used here (eases test stubbing)."""

    messages: Any


class TriageAgent:
    def __init__(
        self,
        client: MessagesParseClient,
        retriever: Retriever,
        *,
        model: str,
        confidence_threshold: float,
        rag_top_k: int = 5,
    ) -> None:
        self.client = client
        self.retriever = retriever
        self.model = model
        self.confidence_threshold = confidence_threshold
        self.rag_top_k = rag_top_k

    def triage(self, ticket: TicketContext) -> TriageResult:
        chunks = self.retriever.retrieve(ticket, self.rag_top_k)
        context = format_context(chunks)
        user_content = (
            f"Ticket #{ticket.ticket_number or ''} "
            f"(priority={ticket.priority}, state={ticket.state})\n"
            f"Title: {ticket.title}\n"
            f"Description: {ticket.description}\n\n"
            f"Knowledge-base context:\n{context}"
        )

        try:
            response = self.client.messages.parse(
                model=self.model,
                max_tokens=2048,
                thinking={"type": "adaptive"},
                system=[
                    {
                        "type": "text",
                        "text": SYSTEM_PROMPT,
                        "cache_control": {"type": "ephemeral"},
                    }
                ],
                messages=[{"role": "user", "content": user_content}],
                output_format=TriageDecision,
            )
        except Exception:  # noqa: BLE001 — any model/transport failure must fail safe.
            logger.exception("triage model call failed for ticket %s; escalating", ticket.ticket_id)
            return _escalate(ticket, "triage model call failed", ["model_error"])

        # Fail safe on a safety refusal — a refused request is not a green light.
        if getattr(response, "stop_reason", None) == "refusal":
            logger.warning("model refused triage for ticket %s; escalating", ticket.ticket_id)
            return _escalate(ticket, "model declined to process the request", ["model_refusal"])

        decision = response.parsed_output
        if decision is None:
            return _escalate(ticket, "model returned no structured decision", ["parse_error"])

        return self.apply_safety_gate(ticket, decision)

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


def _escalate(ticket: TicketContext, reason: str, flags: list[str]) -> TriageResult:
    return TriageResult(
        ticket_id=ticket.ticket_id,
        action="escalate",
        confidence=0.0,
        escalation_reason=reason,
        safety_flags=flags,
    )


def _gate_reason(decision: TriageDecision, threshold: float) -> str:
    if decision.safety_flags:
        return f"safety flags raised: {', '.join(decision.safety_flags)}"
    if decision.confidence < threshold:
        return f"confidence {decision.confidence:.2f} below threshold {threshold:.2f}"
    if not (decision.draft_reply and decision.draft_reply.strip()):
        return "model proposed an auto-answer without a draft reply"
    return "model chose to escalate"
