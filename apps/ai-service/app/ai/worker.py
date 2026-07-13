from __future__ import annotations

import json
import logging

import anthropic
import redis
from sqlalchemy import create_engine

from app.ai.agent import TriageAgent
from app.ai.embeddings import build_embedder
from app.ai.retrieval import VectorRetriever
from app.ai.schemas import TicketContext, TriageResult
from app.ai.ticket_client import TicketServiceClient
from app.ai.vectorstore import VectorStore
from app.core.config import Settings, get_settings

logger = logging.getLogger(__name__)

STREAM_KEY = "ticket-events"
DLQ_STREAM = "ticket-events-dlq"

# Ticket states where triage is pointless — the ticket is finished, so re-triaging
# a transition into one of these would only spend model tokens and post a comment
# on a closed ticket.
TERMINAL_STATES = {"resolved", "closed", "cancelled"}

# Idempotency: after a successful handling we record the event id so a re-delivered
# event (the outbox relay is at-least-once) is dropped instead of triaged twice.
_PROCESSED_PREFIX = "ai-triage:processed:"
_PROCESSED_TTL = 24 * 3600

# Pending-entry recovery: reclaim messages left un-acked by a crashed/failed handler
# after they've been idle this long; give up (dead-letter) after this many deliveries.
_RECLAIM_IDLE_MS = 60_000
_MAX_DELIVERIES = 5


def build_agent(settings: Settings) -> TriageAgent:
    engine = create_engine(settings.database_url, pool_pre_ping=True)
    embedder = build_embedder(settings)
    store = VectorStore(engine, embedder)
    retriever = VectorRetriever(store)
    client = anthropic.Anthropic(api_key=settings.anthropic_api_key)
    return TriageAgent(
        client,
        retriever,
        model=settings.triage_model,
        confidence_threshold=settings.auto_answer_confidence_threshold,
        rag_top_k=settings.rag_top_k,
    )


def apply_result(ticket_client: TicketServiceClient, result: TriageResult) -> None:
    """Apply a triage decision via the ticket API. We only post comments (valid
    from any state); we deliberately avoid state transitions to respect the FSM."""
    if result.action == "auto_answer" and result.draft_reply:
        ticket_client.add_comment(
            result.ticket_id,
            f"[AI-suggested reply]\n\n{result.draft_reply}",
        )
    else:
        reason = result.escalation_reason or "needs human review"
        flags = f" (flags: {', '.join(result.safety_flags)})" if result.safety_flags else ""
        ticket_client.add_comment(
            result.ticket_id,
            f"[AI triage] Escalated to a human: {reason}{flags}",
        )


def process_message(
    agent: TriageAgent, ticket_client: TicketServiceClient, fields: dict[str, str]
) -> None:
    """Triage one ticket event and apply the result. Raises on failure so the
    caller can leave the message un-acked for retry."""
    event_type = fields.get("event_type", "")
    if event_type not in {"ticket.created", "ticket.updated"}:
        logger.debug("ignoring event_type=%s", event_type)
        return
    payload = json.loads(fields["payload"])
    ticket = TicketContext(
        ticket_id=payload["ticket_id"],
        ticket_number=payload.get("ticket_number"),
        title=payload["title"],
        description=payload["description"],
        state=payload["state"],
        priority=payload.get("priority", "low"),
    )
    if ticket.state in TERMINAL_STATES:
        logger.info("skipping triage for ticket %s in terminal state %s", ticket.ticket_id, ticket.state)
        return
    result = agent.triage(ticket)
    logger.info(
        "ticket %s -> %s (confidence %.2f)",
        ticket.ticket_id,
        result.action,
        result.confidence,
    )
    apply_result(ticket_client, result)


def _handle(rdb, agent, ticket_client, message_id, fields) -> bool:
    """Return True if the message was handled and should be acked, False to leave
    it pending for retry. Idempotent on event_id."""
    event_id = fields.get("event_id", "")
    if event_id and rdb.exists(_PROCESSED_PREFIX + event_id):
        logger.info("dropping already-processed event %s", event_id)
        return True  # duplicate delivery — ack and move on
    process_message(agent, ticket_client, fields)
    if event_id:
        rdb.set(_PROCESSED_PREFIX + event_id, "1", ex=_PROCESSED_TTL)
    return True


def _reclaim_stale(rdb, settings: Settings, agent, ticket_client) -> None:
    """Recover messages left pending by a crashed/failed handler. Anything past
    the delivery ceiling is dead-lettered so it can't block the group forever."""
    start = "0-0"
    while True:
        # XAUTOCLAIM returns [next_start_id, [(id, fields), ...], deleted_ids]
        next_start, claimed, _ = rdb.xautoclaim(
            STREAM_KEY,
            settings.consumer_group,
            settings.consumer_name,
            min_idle_time=_RECLAIM_IDLE_MS,
            start_id=start,
            count=10,
        )
        for message_id, fields in claimed:
            pending = rdb.xpending_range(
                STREAM_KEY, settings.consumer_group, min=message_id, max=message_id, count=1
            )
            deliveries = pending[0]["times_delivered"] if pending else 1
            if deliveries > _MAX_DELIVERIES:
                logger.error("dead-lettering message %s after %d deliveries", message_id, deliveries)
                rdb.xadd(DLQ_STREAM, fields)
                rdb.xack(STREAM_KEY, settings.consumer_group, message_id)
                continue
            try:
                if _handle(rdb, agent, ticket_client, message_id, fields):
                    rdb.xack(STREAM_KEY, settings.consumer_group, message_id)
            except Exception:  # noqa: BLE001 — leave pending for the next reclaim.
                logger.exception("reclaim: failed to process %s (delivery %d)", message_id, deliveries)
        if next_start == "0-0":
            break
        start = next_start


def run(settings: Settings | None = None) -> None:
    settings = settings or get_settings()
    # socket_timeout must exceed the xreadgroup block window below (5s), or the
    # client-side socket read can time out right as the server's BLOCK is about
    # to return with no new messages — redis-py's default socket_timeout is 5s.
    rdb = redis.Redis.from_url(settings.redis_url, decode_responses=True, socket_timeout=10)
    agent = build_agent(settings)
    ticket_client = TicketServiceClient(settings)

    # Fail loudly at startup if the callback path is misconfigured, rather than
    # 403-ing on every event with the failure buried in the logs.
    try:
        ticket_client.verify_access()
        logger.info("ticket API callback verified")
    except Exception:  # noqa: BLE001
        logger.exception(
            "ticket API callback check FAILED — triage results won't be applied. "
            "Check AI_SERVICE_ACCOUNT_ID/ROLE, JWT_SECRET, and TICKET_SERVICE_URL."
        )

    try:
        rdb.xgroup_create(STREAM_KEY, settings.consumer_group, id="0", mkstream=True)
    except redis.ResponseError as exc:
        if "BUSYGROUP" not in str(exc):
            raise

    logger.info(
        "triage worker consuming %s as group=%s consumer=%s",
        STREAM_KEY,
        settings.consumer_group,
        settings.consumer_name,
    )
    while True:
        _reclaim_stale(rdb, settings, agent, ticket_client)
        resp = rdb.xreadgroup(
            settings.consumer_group,
            settings.consumer_name,
            {STREAM_KEY: ">"},
            count=10,
            block=5000,
        )
        if not resp:
            continue
        for _stream, messages in resp:
            for message_id, fields in messages:
                try:
                    if _handle(rdb, agent, ticket_client, message_id, fields):
                        rdb.xack(STREAM_KEY, settings.consumer_group, message_id)
                except Exception:  # noqa: BLE001
                    # Leave the message un-acked; it stays in the PEL and is
                    # retried (and eventually dead-lettered) by _reclaim_stale.
                    logger.exception("failed to process message %s; will retry", message_id)


def main() -> None:
    logging.basicConfig(level=logging.INFO)
    run()


if __name__ == "__main__":
    main()
