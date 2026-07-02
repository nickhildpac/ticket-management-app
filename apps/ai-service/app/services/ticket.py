from __future__ import annotations

from typing import Any
from uuid import UUID

from sqlalchemy import exists, select
from sqlalchemy.orm import Session

from app.core.exceptions import BadRequestError, ForbiddenError, NotFoundError
from app.enums.role import UserRole
from app.models.association import TicketAssignee
from app.models.ticket import Ticket
from app.models.user import User
from app.policies.rbac import (
    can_assign_ticket,
    can_delete_ticket,
    can_update_priority,
    can_update_ticket,
    can_update_ticket_state,
    can_view_ticket,
)
from app.policies.ticket_state_machine import can_transition
from app.repositories.ticket import TicketRepository
from app.repositories.user import UserRepository
from app.schemas.ticket import CreateTicketRequest, TicketListQuery, UpdateTicketRequest


class TicketService:
    def __init__(self, session: Session):
        self.session = session
        self.repo = TicketRepository(session)
        self.user_repo = UserRepository(session)

    def _scope_filters(self, current_user: User) -> list[Any]:
        if current_user.role == UserRole.ADMIN:
            return []
        if current_user.role == UserRole.AGENT:
            return [
                exists(
                    select(1).where(
                        TicketAssignee.ticket_id == Ticket.id,
                        TicketAssignee.user_id == current_user.id,
                    )
                )
            ]
        return [Ticket.created_by == current_user.id]

    @staticmethod
    def _parse_uuid(raw: str, field: str) -> UUID:
        try:
            return UUID(raw)
        except ValueError as exc:
            raise BadRequestError(f"invalid {field}") from exc

    def _apply_common_filters(
        self,
        current_user: User,
        query: TicketListQuery,
    ) -> tuple[list[Any], UUID | None]:
        filters: list[Any] = []
        if query.state:
            filters.append(Ticket.state == query.state)
        if query.priority:
            filters.append(Ticket.priority == query.priority)
        if query.ticket_number is not None:
            filters.append(Ticket.ticket_number == query.ticket_number)

        assignee_id: UUID | None = None
        effective_assignee = None
        try:
            effective_assignee = query.effective_assignee()
        except ValueError as exc:
            raise BadRequestError(str(exc)) from exc
        if effective_assignee:
            assignee_id = (
                current_user.id
                if effective_assignee == "me"
                else self._parse_uuid(effective_assignee, "assignee")
            )
            filters.append(
                exists(
                    select(1).where(
                        TicketAssignee.ticket_id == Ticket.id,
                        TicketAssignee.user_id == assignee_id,
                    )
                )
            )

        if query.created_by:
            created_by = self._parse_uuid(query.created_by, "created_by")
            filters.append(Ticket.created_by == created_by)

        return filters, assignee_id

    def _list_filters_for_actor(self, current_user: User, query: TicketListQuery) -> list[Any]:
        filters = self._scope_filters(current_user)
        extra_filters, assignee_id = self._apply_common_filters(current_user, query)

        if current_user.role == UserRole.AGENT:
            if query.created_by:
                raise ForbiddenError("access denied")
            if assignee_id and assignee_id != current_user.id:
                raise ForbiddenError("access denied")

        if current_user.role == UserRole.USER:
            if assignee_id:
                raise ForbiddenError("access denied")
            if (
                query.created_by
                and self._parse_uuid(query.created_by, "created_by") != current_user.id
            ):
                raise ForbiddenError("access denied")

        filters.extend(extra_filters)
        return filters

    def list_ticket_page_for_actor(
        self,
        current_user: User,
        query: TicketListQuery,
    ) -> tuple[list[Ticket], int]:
        filters = self._list_filters_for_actor(current_user, query)
        tickets = self.repo.list_tickets(
            filters=filters,
            limit=query.limit,
            offset=query.offset,
            sort=query.sort,
            order=query.order,
        )
        total = self.repo.count_tickets(filters=filters)
        return tickets, total

    def list_tickets_for_actor(self, current_user: User, query: TicketListQuery) -> list[Ticket]:
        tickets, _ = self.list_ticket_page_for_actor(current_user, query)
        return tickets

    def list_assigned_tickets(self, current_user: User, query: TicketListQuery) -> list[Ticket]:
        assigned_query = query.model_copy(update={"assigned_to": "me", "assignee": None})
        return self.list_tickets_for_actor(current_user, assigned_query)

    def get_ticket_or_404(self, ticket_id: UUID) -> Ticket:
        ticket = self.repo.get_by_id(ticket_id)
        if not ticket:
            raise NotFoundError("ticket not found")
        return ticket

    def get_ticket(self, current_user: User, ticket_id: UUID) -> Ticket:
        ticket = self.get_ticket_or_404(ticket_id)
        if not can_view_ticket(
            current_user.role,
            current_user.id,
            ticket.created_by,
            ticket.assignee_ids,
        ):
            raise ForbiddenError("access denied")
        return ticket

    def get_ticket_by_number(self, current_user: User, ticket_number: int) -> Ticket:
        ticket = self.repo.get_by_number(ticket_number)
        if not ticket:
            raise NotFoundError("ticket not found")
        if not can_view_ticket(
            current_user.role,
            current_user.id,
            ticket.created_by,
            ticket.assignee_ids,
        ):
            raise ForbiddenError("access denied")
        return ticket

    def create_ticket(self, current_user: User, payload: CreateTicketRequest) -> Ticket:
        ticket = Ticket(
            title=payload.title,
            description=payload.description,
            priority=payload.priority,
            created_by=current_user.id,
        )
        self.repo.create(ticket)
        self.repo.set_skills(ticket, payload.skills)
        self.session.commit()
        self.session.refresh(ticket)
        return ticket

    def update_ticket(
        self,
        current_user: User,
        ticket_id: UUID,
        payload: UpdateTicketRequest,
    ) -> Ticket:
        ticket = self.get_ticket_or_404(ticket_id)

        if not can_update_ticket(
            current_user.role,
            current_user.id,
            ticket.created_by,
            ticket.assignee_ids,
        ):
            raise ForbiddenError("access denied")

        changed = False

        if payload.title is not None:
            ticket.title = payload.title
            changed = True

        if payload.description is not None:
            ticket.description = payload.description
            changed = True

        if payload.state is not None:
            if not can_update_ticket_state(
                current_user.role,
                current_user.id,
                ticket.created_by,
                ticket.assignee_ids,
                payload.state,
            ):
                raise ForbiddenError("access denied")
            if not can_transition(ticket.state, payload.state):
                raise BadRequestError("invalid status transition")
            ticket.state = payload.state
            changed = True

        if payload.priority is not None:
            if not can_update_priority(current_user.role):
                raise ForbiddenError("access denied")
            ticket.priority = payload.priority
            changed = True

        if payload.skills is not None:
            self.repo.set_skills(ticket, payload.skills)
            changed = True

        if payload.assigned_to is not None:
            if not can_assign_ticket(current_user.role):
                raise ForbiddenError("access denied")
            assignee_ids = [
                self._parse_uuid(user_id, "assigned_to") for user_id in payload.assigned_to
            ]
            assignees = self.user_repo.get_by_ids(assignee_ids)
            if len(assignees) != len(set(assignee_ids)):
                raise BadRequestError("one or more assignees do not exist")
            self.repo.set_assignees(ticket, assignee_ids)
            changed = True

        if not changed:
            raise BadRequestError("no fields provided to update")

        self.session.commit()
        self.session.refresh(ticket)
        return ticket

    def delete_ticket(self, current_user: User, ticket_id: UUID) -> None:
        if not can_delete_ticket(current_user.role):
            raise ForbiddenError("access denied")
        ticket = self.get_ticket_or_404(ticket_id)
        self.repo.delete(ticket)
        self.session.commit()

    def get_stats(self, current_user: User) -> dict[str, int]:
        filters = self._scope_filters(current_user)
        return self.repo.count_stats(filters=filters, current_user_id=current_user.id)
