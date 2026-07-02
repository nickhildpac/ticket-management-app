from __future__ import annotations

from logging.config import fileConfig

from alembic.script import ScriptDirectory
from sqlalchemy import engine_from_config, inspect, pool, text

from alembic import context
from app.core.config import get_settings
from app.models import (  # noqa: F401
    Base,  # noqa: F401
    association,
    comment,
    refresh_token,
    ticket,
    user,
)

config = context.config

if config.config_file_name is not None:
    fileConfig(config.config_file_name)

settings = get_settings()
config.set_main_option("sqlalchemy.url", settings.database_url)

target_metadata = Base.metadata


def _stamp_head_if_schema_bootstrapped_by_sql(connection) -> bool:
    """Docker Compose mounts Go SQL under initdb.d, which creates tables before Alembic runs.

    If core tables exist but alembic_version does not, record head so upgrade is a no-op.
    """
    insp = inspect(connection)
    if insp.has_table("alembic_version"):
        return False
    if not insp.has_table("users"):
        return False

    script = ScriptDirectory.from_config(config)
    head = script.get_current_head()
    if head is None:
        return False

    connection.execute(
        text(
            """
            CREATE TABLE alembic_version (
                version_num VARCHAR(32) NOT NULL,
                CONSTRAINT alembic_version_pkc PRIMARY KEY (version_num)
            )
            """
        )
    )
    connection.execute(text("INSERT INTO alembic_version (version_num) VALUES (:rev)"), {"rev": head})
    return True


def run_migrations_offline() -> None:
    url = config.get_main_option("sqlalchemy.url")
    context.configure(
        url=url,
        target_metadata=target_metadata,
        literal_binds=True,
        dialect_opts={"paramstyle": "named"},
    )

    with context.begin_transaction():
        context.run_migrations()


def run_migrations_online() -> None:
    connectable = engine_from_config(
        config.get_section(config.config_ini_section, {}),
        prefix="sqlalchemy.",
        poolclass=pool.NullPool,
    )

    with connectable.connect() as connection:
        if _stamp_head_if_schema_bootstrapped_by_sql(connection):
            connection.commit()
            return

        context.configure(connection=connection, target_metadata=target_metadata)

        with context.begin_transaction():
            context.run_migrations()


if context.is_offline_mode():
    run_migrations_offline()
else:
    run_migrations_online()
