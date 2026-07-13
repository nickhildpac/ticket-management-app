from __future__ import annotations

import hashlib
import math
import re
from typing import TYPE_CHECKING, Protocol

if TYPE_CHECKING:
    from app.core.config import Settings

_TOKEN_RE = re.compile(r"[a-z0-9]+")


class Embedder(Protocol):
    """Turns text into a fixed-length vector. Swap the implementation for a real
    embedding model (Voyage AI, sentence-transformers, OpenAI, …) in production."""

    dim: int

    def embed(self, text: str) -> list[float]: ...


class HashingEmbedder:
    """Deterministic, dependency-free bag-of-words hashing embedder.

    Good enough to exercise the RAG plumbing offline and in tests. It is *not* a
    semantic model — replace it with a real embedding provider for quality
    retrieval. Kept behind the ``Embedder`` protocol so the swap is one line.
    """

    def __init__(self, dim: int = 384) -> None:
        self.dim = dim

    def embed(self, text: str) -> list[float]:
        vec = [0.0] * self.dim
        for token in _TOKEN_RE.findall(text.lower()):
            digest = hashlib.md5(token.encode()).digest()
            bucket = int.from_bytes(digest[:4], "big") % self.dim
            sign = 1.0 if digest[4] & 1 else -1.0
            vec[bucket] += sign
        norm = math.sqrt(sum(v * v for v in vec))
        if norm > 0:
            vec = [v / norm for v in vec]
        return vec


class OpenAIEmbedder:
    """Real semantic embeddings via the OpenAI embeddings API.

    Uses the `dimensions` param to truncate text-embedding-3-* output to
    ``dim``, so it can produce vectors matching the existing pgvector column
    (kb_chunks.embedding is a fixed vector(EMBEDDING_DIM)) without a schema
    migration.
    """

    def __init__(self, api_key: str, model: str = "text-embedding-3-small", dim: int = 384) -> None:
        from openai import OpenAI

        self._client = OpenAI(api_key=api_key)
        self.model = model
        self.dim = dim

    def embed(self, text: str) -> list[float]:
        response = self._client.embeddings.create(
            model=self.model, input=text, dimensions=self.dim
        )
        return response.data[0].embedding


def build_embedder(settings: "Settings") -> Embedder:
    """Pick the configured embedder: OpenAI when an API key is set, otherwise
    fall back to the offline hashing embedder (tests / no-key local runs)."""
    if settings.openai_api_key:
        return OpenAIEmbedder(
            api_key=settings.openai_api_key,
            model=settings.embedding_model,
            dim=settings.embedding_dim,
        )
    return HashingEmbedder(dim=settings.embedding_dim)
