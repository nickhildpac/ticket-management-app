from __future__ import annotations

from typing import Protocol

from app.ai.schemas import KBChunk, TicketContext
from app.ai.vectorstore import VectorStore


class Retriever(Protocol):
    def retrieve(self, ticket: TicketContext, k: int) -> list[KBChunk]: ...


class VectorRetriever:
    """Retrieves knowledge-base passages relevant to a ticket via the vector store."""

    def __init__(self, store: VectorStore) -> None:
        self.store = store

    def retrieve(self, ticket: TicketContext, k: int) -> list[KBChunk]:
        query = f"{ticket.title}\n\n{ticket.description}"
        return self.store.search(query, k)


def format_context(chunks: list[KBChunk]) -> str:
    """Render retrieved chunks into a grounded context block for the prompt."""
    if not chunks:
        return "(no relevant knowledge-base passages were found)"
    parts = []
    for i, chunk in enumerate(chunks, start=1):
        parts.append(f"[{i}] source={chunk.source}\n{chunk.content}")
    return "\n\n".join(parts)
