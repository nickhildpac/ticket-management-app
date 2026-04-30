from app.enums.ticket_state import TicketState
from app.policies.ticket_state_machine import can_transition, get_valid_transitions


def test_in_progress_only_transitions_to_resolved():
    assert can_transition(TicketState.IN_PROGRESS, TicketState.RESOLVED)
    assert not can_transition(TicketState.IN_PROGRESS, TicketState.PENDING)
    assert not can_transition(TicketState.IN_PROGRESS, TicketState.OPEN)


def test_open_transitions():
    assert can_transition(TicketState.OPEN, TicketState.PENDING)
    assert can_transition(TicketState.OPEN, TicketState.IN_PROGRESS)
    assert can_transition(TicketState.OPEN, TicketState.CANCELLED)
    assert not can_transition(TicketState.OPEN, TicketState.CLOSED)


def test_get_valid_transitions_includes_current_state():
    valid = get_valid_transitions(TicketState.RESOLVED)
    assert TicketState.RESOLVED in valid
    assert TicketState.CLOSED in valid
