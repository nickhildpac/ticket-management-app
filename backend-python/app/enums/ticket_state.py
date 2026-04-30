from __future__ import annotations

from enum import StrEnum


class TicketState(StrEnum):
    OPEN = "open"
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    RESOLVED = "resolved"
    CLOSED = "closed"
    CANCELLED = "cancelled"


def normalize_ticket_state_input(value: str) -> TicketState:
    normalized = value.strip().lower().replace(" ", "_")
    if normalized == "cancel":
        normalized = "cancelled"
    return TicketState(normalized)


def ticket_state_to_wire(value: TicketState) -> str:
    if value == TicketState.IN_PROGRESS:
        return "in progress"
    return value.value
