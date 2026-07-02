from __future__ import annotations

from typing import TYPE_CHECKING
from uuid import UUID, uuid4

from sqlalchemy import Boolean, String, text
from sqlalchemy.dialects.postgresql import UUID as PG_UUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.enums.role import UserRole
from app.models.base import Base, TimestampMixin
from app.models.types import user_role_enum

if TYPE_CHECKING:
    from app.models.association import TicketAssignee, UserSkill
    from app.models.comment import Comment
    from app.models.refresh_token import RefreshToken
    from app.models.ticket import Ticket


class User(Base, TimestampMixin):
    __tablename__ = "users"

    id: Mapped[UUID] = mapped_column(
        PG_UUID(as_uuid=True),
        primary_key=True,
        default=uuid4,
        server_default=text("gen_random_uuid()"),
    )
    email: Mapped[str] = mapped_column(String(320), unique=True, index=True, nullable=False)
    password_hash: Mapped[str] = mapped_column(String, nullable=False)
    first_name: Mapped[str] = mapped_column(String(100), nullable=False)
    last_name: Mapped[str] = mapped_column(String(100), nullable=False)
    role: Mapped[UserRole] = mapped_column(
        user_role_enum,
        nullable=False,
        default=UserRole.USER,
        server_default=UserRole.USER.value,
    )
    is_active: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True, server_default=text("true"))

    skills: Mapped[list[UserSkill]] = relationship(
        back_populates="user", cascade="all, delete-orphan", lazy="selectin"
    )

    created_tickets: Mapped[list[Ticket]] = relationship(
        "Ticket", back_populates="creator", foreign_keys="Ticket.created_by"
    )
    assigned_ticket_links: Mapped[list[TicketAssignee]] = relationship(
        back_populates="user", cascade="all, delete-orphan"
    )
    assigned_tickets: Mapped[list[Ticket]] = relationship(
        "Ticket",
        secondary="ticket_assignees",
        back_populates="assignees",
        viewonly=True,
    )
    comments: Mapped[list[Comment]] = relationship("Comment", back_populates="creator")
    refresh_tokens: Mapped[list[RefreshToken]] = relationship(
        "RefreshToken", back_populates="user", cascade="all, delete-orphan"
    )

    @property
    def skill_names(self) -> list[str]:
        return [skill.skill for skill in self.skills]
