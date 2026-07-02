import { useState } from "react";
import { formatDistanceToNow } from "date-fns";
import { MaterialSymbol } from "@/components/material-symbol";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Skeleton } from "@/components/ui/skeleton";
import { useComments } from "../queries";
import type { Ticket } from "@/lib/types";
import { cn } from "@/lib/utils";

type Filter = "all" | "comments" | "history";

type TicketActivityLedgerProps = {
    ticket: Ticket;
};

function initials(name: string) {
    const p = name.trim().split(/\s+/).filter(Boolean);
    if (p.length === 0) return "?";
    if (p.length === 1) return p[0].slice(0, 2).toUpperCase();
    return (p[0][0] + p[p.length - 1][0]).toUpperCase();
}

export function TicketActivityLedger({ ticket }: TicketActivityLedgerProps) {
    const [filter, setFilter] = useState<Filter>("all");
    const { data: comments, isLoading } = useComments(ticket.id);

    const showHistory = filter === "all" || filter === "history";
    const showComments = filter === "all" || filter === "comments";

    return (
        <div className="relative overflow-hidden rounded-xl bg-surface-container-low p-8">
            <div className="mb-10 flex items-center justify-between gap-4">
                <h2 className="flex items-center gap-2 text-lg font-semibold text-on-surface">
                    <MaterialSymbol name="history" className="text-primary" size="sm" />
                    Activity ledger
                </h2>
                <div className="flex flex-wrap gap-2">
                    {(["all", "comments", "history"] as const).map((f) => (
                        <button
                            key={f}
                            type="button"
                            onClick={() => setFilter(f)}
                            className={cn(
                                "rounded-full border px-3 py-1 text-xs transition-colors",
                                filter === f
                                    ? "border-outline-variant/15 bg-surface-container-highest text-on-surface-variant"
                                    : "border-transparent text-on-surface-variant hover:text-on-surface"
                            )}
                        >
                            {f.charAt(0).toUpperCase() + f.slice(1)}
                        </button>
                    ))}
                </div>
            </div>

            <div className="relative space-y-12 before:absolute before:bottom-2 before:left-[19px] before:top-2 before:w-px before:bg-gradient-to-b before:from-primary/40 before:via-outline-variant/20 before:to-transparent">
                {showHistory ? (
                    <>
                        <div className="relative pl-12">
                            <div className="absolute left-0 top-0 z-10 flex h-10 w-10 items-center justify-center rounded-full border border-outline-variant/15 bg-surface-container-highest">
                                <MaterialSymbol name="robot_2" className="text-on-surface-variant" size="sm" />
                            </div>
                            <div className="space-y-1">
                                <p className="text-sm text-on-surface-variant">
                                    <span className="font-semibold text-on-surface">System</span> logged ticket creation.
                                </p>
                                <time className="text-[0.625rem] font-bold uppercase tracking-widest text-outline">
                                    {formatDistanceToNow(new Date(ticket.created_at), { addSuffix: true })}
                                </time>
                            </div>
                        </div>
                        <div className="relative pl-12">
                            <div className="absolute left-0 top-0 z-10 flex h-10 w-10 items-center justify-center rounded-full border border-outline-variant/15 bg-surface-container-highest">
                                <MaterialSymbol name="update" className="text-on-surface-variant" size="sm" />
                            </div>
                            <div className="space-y-1">
                                <p className="text-sm text-on-surface-variant">
                                    <span className="font-semibold text-on-surface">Ticket</span> last updated.
                                </p>
                                <time className="text-[0.625rem] font-bold uppercase tracking-widest text-outline">
                                    {formatDistanceToNow(new Date(ticket.updated_at), { addSuffix: true })}
                                </time>
                            </div>
                        </div>
                    </>
                ) : null}

                {showComments ? (
                    isLoading ? (
                        <div className="space-y-4 pl-12">
                            <Skeleton className="h-20 w-full bg-surface-container-highest" />
                            <Skeleton className="h-20 w-full bg-surface-container-highest" />
                        </div>
                    ) : comments?.length === 0 ? (
                        <p className="pl-12 text-sm text-on-surface-variant">No comments yet.</p>
                    ) : (
                        comments?.map((comment) => {
                            const who = comment.creator
                                ? `${comment.creator.first_name} ${comment.creator.last_name}`
                                : comment.created_by;
                            return (
                                <div key={comment.id} className="relative pl-12">
                                    <div className="absolute left-0 top-0 z-10 rounded-full border border-primary/20 p-0.5">
                                        <Avatar className="h-9 w-9">
                                            <AvatarFallback className="bg-surface-container-highest text-xs text-on-surface">
                                                {initials(who)}
                                            </AvatarFallback>
                                        </Avatar>
                                    </div>
                                    <div className="rounded-xl border border-outline-variant/5 bg-surface-container-highest/40 p-5">
                                        <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
                                            <span className="text-sm font-bold text-on-surface">{who}</span>
                                            <time className="text-[0.625rem] font-bold uppercase tracking-widest text-outline">
                                                {formatDistanceToNow(new Date(comment.created_at), {
                                                    addSuffix: true,
                                                })}
                                            </time>
                                        </div>
                                        <p className="whitespace-pre-wrap text-sm leading-relaxed text-on-surface-variant">
                                            {comment.description}
                                        </p>
                                    </div>
                                </div>
                            );
                        })
                    )
                ) : null}
            </div>
        </div>
    );
}
