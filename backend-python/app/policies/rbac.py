from __future__ import annotations

from uuid import UUID

from app.enums.role import UserRole
from app.enums.ticket_state import TicketState


def can_manage_users(role: UserRole) -> bool:
    return role == UserRole.ADMIN


def can_delete_ticket(role: UserRole) -> bool:
    return role == UserRole.ADMIN


def can_assign_ticket(role: UserRole) -> bool:
    return role == UserRole.ADMIN


def can_update_priority(role: UserRole) -> bool:
    return role == UserRole.ADMIN


def is_assigned_to(ticket_assignee_ids: list[UUID], user_id: UUID) -> bool:
    return user_id in ticket_assignee_ids


def can_view_ticket(role: UserRole, actor_id: UUID, ticket_creator_id: UUID, ticket_assignee_ids: list[UUID]) -> bool:
    if role == UserRole.ADMIN:
        return True
    if role == UserRole.AGENT:
        return is_assigned_to(ticket_assignee_ids, actor_id)
    return ticket_creator_id == actor_id


def can_update_ticket(role: UserRole, actor_id: UUID, ticket_creator_id: UUID, ticket_assignee_ids: list[UUID]) -> bool:
    return can_view_ticket(role, actor_id, ticket_creator_id, ticket_assignee_ids)


def can_update_ticket_state(
    role: UserRole,
    actor_id: UUID,
    ticket_creator_id: UUID,
    ticket_assignee_ids: list[UUID],
    target_state: TicketState,
) -> bool:
    if role == UserRole.ADMIN:
        return True
    if role == UserRole.AGENT:
        return is_assigned_to(ticket_assignee_ids, actor_id)
    if ticket_creator_id != actor_id:
        return False
    return target_state in {TicketState.CLOSED, TicketState.CANCELLED}
