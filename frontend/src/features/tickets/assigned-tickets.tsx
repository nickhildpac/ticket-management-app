import { AppShell } from "@/app/shell";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { Ticket } from "@/lib/types";
import { useState, useEffect, useMemo } from "react";
import { getTicketStateString, getTicketPriorityString } from "@/lib/types";
import {
    QueuePageHeader,
    QueueFilterBar,
    QueueTableHeader,
    TicketQueueRow,
    QueueRowSkeleton,
    QueuePagination,
} from "./components/queue-ui";

export function AssignedTicketsList() {
    const [page, setPage] = useState(1);
    const [state, setState] = useState<string>("all");
    const [q, setQ] = useState("");
    const [searchQuery, setSearchQuery] = useState("");

    useEffect(() => {
        const timer = setTimeout(() => {
            setSearchQuery(q);
            setPage(1);
        }, 500);
        return () => clearTimeout(timer);
    }, [q]);

    const handleStateChange = (newState: string) => {
        setState(newState);
        setPage(1);
    };

    const { data, isLoading, isError } = useQuery({
        queryKey: ["tickets", "assigned", page, state, searchQuery],
        queryFn: async () => {
            const queryParams = new URLSearchParams();
            if (page) queryParams.append("page", String(page));
            if (state && state !== "all") queryParams.append("state", state);
            if (searchQuery) queryParams.append("q", searchQuery);

            const response = await api<Ticket[]>(`/api/v1/ticket/assigned?${queryParams.toString()}`);
            if (Array.isArray(response)) {
                return { items: response, total: response.length, page: 1, pageSize: 20 };
            }
            return response;
        },
    });

    const stats = useMemo(() => {
        const items = data?.items ?? [];
        const open = items.filter((t) => {
            const s = typeof t.state === "number" ? getTicketStateString(t.state) : t.state;
            return s === "open";
        }).length;
        const critical = items.filter((t) => {
            const p = typeof t.priority === "number" ? getTicketPriorityString(t.priority) : t.priority;
            return p === "critical";
        }).length;
        return [
            { label: "Open", value: open, icon: "pending_actions" as const },
            {
                label: "Critical",
                value: critical,
                icon: "priority_high" as const,
                iconClassName: "bg-orange-500/20 text-orange-400",
            },
        ];
    }, [data?.items]);

    const metaRight =
        data?.items != null ? `${data.total ?? data.items.length} tickets on this page` : undefined;

    return (
        <AppShell>
            <QueuePageHeader
                title="My Assigned Tickets"
                subtitle="Tickets assigned to you for resolution."
                stats={stats}
            />

            <QueueFilterBar
                searchValue={q}
                onSearchChange={setQ}
                state={state}
                onStateChange={handleStateChange}
                metaRight={metaRight}
            />

            <QueueTableHeader variant="assigned" />

            <div className="space-y-3">
                {isLoading ? (
                    Array.from({ length: 5 }).map((_, i) => (
                        <QueueRowSkeleton key={i} showTicketNumber={false} />
                    ))
                ) : isError ? (
                    <p className="rounded-2xl border border-error/30 bg-error/10 p-6 text-center text-error">
                        Failed to load tickets
                    </p>
                ) : data?.items?.length === 0 ? (
                    <p className="rounded-2xl border border-outline-variant/30 bg-surface-container-low p-6 text-center text-on-surface-variant">
                        No assigned tickets found
                    </p>
                ) : (
                    data?.items?.map((ticket) => (
                        <TicketQueueRow key={ticket.id} ticket={ticket} showTicketNumber={false} />
                    ))
                )}
            </div>

            <QueuePagination
                page={page}
                onPageChange={setPage}
                disabledPrev={page === 1 || isLoading}
                disabledNext={
                    !data || !data.items || data.items.length < (data.pageSize || 10) || isLoading
                }
            />
        </AppShell>
    );
}
