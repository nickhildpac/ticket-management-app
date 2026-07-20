from __future__ import annotations

import json

import pytest

from app.ai.agent import TriageAgent
from app.ai.schemas import KBChunk, TriageResult
from app.ai.worker import _PROCESSED_PREFIX, _handle, apply_result, process_message


class StubRedis:
    """Just enough of the redis client for _handle: the dedupe key store."""

    def __init__(self) -> None:
        self.store: dict[str, str] = {}

    def exists(self, key: str) -> int:
        return 1 if key in self.store else 0

    def set(self, key: str, value: str, ex: int | None = None) -> None:
        self.store[key] = value


class StubAgent:
    def __init__(self, result: TriageResult | Exception) -> None:
        self._result = result
        self.triaged: list[str] = []

    def triage(self, ticket) -> TriageResult:
        if isinstance(self._result, Exception):
            raise self._result
        self.triaged.append(ticket.ticket_id)
        return self._result


class StubTicketClient:
    def __init__(self, fail: bool = False) -> None:
        self.fail = fail
        self.comments: list[tuple[str, str]] = []

    def add_comment(self, ticket_id: str, description: str) -> None:
        if self.fail:
            raise RuntimeError("ticket API unavailable")
        self.comments.append((ticket_id, description))


def _fields(event_id: str = "evt-1", state: str = "open") -> dict[str, str]:
    return {
        "event_id": event_id,
        "event_type": "ticket.created",
        "payload": json.dumps(
            {
                "ticket_id": "11111111-1111-1111-1111-111111111111",
                "ticket_number": 42,
                "title": "Password reset",
                "description": "I forgot my password.",
                "state": state,
                "priority": "low",
            }
        ),
    }


def _escalation() -> TriageResult:
    return TriageResult(
        ticket_id="11111111-1111-1111-1111-111111111111",
        action="escalate",
        confidence=0.0,
        escalation_reason="needs human review",
    )


def test_handle_success_marks_processed_and_acks():
    rdb = StubRedis()
    agent = StubAgent(_escalation())
    client = StubTicketClient()

    assert _handle(rdb, agent, client, "1-0", _fields()) is True
    assert _PROCESSED_PREFIX + "evt-1" in rdb.store
    assert len(client.comments) == 1


def test_handle_drops_duplicate_event_without_retriaging():
    rdb = StubRedis()
    rdb.store[_PROCESSED_PREFIX + "evt-1"] = "1"
    agent = StubAgent(_escalation())
    client = StubTicketClient()

    assert _handle(rdb, agent, client, "1-0", _fields()) is True  # ack the duplicate
    assert agent.triaged == []
    assert client.comments == []


def test_handle_failure_raises_and_does_not_mark_processed():
    rdb = StubRedis()
    agent = StubAgent(_escalation())
    client = StubTicketClient(fail=True)

    with pytest.raises(RuntimeError):
        _handle(rdb, agent, client, "1-0", _fields())
    # Not marked processed: the message stays pending and a retry re-triages.
    assert _PROCESSED_PREFIX + "evt-1" not in rdb.store


def test_process_message_skips_terminal_states():
    agent = StubAgent(_escalation())
    client = StubTicketClient()

    process_message(agent, client, _fields(state="closed"))
    assert agent.triaged == []
    assert client.comments == []


def test_process_message_ignores_unknown_event_types():
    agent = StubAgent(_escalation())
    client = StubTicketClient()

    process_message(agent, client, {"event_type": "ticket.deleted", "payload": "{}"})
    assert agent.triaged == []


def test_apply_result_posts_ai_reply_for_auto_answer():
    client = StubTicketClient()
    apply_result(
        client,
        TriageResult(
            ticket_id="t-1", action="auto_answer", confidence=0.9, draft_reply="Use the reset link."
        ),
    )
    assert client.comments[0][0] == "t-1"
    assert "AI-suggested reply" in client.comments[0][1]


def test_apply_result_escalation_comment_hides_internal_flags():
    client = StubTicketClient()
    apply_result(
        client,
        TriageResult(
            ticket_id="t-1",
            action="escalate",
            confidence=0.0,
            escalation_reason="This request needs review by our team before we can respond.",
            safety_flags=["refund_or_cancellation"],
        ),
    )
    comment = client.comments[0][1]
    assert "Escalated to a human" in comment
    # The end user must not see raw safety-flag slugs in the comment.
    assert "refund_or_cancellation" not in comment
    assert "flags:" not in comment


def test_created_ticket_drives_the_tool_loop_and_posts_ai_reply():
    """End-to-end through the worker: a created ticket runs the agentic loop, the
    model's search_docs tool hits the store, and an auto-answer comment is posted."""

    class RecordingStore:
        def __init__(self) -> None:
            self.semantic_calls = []
            self.keyword_calls = []

        def search_semantic(self, query, k):
            self.semantic_calls.append((query, k))
            return [
                KBChunk(
                    id=1,
                    source="kb/password-reset.md",
                    content="Use the reset link on the login page.",
                    distance=0.08,
                )
            ]

        def search_keyword(self, query, k):
            self.keyword_calls.append((query, k))
            return [
                KBChunk(
                    id=1,
                    source="kb/password-reset.md",
                    content="Use the reset link on the login page.",
                    distance=0.05,
                )
            ]

    class StubReranker:
        def score(self, query, passages):
            return [1.0] * len(passages)

    # Fake Tool Runner: the model searches, then drafts a cited reply.
    class FakeMessage:
        stop_reason = "tool_use"

    class FakeRunner:
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
            return FakeMessage()

        def generate_tool_call_response(self):
            tool, kwargs = self._script[self._i]
            self._by_name[tool](**kwargs)
            return {"role": "user", "content": "ok"}

    script = [
        ("search_docs", {"query": "reset password"}),
        ("draft_reply", {"reply": "Use the reset link [1].", "confidence": 0.95}),
    ]

    def runner_factory(funcs, *, system, messages, max_tokens):
        return FakeRunner(funcs, script)

    store = RecordingStore()
    agent = TriageAgent(
        object(),
        store,
        StubReranker(),
        model="test-model",
        confidence_threshold=0.75,
        candidate_k=3,
        runner_factory=runner_factory,
    )
    ticket_client = StubTicketClient()

    process_message(agent, ticket_client, _fields())

    # The model's search_docs call reached both retrieval lanes with candidate_k.
    assert len(store.semantic_calls) == 1
    assert len(store.keyword_calls) == 1
    assert store.semantic_calls[0] == ("reset password", 3)
    # ...and the drafted reply was posted as a suggested comment.
    assert len(ticket_client.comments) == 1
    assert "AI-suggested reply" in ticket_client.comments[0][1]
    assert "Use the reset link [1]." in ticket_client.comments[0][1]
