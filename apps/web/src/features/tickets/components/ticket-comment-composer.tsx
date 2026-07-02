import { forwardRef, useState } from "react";
import { MaterialSymbol } from "@/components/material-symbol";
import { useCreateComment } from "../queries";
import { cn } from "@/lib/utils";

type TicketCommentComposerProps = {
    ticketId: string;
};

export const TicketCommentComposer = forwardRef<HTMLTextAreaElement, TicketCommentComposerProps>(
    function TicketCommentComposer({ ticketId }, ref) {
        const [body, setBody] = useState("");
        const createComment = useCreateComment();

        const submit = () => {
            const t = body.trim();
            if (!t) return;
            createComment.mutate(
                { ticketId, body: t },
                {
                    onSuccess: () => setBody(""),
                }
            );
        };

        return (
            <div className="mt-12 rounded-xl border border-outline-variant/10 bg-surface-container-highest transition-all focus-within:border-primary/40 focus-within:ring-1 focus-within:ring-primary/10">
                <div className="flex items-center gap-4 border-b border-outline-variant/5 px-4 py-3">
                    <button type="button" className="p-1 text-on-surface-variant hover:text-primary" aria-hidden tabIndex={-1}>
                        <MaterialSymbol name="format_bold" size="sm" />
                    </button>
                    <button type="button" className="p-1 text-on-surface-variant hover:text-primary" aria-hidden tabIndex={-1}>
                        <MaterialSymbol name="format_italic" size="sm" />
                    </button>
                    <button type="button" className="p-1 text-on-surface-variant hover:text-primary" aria-hidden tabIndex={-1}>
                        <MaterialSymbol name="link" size="sm" />
                    </button>
                    <button type="button" className="p-1 text-on-surface-variant hover:text-primary" aria-hidden tabIndex={-1}>
                        <MaterialSymbol name="code" size="sm" />
                    </button>
                    <div className="mx-1 h-4 w-px bg-outline-variant/20" />
                    <button type="button" className="p-1 text-on-surface-variant hover:text-primary" aria-hidden tabIndex={-1}>
                        <MaterialSymbol name="attach_file" size="sm" />
                    </button>
                </div>
                <textarea
                    ref={ref}
                    rows={4}
                    value={body}
                    onChange={(e) => setBody(e.target.value)}
                    placeholder="Write a response or internal note..."
                    className="w-full resize-none border-none bg-transparent p-4 text-sm text-on-surface focus:ring-0"
                />
                <div className="flex flex-col gap-2 border-t border-outline-variant/5 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
                    <label
                        className="flex cursor-not-allowed items-center gap-2 opacity-60"
                        title="Internal notes are not available yet."
                    >
                        <input type="checkbox" disabled className="rounded border-outline-variant/30" />
                        <span className="text-xs text-on-surface-variant">Mark as internal note</span>
                    </label>
                    <button
                        type="button"
                        onClick={submit}
                        disabled={createComment.isPending || !body.trim()}
                        className={cn(
                            "glow-button rounded-lg px-5 py-2 text-xs font-bold uppercase tracking-widest text-on-primary disabled:opacity-50"
                        )}
                    >
                        {createComment.isPending ? "Posting…" : "Post response"}
                    </button>
                </div>
                {createComment.isError ? (
                    <p className="border-t border-outline-variant/5 px-4 py-2 text-xs text-error">
                        {(createComment.error as Error)?.message ?? "Failed to post comment."}
                    </p>
                ) : null}
            </div>
        );
    }
);
