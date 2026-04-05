import { MaterialSymbol } from "@/components/material-symbol";
import type { Ticket } from "@/lib/types";

type TicketRequestorCardProps = {
    ticket: Ticket;
};

export function TicketRequestorCard({ ticket }: TicketRequestorCardProps) {
    const c = ticket.creator;
    if (!c) {
        return (
            <section className="relative overflow-hidden rounded-xl border border-outline-variant/5 bg-surface-container-low p-6">
                <h3 className="mb-6 text-[0.6875rem] font-bold uppercase tracking-[0.1em] text-outline">
                    Requestor identity
                </h3>
                <p className="text-xs text-on-surface-variant">Reporter information unavailable.</p>
            </section>
        );
    }

    const name = `${c.first_name} ${c.last_name}`.trim();

    return (
        <section className="relative overflow-hidden rounded-xl border border-outline-variant/5 bg-surface-container-low p-6">
            <div className="absolute -right-4 -top-4 h-24 w-24 rounded-full bg-primary/5 blur-2xl" />
            <h3 className="mb-6 text-[0.6875rem] font-bold uppercase tracking-[0.1em] text-outline">
                Requestor identity
            </h3>
            <div className="mb-6 flex items-center gap-4">
                <div className="flex h-12 w-12 items-center justify-center rounded-full border border-primary/20 bg-surface-container-highest font-headline text-sm font-bold text-primary">
                    {name
                        .split(/\s+/)
                        .map((p) => p[0])
                        .join("")
                        .slice(0, 2)
                        .toUpperCase()}
                </div>
                <div>
                    <p className="font-bold text-on-surface">{name}</p>
                    <p className="text-[11px] text-on-surface-variant">Ticket reporter</p>
                </div>
            </div>
            <div className="space-y-3">
                <div className="flex items-center gap-3 text-xs text-on-surface-variant">
                    <MaterialSymbol name="mail" size="sm" />
                    {c.email}
                </div>
            </div>
        </section>
    );
}
