from __future__ import annotations

from datetime import datetime

from sqlalchemy import select
from sqlalchemy.orm import Session

from app.models.refresh_token import RefreshToken


class RefreshTokenRepository:
    def __init__(self, session: Session):
        self.session = session

    def create(self, token: RefreshToken) -> RefreshToken:
        self.session.add(token)
        self.session.flush()
        self.session.refresh(token)
        return token

    def get_active_by_hash(self, token_hash: str, now: datetime) -> RefreshToken | None:
        stmt = select(RefreshToken).where(
            RefreshToken.token_hash == token_hash,
            RefreshToken.revoked_at.is_(None),
            RefreshToken.expires_at > now,
        )
        return self.session.scalar(stmt)
