from __future__ import annotations

import logging
import math
from typing import Protocol

import httpx

logger = logging.getLogger(__name__)

DEFAULT_OPENROUTER_RERANK_URL = "https://openrouter.ai/api/v1/rerank"


class Reranker(Protocol):
    """Scores (query, passage) pairs; higher is more relevant."""

    def score(self, query: str, passages: list[str]) -> list[float]: ...


class OpenRouterReranker:
    """Cohere-style re-rank via OpenRouter's ``/api/v1/rerank`` endpoint.

    Default model is ``cohere/rerank-v3.5``. Scores are returned aligned to the
    input passage order (API results are ranked; we map ``index`` → score).
    """

    def __init__(
        self,
        api_key: str,
        *,
        model: str = "cohere/rerank-v3.5",
        base_url: str = DEFAULT_OPENROUTER_RERANK_URL,
        client: httpx.Client | None = None,
        timeout: float = 30.0,
    ) -> None:
        self.api_key = api_key
        self.model = model
        self.base_url = base_url.rstrip("/")
        self.client = client or httpx.Client(timeout=timeout)

    def score(self, query: str, passages: list[str]) -> list[float]:
        if not passages:
            return []
        if not self.api_key.strip():
            raise RuntimeError(
                "OPENROUTER_API_KEY is empty — cross-encoder re-rank requires it "
                "(model=cohere/rerank-v3.5 via OpenRouter)."
            )

        resp = self.client.post(
            self.base_url,
            headers={
                "Authorization": f"Bearer {self.api_key}",
                "Content-Type": "application/json",
            },
            json={
                "model": self.model,
                "query": query,
                "documents": passages,
                "top_n": len(passages),
            },
        )
        resp.raise_for_status()
        payload = resp.json()
        results = payload.get("results") or []

        # API returns ranked results with original document indices; fill a
        # dense score vector so callers can zip against the input passages.
        scores = [0.0] * len(passages)
        for item in results:
            idx = int(item["index"])
            if 0 <= idx < len(passages):
                scores[idx] = float(item["relevance_score"])
        return scores


def score_to_distance(score: float) -> float:
    """Map a higher-is-better re-ranker score to a lower-is-better distance."""
    # Works for both logits and [0, 1] relevance scores (monotone decreasing).
    prob = 1.0 / (1.0 + math.exp(-score))
    return 1.0 - prob
