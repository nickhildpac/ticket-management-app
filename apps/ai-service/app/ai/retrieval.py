from __future__ import annotations

from typing import Protocol

from app.ai.rerank import Reranker, score_to_distance
from app.ai.schemas import KBChunk, TicketContext
from app.ai.vectorstore import VectorStore


class Retriever(Protocol):
    def retrieve(self, ticket: TicketContext, k: int) -> list[KBChunk]: ...


def reciprocal_rank_fusion(
    ranked_lists: list[list[KBChunk]],
    *,
    rrf_k: int = 60,
) -> list[KBChunk]:
    """Fuse multiple ranked lists with Reciprocal Rank Fusion.

    ``score(doc) = Σ 1 / (rrf_k + rank)`` over lists where the doc appears
    (1-based ranks). Docs are keyed by ``id`` when present, else by
    ``(source, content)``. Returns chunks sorted by RRF score descending,
    with ``distance`` set to ``1 / (1 + rrf_score)`` (lower is better).
    """
    scores: dict[object, float] = {}
    best: dict[object, KBChunk] = {}

    for ranked in ranked_lists:
        for rank, chunk in enumerate(ranked, start=1):
            key = chunk.id if chunk.id is not None else (chunk.source, chunk.content)
            scores[key] = scores.get(key, 0.0) + 1.0 / (rrf_k + rank)
            # Prefer a copy that carries an id when available.
            prev = best.get(key)
            if prev is None or (prev.id is None and chunk.id is not None):
                best[key] = chunk

    fused = sorted(scores.items(), key=lambda item: item[1], reverse=True)
    out: list[KBChunk] = []
    for key, rrf_score in fused:
        chunk = best[key]
        out.append(
            KBChunk(
                id=chunk.id,
                content=chunk.content,
                source=chunk.source,
                distance=1.0 / (1.0 + rrf_score),
            )
        )
    return out


class VectorRetriever:
    """Retrieves knowledge-base passages via pure semantic (vector) search.

    A convenience wrapper for tests/simple callers. The triage agent instead
    calls the underlying primitives (``search_semantic``/``search_keyword`` +
    ``reciprocal_rank_fusion`` + the re-ranker) directly from its ``search_docs``
    and ``rerank_results`` tools.
    """

    def __init__(self, store: VectorStore) -> None:
        self.store = store

    def retrieve(self, ticket: TicketContext, k: int) -> list[KBChunk]:
        query = f"{ticket.title}\n\n{ticket.description}"
        return self.store.search(query, k)


class HybridRetriever:
    """Semantic + keyword retrieval → RRF fusion → CrossEncoder re-rank.

    Each lane fetches ``candidate_k`` hits; RRF merges them; the top
    ``rerank_pool`` fused docs are scored by the re-ranker; the best ``k`` are
    returned. Kept as a one-shot retriever for tests/simple callers; the triage
    agent runs the same pipeline in stages across its tool calls.
    """

    def __init__(
        self,
        store: VectorStore,
        reranker: Reranker,
        *,
        candidate_k: int = 20,
        rerank_pool: int = 10,
        rrf_k: int = 60,
    ) -> None:
        self.store = store
        self.reranker = reranker
        self.candidate_k = candidate_k
        self.rerank_pool = rerank_pool
        self.rrf_k = rrf_k

    def retrieve(self, ticket: TicketContext, k: int) -> list[KBChunk]:
        query = f"{ticket.title}\n\n{ticket.description}"
        semantic = self.store.search_semantic(query, self.candidate_k)
        keyword = self.store.search_keyword(query, self.candidate_k)
        fused = reciprocal_rank_fusion([semantic, keyword], rrf_k=self.rrf_k)
        if not fused:
            return []

        pool = fused[: self.rerank_pool]
        scores = self.reranker.score(query, [c.content for c in pool])
        scored = sorted(
            zip(pool, scores, strict=True),
            key=lambda item: item[1],
            reverse=True,
        )
        return [
            KBChunk(
                id=chunk.id,
                content=chunk.content,
                source=chunk.source,
                distance=score_to_distance(score),
            )
            for chunk, score in scored[:k]
        ]


def format_context(chunks: list[KBChunk]) -> str:
    """Render retrieved chunks into a grounded context block for the prompt."""
    if not chunks:
        return "(no relevant knowledge-base passages were found)"
    parts = []
    for i, chunk in enumerate(chunks, start=1):
        parts.append(f"[{i}] source={chunk.source}\n{chunk.content}")
    return "\n\n".join(parts)


def format_candidates(chunks: list[KBChunk]) -> str:
    """Render chunks for a tool result the model can cite.

    Labels each passage with its KB ``id`` (so the model can cite ``[id]`` and
    ``rerank_results`` can reference the same candidates); falls back to a 1-based
    index for chunks without an id (e.g. stubs).
    """
    if not chunks:
        return "(no relevant knowledge-base passages were found)"
    parts = []
    for i, chunk in enumerate(chunks, start=1):
        label = chunk.id if chunk.id is not None else i
        parts.append(
            f"[{label}] source={chunk.source} (distance={chunk.distance:.4f})\n{chunk.content}"
        )
    return "\n\n".join(parts)
