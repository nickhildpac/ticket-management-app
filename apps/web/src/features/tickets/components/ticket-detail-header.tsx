import { MaterialSymbol } from "@/components/material-symbol";
import { formatDistanceToNow } from "date-fns";
import type { Ticket } from "@/lib/types";
import { cn } from "@/lib/utils";
import {
    formatStateLabel,
    getValidNextStates,
    normalizeTicketState,
    type ApiTicketState,
} from "@/lib/ticket-transitions";

type TicketDetailHeaderProps = {
    ticket: Ticket;
    onReply: () => void;
    onTransition: (state: ApiTicketState) => void;
};

function priorityTone(p: string): { dot: string; label: string; labelClass: string } {
    const x = String(p).toLowerCase();
    if (x === "critical") {
        return {
            dot: "bg-error shadow-[0_0_8px_hsl(var(--mds-error)/0.4)]",
            label: "Critical priority",
            labelClass: "text-error",
        };
    }
    if (x === "high") {
        return {
            dot: "bg-tertiary shadow-[0_0_8px_hsl(var(--mds-tertiary)/0.35)]",
            label: "High priority",
            labelClass: "text-tertiary",
        };
    }
    return {
        dot: "bg-on-surface-variant/60",
        label: `${x} priority`,
        labelClass: "text-on-surface-variant",
    };
}

export function TicketDetailHeader({ ticket, onReply, onTransition }: TicketDetailHeaderProps) {
    const norm = normalizeTicketState(ticket.state);
    const next = getValidNextStates(ticket.state);
    const canResolve = next.includes("resolved");
    const tone = priorityTone(String(ticket.priority));

    return (
        <div className="mb-8 flex flex-col justify-between gap-6 md:flex-row md:items-end">
            <div className="space-y-2">
                <div className="flex flex-wrap items-center gap-3">
                    {ticket.category ? (
                        <span className="rounded bg-surface-container-high px-2 py-0.5 text-[0.6875rem] font-bold uppercase tracking-[0.05em] text-on-surface-variant">
                            {ticket.category}
                        </span>
                    ) : null}
                    <div className="flex items-center gap-1.5">
                        <span className={cn("h-2 w-2 rounded-full", tone.dot)} />
                        <span
                            className={cn(
                                "text-[0.6875rem] font-bold uppercase tracking-[0.05em]",
                                tone.labelClass
                            )}
                        >
                            {tone.label}
                        </span>
                    </div>
                </div>
                <h1 className="font-headline text-3xl font-extrabold tracking-tight text-on-surface">
                    Ticket #{ticket.ticket_number}: {ticket.title}
                </h1>
                <div className="flex flex-wrap items-center gap-4 text-sm text-on-surface-variant">
                    <span className="flex items-center gap-1.5">
                        <MaterialSymbol name="schedule" size="sm" className="text-on-surface-variant" />
                        Created {formatDistanceToNow(new Date(ticket.created_at), { addSuffix: true })}
                    </span>
                    <span className="flex items-center gap-1.5">
                        <MaterialSymbol name="radio_button_checked" size="sm" />
                        Status:{" "}
                        <span className="text-primary">{formatStateLabel(norm)}</span>
                    </span>
                </div>
            </div>
            <div className="flex flex-wrap gap-3">
                {canResolve ? (
                    <button
                        type="button"
                        onClick={() => onTransition("resolved")}
                        className="rounded-lg border border-outline-variant/15 px-5 py-2.5 text-xs font-bold uppercase tracking-wider text-primary transition-all hover:bg-primary/5"
                    >
                        Resolve ticket
                    </button>
                ) : null}
                <button
                    type="button"
                    onClick={onReply}
                    className="glow-button rounded-lg px-6 py-2.5 text-xs font-bold uppercase tracking-wider text-primary-foreground shadow-lg shadow-primary/20"
                >
                    Reply
                </button>
            </div>
        </div>
    );
}
