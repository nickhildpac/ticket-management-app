import { useCallback, useMemo, useRef, useState } from "react";
import { AppShell } from "@/app/shell";
import { useParams } from "@tanstack/react-router";
import { useTicket, useUpdateTicket } from "./queries";
import { useUsersForAssignment } from "@/features/admin/queries";
import { useUser } from "@/app/user-context";
import { Skeleton } from "@/components/ui/skeleton";
import { TicketDetailHeader } from "./components/ticket-detail-header";
import { TicketActivityLedger } from "./components/ticket-activity-ledger";
import { TicketCommentComposer } from "./components/ticket-comment-composer";
import { TicketSidebarMetadata } from "./components/ticket-sidebar-metadata";
import { TicketRequestorCard } from "./components/ticket-requestor-card";
import { TicketOperationsHub } from "./components/ticket-operations-hub";
import type { Ticket } from "@/lib/types";
import {
    filterStatesForEndUser,
    getValidNextStates,
    normalizeTicketState,
    type ApiTicketState,
} from "@/lib/ticket-transitions";

function TicketDetailBody({ ticket }: { ticket: Ticket }) {
    const { user } = useUser();
    const { data: users } = useUsersForAssignment();
    const updateTicket = useUpdateTicket();
    const composerRef = useRef<HTMLTextAreaElement>(null);

    const isAdmin = user?.role === "admin";
    const isAgent = user?.role === "agent";
    const isCreator = user?.id === ticket.created_by;
    const isAssignee = !!(user?.id && ticket.assigned_to?.includes(user.id));
    const normState = normalizeTicketState(ticket.state);
    const canCancel = isCreator && normState !== "cancelled" && normState !== "closed";

    const [assignedTo, setAssignedTo] = useState(() => ticket.assigned_to || []);
    const [skills, setSkills] = useState(() => ticket.skills || []);
    const [state, setState] = useState<string>(() => normalizeTicketState(ticket.state));
    const [priority, setPriority] = useState(() => String(ticket.priority));
    const [cancelDialogOpen, setCancelDialogOpen] = useState(false);

    const hasChanges = useMemo(() => {
        const assigneeChanged =
            JSON.stringify(assignedTo) !== JSON.stringify(ticket.assigned_to || []);
        const stateChanged = state !== normalizeTicketState(ticket.state);
        const priorityChanged = priority !== String(ticket.priority);
        const skillsChanged = JSON.stringify(skills) !== JSON.stringify(ticket.skills || []);
        return assigneeChanged || stateChanged || priorityChanged || skillsChanged;
    }, [assignedTo, state, priority, skills, ticket]);

    const handleUpdate = () => {
        if (!hasChanges) return;
        const patch: Partial<Ticket> = {};
        if (JSON.stringify(assignedTo) !== JSON.stringify(ticket.assigned_to || [])) {
            patch.assigned_to = assignedTo;
        }
        if (state !== String(ticket.state)) patch.state = state as Ticket["state"];
        if (priority !== String(ticket.priority)) patch.priority = priority as Ticket["priority"];
        if (JSON.stringify(skills) !== JSON.stringify(ticket.skills || [])) patch.skills = skills;

        updateTicket.mutate(
            { id: ticket.id, patch },
            {
                onSuccess: () => {
                    /* key bump from updated_at resets local draft */
                },
            }
        );
    };

    const handleCancelTicket = () => {
        updateTicket.mutate(
            { id: ticket.id, patch: { state: "cancelled" } },
            {
                onSuccess: () => {
                    setCancelDialogOpen(false);
                },
            }
        );
    };

    const nextStatesRaw = getValidNextStates(ticket.state);
    const nextStates: ApiTicketState[] =
        isAdmin || (isAgent && isAssignee)
            ? nextStatesRaw
            : user?.role === "user" && isCreator
              ? filterStatesForEndUser(nextStatesRaw)
              : [];

    const requestTransition = useCallback(
        (target: ApiTicketState) => {
            if (target === "cancelled") {
                setCancelDialogOpen(true);
                return;
            }
            updateTicket.mutate({ id: ticket.id, patch: { state: target } });
        },
        [ticket.id, updateTicket]
    );

    const scrollToComposer = () => {
        composerRef.current?.focus();
        composerRef.current?.scrollIntoView({ behavior: "smooth", block: "center" });
    };

    const scrollToTransfer = () => {
        document.getElementById("ticket-assign-section")?.scrollIntoView({
            behavior: "smooth",
            block: "start",
        });
    };

    const saveError =
        updateTicket.isError && updateTicket.error
            ? (updateTicket.error as Error).message
            : null;

    return (
        <div className="mx-auto max-w-7xl">
            <TicketDetailHeader
                ticket={ticket}
                onReply={scrollToComposer}
                onTransition={requestTransition}
            />

            <div className="grid grid-cols-1 gap-8 lg:grid-cols-12">
                <div className="space-y-6 lg:col-span-8">
                    <div className="rounded-xl border border-outline-variant/5 bg-surface-container-low p-6">
                        <h3 className="mb-3 text-[0.6875rem] font-bold uppercase tracking-[0.1em] text-outline">
                            Description
                        </h3>
                        <p className="whitespace-pre-wrap text-sm leading-relaxed text-on-surface-variant">
                            {ticket.description}
                        </p>
                    </div>
                    <TicketActivityLedger ticket={ticket} />
                    <TicketCommentComposer ref={composerRef} ticketId={ticket.id} />
                </div>

                <div className="space-y-6 lg:col-span-4">
                    <TicketSidebarMetadata
                        ticket={ticket}
                        users={users}
                        isAdmin={isAdmin}
                        assignedTo={assignedTo}
                        setAssignedTo={setAssignedTo}
                        skills={skills}
                        setSkills={setSkills}
                        state={state}
                        setState={setState}
                        priority={priority}
                        setPriority={setPriority}
                        hasChanges={hasChanges}
                        onSave={handleUpdate}
                        savePending={updateTicket.isPending}
                        saveError={saveError}
                        canCancel={canCancel}
                        cancelDialogOpen={cancelDialogOpen}
                        setCancelDialogOpen={setCancelDialogOpen}
                        onCancelTicket={handleCancelTicket}
                    />
                    <TicketRequestorCard ticket={ticket} />
                    <TicketOperationsHub
                        nextStates={nextStates}
                        isAdmin={isAdmin}
                        onSelectTransition={requestTransition}
                        onScrollToTransfer={scrollToTransfer}
                        updateError={saveError}
                        updatePending={updateTicket.isPending}
                    />
                </div>
            </div>
        </div>
    );
}

export function TicketDetails() {
    const params = useParams({ strict: false });
    const id = params.id as string;
    const { data: ticket, isLoading } = useTicket(id);

    if (isLoading) {
        return (
            <AppShell>
                <div className="mx-auto max-w-7xl space-y-4">
                    <Skeleton className="h-10 w-2/3 bg-surface-container-highest" />
                    <Skeleton className="h-[240px] w-full bg-surface-container-highest" />
                </div>
            </AppShell>
        );
    }
    if (!ticket) {
        return (
            <AppShell>
                <p className="text-on-surface-variant">Ticket not found</p>
            </AppShell>
        );
    }

    return (
        <AppShell>
            <TicketDetailBody key={`${ticket.id}-${ticket.updated_at}`} ticket={ticket} />
        </AppShell>
    );
}
