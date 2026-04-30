from __future__ import annotations

from uuid import UUID

from sqlalchemy import Select, and_, exists, func, select
from sqlalchemy.orm import Session, joinedload, selectinload

from app.enums.ticket_state import TicketState
from app.models.association import TicketAssignee, TicketSkill
from app.models.ticket import Ticket


class TicketRepository:
    def __init__(self, session: Session):
        self.session = session

    def _base_select(self) -> Select[tuple[Ticket]]:
        return (
            select(Ticket)
            .where(Ticket.deleted_at.is_(None))
            .options(
                joinedload(Ticket.creator),
                selectinload(Ticket.skills),
                selectinload(Ticket.assignee_links),
            )
        )

    def create(self, ticket: Ticket) -> Ticket:
        self.session.add(ticket)
        self.session.flush()
        self.session.refresh(ticket)
        return ticket

    def get_by_id(self, ticket_id: UUID) -> Ticket | None:
        stmt = self._base_select().where(Ticket.id == ticket_id)
        return self.session.scalar(stmt)

    def get_by_number(self, ticket_number: int) -> Ticket | None:
        stmt = self._base_select().where(Ticket.ticket_number == ticket_number)
        return self.session.scalar(stmt)

    def list_tickets(
        self,
        *,
        filters: list,
        limit: int,
        offset: int,
        sort: str,
        order: str,
    ) -> list[Ticket]:
        sort_column = Ticket.ticket_number if sort == "ticket_number" else Ticket.created_at
        ordering = sort_column.asc() if order == "asc" else sort_column.desc()

        stmt = (
            self._base_select()
            .where(*filters)
            .order_by(ordering)
            .limit(limit)
            .offset(offset)
        )
        return list(self.session.scalars(stmt).unique().all())

    def delete(self, ticket: Ticket) -> None:
        self.session.delete(ticket)

    def set_assignees(self, ticket: Ticket, assignee_ids: list[UUID]) -> None:
        ticket.assignee_links = [
            TicketAssignee(ticket_id=ticket.id, user_id=user_id)
            for user_id in sorted(set(assignee_ids), key=lambda item: str(item))
        ]

    def set_skills(self, ticket: Ticket, skills: list[str]) -> None:
        ticket.skills = [TicketSkill(ticket_id=ticket.id, skill=skill) for skill in skills]

    def count_stats(self, *, filters: list, current_user_id: UUID) -> dict[str, int]:
        mine_condition = exists(
            select(1).where(
                TicketAssignee.ticket_id == Ticket.id,
                TicketAssignee.user_id == current_user_id,
            )
        )

        def count(extra_filter=None) -> int:
            criteria = [Ticket.deleted_at.is_(None), *filters]
            if extra_filter is not None:
                criteria.append(extra_filter)
            stmt = select(func.count(Ticket.id)).where(and_(*criteria))
            return int(self.session.scalar(stmt) or 0)

        return {
            "total": count(),
            "open": count(Ticket.state == TicketState.OPEN),
            "pending": count(Ticket.state == TicketState.PENDING),
            "resolved": count(Ticket.state == TicketState.RESOLVED),
            "mine": count(mine_condition),
        }
