from __future__ import annotations

from app.enums.ticket_state import TicketState


ALLOWED_TRANSITIONS: dict[TicketState, set[TicketState]] = {
    TicketState.OPEN: {TicketState.PENDING, TicketState.IN_PROGRESS, TicketState.CANCELLED},
    TicketState.PENDING: {
        TicketState.OPEN,
        TicketState.IN_PROGRESS,
        TicketState.RESOLVED,
        TicketState.CANCELLED,
    },
    TicketState.RESOLVED: {
        TicketState.OPEN,
        TicketState.PENDING,
        TicketState.CLOSED,
        TicketState.CANCELLED,
    },
    TicketState.IN_PROGRESS: {TicketState.RESOLVED},
    TicketState.CLOSED: set(),
    TicketState.CANCELLED: set(),
}


def can_transition(from_state: TicketState, to_state: TicketState) -> bool:
    if from_state == to_state:
        return True
    return to_state in ALLOWED_TRANSITIONS.get(from_state, set())


def get_valid_transitions(from_state: TicketState) -> set[TicketState]:
    return {from_state, *ALLOWED_TRANSITIONS.get(from_state, set())}
