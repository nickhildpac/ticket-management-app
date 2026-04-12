import type { Ticket } from "@/lib/types";
import { getTicketStateString } from "@/lib/types";

/** API / domain state strings (parity with backend/internal/domain/ticket.go). */
export type ApiTicketState =
    | "open"
    | "pending"
    | "in progress"
    | "resolved"
    | "closed"
    | "cancelled";

/**
 * allowedTransitions[from] = states reachable in one step (excluding staying put).
 * Mirrors backend/internal/domain/ticket.go allowedTransitions. TicketStateInProgress has no map entry there,
 * so no outbound transitions except identity (handled in canTransition on backend).
 */
const ALLOWED: Record<ApiTicketState, readonly ApiTicketState[]> = {
    open: ["pending", "cancelled", "in progress"],
    pending: ["open", "in progress", "resolved", "cancelled"],
    "in progress": [],
    resolved: ["open", "pending", "closed", "cancelled"],
    closed: [],
    cancelled: [],
};

const STATE_ORDER: readonly ApiTicketState[] = [
    "open",
    "pending",
    "in progress",
    "resolved",
    "closed",
    "cancelled",
];

export function normalizeTicketState(state: Ticket["state"]): ApiTicketState {
    if (typeof state === "number") {
        return getTicketStateString(state);
    }
    const raw = String(state).toLowerCase().trim();
    if (raw === "in progress" || raw === "in_progress") return "in progress";
    if (
        raw === "open" ||
        raw === "pending" ||
        raw === "resolved" ||
        raw === "closed" ||
        raw === "cancelled"
    ) {
        return raw;
    }
    return "open";
}

export function getValidNextStates(current: Ticket["state"]): ApiTicketState[] {
    const key = normalizeTicketState(current);
    const next = ALLOWED[key];
    return next ? [...next] : [];
}

/** Current state plus all states allowed by the domain FSM (for admin State dropdown on ticket detail). */
export function getValidTransitionTargets(current: Ticket["state"]): ApiTicketState[] {
    const key = normalizeTicketState(current);
    const reachable = new Set<ApiTicketState>([key, ...(ALLOWED[key] ?? [])]);
    return STATE_ORDER.filter((s) => reachable.has(s));
}

export function canTransition(from: Ticket["state"], to: ApiTicketState): boolean {
    const fromKey = normalizeTicketState(from);
    if (fromKey === to) return true;
    return ALLOWED[fromKey]?.includes(to) ?? false;
}

const LABELS: Record<ApiTicketState, string> = {
    open: "Open",
    pending: "Pending",
    "in progress": "In progress",
    resolved: "Resolved",
    closed: "Closed",
    cancelled: "Cancelled",
};

export function stateActionLabel(target: ApiTicketState): string {
    switch (target) {
        case "resolved":
            return "Resolve ticket";
        case "pending":
            return "Mark pending";
        case "open":
            return "Reopen";
        case "closed":
            return "Close ticket";
        case "cancelled":
            return "Cancel ticket";
        case "in progress":
            return "Start work";
        default:
            return LABELS[target];
    }
}

export function formatStateLabel(state: ApiTicketState): string {
    return LABELS[state];
}

/** End users may only move to closed or cancelled (backend CanUpdateTicketState). */
export function filterStatesForEndUser(next: ApiTicketState[]): ApiTicketState[] {
    return next.filter((s) => s === "closed" || s === "cancelled");
}
