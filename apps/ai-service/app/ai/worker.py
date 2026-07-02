from __future__ import annotations

import json
import logging

import anthropic
import redis
from sqlalchemy import create_engine

from app.ai.agent import TriageAgent
from app.ai.embeddings import HashingEmbedder
from app.ai.retrieval import VectorRetriever
from app.ai.schemas import TicketContext, TriageResult
from app.ai.ticket_client import TicketServiceClient
from app.ai.vectorstore import VectorStore
from app.core.config import Settings, get_settings

logger = logging.getLogger(__name__)

STREAM_KEY = "ticket-events"
CONSUMER_NAME = "worker-1"


def build_agent(settings: Settings) -> TriageAgent:
    engine = create_engine(settings.database_url, pool_pre_ping=True)
    embedder = HashingEmbedder(dim=settings.embedding_dim)
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
    result = agent.triage(ticket)
    logger.info(
        "ticket %s -> %s (confidence %.2f)",
        ticket.ticket_id,
        result.action,
        result.confidence,
    )
    apply_result(ticket_client, result)


def run(settings: Settings | None = None) -> None:
    settings = settings or get_settings()
    rdb = redis.Redis.from_url(settings.redis_url, decode_responses=True)
    agent = build_agent(settings)
    ticket_client = TicketServiceClient(settings)

    # Create the consumer group (idempotent). mkstream so we don't need the
    # stream to pre-exist.
    try:
        rdb.xgroup_create(STREAM_KEY, settings.consumer_group, id="0", mkstream=True)
    except redis.ResponseError as exc:
        if "BUSYGROUP" not in str(exc):
            raise

    logger.info("triage worker consuming %s as group=%s", STREAM_KEY, settings.consumer_group)
    while True:
        resp = rdb.xreadgroup(
            settings.consumer_group,
            CONSUMER_NAME,
            {STREAM_KEY: ">"},
            count=10,
            block=5000,
        )
        if not resp:
            continue
        for _stream, messages in resp:
            for message_id, fields in messages:
                try:
                    process_message(agent, ticket_client, fields)
                except Exception:  # noqa: BLE001 — never let one bad event kill the loop.
                    logger.exception("failed to process message %s", message_id)
                finally:
                    # At-least-once: we ack after handling. A crash before ack
                    # re-delivers the event; applying a comment twice is tolerable.
                    rdb.xack(STREAM_KEY, settings.consumer_group, message_id)


def main() -> None:
    logging.basicConfig(level=logging.INFO)
    run()


if __name__ == "__main__":
    main()
