from __future__ import annotations

import pytest
from sqlalchemy import inspect
from sqlalchemy.orm import Session

pytestmark = pytest.mark.integration


def test_initial_migration_creates_required_tables(db_session: Session):
    inspector = inspect(db_session.bind)
    tables = set(inspector.get_table_names())
    expected = {
        "users",
        "user_skills",
        "tickets",
        "ticket_skills",
        "ticket_assignees",
        "comments",
        "refresh_tokens",
    }
    assert expected.issubset(tables)
