"""Exercises the generated state-machine contract (contracts/ticket_state_machine.json
via `make contracts`) so a drifted artifact fails fast in this service too, not just
in the Go/TS consumers."""

from app.generated.ticket_state_contract import (
    ALLOWED_TRANSITION_VALUES,
    TICKET_STATE_ALIASES,
    TICKET_STATE_ENUM_MEMBERS,
)


def can_transition(from_state: str, to_state: str) -> bool:
    if from_state == to_state:
        return True
    return to_state in ALLOWED_TRANSITION_VALUES.get(from_state, [])


def test_in_progress_only_transitions_to_resolved():
    assert can_transition("in_progress", "resolved")
    assert not can_transition("in_progress", "pending")
    assert not can_transition("in_progress", "open")


def test_open_transitions():
    assert can_transition("open", "pending")
    assert can_transition("open", "in_progress")
    assert can_transition("open", "cancelled")
    assert not can_transition("open", "resolved")
    assert not can_transition("open", "closed")


def test_terminal_states_have_no_transitions():
    assert not ALLOWED_TRANSITION_VALUES.get("closed")
    assert not ALLOWED_TRANSITION_VALUES.get("cancelled")


def test_same_state_is_always_allowed():
    for wire in TICKET_STATE_ENUM_MEMBERS.values():
        assert can_transition(wire, wire)


def test_aliases_resolve_every_wire_value():
    for wire in TICKET_STATE_ENUM_MEMBERS.values():
        assert TICKET_STATE_ALIASES[wire] == wire
