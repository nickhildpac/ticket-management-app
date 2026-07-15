from __future__ import annotations

from sqlalchemy import Column, Integer, String, Table, Text, select, text
from sqlalchemy.engine import Engine

from app.ai.embeddings import Embedder
from app.ai.schemas import KBChunk

try:  # pgvector is required at runtime; guarded so imports don't explode in bare envs.
    from pgvector.sqlalchemy import Vector
except ImportError:  # pragma: no cover
    Vector = None  # type: ignore[assignment]


class VectorStore:
    """pgvector-backed knowledge-base store living in the shared Postgres.

    The table is created by an Alembic migration (kb_chunks). This class only
    reads/writes rows and runs semantic (cosine) and keyword (FTS) search.
    """

    def __init__(self, engine: Engine, embedder: Embedder) -> None:
        if Vector is None:  # pragma: no cover
            raise RuntimeError("pgvector is not installed; add 'pgvector' to dependencies")
        self.engine = engine
        self.embedder = embedder
        self.table = Table(
            "kb_chunks",
            _metadata(),
            Column("id", Integer, primary_key=True),
            Column("source", String, nullable=False),
            Column("content", Text, nullable=False),
            Column("embedding", Vector(embedder.dim), nullable=False),
            extend_existing=True,
        )

    def add(self, source: str, content: str) -> None:
        embedding = self.embedder.embed(content)
        with self.engine.begin() as conn:
            conn.execute(
                self.table.insert().values(source=source, content=content, embedding=embedding)
            )

    def search(self, query: str, k: int) -> list[KBChunk]:
        """Backward-compatible alias for semantic search."""
        return self.search_semantic(query, k)

    def search_semantic(self, query: str, k: int) -> list[KBChunk]:
        embedding = self.embedder.embed(query)
        distance = self.table.c.embedding.cosine_distance(embedding)
        stmt = (
            select(
                self.table.c.id,
                self.table.c.content,
                self.table.c.source,
                distance.label("distance"),
            )
            .order_by(distance)
            .limit(k)
        )
        with self.engine.connect() as conn:
            rows = conn.execute(stmt).all()
        return [
            KBChunk(
                id=int(r.id),
                content=r.content,
                source=r.source,
                distance=float(r.distance),
            )
            for r in rows
        ]

    def search_keyword(self, query: str, k: int) -> list[KBChunk]:
        """Postgres full-text search ranked by ``ts_rank_cd``.

        Returns an empty list when the query has no useful terms (so
        ``plainto_tsquery`` would match nothing useful / raise noise).
        Distance is set to ``1 - rank`` (clipped) so lower-is-better still holds.
        """
        q = (query or "").strip()
        if not q:
            return []

        # Use a single SQL statement so the GIN expression index can be used.
        # Rank is higher-is-better; we invert for KBChunk.distance consistency.
        sql = text(
            """
            SELECT id, content, source,
                   ts_rank_cd(to_tsvector('english', content), query) AS rank
            FROM kb_chunks,
                 plainto_tsquery('english', :q) AS query
            WHERE to_tsvector('english', content) @@ query
            ORDER BY rank DESC
            LIMIT :k
            """
        )
        with self.engine.connect() as conn:
            rows = conn.execute(sql, {"q": q, "k": k}).all()
        return [
            KBChunk(
                id=int(r.id),
                content=r.content,
                source=r.source,
                distance=max(0.0, 1.0 - float(r.rank)),
            )
            for r in rows
        ]


def _metadata():
    from sqlalchemy import MetaData

    return MetaData()
