"""add pgvector kb_chunks table for RAG

Revision ID: 0002_add_kb_chunks
Revises: 0001_initial_spec1
Create Date: 2026-07-02 00:00:00

Note: under ADR 0002 the Go ticket-service owns the ticket/user/comment schema.
This service's Alembic history now only adds AI-owned tables. The embedding
dimension (384) must match Settings.embedding_dim; changing it needs a new
migration.
"""

from typing import Sequence

from alembic import op

revision: str = "0002_add_kb_chunks"
down_revision: str | None = "0001_initial_spec1"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None

EMBEDDING_DIM = 384


def upgrade() -> None:
    op.execute("CREATE EXTENSION IF NOT EXISTS vector")
    op.execute(
        f"""
        CREATE TABLE IF NOT EXISTS kb_chunks (
            id        SERIAL PRIMARY KEY,
            source    TEXT NOT NULL,
            content   TEXT NOT NULL,
            embedding vector({EMBEDDING_DIM}) NOT NULL,
            created_at timestamptz NOT NULL DEFAULT now()
        )
        """
    )
    # Approximate-nearest-neighbour index over cosine distance.
    op.execute(
        "CREATE INDEX IF NOT EXISTS idx_kb_chunks_embedding "
        "ON kb_chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)"
    )


def downgrade() -> None:
    op.execute("DROP INDEX IF EXISTS idx_kb_chunks_embedding")
    op.execute("DROP TABLE IF EXISTS kb_chunks")
