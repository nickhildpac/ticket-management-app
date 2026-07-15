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
        ge=0.0, le=1.0, description="Confidence in the drafted answer, 0..1."
    )
    draft_reply: str | None = Field(
        default=None, description="Proposed reply to the customer when action is auto_answer."
    )
    escalation_reason: str | None = Field(
        default=None, description="Why a human is needed when action is escalate."
    )
    safety_flags: list[str] = Field(
        default_factory=list,
        description="Non-empty if the request touches anything unsafe/sensitive "
        "(account changes, refunds, legal, security, self-harm, PII exposure, etc.).",
    )


class TriageResult(BaseModel):
    """Final decision after applying the deterministic safety gate."""

    ticket_id: str
    action: Literal["auto_answer", "escalate"]
    confidence: float
    draft_reply: str | None = None
    escalation_reason: str | None = None
    safety_flags: list[str] = Field(default_factory=list)
