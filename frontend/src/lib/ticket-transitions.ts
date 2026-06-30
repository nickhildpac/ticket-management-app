import type { Ticket } from "@/lib/types";
import { getTicketStateString } from "@/lib/types";
import {
    TICKET_STATE_ACTION_LABELS,
    TICKET_STATE_ALIASES,
    TICKET_STATE_LABELS,
    TICKET_STATE_ORDER,
    TICKET_STATE_TRANSITIONS,
    type ApiTicketState,
} from "@/lib/ticket-state-contract.generated";

export type { ApiTicketState };

export function normalizeTicketState(state: Ticket["state"]): ApiTicketState {
    if (typeof state === "number") {
        return getTicketStateString(state);
    }
    const raw = String(state).toLowerCase().trim();
    if (Object.prototype.hasOwnProperty.call(TICKET_STATE_ALIASES, raw)) {
        return TICKET_STATE_ALIASES[raw as keyof typeof TICKET_STATE_ALIASES];
    }
    return "open";
}

export function getValidNextStates(current: Ticket["state"]): ApiTicketState[] {
    const key = normalizeTicketState(current);
    const next = TICKET_STATE_TRANSITIONS[key];
    return next ? [...next] : [];
}

/** Current state plus all states allowed by the domain FSM (for admin State dropdown on ticket detail). */
export function getValidTransitionTargets(current: Ticket["state"]): ApiTicketState[] {
    const key = normalizeTicketState(current);
    const reachable = new Set<ApiTicketState>([key, ...(TICKET_STATE_TRANSITIONS[key] ?? [])]);
    return TICKET_STATE_ORDER.filter((s) => reachable.has(s));
}

export function canTransition(from: Ticket["state"], to: ApiTicketState): boolean {
    const fromKey = normalizeTicketState(from);
    if (fromKey === to) return true;
    return (TICKET_STATE_TRANSITIONS[fromKey] as readonly ApiTicketState[]).includes(to);
}

export function stateActionLabel(target: ApiTicketState): string {
    return TICKET_STATE_ACTION_LABELS[target] ?? TICKET_STATE_LABELS[target];
}

export function formatStateLabel(state: ApiTicketState): string {
    return TICKET_STATE_LABELS[state];
}

/** End users may only move to closed or cancelled (backend CanUpdateTicketState). */
export function filterStatesForEndUser(next: ApiTicketState[]): ApiTicketState[] {
    return next.filter((s) => s === "closed" || s === "cancelled");
}
