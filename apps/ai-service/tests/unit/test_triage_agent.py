from __future__ import annotations

import logging

from app.ai.agent import TriageAgent
from app.ai.schemas import KBChunk, TicketContext, TriageDecision

TICKET = TicketContext(
    ticket_id="11111111-1111-1111-1111-111111111111",
    ticket_number=42,
    title="How do I reset my password?",
    description="I forgot my password and can't log in.",
    state="open",
    priority="low",
)


class StubRetriever:
    def retrieve(self, ticket, k):
        return [
            KBChunk(
                content="Use the reset link on the login page.",
                source="kb/faq.md",
                distance=0.1,
            )
        ]


class _Messages:
    def __init__(self, response):
        self._response = response

    def parse(self, **_kwargs):
        if isinstance(self._response, Exception):
            raise self._response
        return self._response


class StubClient:
    def __init__(self, response):
        self.messages = _Messages(response)


class StubResponse:
    def __init__(self, parsed_output, stop_reason="end_turn"):
        self.parsed_output = parsed_output
        self.stop_reason = stop_reason


def _agent(response, threshold=0.75):
    return TriageAgent(
        StubClient(response),
        StubRetriever(),
        model="claude-opus-4-8",
        confidence_threshold=threshold,
    )


def test_auto_answer_when_confident_and_clean():
    decision = TriageDecision(
        action="auto_answer", confidence=0.9, draft_reply="Use the reset link."
    )
    result = _agent(StubResponse(decision)).triage(TICKET)
    assert result.action == "auto_answer"
    assert result.draft_reply == "Use the reset link."


def test_escalates_when_below_confidence_threshold():
    decision = TriageDecision(
        action="auto_answer", confidence=0.5, draft_reply="Maybe try resetting."
    )
    result = _agent(StubResponse(decision)).triage(TICKET)
    assert result.action == "escalate"
    assert "confidence" in (result.escalation_reason or "")


def test_escalates_when_safety_flags_present():
    decision = TriageDecision(
        action="auto_answer",
        confidence=0.99,
        draft_reply="I refunded your account.",
        safety_flags=["refund"],
    )
    result = _agent(StubResponse(decision)).triage(TICKET)
    assert result.action == "escalate"
    assert result.safety_flags == ["refund"]


def test_escalates_when_auto_answer_has_no_draft():
    decision = TriageDecision(action="auto_answer", confidence=0.95, draft_reply="  ")
    result = _agent(StubResponse(decision)).triage(TICKET)
    assert result.action == "escalate"


def test_refusal_stop_reason_escalates():
    decision = TriageDecision(action="auto_answer", confidence=0.99, draft_reply="anything")
    result = _agent(StubResponse(decision, stop_reason="refusal")).triage(TICKET)
    assert result.action == "escalate"
    assert "model_refusal" in result.safety_flags


def test_model_exception_fails_safe_to_escalate():
    result = _agent(RuntimeError("boom")).triage(TICKET)
    assert result.action == "escalate"
    assert "model_error" in result.safety_flags


def test_retrieved_chunks_are_logged(caplog):
    decision = TriageDecision(action="escalate", confidence=0.4)

    with caplog.at_level(logging.INFO, logger="app.ai.agent"):
        _agent(StubResponse(decision)).triage(TICKET)

    assert "retrieved 1 knowledge-base chunk(s)" in caplog.text
    assert "source=kb/faq.md" in caplog.text
    assert "distance=0.1000" in caplog.text
    assert "Use the reset link on the login page." in caplog.text
