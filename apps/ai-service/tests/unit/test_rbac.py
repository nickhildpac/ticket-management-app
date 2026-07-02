from uuid import uuid4

from app.enums.role import UserRole
from app.enums.ticket_state import TicketState
from app.policies.rbac import (
    can_delete_ticket,
    can_manage_users,
    can_update_ticket_state,
    can_view_ticket,
)


def test_admin_has_full_management_access():
    assert can_manage_users(UserRole.ADMIN)
    assert can_delete_ticket(UserRole.ADMIN)


def test_agent_can_only_view_assigned_tickets():
    actor_id = uuid4()
    creator_id = uuid4()
    assert can_view_ticket(UserRole.AGENT, actor_id, creator_id, [actor_id])
    assert not can_view_ticket(UserRole.AGENT, actor_id, creator_id, [])


def test_end_user_state_updates_limited_to_closed_or_cancelled():
    actor_id = uuid4()
    creator_id = actor_id
    assert can_update_ticket_state(UserRole.USER, actor_id, creator_id, [], TicketState.CLOSED)
    assert can_update_ticket_state(UserRole.USER, actor_id, creator_id, [], TicketState.CANCELLED)
    assert not can_update_ticket_state(UserRole.USER, actor_id, creator_id, [], TicketState.RESOLVED)
