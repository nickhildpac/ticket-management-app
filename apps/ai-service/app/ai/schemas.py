from __future__ import annotations

from typing import Literal

from pydantic import BaseModel, Field


class TicketContext(BaseModel):
    """The ticket fields the triage agent reasons over (from a ticket event)."""

    ticket_id: str
    ticket_number: int | None = None
    title: str
    description: str
    state: str
    priority: str = "low"


class KBChunk(BaseModel):
    """A retrieved knowledge-base passage with its similarity distance.

    ``id`` identifies the row for hybrid fusion (RRF); optional so stubs/tests
    can omit it. ``distance`` is a lower-is-better score (cosine distance or a
    monotone transform of the re-ranker score).
    """

    content: str
    source: str
    distance: float
    id: int | None = None


class TriageDecision(BaseModel):
    """Structured output the model returns for a triage request.

    This is the model's *proposal*; the deterministic safety gate in
    ``agent.apply_safety_gate`` has the final say on whether we auto-answer.
    """

    action: Literal["auto_answer", "escalate"] = Field(
        description="Whether the ticket can be answered safely from the knowledge base."
    )
    confidence: float = Field(
        ge=0.0, le=1.0, description="Confidence that the drafted answer is correct and "
        "fully grounded in the retrieved knowledge base, 0..1.",
    )
    draft_reply: str | None = Field(
        default=None,
        description="Proposed reply to the customer when action is auto_answer. Every "
        "step/claim must be grounded in the retrieved passages and cite the [i] source "
        "it relied on.",
    )
    escalation_reason: str | None = Field(
        default=None,
        description="Brief, customer-safe explanation of why a human is needed when action "
        "is escalate. This is shown to the end user, so do not include internal metrics "
        "(confidence scores, thresholds) or safety-flag slugs.",
    )
    safety_flags: list[str] = Field(
        default_factory=list,
        description="Empty unless the ticket needs a sensitive action/decision or a "
        "human/Anthropic-side remedy. Use only these values: account_or_billing_change, "
        "refund_or_cancellation, security_incident_or_compromise, credential_or_key_leak, "
        "account_access_lockout, data_deletion_or_pii, legal_or_compliance, "
        "user_distress_or_self_harm, requires_anthropic_side_action. Do NOT flag "
        "documented self-service troubleshooting (e.g. an access toggle or API-key scope).",
    )


class TriageResult(BaseModel):
    """Final decision after applying the deterministic safety gate."""

    ticket_id: str
    action: Literal["auto_answer", "escalate"]
    confidence: float
    draft_reply: str | None = None
    escalation_reason: str | None = None
    safety_flags: list[str] = Field(default_factory=list)
