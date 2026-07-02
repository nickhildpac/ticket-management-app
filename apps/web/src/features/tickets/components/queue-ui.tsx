import { Link } from "@tanstack/react-router";
import { formatDistanceToNow } from "date-fns";
import type { Ticket, UserInfo } from "@/lib/types";
import { getTicketStateString, getTicketPriorityString } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { MaterialSymbol } from "@/components/material-symbol";
import { cn } from "@/lib/utils";
import type { TicketListSortField, TicketListSortOrder } from "../queue-state";

type StatChip = { label: string; value: number; icon: string; iconClassName?: string };

export function QueuePageHeader({
    title,
    subtitle,
    stats,
}: {
    title: string;
    subtitle: string;
    stats?: StatChip[];
}) {
    return (
        <div className="mb-8 flex flex-col justify-between gap-4 md:flex-row md:items-end">
            <div>
                <h2 className="font-headline text-3xl font-extrabold tracking-tight text-on-background">{title}</h2>
                <p className="mt-1 font-body text-on-surface-variant">{subtitle}</p>
            </div>
            {stats && stats.length > 0 && (
                <div className="flex flex-wrap gap-2">
                    {stats.map((s) => (
                        <div
                            key={s.label}
                            className="flex items-center gap-3 rounded-xl border border-outline-variant/30 bg-surface-container px-4 py-3"
                        >
                            <div
                                className={cn(
                                    "flex h-10 w-10 items-center justify-center rounded-lg bg-primary/20 text-primary",
                                    s.iconClassName
                                )}
                            >
                                <MaterialSymbol name={s.icon} />
                            </div>
                            <div>
                                <p className="text-[10px] font-bold uppercase tracking-wider text-on-surface-variant">
                                    {s.label}
                                </p>
                                <p className="text-lg font-bold leading-none text-on-surface">{String(s.value).padStart(2, "0")}</p>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}

type QueueFilterBarProps = {
    searchValue: string;
    onSearchChange: (v: string) => void;
    state: string;
    onStateChange: (v: string) => void;
    /** When set with onPriorityChange, shows priority filter (GET /tickets only). */
    priority?: string;
    onPriorityChange?: (v: string) => void;
    metaRight?: string;
};

export function QueueFilterBar({
    searchValue,
    onSearchChange,
    state,
    onStateChange,
    priority = "all",
    onPriorityChange,
    metaRight,
}: QueueFilterBarProps) {
    return (
        <div className="mb-6 flex flex-wrap items-center gap-2 rounded-2xl border border-outline-variant/50 bg-surface-container-low p-2 shadow-sm">
            <div className="relative min-w-[200px] flex-1 sm:max-w-xs">
                <MaterialSymbol
                    name="search"
                    className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant"
                />
                <Input
                    placeholder="Search by ticket number…"
                    value={searchValue}
                    onChange={(e) => onSearchChange(e.target.value)}
                    className="rounded-xl border-0 bg-surface-container pl-10 focus-visible:ring-primary/20"
                />
            </div>
            <Select value={state} onValueChange={onStateChange}>
                <SelectTrigger className="relative h-auto min-w-[160px] rounded-xl border-0 bg-surface-container py-2.5 pl-10 pr-8 font-medium text-on-surface shadow-none hover:bg-surface-container-highest">
                    <MaterialSymbol
                        name="filter_list"
                        className="pointer-events-none absolute left-3 top-1/2 !text-lg -translate-y-1/2 text-on-surface-variant"
                    />
                    <SelectValue placeholder="All Status" />
                </SelectTrigger>
                <SelectContent>
                    <SelectItem value="all">All Statuses</SelectItem>
                    <SelectItem value="open">Open</SelectItem>
                    <SelectItem value="pending">Pending</SelectItem>
                    <SelectItem value="in_progress">In progress</SelectItem>
                    <SelectItem value="resolved">Resolved</SelectItem>
                    <SelectItem value="closed">Closed</SelectItem>
                    <SelectItem value="cancelled">Cancelled</SelectItem>
                </SelectContent>
            </Select>
            {onPriorityChange ? (
                <Select value={priority} onValueChange={onPriorityChange}>
                    <SelectTrigger className="relative h-auto min-w-[160px] rounded-xl border-0 bg-surface-container py-2.5 pl-10 pr-8 font-medium text-on-surface shadow-none hover:bg-surface-container-highest sm:flex">
                        <MaterialSymbol
                            name="bolt"
                            className="pointer-events-none absolute left-3 top-1/2 !text-lg -translate-y-1/2 text-on-surface-variant"
                        />
                        <SelectValue placeholder="All priorities" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="all">All priorities</SelectItem>
                        <SelectItem value="low">Low</SelectItem>
                        <SelectItem value="medium">Medium</SelectItem>
                        <SelectItem value="high">High</SelectItem>
                        <SelectItem value="critical">Critical</SelectItem>
                    </SelectContent>
                </Select>
            ) : null}
            <button
                type="button"
                className="hidden items-center gap-2 rounded-xl bg-surface-container px-4 py-2 text-sm font-medium text-on-surface transition-colors hover:bg-surface-container-highest md:flex"
            >
                <MaterialSymbol name="category" className="!text-lg" />
                Category
                <MaterialSymbol name="expand_more" className="!text-sm opacity-70" />
            </button>
            <div className="mx-2 hidden h-6 w-px bg-outline-variant lg:block" />
            <div className="hidden items-center gap-1 rounded-xl bg-surface-container p-1 lg:flex">
                <span className="rounded-lg bg-surface-container-highest px-4 py-1.5 text-sm font-bold text-primary shadow-sm">
                    Table View
                </span>
                <span className="cursor-not-allowed rounded-lg px-4 py-1.5 text-sm font-medium text-on-surface-variant opacity-50">
                    Kanban
                </span>
            </div>
            {metaRight && (
                <div className="ml-auto flex items-center gap-2 pr-2">
                    <span className="text-xs font-medium italic text-on-surface-variant">{metaRight}</span>
                </div>
            )}
        </div>
    );
}

function stateLeftBorder(stateStr: string) {
    switch (stateStr) {
        case "open":
            return "border-l-primary";
        case "pending":
            return "border-l-amber-500";
        case "in_progress":
            return "border-l-sky-500";
        case "resolved":
            return "border-l-emerald-500/50";
        default:
            return "border-l-outline-variant";
    }
}

function statusPillClass(stateStr: string) {
    switch (stateStr) {
        case "open":
            return "border-primary/30 bg-primary/20 text-primary";
        case "pending":
            return "border-amber-500/30 bg-amber-500/20 text-amber-500";
        case "in_progress":
            return "border-sky-500/30 bg-sky-500/15 text-sky-600 dark:text-sky-400";
        case "resolved":
            return "border-emerald-500/20 bg-emerald-500/10 text-emerald-500/80";
        case "closed":
        case "cancelled":
            return "border-outline-variant bg-surface-container-highest text-on-surface-variant";
        default:
            return "border-outline-variant bg-surface-container-highest text-on-surface-variant";
    }
}

function priorityPillClass(priorityStr: string) {
    switch (priorityStr) {
        case "critical":
            return "border-error/30 bg-error/20 font-black uppercase text-error";
        case "high":
            return "border-orange-500/30 bg-orange-500/20 font-black uppercase text-orange-400";
        case "medium":
        case "low":
            return "border-outline-variant bg-surface-container-highest font-black uppercase text-on-surface-variant";
        default:
            return "border-outline-variant bg-surface-container-highest text-on-surface-variant";
    }
}

function creatorDisplayName(creator: UserInfo | undefined) {
    if (!creator) return "Unknown";
    const name = `${creator.first_name ?? ""} ${creator.last_name ?? ""}`.trim();
    if (name) return name;
    if (creator.email) return creator.email;
    return "Unknown";
}

function creatorInitials(creator: UserInfo | undefined) {
    if (!creator) return "?";
    const a = creator.first_name?.charAt(0) ?? "";
    const b = creator.last_name?.charAt(0) ?? "";
    if (a || b) return `${a}${b}`.toUpperCase();
    if (creator.email) return creator.email.charAt(0).toUpperCase();
    return "?";
}

export function TicketQueueRow({
    ticket,
    showTicketNumber = true,
}: {
    ticket: Ticket;
    showTicketNumber?: boolean;
}) {
    const stateStr = typeof ticket.state === "number" ? getTicketStateString(ticket.state) : ticket.state;
    const priorityStr =
        typeof ticket.priority === "number" ? getTicketPriorityString(ticket.priority) : ticket.priority;
    const dimmed = stateStr === "resolved" || stateStr === "closed" || stateStr === "cancelled";
    const relative = formatDistanceToNow(new Date(ticket.created_at), { addSuffix: true });

    return (
        <Link
            to="/tickets/$id"
            params={{ id: ticket.id }}
            className={cn(
                "group grid grid-cols-12 items-center gap-4 rounded-2xl border border-outline-variant/30 border-l-4 bg-surface-container-low p-4 shadow-lg transition-all hover:bg-surface-container-high sm:p-6",
                stateLeftBorder(stateStr),
                dimmed && "opacity-80"
            )}
        >
            {showTicketNumber && (
                <div
                    className={cn(
                        "col-span-12 font-headline text-sm font-bold text-primary sm:col-span-1",
                        dimmed && "text-on-surface-variant/50"
                    )}
                >
                    #{ticket.ticket_number}
                </div>
            )}
            <div className={cn("col-span-12 sm:col-span-2", !showTicketNumber && "sm:col-span-2")}>
                <span
                    className={cn(
                        "inline-flex items-center rounded-full border px-2.5 py-1 text-xs font-bold",
                        statusPillClass(stateStr)
                    )}
                >
                    <span
                        className={cn(
                            "mr-2 h-1.5 w-1.5 shrink-0 rounded-full",
                            stateStr === "open" && "bg-primary",
                            stateStr === "pending" && "bg-amber-500",
                            stateStr === "in_progress" && "bg-sky-500",
                            stateStr === "resolved" && "bg-emerald-500/50",
                            (stateStr === "closed" || stateStr === "cancelled") && "bg-on-surface-variant"
                        )}
                    />
                    <span>{stateStr === "in_progress" ? "In progress" : stateStr}</span>
                </span>
            </div>
            <div
                className={cn(
                    "col-span-12 min-w-0 sm:col-span-4",
                    !showTicketNumber && "sm:col-span-5"
                )}
            >
                <h4
                    className={cn(
                        "truncate text-sm font-bold text-on-surface transition-colors group-hover:text-primary",
                        dimmed && "opacity-90"
                    )}
                >
                    {ticket.title}
                </h4>
                <p className="mt-0.5 truncate text-xs text-on-surface-variant">
                    {ticket.description?.substring(0, 100)}
                    {ticket.description && ticket.description.length > 100 ? "…" : ""}
                </p>
            </div>
            <div className="col-span-12 flex items-center gap-2 sm:col-span-2">
                <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-surface-container-highest text-[10px] font-bold text-primary">
                    {creatorInitials(ticket.creator)}
                </div>
                <span className="truncate text-sm font-medium text-on-surface-variant">
                    {creatorDisplayName(ticket.creator)}
                </span>
            </div>
            <div className="col-span-6 flex justify-center sm:col-span-1 sm:justify-center">
                <span
                    className={cn(
                        "inline-flex items-center rounded-full border px-2.5 py-1 text-[10px]",
                        priorityPillClass(priorityStr)
                    )}
                >
                    {priorityStr}
                </span>
            </div>
            <div className="col-span-6 text-right sm:col-span-2">
                <span className="text-xs font-medium text-on-surface-variant" title={ticket.created_at}>
                    {relative}
                </span>
            </div>
        </Link>
    );
}

function SortableHeader({
    label,
    active,
    order,
    className,
    onClick,
}: {
    label: string;
    active: boolean;
    order: TicketListSortOrder;
    className?: string;
    onClick: () => void;
}) {
    return (
        <button
            type="button"
            onClick={onClick}
            className={cn(
                "inline-flex items-center gap-0.5 text-left transition-colors hover:text-primary",
                active && "text-primary",
                className
            )}
        >
            {label}
            {active ? (
                <MaterialSymbol
                    name={order === "asc" ? "arrow_upward" : "arrow_downward"}
                    className="!text-sm opacity-90"
                />
            ) : (
                <MaterialSymbol name="unfold_more" className="!text-sm opacity-40" />
            )}
        </button>
    );
}

export function QueueTableHeader({
    variant,
    sortBy,
    sortOrder,
    onSortClick,
}: {
    variant: "withId" | "assigned";
    sortBy: TicketListSortField;
    sortOrder: TicketListSortOrder;
    onSortClick: (field: TicketListSortField) => void;
}) {
    if (variant === "assigned") {
        return (
            <div className="mb-2 grid grid-cols-12 gap-4 px-2 py-2 font-label text-[10px] font-bold uppercase tracking-widest text-on-surface-variant sm:px-6">
                <div className="col-span-1">
                    <SortableHeader
                        label="Ticket #"
                        active={sortBy === "ticket_number"}
                        order={sortOrder}
                        onClick={() => onSortClick("ticket_number")}
                    />
                </div>
                <div className="col-span-2">Status</div>
                <div className="col-span-4">Subject</div>
                <div className="col-span-2">Created by</div>
                <div className="col-span-1 text-center">Priority</div>
                <div className="col-span-2 text-right">
                    <SortableHeader
                        label="Created"
                        active={sortBy === "created_at"}
                        order={sortOrder}
                        className="ml-auto"
                        onClick={() => onSortClick("created_at")}
                    />
                </div>
            </div>
        );
    }
    return (
        <div className="mb-2 grid grid-cols-12 gap-4 px-2 py-2 font-label text-[10px] font-bold uppercase tracking-widest text-on-surface-variant sm:px-6">
            <div className="col-span-1">
                <SortableHeader
                    label="Ticket #"
                    active={sortBy === "ticket_number"}
                    order={sortOrder}
                    onClick={() => onSortClick("ticket_number")}
                />
            </div>
            <div className="col-span-2">Status</div>
            <div className="col-span-4">Subject</div>
            <div className="col-span-2">Created by</div>
            <div className="col-span-1 text-center">Priority</div>
            <div className="col-span-2 text-right">
                <SortableHeader
                    label="Created"
                    active={sortBy === "created_at"}
                    order={sortOrder}
                    className="ml-auto"
                    onClick={() => onSortClick("created_at")}
                />
            </div>
        </div>
    );
}

export function QueueRowSkeleton({ showTicketNumber = true }: { showTicketNumber?: boolean }) {
    return (
        <div className="grid grid-cols-12 gap-4 rounded-2xl border border-outline-variant/30 bg-surface-container-low p-6">
            {showTicketNumber ? (
                <>
                    <div className="col-span-1">
                        <Skeleton className="h-4 w-16 bg-surface-container-highest" />
                    </div>
                    <div className="col-span-2">
                        <Skeleton className="h-6 w-24 rounded-full bg-surface-container-highest" />
                    </div>
                    <div className="col-span-4">
                        <Skeleton className="h-4 w-full bg-surface-container-highest" />
                        <Skeleton className="mt-2 h-3 w-[75%] bg-surface-container-highest" />
                    </div>
                    <div className="col-span-2">
                        <Skeleton className="h-4 w-32 bg-surface-container-highest" />
                    </div>
                    <div className="col-span-1 flex justify-center">
                        <Skeleton className="h-6 w-16 rounded-full bg-surface-container-highest" />
                    </div>
                    <div className="col-span-2">
                        <Skeleton className="ml-auto h-4 w-20 bg-surface-container-highest" />
                    </div>
                </>
            ) : (
                <>
                    <div className="col-span-2">
                        <Skeleton className="h-6 w-24 rounded-full bg-surface-container-highest" />
                    </div>
                    <div className="col-span-5">
                        <Skeleton className="h-4 w-full bg-surface-container-highest" />
                        <Skeleton className="mt-2 h-3 w-[75%] bg-surface-container-highest" />
                    </div>
                    <div className="col-span-2">
                        <Skeleton className="h-4 w-32 bg-surface-container-highest" />
                    </div>
                    <div className="col-span-1 flex justify-center">
                        <Skeleton className="h-6 w-16 rounded-full bg-surface-container-highest" />
                    </div>
                    <div className="col-span-2">
                        <Skeleton className="ml-auto h-4 w-20 bg-surface-container-highest" />
                    </div>
                </>
            )}
        </div>
    );
}

type QueuePaginationProps = {
    page: number;
    onPageChange: (p: number) => void;
    disabledPrev: boolean;
    disabledNext: boolean;
    pageSize?: number;
    onPageSizeChange?: (n: number) => void;
};

export function QueuePagination({
    page,
    onPageChange,
    disabledPrev,
    disabledNext,
    pageSize = 20,
    onPageSizeChange,
}: QueuePaginationProps) {
    return (
        <div className="mt-8 flex flex-col items-stretch justify-between gap-4 sm:flex-row sm:items-center">
            <div className="flex items-center justify-center gap-2 sm:justify-start">
                <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    className="h-10 w-10 rounded-xl border-outline-variant/30 bg-surface-container text-on-surface-variant hover:text-primary"
                    disabled={disabledPrev}
                    onClick={() => onPageChange(1)}
                    aria-label="First page"
                >
                    <MaterialSymbol name="first_page" />
                </Button>
                <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    className="h-10 w-10 rounded-xl border-outline-variant/30 bg-surface-container text-on-surface-variant hover:text-primary"
                    disabled={disabledPrev}
                    onClick={() => onPageChange(page - 1)}
                    aria-label="Previous page"
                >
                    <MaterialSymbol name="chevron_left" />
                </Button>
                <div className="flex items-center gap-1">
                    <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary text-sm font-bold text-on-primary">
                        {page}
                    </span>
                </div>
                <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    className="h-10 w-10 rounded-xl border-outline-variant/30 bg-surface-container text-on-surface-variant hover:text-primary"
                    disabled={disabledNext}
                    onClick={() => onPageChange(page + 1)}
                    aria-label="Next page"
                >
                    <MaterialSymbol name="chevron_right" />
                </Button>
            </div>
            {onPageSizeChange ? (
                <div className="flex items-center justify-center gap-4 sm:justify-end">
                    <span className="text-sm font-medium text-on-surface-variant">Rows per page:</span>
                    <select
                        className="cursor-pointer rounded-xl border border-outline-variant/30 bg-surface-container py-2 pl-4 pr-10 text-sm font-bold text-primary focus:ring-0"
                        value={pageSize}
                        onChange={(e) => onPageSizeChange(Number(e.target.value))}
                    >
                        {[10, 25, 50, 100].map((n) => (
                            <option key={n} value={n}>
                                {n}
                            </option>
                        ))}
                    </select>
                </div>
            ) : null}
        </div>
    );
}
