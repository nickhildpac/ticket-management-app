import { AppShell } from "@/app/shell";
import { TICKET_LIST_PAGE_SIZE, useTickets } from "./queries";
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

export function TicketList() {
    const [page, setPage] = useState(1);
    const [state, setState] = useState<string>("all");
    const [priority, setPriority] = useState<string>("all");
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

    const handlePriorityChange = (newPriority: string) => {
        setPriority(newPriority);
        setPage(1);
    };

    const { data, isLoading, isError } = useTickets({
        page,
        state,
        priority,
        search: searchQuery,
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
                title="Ticket Queue"
                subtitle="Manage and prioritize your support requests."
                stats={stats}
            />

            <QueueFilterBar
                searchValue={q}
                onSearchChange={setQ}
                state={state}
                onStateChange={handleStateChange}
                priority={priority}
                onPriorityChange={handlePriorityChange}
                metaRight={metaRight}
            />

            <QueueTableHeader variant="withId" />

            <div className="space-y-3">
                {isLoading ? (
                    Array.from({ length: 5 }).map((_, i) => <QueueRowSkeleton key={i} />)
                ) : isError ? (
                    <p className="rounded-2xl border border-error/30 bg-error/10 p-6 text-center text-error">
                        Failed to load tickets
                    </p>
                ) : data?.items?.length === 0 ? (
                    <p className="rounded-2xl border border-outline-variant/30 bg-surface-container-low p-6 text-center text-on-surface-variant">
                        No tickets found
                    </p>
                ) : (
                    data?.items?.map((ticket) => (
                        <TicketQueueRow key={ticket.id} ticket={ticket} showTicketNumber />
                    ))
                )}
            </div>

            <QueuePagination
                page={page}
                onPageChange={setPage}
                disabledPrev={page === 1 || isLoading}
                disabledNext={
                    !data ||
                    !data.items ||
                    data.items.length < (data.pageSize || TICKET_LIST_PAGE_SIZE) ||
                    isLoading
                }
            />
        </AppShell>
    );
}
