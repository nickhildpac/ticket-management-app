from __future__ import annotations

import json

import pytest

from app.ai.agent import TriageAgent
from app.ai.retrieval import VectorRetriever
from app.ai.schemas import KBChunk, TriageDecision, TriageResult
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


def test_created_ticket_retrieves_chunks_and_adds_them_to_model_context():
    class RecordingStore:
        def __init__(self) -> None:
            self.calls = []

        def search(self, query, k):
            self.calls.append((query, k))
            return [
                KBChunk(
                    source="kb/password-reset.md",
                    content="Use the reset link on the login page.",
                    distance=0.08,
                )
            ]

    class RecordingMessages:
        def __init__(self) -> None:
            self.parse_kwargs = None

        def parse(self, **kwargs):
            self.parse_kwargs = kwargs
            return type(
                "Response",
                (),
                {
                    "stop_reason": "end_turn",
                    "parsed_output": TriageDecision(
                        action="auto_answer",
                        confidence=0.95,
                        draft_reply="Use the reset link.",
                    ),
                },
            )()

    class RecordingClient:
        def __init__(self) -> None:
            self.messages = RecordingMessages()

    store = RecordingStore()
    retriever = VectorRetriever(store)
    model_client = RecordingClient()
    agent = TriageAgent(
        model_client,
        retriever,
        model="test-model",
        confidence_threshold=0.75,
        rag_top_k=3,
    )
    ticket_client = StubTicketClient()

    process_message(agent, ticket_client, _fields())

    assert len(store.calls) == 1
    retrieval_query, top_k = store.calls[0]
    assert retrieval_query == "Password reset\n\nI forgot my password."
    assert top_k == 3

    prompt = model_client.messages.parse_kwargs["messages"][0]["content"]
    assert "source=kb/password-reset.md" in prompt
    assert "Use the reset link on the login page." in prompt
    assert len(ticket_client.comments) == 1
    assert "AI-suggested reply" in ticket_client.comments[0][1]
