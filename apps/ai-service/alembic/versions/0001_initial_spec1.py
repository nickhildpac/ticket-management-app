"""baseline (no-op)

Revision ID: 0001_initial_spec1
Revises:
Create Date: 2026-04-12 00:00:00

Under ADR 0002 the Go ticket-service owns the ticket/user/comment schema and
creates it via its own migrations. This revision — which used to create that
schema back when FastAPI was authoritative — is now an intentional no-op baseline
so Alembic (which here only manages AI-owned tables like kb_chunks, see 0002)
does not conflict with the tables the Go service creates.

The original DDL is preserved in git history if ever needed.
"""

from typing import Sequence

# revision identifiers, used by Alembic.
revision: str = "0001_initial_spec1"
down_revision: str | None = None
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    pass


def downgrade() -> None:
    pass
