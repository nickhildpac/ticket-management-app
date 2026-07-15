"""add GIN full-text search index on kb_chunks.content

Revision ID: 0003_kb_chunks_fts
Revises: 0002_add_kb_chunks
Create Date: 2026-07-14 00:00:00

Supports the keyword lane of hybrid RAG retrieval (semantic + FTS + RRF).
"""

from typing import Sequence

from alembic import op

revision: str = "0003_kb_chunks_fts"
down_revision: str | None = "0002_add_kb_chunks"
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    op.execute(
        "CREATE INDEX IF NOT EXISTS idx_kb_chunks_content_fts "
        "ON kb_chunks USING GIN (to_tsvector('english', content))"
    )


def downgrade() -> None:
    op.execute("DROP INDEX IF EXISTS idx_kb_chunks_content_fts")
