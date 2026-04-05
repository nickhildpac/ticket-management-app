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
 * Mirrors backend allowedTransitions map. States not listed as keys (e.g. in progress) have no outbound edges in domain.
 */
const ALLOWED: Record<ApiTicketState, readonly ApiTicketState[]> = {
    open: ["pending", "cancelled", "in progress"],
    pending: ["open", "in progress", "resolved", "cancelled"],
    "in progress": [],
    resolved: ["open", "pending", "closed", "cancelled"],
    closed: [],
    cancelled: [],
};

export function normalizeTicketState(state: Ticket["state"]): ApiTicketState {
    if (typeof state === "number") {
        const s = getTicketStateString(state);
        if (s === "open" || s === "pending" || s === "resolved" || s === "closed" || s === "cancelled") {
            return s;
        }
        return "open";
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
