from __future__ import annotations

from uuid import UUID

from sqlalchemy import select
from sqlalchemy.orm import Session, joinedload

from app.models.comment import Comment


class CommentRepository:
    def __init__(self, session: Session):
        self.session = session

    def create(self, comment: Comment) -> Comment:
        self.session.add(comment)
        self.session.flush()
        self.session.refresh(comment)
        return comment

    def get_by_id(self, comment_id: UUID) -> Comment | None:
        stmt = (
            select(Comment)
            .where(Comment.id == comment_id)
            .options(joinedload(Comment.creator), joinedload(Comment.ticket))
        )
        return self.session.scalar(stmt)

    def list_by_ticket(self, ticket_id: UUID, limit: int = 100, offset: int = 0) -> list[Comment]:
        stmt = (
            select(Comment)
            .where(Comment.ticket_id == ticket_id)
            .order_by(Comment.created_at.asc())
            .limit(limit)
            .offset(offset)
            .options(joinedload(Comment.creator))
        )
        return list(self.session.scalars(stmt).all())
