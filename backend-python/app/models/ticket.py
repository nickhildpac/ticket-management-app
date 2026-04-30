from __future__ import annotations

from datetime import datetime
from typing import TYPE_CHECKING
from uuid import UUID, uuid4

from sqlalchemy import BigInteger, DateTime, ForeignKey, String, Text, text
from sqlalchemy.dialects.postgresql import UUID as PG_UUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.enums.priority import TicketPriority
from app.enums.ticket_state import TicketState
from app.models.base import Base, TimestampMixin
from app.models.types import ticket_priority_enum, ticket_state_enum

if TYPE_CHECKING:
    from app.models.association import TicketAssignee, TicketSkill
    from app.models.comment import Comment
    from app.models.user import User


class Ticket(Base, TimestampMixin):
    __tablename__ = "tickets"

    id: Mapped[UUID] = mapped_column(
        PG_UUID(as_uuid=True),
        primary_key=True,
        default=uuid4,
        server_default=text("gen_random_uuid()"),
    )
    ticket_number: Mapped[int] = mapped_column(
        BigInteger,
        nullable=False,
        unique=True,
        server_default=text("nextval('ticket_number_seq')"),
    )
    title: Mapped[str] = mapped_column(String(200), nullable=False)
    description: Mapped[str] = mapped_column(Text, nullable=False)
    priority: Mapped[TicketPriority] = mapped_column(
        ticket_priority_enum,
        nullable=False,
        default=TicketPriority.LOW,
        server_default=TicketPriority.LOW.value,
    )
    state: Mapped[TicketState] = mapped_column(
        ticket_state_enum,
        nullable=False,
        default=TicketState.OPEN,
        server_default=TicketState.OPEN.value,
    )
    created_by: Mapped[UUID] = mapped_column(
        PG_UUID(as_uuid=True), ForeignKey("users.id", ondelete="RESTRICT"), index=True, nullable=False
    )
    deleted_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)

    creator: Mapped["User"] = relationship("User", back_populates="created_tickets", lazy="joined")

    assignee_links: Mapped[list["TicketAssignee"]] = relationship(
        back_populates="ticket", cascade="all, delete-orphan", lazy="selectin"
    )
    assignees: Mapped[list["User"]] = relationship(
        "User",
        secondary="ticket_assignees",
        back_populates="assigned_tickets",
        viewonly=True,
        lazy="selectin",
    )
    skills: Mapped[list["TicketSkill"]] = relationship(
        back_populates="ticket", cascade="all, delete-orphan", lazy="selectin"
    )
    comments: Mapped[list["Comment"]] = relationship(
        back_populates="ticket", cascade="all, delete-orphan"
    )

    @property
    def assignee_ids(self) -> list[UUID]:
        return [link.user_id for link in self.assignee_links]

    @property
    def skill_names(self) -> list[str]:
        return [skill.skill for skill in self.skills]
