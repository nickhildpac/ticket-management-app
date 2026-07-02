from __future__ import annotations

from sqlalchemy import Column, Integer, String, Table, Text, select
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
    reads/writes rows and runs cosine-distance similarity search.
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
        embedding = self.embedder.embed(query)
        distance = self.table.c.embedding.cosine_distance(embedding)
        stmt = (
            select(self.table.c.content, self.table.c.source, distance.label("distance"))
            .order_by(distance)
            .limit(k)
        )
        with self.engine.connect() as conn:
            rows = conn.execute(stmt).all()
        return [
            KBChunk(content=r.content, source=r.source, distance=float(r.distance))
            for r in rows
        ]


def _metadata():
    from sqlalchemy import MetaData

    return MetaData()
