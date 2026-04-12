import { AppShell } from "@/app/shell";
import {
    TICKET_LIST_PAGE_SIZE,
    useTickets,
} from "./queries";
import { useMemo } from "react";
import { getTicketStateString, getTicketPriorityString } from "@/lib/types";
import { useUserTicketQueueControls } from "./queue-state";
import {
    QueuePageHeader,
    QueueFilterBar,
    QueueTableHeader,
    TicketQueueRow,
    QueueRowSkeleton,
    QueuePagination,
} from "./components/queue-ui";

export function TicketList() {
    const {
        page,
        setPage,
        state,
        priority,
        searchValue,
        searchQuery,
        sortBy,
        sortOrder,
        setSearchValue,
        setStateFilter,
        setPriorityFilter,
        handleSortClick,
    } = useUserTicketQueueControls();

    const { data, isLoading, isError } = useTickets({
        page,
        state,
        priority,
        search: searchQuery,
        sortBy,
        sortOrder,
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
                searchValue={searchValue}
                onSearchChange={setSearchValue}
                state={state}
                onStateChange={setStateFilter}
                priority={priority}
                onPriorityChange={setPriorityFilter}
                metaRight={metaRight}
            />

            <QueueTableHeader
                variant="withId"
                sortBy={sortBy}
                sortOrder={sortOrder}
                onSortClick={handleSortClick}
            />

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
