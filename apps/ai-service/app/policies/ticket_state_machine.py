from __future__ import annotations

from app.enums.ticket_state import TicketState
from app.generated.ticket_state_contract import ALLOWED_TRANSITION_VALUES

ALLOWED_TRANSITIONS: dict[TicketState, set[TicketState]] = {
    TicketState(state): {TicketState(target) for target in targets}
    for state, targets in ALLOWED_TRANSITION_VALUES.items()
}


def can_transition(from_state: TicketState, to_state: TicketState) -> bool:
    if from_state == to_state:
        return True
    return to_state in ALLOWED_TRANSITIONS.get(from_state, set())


def get_valid_transitions(from_state: TicketState) -> set[TicketState]:
    return {from_state, *ALLOWED_TRANSITIONS.get(from_state, set())}
