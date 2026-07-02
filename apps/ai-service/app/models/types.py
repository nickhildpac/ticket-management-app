from __future__ import annotations

from sqlalchemy import Enum

from app.enums.priority import TicketPriority
from app.enums.role import UserRole
from app.enums.ticket_state import TicketState


def enum_values(enum_cls: type) -> list[str]:
    return [item.value for item in enum_cls]


user_role_enum = Enum(UserRole, name="user_role", values_callable=enum_values)
ticket_priority_enum = Enum(TicketPriority, name="ticket_priority", values_callable=enum_values)
ticket_state_enum = Enum(TicketState, name="ticket_state", values_callable=enum_values)
