from __future__ import annotations

import pytest

from app.ai.rerank import OpenRouterReranker, score_to_distance
from app.ai.retrieval import HybridRetriever, reciprocal_rank_fusion
from app.ai.schemas import KBChunk, TicketContext


def _chunk(id: int, content: str, source: str = "kb.md", distance: float = 0.1) -> KBChunk:
    return KBChunk(id=id, content=content, source=source, distance=distance)


def test_rrf_prefers_docs_ranked_high_in_both_lists():
    # Doc 1 is #1 in both → highest RRF; doc 2 only in semantic; doc 3 only in keyword.
    semantic = [_chunk(1, "a"), _chunk(2, "b")]
    keyword = [_chunk(1, "a"), _chunk(3, "c")]
    fused = reciprocal_rank_fusion([semantic, keyword], rrf_k=60)
    assert [c.id for c in fused] == [1, 2, 3]
    # Doc 1 distance must be strictly better (lower) than exclusive docs.
    assert fused[0].distance < fused[1].distance
    assert fused[0].distance < fused[2].distance


def test_rrf_handles_empty_lanes():
    assert reciprocal_rank_fusion([[], []], rrf_k=60) == []
    only_semantic = reciprocal_rank_fusion([[_chunk(9, "solo")], []], rrf_k=60)
    assert [c.id for c in only_semantic] == [9]


def test_rrf_falls_back_to_source_content_key_without_id():
    a = KBChunk(content="same", source="s", distance=0.2)
    b = KBChunk(content="same", source="s", distance=0.1)
    fused = reciprocal_rank_fusion([[a], [b]], rrf_k=60)
    assert len(fused) == 1
    assert fused[0].content == "same"


def test_score_to_distance_is_monotone_inverted():
    assert score_to_distance(5.0) < score_to_distance(0.0) < score_to_distance(-5.0)


def test_openrouter_reranker_maps_scores_to_input_order():
    class FakeResponse:
        def raise_for_status(self):
            return None

        def json(self):
            # Ranked order differs from input; indices map back.
            return {
                "results": [
                    {"index": 1, "relevance_score": 0.95},
                    {"index": 0, "relevance_score": 0.4},
                    {"index": 2, "relevance_score": 0.1},
                ]
            }

    class FakeClient:
        def __init__(self) -> None:
            self.last_json = None

        def post(self, url, headers=None, json=None):
            self.last_json = json
            return FakeResponse()

    client = FakeClient()
    reranker = OpenRouterReranker("test-key", model="cohere/rerank-v3.5", client=client)
    scores = reranker.score("password reset", ["a", "b", "c"])
    assert scores == [0.4, 0.95, 0.1]
    assert client.last_json["model"] == "cohere/rerank-v3.5"
    assert client.last_json["top_n"] == 3


def test_openrouter_reranker_requires_api_key():
    reranker = OpenRouterReranker("")
    with pytest.raises(RuntimeError, match="OPENROUTER_API_KEY"):
        reranker.score("q", ["doc"])


def test_hybrid_retriever_reranks_to_final_top_k():
    class StubStore:
        def search_semantic(self, query, k):
            return [
                _chunk(1, "password reset link", distance=0.1),
                _chunk(2, "billing invoice pdf", distance=0.2),
                _chunk(3, "forgot password steps", distance=0.3),
                _chunk(4, "unrelated hardware tip", distance=0.4),
            ][:k]

        def search_keyword(self, query, k):
            return [
                _chunk(3, "forgot password steps", distance=0.05),
                _chunk(1, "password reset link", distance=0.1),
                _chunk(5, "password manager export", distance=0.2),
            ][:k]

    class StubReranker:
        def score(self, query, passages):
            # Prefer "forgot password steps", then reset link, then others low.
            out = []
            for p in passages:
                if "forgot password" in p:
                    out.append(4.0)
                elif "reset link" in p:
                    out.append(2.0)
                else:
                    out.append(-1.0)
            return out

    retriever = HybridRetriever(
        StubStore(),
        StubReranker(),
        candidate_k=10,
        rerank_pool=5,
        rrf_k=60,
    )
    ticket = TicketContext(
        ticket_id="t1",
        title="Password reset",
        description="I forgot my password",
        state="open",
        priority="low",
    )
    result = retriever.retrieve(ticket, k=3)
    assert len(result) == 3
    assert result[0].id == 3
    assert result[1].id == 1
    assert result[0].distance < result[1].distance
