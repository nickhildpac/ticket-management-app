from __future__ import annotations

from uuid import UUID

from sqlalchemy import select
from sqlalchemy.orm import Session, selectinload

from app.enums.role import UserRole
from app.models.association import UserSkill
from app.models.user import User


class UserRepository:
    def __init__(self, session: Session):
        self.session = session

    def create(self, user: User) -> User:
        self.session.add(user)
        self.session.flush()
        self.session.refresh(user)
        return user

    def get_by_id(self, user_id: UUID) -> User | None:
        stmt = select(User).where(User.id == user_id).options(selectinload(User.skills))
        return self.session.scalar(stmt)

    def get_by_email(self, email: str) -> User | None:
        stmt = select(User).where(User.email == email).options(selectinload(User.skills))
        return self.session.scalar(stmt)

    def list_all(self) -> list[User]:
        stmt = select(User).order_by(User.created_at.desc()).options(selectinload(User.skills))
        return list(self.session.scalars(stmt).all())

    def list_assignment_candidates(self) -> list[User]:
        stmt = (
            select(User)
            .where(User.is_active.is_(True), User.role.in_([UserRole.AGENT, UserRole.ADMIN]))
            .order_by(User.created_at.desc())
            .options(selectinload(User.skills))
        )
        return list(self.session.scalars(stmt).all())

    def delete(self, user: User) -> None:
        self.session.delete(user)

    def set_skills(self, user: User, skills: list[str]) -> None:
        user.skills = [UserSkill(user_id=user.id, skill=skill) for skill in skills]

    def get_by_ids(self, user_ids: list[UUID]) -> list[User]:
        if not user_ids:
            return []
        stmt = select(User).where(User.id.in_(user_ids))
        return list(self.session.scalars(stmt).all())
