import { useCallback, useEffect, useState } from "react";
import { cn } from "@/lib/utils";

const storageKey = (ticketId: string) => `ticket-tasks:${ticketId}`;

type LocalTask = { id: string; label: string; done: boolean };

function loadTasks(ticketId: string): LocalTask[] {
    try {
        const raw = localStorage.getItem(storageKey(ticketId));
        if (!raw) return [];
        const parsed = JSON.parse(raw) as LocalTask[];
        return Array.isArray(parsed) ? parsed : [];
    } catch {
        return [];
    }
}

function saveTasks(ticketId: string, tasks: LocalTask[]) {
    localStorage.setItem(storageKey(ticketId), JSON.stringify(tasks));
}

type TicketActionItemsLocalProps = {
    ticketId: string;
};

export function TicketActionItemsLocal({ ticketId }: TicketActionItemsLocalProps) {
    const [tasks, setTasks] = useState<LocalTask[]>(() => loadTasks(ticketId));

    useEffect(() => {
        setTasks(loadTasks(ticketId));
    }, [ticketId]);

    const persist = useCallback(
        (next: LocalTask[]) => {
            setTasks(next);
            saveTasks(ticketId, next);
        },
        [ticketId]
    );

    const toggle = (id: string) => {
        persist(tasks.map((t) => (t.id === id ? { ...t, done: !t.done } : t)));
    };

    const addTask = () => {
        const label = window.prompt("Task label");
        if (!label?.trim()) return;
        persist([
            ...tasks,
            { id: crypto.randomUUID(), label: label.trim(), done: false },
        ]);
    };

    return (
        <section className="rounded-xl border border-outline-variant/5 bg-surface-container-low p-6">
            <h3 className="mb-6 text-[0.6875rem] font-bold uppercase tracking-[0.1em] text-outline">
                Action items
            </h3>
            <div className="space-y-4">
                {tasks.length === 0 ? (
                    <p className="text-xs text-on-surface-variant">No tasks yet. Add one below (saved in this browser).</p>
                ) : (
                    tasks.map((t) => (
                        <label
                            key={t.id}
                            className="group flex cursor-pointer items-center gap-3"
                        >
                            <input
                                type="checkbox"
                                checked={t.done}
                                onChange={() => toggle(t.id)}
                                className={cn(
                                    "h-5 w-5 rounded border-outline-variant/30 bg-surface-container-highest text-primary focus:ring-primary/20"
                                )}
                            />
                            <span
                                className={cn(
                                    "text-xs text-on-surface-variant transition-colors group-hover:text-on-surface",
                                    t.done && "line-through opacity-60"
                                )}
                            >
                                {t.label}
                            </span>
                        </label>
                    ))
                )}
            </div>
            <button
                type="button"
                onClick={addTask}
                className="mt-6 w-full rounded-lg border border-dashed border-outline-variant/20 py-2 text-[10px] font-bold uppercase tracking-widest text-outline transition-all hover:border-primary/40 hover:text-primary"
            >
                + Add task
            </button>
        </section>
    );
}
