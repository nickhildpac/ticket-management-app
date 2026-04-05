import { Loader2, Tag, X } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { Ticket, User } from "@/lib/types";
import { formatStateLabel, normalizeTicketState } from "@/lib/ticket-transitions";

const VALID_SKILLS = [
    "incident-management",
    "major-incident",
    "root-cause-analysis",
    "log-analysis",
    "production-support",
    "sla-management",
    "post-incident-review",
] as const;

const STATE_OPTIONS = [
    "open",
    "pending",
    "in progress",
    "resolved",
    "closed",
    "cancelled",
] as const;

type TicketSidebarMetadataProps = {
    ticket: Ticket;
    users: User[] | undefined;
    isAdmin: boolean;
    assignedTo: string[];
    setAssignedTo: (v: string[]) => void;
    skills: string[];
    setSkills: (v: string[]) => void;
    state: string;
    setState: (v: string) => void;
    priority: string;
    setPriority: (v: string) => void;
    hasChanges: boolean;
    onSave: () => void;
    savePending: boolean;
    saveError: string | null;
    canCancel: boolean;
    cancelDialogOpen: boolean;
    setCancelDialogOpen: (v: boolean) => void;
    onCancelTicket: () => void;
};

export function TicketSidebarMetadata({
    ticket,
    users,
    isAdmin,
    assignedTo,
    setAssignedTo,
    skills,
    setSkills,
    state,
    setState,
    priority,
    setPriority,
    hasChanges,
    onSave,
    savePending,
    saveError,
    canCancel,
    cancelDialogOpen,
    setCancelDialogOpen,
    onCancelTicket,
}: TicketSidebarMetadataProps) {
    const handleAddAssignee = (userId: string) => {
        if (!assignedTo.includes(userId)) setAssignedTo([...assignedTo, userId]);
    };
    const handleRemoveAssignee = (userId: string) => {
        setAssignedTo(assignedTo.filter((id) => id !== userId));
    };
    const handleAddSkill = (skill: string) => {
        if (!skills.includes(skill)) setSkills([...skills, skill]);
    };
    const handleRemoveSkill = (skill: string) => {
        setSkills(skills.filter((s) => s !== skill));
    };

    const created = new Date(ticket.created_at);
    const updated = new Date(ticket.updated_at);

    return (
        <section className="rounded-xl border border-outline-variant/5 bg-surface-container-low p-6">
            <h3 className="mb-6 text-[0.6875rem] font-bold uppercase tracking-[0.1em] text-outline">
                Metadata ledger
            </h3>
            <div className="space-y-5">
                <div className="flex items-center justify-between gap-2">
                    <span className="text-xs text-on-surface-variant">Created</span>
                    <span className="text-right text-xs font-semibold text-on-surface">
                        {created.toLocaleString()}
                    </span>
                </div>
                <div className="flex items-center justify-between gap-2">
                    <span className="text-xs text-on-surface-variant">Updated</span>
                    <span className="text-right text-xs font-semibold text-on-surface">
                        {updated.toLocaleString()}
                    </span>
                </div>
                {ticket.category ? (
                    <div className="flex items-center justify-between gap-2">
                        <span className="text-xs text-on-surface-variant">Category</span>
                        <span className="rounded bg-surface-container-highest px-2 py-1 text-[10px] font-bold uppercase text-primary">
                            {ticket.category}
                        </span>
                    </div>
                ) : null}
                <div className="flex items-center justify-between gap-2">
                    <span className="text-xs text-on-surface-variant">Current status</span>
                    <span className="text-xs font-semibold capitalize text-on-surface">
                        {formatStateLabel(normalizeTicketState(ticket.state))}
                    </span>
                </div>
                <p className="text-[11px] text-on-surface-variant/80">SLA not configured.</p>
            </div>

            <div id="ticket-assign-section" className="mt-8 space-y-6 border-t border-outline-variant/10 pt-8">
                <h4 className="text-xs font-bold uppercase tracking-wider text-outline">Assignment &amp; routing</h4>
                <div className="space-y-2">
                    <label className="text-[11px] text-on-surface-variant">Assigned to</label>
                    {isAdmin ? (
                        <>
                            {assignedTo.length > 0 ? (
                                <div className="mb-2 flex flex-wrap gap-2">
                                    {assignedTo.map((userId) => {
                                        const u = users?.find((x) => x.id === userId);
                                        return (
                                            <Badge
                                                key={userId}
                                                variant="secondary"
                                                className="gap-1 border-outline-variant/20 bg-surface-container-highest pr-1"
                                            >
                                                <span>
                                                    {u ? `${u.first_name} ${u.last_name}` : "Unknown"}
                                                </span>
                                                <button
                                                    type="button"
                                                    onClick={() => handleRemoveAssignee(userId)}
                                                    className="ml-1 rounded-full p-0.5 hover:bg-on-surface/10"
                                                >
                                                    <X className="h-3 w-3" />
                                                </button>
                                            </Badge>
                                        );
                                    })}
                                </div>
                            ) : null}
                            <Select onValueChange={handleAddAssignee}>
                                <SelectTrigger className="border-outline-variant/20 bg-surface-container-highest">
                                    <SelectValue placeholder="Add assignee…" />
                                </SelectTrigger>
                                <SelectContent>
                                    {users
                                        ?.filter((u) => !assignedTo.includes(u.id))
                                        .map((u) => (
                                            <SelectItem key={u.id} value={u.id}>
                                                {u.first_name} {u.last_name} ({u.email})
                                            </SelectItem>
                                        ))}
                                </SelectContent>
                            </Select>
                        </>
                    ) : (
                        <div className="flex flex-wrap gap-2">
                            {ticket.assigned_to?.length ? (
                                ticket.assigned_to.map((userId) => {
                                    const u = users?.find((x) => x.id === userId);
                                    return (
                                        <Badge key={userId} variant="secondary" className="bg-surface-container-highest">
                                            {u ? `${u.first_name} ${u.last_name}` : "Unknown"}
                                        </Badge>
                                    );
                                })
                            ) : (
                                <span className="text-xs text-on-surface-variant">Unassigned</span>
                            )}
                        </div>
                    )}
                </div>

                <div className="space-y-2">
                    <label className="text-[11px] text-on-surface-variant">Skills</label>
                    {isAdmin ? (
                        <>
                            {skills.length > 0 ? (
                                <div className="mb-2 flex flex-wrap gap-2">
                                    {skills.map((skill) => (
                                        <Badge
                                            key={skill}
                                            variant="outline"
                                            className="gap-1 border-primary/20 bg-primary/5 pr-1 text-primary"
                                        >
                                            <span>{skill.replace(/-/g, " ")}</span>
                                            <button
                                                type="button"
                                                onClick={() => handleRemoveSkill(skill)}
                                                className="ml-1 rounded-full p-0.5 hover:bg-primary/15"
                                            >
                                                <X className="h-3 w-3" />
                                            </button>
                                        </Badge>
                                    ))}
                                </div>
                            ) : null}
                            <Select onValueChange={handleAddSkill}>
                                <SelectTrigger className="border-outline-variant/20 bg-surface-container-highest">
                                    <div className="flex items-center gap-2">
                                        <Tag className="h-3.5 w-3.5 text-on-surface-variant" />
                                        <SelectValue placeholder="Add skill…" />
                                    </div>
                                </SelectTrigger>
                                <SelectContent>
                                    {VALID_SKILLS.filter((s) => !skills.includes(s)).map((skill) => (
                                        <SelectItem key={skill} value={skill}>
                                            {skill.replace(/-/g, " ")}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </>
                    ) : (
                        <div className="flex flex-wrap gap-2">
                            {ticket.skills?.length ? (
                                ticket.skills.map((skill) => (
                                    <Badge
                                        key={skill}
                                        variant="outline"
                                        className="border-primary/20 bg-primary/5 text-primary"
                                    >
                                        {skill.replace(/-/g, " ")}
                                    </Badge>
                                ))
                            ) : (
                                <span className="text-xs italic text-on-surface-variant">None</span>
                            )}
                        </div>
                    )}
                </div>

                <div className="grid grid-cols-2 items-center gap-2">
                    <span className="text-xs text-on-surface-variant">Priority</span>
                    {isAdmin ? (
                        <Select value={priority} onValueChange={setPriority}>
                            <SelectTrigger className="h-9 border-outline-variant/20 bg-surface-container-highest">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="low">Low</SelectItem>
                                <SelectItem value="medium">Medium</SelectItem>
                                <SelectItem value="high">High</SelectItem>
                                <SelectItem value="critical">Critical</SelectItem>
                            </SelectContent>
                        </Select>
                    ) : (
                        <span className="text-right text-xs font-semibold capitalize text-on-surface">
                            {ticket.priority}
                        </span>
                    )}
                </div>
                <div className="grid grid-cols-2 items-center gap-2">
                    <span className="text-xs text-on-surface-variant">State</span>
                    {isAdmin ? (
                        <Select value={state} onValueChange={setState}>
                            <SelectTrigger className="h-9 border-outline-variant/20 bg-surface-container-highest">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                {STATE_OPTIONS.map((s) => (
                                    <SelectItem key={s} value={s}>
                                        {formatStateLabel(s)}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    ) : (
                        <span className="text-right text-xs font-semibold text-on-surface">
                            {formatStateLabel(normalizeTicketState(ticket.state))}
                        </span>
                    )}
                </div>
            </div>

            {isAdmin ? (
                <div className="mt-6 space-y-3 border-t border-outline-variant/10 pt-6">
                    <Button
                        onClick={onSave}
                        disabled={!hasChanges || savePending}
                        className="glow-button w-full border-0 font-bold uppercase tracking-wide text-on-primary"
                    >
                        {savePending ? (
                            <>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                Updating…
                            </>
                        ) : (
                            "Save changes"
                        )}
                    </Button>
                    {saveError ? <p className="text-center text-xs text-error">{saveError}</p> : null}
                </div>
            ) : null}

            {canCancel ? (
                <div className="mt-4">
                    <Dialog open={cancelDialogOpen} onOpenChange={setCancelDialogOpen}>
                        <DialogTrigger asChild>
                            <Button variant="destructive" className="w-full">
                                Cancel ticket
                            </Button>
                        </DialogTrigger>
                        <DialogContent>
                            <DialogHeader>
                                <DialogTitle>Cancel ticket</DialogTitle>
                                <DialogDescription>
                                    Are you sure? This action cannot be undone.
                                </DialogDescription>
                            </DialogHeader>
                            <DialogFooter>
                                <Button variant="outline" onClick={() => setCancelDialogOpen(false)}>
                                    Go back
                                </Button>
                                <Button variant="destructive" onClick={onCancelTicket} disabled={savePending}>
                                    {savePending ? (
                                        <>
                                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                            Cancelling…
                                        </>
                                    ) : (
                                        "Yes, cancel"
                                    )}
                                </Button>
                            </DialogFooter>
                        </DialogContent>
                    </Dialog>
                </div>
            ) : null}
        </section>
    );
}
