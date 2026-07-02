import { useEffect, useState } from "react";
import { useNavigate, useSearch } from "@tanstack/react-router";
import type { SearchSchemaInput } from "@tanstack/react-router";

export const TICKET_LIST_PAGE_SIZE = 20;

export type TicketListSortField = "ticket_number" | "created_at";
export type TicketListSortOrder = "asc" | "desc";

export type TicketQueueSearch = {
    page: number;
    state: string;
    priority: string;
    q: string;
    sortBy: TicketListSortField;
    sortOrder: TicketListSortOrder;
};

type TicketQueueSearchInput = SearchSchemaInput & {
    page?: number | string;
    state?: string;
    priority?: string;
    q?: string;
    sortBy?: TicketListSortField;
    sortOrder?: TicketListSortOrder;
};

const DEFAULT_QUEUE_SEARCH: TicketQueueSearch = {
    page: 1,
    state: "all",
    priority: "all",
    q: "",
    sortBy: "created_at",
    sortOrder: "desc",
};

const validStates = new Set([
    "all",
    "open",
    "pending",
    "in_progress",
    "resolved",
    "closed",
    "cancelled",
]);

const validPriorities = new Set(["all", "low", "medium", "high", "critical"]);

function parsePage(value: unknown): number {
    if (typeof value === "number" && Number.isInteger(value) && value > 0) {
        return value;
    }

    if (typeof value === "string") {
        const parsed = Number.parseInt(value, 10);
        if (Number.isInteger(parsed) && parsed > 0) {
            return parsed;
        }
    }

    return DEFAULT_QUEUE_SEARCH.page;
}

function parseState(value: unknown): string {
    if (typeof value === "string" && validStates.has(value)) {
        return value;
    }
    return DEFAULT_QUEUE_SEARCH.state;
}

function parsePriority(value: unknown): string {
    if (typeof value === "string" && validPriorities.has(value)) {
        return value;
    }
    return DEFAULT_QUEUE_SEARCH.priority;
}

function parseSearchText(value: unknown): string {
    return typeof value === "string" ? value : DEFAULT_QUEUE_SEARCH.q;
}

function parseSortField(value: unknown): TicketListSortField {
    return value === "ticket_number" ? "ticket_number" : DEFAULT_QUEUE_SEARCH.sortBy;
}

function parseSortOrder(value: unknown): TicketListSortOrder {
    return value === "asc" || value === "desc" ? value : DEFAULT_QUEUE_SEARCH.sortOrder;
}

export function validateTicketQueueSearch(search: TicketQueueSearchInput): TicketQueueSearch {
    return {
        page: parsePage(search.page),
        state: parseState(search.state),
        priority: parsePriority(search.priority),
        q: parseSearchText(search.q),
        sortBy: parseSortField(search.sortBy),
        sortOrder: parseSortOrder(search.sortOrder),
    };
}

type SearchUpdateFn = (
    updater: (current: TicketQueueSearch) => TicketQueueSearch,
    options?: { replace?: boolean }
) => void;

function useTicketQueueControlsState(search: TicketQueueSearch, updateSearch: SearchUpdateFn) {
    const [debouncedQuery, setDebouncedQuery] = useState(search.q);

    useEffect(() => {
        const timer = window.setTimeout(() => {
            setDebouncedQuery(search.q);
        }, 500);

        return () => window.clearTimeout(timer);
    }, [search.q]);

    return {
        page: search.page,
        state: search.state,
        priority: search.priority,
        searchValue: search.q,
        searchQuery: debouncedQuery,
        sortBy: search.sortBy,
        sortOrder: search.sortOrder,
        setPage: (page: number) => {
            updateSearch((current) => ({
                ...current,
                page: Math.max(1, page),
            }));
        },
        setSearchValue: (value: string) => {
            updateSearch(
                (current) => ({
                    ...current,
                    q: value,
                    page: 1,
                }),
                { replace: true }
            );
        },
        setStateFilter: (value: string) => {
            updateSearch((current) => ({
                ...current,
                state: value,
                page: 1,
            }));
        },
        setPriorityFilter: (value: string) => {
            updateSearch((current) => ({
                ...current,
                priority: value,
                page: 1,
            }));
        },
        handleSortClick: (field: TicketListSortField) => {
            updateSearch((current) => ({
                ...current,
                page: 1,
                sortBy: field,
                sortOrder:
                    field === current.sortBy
                        ? current.sortOrder === "asc"
                            ? "desc"
                            : "asc"
                        : "desc",
            }));
        },
    };
}

export function useUserTicketQueueControls() {
    const search = useSearch({ from: "/tickets" });
    const navigate = useNavigate({ from: "/tickets" });

    return useTicketQueueControlsState(search, (updater, options) => {
        void navigate({
            to: ".",
            search: (current) => updater(current),
            replace: options?.replace,
            resetScroll: false,
        });
    });
}

export function useAllTicketQueueControls() {
    const search = useSearch({ from: "/tickets/all" });
    const navigate = useNavigate({ from: "/tickets/all" });

    return useTicketQueueControlsState(search, (updater, options) => {
        void navigate({
            to: ".",
            search: (current) => updater(current),
            replace: options?.replace,
            resetScroll: false,
        });
    });
}

export function useAssignedTicketQueueControls() {
    const search = useSearch({ from: "/tickets/assigned" });
    const navigate = useNavigate({ from: "/tickets/assigned" });

    return useTicketQueueControlsState(search, (updater, options) => {
        void navigate({
            to: ".",
            search: (current) => updater(current),
            replace: options?.replace,
            resetScroll: false,
        });
    });
}
