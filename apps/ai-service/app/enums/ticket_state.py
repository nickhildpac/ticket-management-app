from __future__ import annotations

from enum import StrEnum

from app.generated.ticket_state_contract import TICKET_STATE_ALIASES, TICKET_STATE_ENUM_MEMBERS

TicketState = StrEnum("TicketState", TICKET_STATE_ENUM_MEMBERS)


def normalize_ticket_state_input(value: str) -> TicketState:
    normalized = value.strip().lower().replace(" ", "_")
    return TicketState(TICKET_STATE_ALIASES.get(normalized, normalized))


def ticket_state_to_wire(value: TicketState) -> str:
    return value.value
