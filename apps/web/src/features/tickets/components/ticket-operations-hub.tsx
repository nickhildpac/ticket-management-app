import { useState } from "react";
import { MaterialSymbol } from "@/components/material-symbol";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { stateActionLabel, type ApiTicketState } from "@/lib/ticket-transitions";

type TicketOperationsHubProps = {
    nextStates: ApiTicketState[];
    isAdmin: boolean;
    onSelectTransition: (state: ApiTicketState) => void;
    onScrollToTransfer: () => void;
    updateError: string | null;
    updatePending: boolean;
};

export function TicketOperationsHub({
    nextStates,
    isAdmin,
    onSelectTransition,
    onScrollToTransfer,
    updateError,
    updatePending,
}: TicketOperationsHubProps) {
    const [open, setOpen] = useState(false);

    if (nextStates.length === 0 && !isAdmin) return null;

    return (
        <section className="rounded-xl border border-outline-variant/5 bg-surface-container-low p-6">
            <h3 className="mb-6 text-[0.6875rem] font-bold uppercase tracking-[0.1em] text-outline">
                Operations hub
            </h3>
            <div className="grid grid-cols-1 gap-3">
                {nextStates.length > 0 ? (
                    <Dialog open={open} onOpenChange={setOpen}>
                        <DialogTrigger asChild>
                            <button
                                type="button"
                                disabled={updatePending}
                                className="flex items-center justify-between rounded-lg border border-outline-variant/10 bg-surface-container-highest/40 p-3 text-left text-sm transition-all hover:border-primary/40 disabled:opacity-50"
                            >
                                <span className="flex items-center gap-3">
                                    <MaterialSymbol name="sync_alt" className="text-primary" size="sm" />
                                    Change status
                                </span>
                                <MaterialSymbol name="chevron_right" className="text-on-surface-variant" size="sm" />
                            </button>
                        </DialogTrigger>
                        <DialogContent className="border-outline-variant/20 bg-surface-container-low text-on-surface">
                            <DialogHeader>
                                <DialogTitle className="font-headline">Change status</DialogTitle>
                            </DialogHeader>
                            <div className="flex flex-col gap-2 pt-2">
                                {nextStates.map((s) => (
                                    <button
                                        key={s}
                                        type="button"
                                        className="rounded-lg border border-outline-variant/15 bg-surface-container-highest/40 px-4 py-3 text-left text-sm font-medium hover:border-primary/40"
                                        onClick={() => {
                                            setOpen(false);
                                            onSelectTransition(s);
                                        }}
                                    >
                                        {stateActionLabel(s)}
                                    </button>
                                ))}
                            </div>
                        </DialogContent>
                    </Dialog>
                ) : null}

                {isAdmin ? (
                    <button
                        type="button"
                        onClick={onScrollToTransfer}
                        className="flex items-center justify-between rounded-lg border border-outline-variant/10 bg-surface-container-highest/40 p-3 text-left text-sm transition-all hover:border-primary/40"
                    >
                        <span className="flex items-center gap-3">
                            <MaterialSymbol name="forward_to_inbox" className="text-primary" size="sm" />
                            Transfer ticket
                        </span>
                        <MaterialSymbol name="chevron_right" className="text-on-surface-variant" size="sm" />
                    </button>
                ) : null}
            </div>
            {updateError ? <p className="mt-4 text-xs text-error">{updateError}</p> : null}
        </section>
    );
}
