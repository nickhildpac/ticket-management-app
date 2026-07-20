from __future__ import annotations

import logging

from app.ai.agent import TriageAgent
from app.ai.schemas import KBChunk, TicketContext

TICKET = TicketContext(
    ticket_id="11111111-1111-1111-1111-111111111111",
    ticket_number=42,
    title="How do I reset my password?",
    description="I forgot my password and can't log in.",
    state="open",
    priority="low",
)


class StubStore:
    """A one-passage KB so search_docs/rerank_results run against real code."""

    _CHUNK = KBChunk(
        id=1,
        content="Use the reset link on the login page.",
        source="kb/faq.md",
        distance=0.1,
    )

    def search_semantic(self, query, k):
        return [self._CHUNK]

    def search_keyword(self, query, k):
        return [KBChunk(id=1, content=self._CHUNK.content, source=self._CHUNK.source, distance=0.2)]


class StubReranker:
    def score(self, query, passages):
        return [1.0 for _ in passages]


class FakeMessage:
    def __init__(self, stop_reason="tool_use"):
        self.stop_reason = stop_reason
        self.content = []


class FakeRunner:
    """Stands in for the SDK Tool Runner: yields one assistant message per scripted
    step and, when the triage loop calls ``generate_tool_call_response()``, executes
    that step's tool against the agent's real closures (setting session state)."""

    def __init__(self, funcs, script):
        self._by_name = {f.__name__: f for f in funcs}
        self._script = script
        self._i = -1

    def __iter__(self):
        return self

    def __next__(self):
        self._i += 1
        if self._i >= len(self._script):
            raise StopIteration
        return FakeMessage(stop_reason=self._script[self._i].get("stop_reason", "tool_use"))

    def generate_tool_call_response(self):
        step = self._script[self._i]
        tool = step.get("tool")
        if tool is None:
            return None
        self._by_name[tool](**step.get("kwargs", {}))
        return {"role": "user", "content": "tool result"}


def _factory(script):
    def factory(funcs, *, system, messages, max_tokens):
        return FakeRunner(funcs, script)

    return factory


def _raising_factory(exc):
    def factory(funcs, *, system, messages, max_tokens):
        raise exc

    return factory


def _agent(script_or_factory, threshold=0.75):
    factory = script_or_factory if callable(script_or_factory) else _factory(script_or_factory)
    return TriageAgent(
        object(),  # client is bypassed by the injected runner factory
        StubStore(),
        StubReranker(),
        model="claude-opus-4-8",
        confidence_threshold=threshold,
        runner_factory=factory,
    )


def _draft(**kwargs):
    return [{"tool": "draft_reply", "kwargs": kwargs}]


def test_auto_answer_when_confident_and_clean():
    result = _agent(_draft(reply="Use the reset link.", confidence=0.9)).triage(TICKET)
    assert result.action == "auto_answer"
    assert result.draft_reply == "Use the reset link."


def test_escalates_when_below_confidence_threshold():
    result = _agent(_draft(reply="Maybe try resetting.", confidence=0.72)).triage(TICKET)
    assert result.action == "escalate"
    # The score is still available on the result for logging...
    assert result.confidence == 0.72
    # ...but the customer-facing reason must not leak the metric or threshold.
    reason = result.escalation_reason or ""
    assert "0.72" not in reason
    assert "0.75" not in reason
    assert "threshold" not in reason.lower()


def test_gate_escalation_reason_hides_safety_flag_slugs():
    result = _agent(
        _draft(reply="I refunded your account.", confidence=0.99, safety_flags=["refund_or_cancellation"])
    ).triage(TICKET)
    assert result.action == "escalate"
    # Flags are preserved on the result (for logging) but not surfaced in the reason text.
    assert result.safety_flags == ["refund_or_cancellation"]
    assert "refund_or_cancellation" not in (result.escalation_reason or "")


def test_escalates_when_safety_flags_present():
    result = _agent(
        _draft(reply="I refunded your account.", confidence=0.99, safety_flags=["refund_or_cancellation"])
    ).triage(TICKET)
    assert result.action == "escalate"
    assert result.safety_flags == ["refund_or_cancellation"]


def test_documented_self_service_auto_answers_without_flags():
    # A 403-style ticket whose fix is documented self-service configuration:
    # no sensitive action, a cited draft, no safety flags -> auto_answer.
    result = _agent(
        _draft(
            reply="Enable the Analytics API Access toggle and confirm the key has the "
            "read:analytics scope [1]. A human will follow up if that doesn't resolve it.",
            confidence=0.85,
            safety_flags=[],
        )
    ).triage(TICKET)
    assert result.action == "auto_answer"
    assert result.safety_flags == []
    assert "[1]" in (result.draft_reply or "")


def test_account_lockout_escalates_despite_high_confidence():
    result = _agent(
        _draft(
            reply="Regaining access requires an account action.",
            confidence=0.98,
            safety_flags=["account_access_lockout"],
        )
    ).triage(TICKET)
    assert result.action == "escalate"
    assert result.safety_flags == ["account_access_lockout"]


def test_escalates_when_auto_answer_has_no_draft():
    result = _agent(_draft(reply="  ", confidence=0.95)).triage(TICKET)
    assert result.action == "escalate"


def test_escalate_ticket_tool_hands_off_to_human():
    script = [{"tool": "escalate_ticket", "kwargs": {"reason": "A teammate will review this refund."}}]
    result = _agent(script).triage(TICKET)
    assert result.action == "escalate"
    assert result.escalation_reason == "A teammate will review this refund."


def test_search_then_rerank_then_draft():
    script = [
        {"tool": "search_docs", "kwargs": {"query": "reset password"}},
        {"tool": "rerank_results", "kwargs": {"query": "reset password"}},
        {"tool": "draft_reply", "kwargs": {"reply": "Use the reset link [1].", "confidence": 0.9}},
    ]
    result = _agent(script).triage(TICKET)
    assert result.action == "auto_answer"
    assert result.draft_reply == "Use the reset link [1]."


def test_refusal_stop_reason_escalates():
    result = _agent([{"stop_reason": "refusal"}]).triage(TICKET)
    assert result.action == "escalate"
    assert "model_refusal" in result.safety_flags


def test_model_exception_fails_safe_to_escalate():
    result = _agent(_raising_factory(RuntimeError("boom"))).triage(TICKET)
    assert result.action == "escalate"
    assert "model_error" in result.safety_flags


def test_no_terminal_decision_fails_safe_to_escalate():
    # The model searches but never makes a terminal decision.
    script = [{"tool": "search_docs", "kwargs": {"query": "reset password"}}]
    result = _agent(script).triage(TICKET)
    assert result.action == "escalate"
    assert "no_decision" in result.safety_flags


def test_max_iterations_without_decision_escalates():
    # More search turns than the iteration cap, no terminal decision -> fail safe.
    script = [{"tool": "search_docs", "kwargs": {"query": "reset password"}} for _ in range(10)]
    agent = TriageAgent(
        object(),
        StubStore(),
        StubReranker(),
        model="claude-opus-4-8",
        confidence_threshold=0.75,
        max_iterations=3,
        runner_factory=_factory(script),
    )
    result = agent.triage(TICKET)
    assert result.action == "escalate"
    assert "no_decision" in result.safety_flags


def test_search_docs_logs_retrieved_passages(caplog):
    script = [
        {"tool": "search_docs", "kwargs": {"query": "reset password"}},
        {"tool": "escalate_ticket", "kwargs": {"reason": "A teammate will follow up."}},
    ]
    with caplog.at_level(logging.INFO, logger="app.ai.agent"):
        _agent(script).triage(TICKET)

    assert "search_docs" in caplog.text
    assert "returned 1 passage(s)" in caplog.text
    assert "source=kb/faq.md" in caplog.text
    assert "Use the reset link on the login page." in caplog.text
