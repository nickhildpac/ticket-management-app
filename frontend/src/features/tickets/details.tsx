import { AppShell } from "@/app/shell";
import { useParams } from "@tanstack/react-router";
import { useTicket, useUpdateTicket } from "./queries";
import { useUsersForAssignment } from "@/features/admin/queries";
import { useUser } from "@/app/user-context";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { CommentsSection } from "./components/comments";
import { useState, useEffect } from "react";
import { Loader2, X, Tag } from "lucide-react";

const VALID_SKILLS = [
    "incident-management",
    "major-incident",
    "root-cause-analysis",
    "log-analysis",
    "production-support",
    "sla-management",
    "post-incident-review",
];

export function TicketDetails() {
    const params = useParams({ strict: false });
    const id = params.id as string;
    const { user } = useUser();
    const { data: ticket, isLoading } = useTicket(id);
    const { data: users } = useUsersForAssignment();
    const updateTicket = useUpdateTicket();

    const isAdmin = user?.role === "admin";
    const isCreator = user?.id === ticket?.created_by;
    const canCancel = isCreator && ticket?.state !== 'cancelled' && ticket?.state !== 'closed';

    // Local state for editable fields
    const [description, setDescription] = useState("");
    const [assignedTo, setAssignedTo] = useState<string[]>([]);
    const [skills, setSkills] = useState<string[]>([]);
    const [state, setState] = useState("");
    const [priority, setPriority] = useState("");
    const [hasChanges, setHasChanges] = useState(false);
    const [cancelDialogOpen, setCancelDialogOpen] = useState(false);

    // Initialize local state when ticket data loads
    useEffect(() => {
        if (ticket) {
            setDescription(ticket.description);
            setAssignedTo(ticket.assigned_to || []);
            setSkills(ticket.skills || []);
            setState(String(ticket.state));
            setPriority(String(ticket.priority));
        }
    }, [ticket]);

    // Track changes
    useEffect(() => {
        if (ticket) {
            const descChanged = description !== ticket.description;
            const assigneeChanged = JSON.stringify(assignedTo) !== JSON.stringify(ticket.assigned_to || []);
            const stateChanged = state !== ticket.state;
            const priorityChanged = priority !== ticket.priority;
            const skillsChanged = JSON.stringify(skills) !== JSON.stringify(ticket.skills || []);
            setHasChanges(descChanged || assigneeChanged || stateChanged || priorityChanged || skillsChanged);
        }
    }, [description, assignedTo, state, priority, skills, ticket]);

    const handleUpdate = () => {
        if (!ticket || !hasChanges) return;

        const patch: any = {};
        if (description !== ticket.description) patch.description = description;
        if (JSON.stringify(assignedTo) !== JSON.stringify(ticket.assigned_to || [])) patch.assigned_to = assignedTo;
        if (state !== ticket.state) patch.state = state;
        if (priority !== ticket.priority) patch.priority = priority;
        if (JSON.stringify(skills) !== JSON.stringify(ticket.skills || [])) patch.skills = skills;

        updateTicket.mutate(
            { id: ticket.id, patch },
            {
                onSuccess: () => {
                    setHasChanges(false);
                }
            }
        );
    };

    const handleAddAssignee = (userId: string) => {
        if (!assignedTo.includes(userId)) {
            setAssignedTo([...assignedTo, userId]);
        }
    };

    const handleRemoveAssignee = (userId: string) => {
        setAssignedTo(assignedTo.filter(id => id !== userId));
    };

    const handleAddSkill = (skill: string) => {
        if (!skills.includes(skill)) {
            setSkills([...skills, skill]);
        }
    };

    const handleRemoveSkill = (skill: string) => {
        setSkills(skills.filter(s => s !== skill));
    };

    const handleCancelTicket = () => {
        if (!ticket) return;
        updateTicket.mutate(
            { id: ticket.id, patch: { state: 'cancelled' } },
            {
                onSuccess: () => {
                    setCancelDialogOpen(false);
                    setHasChanges(false);
                }
            }
        );
    };

    if (isLoading) return <AppShell><div className="space-y-4"><Skeleton className="h-8 w-1/3" /><Skeleton className="h-[200px] w-full" /></div></AppShell>;
    if (!ticket) return <AppShell><div>Ticket not found</div></AppShell>;

    return (
        <AppShell>
            <div className="flex flex-col gap-6">
                {/* Header */}
                <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
                    <div>
                        <div className="flex items-center gap-3 mb-2">
                            <span className="text-muted-foreground font-mono">#{ticket.ticket_number}</span>
                            <h1 className="text-2xl font-bold">{ticket.title}</h1>
                            {isAdmin ? (
                                <Badge variant={state === 'open' ? 'default' : 'secondary'}>{state}</Badge>
                            ) : (
                                <Badge variant={ticket.state === 'open' ? 'default' : 'secondary'}>{ticket.state}</Badge>
                            )}
                        </div>
                        <div className="text-muted-foreground text-sm">
                            Created by {ticket.creator ? `${ticket.creator.first_name} ${ticket.creator.last_name}` : ticket.created_by} on {new Date(ticket.created_at).toLocaleString()}
                        </div>
                    </div>
                    {isAdmin && (
                        <div className="flex items-center gap-3">
                            <Select value={state} onValueChange={setState}>
                                <SelectTrigger className="w-[140px]">
                                    <SelectValue placeholder="Status" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="open">Open</SelectItem>
                                    <SelectItem value="pending">Pending</SelectItem>
                                    <SelectItem value="resolved">Resolved</SelectItem>
                                    <SelectItem value="closed">Closed</SelectItem>
                                    <SelectItem value="cancelled">Cancelled</SelectItem>
                                </SelectContent>
                            </Select>
                            <Select value={priority} onValueChange={setPriority}>
                                <SelectTrigger className="w-[140px]">
                                    <SelectValue placeholder="Priority" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="low">Low</SelectItem>
                                    <SelectItem value="medium">Medium</SelectItem>
                                    <SelectItem value="high">High</SelectItem>
                                    <SelectItem value="critical">Critical</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    )}
                </div>

                {/* Content */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                    <div className="md:col-span-2 space-y-6">
                        <Card>
                            <CardHeader><CardTitle>Description</CardTitle></CardHeader>
                            <CardContent>
                                <Textarea
                                    value={description}
                                    onChange={(e) => setDescription(e.target.value)}
                                    className="min-h-[150px] font-mono text-sm"
                                    placeholder="Enter description..."
                                />
                            </CardContent>
                        </Card>

                        <Tabs defaultValue="activity">
                            <TabsList>
                                <TabsTrigger value="activity">Activity</TabsTrigger>
                                <TabsTrigger value="comments">Comments</TabsTrigger>
                            </TabsList>
                            <TabsContent value="activity">
                                <div className="p-4 text-center text-muted-foreground border rounded-md">Activity Log functionality not implemented yet.</div>
                            </TabsContent>
                            <TabsContent value="comments">
                                <CommentsSection ticketId={ticket.id} />
                            </TabsContent>
                        </Tabs>
                    </div>
                    <div className="space-y-6">
                        <Card>
                            <CardHeader><CardTitle className="text-base">Details</CardTitle></CardHeader>
                            <CardContent className="space-y-4 text-sm">
                                <div className="space-y-2">
                                    <label className="text-muted-foreground text-xs">Assigned To</label>
                                    {isAdmin ? (
                                        <>
                                            {/* Selected users chips */}
                                            {assignedTo.length > 0 && (
                                                <div className="flex flex-wrap gap-2 mb-2">
                                                    {assignedTo.map((userId) => {
                                                        const assignedUser = users?.find(u => u.id === userId);
                                                        return (
                                                            <Badge
                                                                key={userId}
                                                                variant="secondary"
                                                                className="gap-1 pr-1"
                                                            >
                                                                <span>
                                                                    {assignedUser
                                                                        ? `${assignedUser.first_name} ${assignedUser.last_name}`
                                                                        : "Unknown"}
                                                                </span>
                                                                <button
                                                                    onClick={() => handleRemoveAssignee(userId)}
                                                                    className="ml-1 rounded-full hover:bg-muted-foreground/20 p-0.5"
                                                                >
                                                                    <X className="h-3 w-3" />
                                                                </button>
                                                            </Badge>
                                                        );
                                                    })}
                                                </div>
                                            )}
                                            {/* Dropdown to add users */}
                                            <Select onValueChange={handleAddAssignee}>
                                                <SelectTrigger>
                                                    <SelectValue placeholder="Add assignee..." />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    {users
                                                        ?.filter(u => !assignedTo.includes(u.id))
                                                        .map((user) => (
                                                            <SelectItem key={user.id} value={user.id}>
                                                                {user.first_name} {user.last_name} ({user.email}) - {user.role}
                                                            </SelectItem>
                                                        ))}
                                                </SelectContent>
                                            </Select>
                                        </>
                                    ) : (
                                        <div className="space-y-1">
                                            {(ticket.assigned_to && ticket.assigned_to.length > 0) ? (
                                                ticket.assigned_to.map((userId) => {
                                                    const assignedUser = users?.find(u => u.id === userId);
                                                    return (
                                                        <Badge key={userId} variant="secondary">
                                                            {assignedUser
                                                                ? `${assignedUser.first_name} ${assignedUser.last_name}`
                                                                : "Unknown"}
                                                        </Badge>
                                                    );
                                                })
                                            ) : (
                                                <span className="text-muted-foreground">Unassigned</span>
                                            )}
                                        </div>
                                    )}
                                </div>

                                <div className="space-y-2">
                                    <label className="text-muted-foreground text-xs">Skills Required</label>
                                    {isAdmin ? (
                                        <>
                                            {/* Selected skills chips */}
                                            {skills.length > 0 && (
                                                <div className="flex flex-wrap gap-2 mb-2">
                                                    {skills.map((skill) => (
                                                        <Badge
                                                            key={skill}
                                                            variant="outline"
                                                            className="gap-1 pr-1 bg-primary/5 border-primary/20 text-primary"
                                                        >
                                                            <span>{skill.replace(/-/g, ' ')}</span>
                                                            <button
                                                                onClick={() => handleRemoveSkill(skill)}
                                                                className="ml-1 rounded-full hover:bg-primary/20 p-0.5"
                                                            >
                                                                <X className="h-3 w-3" />
                                                            </button>
                                                        </Badge>
                                                    ))}
                                                </div>
                                            )}
                                            {/* Dropdown to add skills */}
                                            <Select onValueChange={handleAddSkill}>
                                                <SelectTrigger className="h-9">
                                                    <div className="flex items-center gap-2">
                                                        <Tag className="h-3.5 w-3.5 text-muted-foreground" />
                                                        <SelectValue placeholder="Add skill..." />
                                                    </div>
                                                </SelectTrigger>
                                                <SelectContent>
                                                    {VALID_SKILLS
                                                        .filter(s => !skills.includes(s))
                                                        .map((skill) => (
                                                            <SelectItem key={skill} value={skill}>
                                                                {skill.replace(/-/g, ' ')}
                                                            </SelectItem>
                                                        ))}
                                                </SelectContent>
                                            </Select>
                                        </>
                                    ) : (
                                        <div className="flex flex-wrap gap-2">
                                            {(ticket.skills && ticket.skills.length > 0) ? (
                                                ticket.skills.map((skill) => (
                                                    <Badge key={skill} variant="outline" className="bg-primary/5 border-primary/20 text-primary">
                                                        {skill.replace(/-/g, ' ')}
                                                    </Badge>
                                                ))
                                            ) : (
                                                <span className="text-muted-foreground italic">None required</span>
                                            )}
                                        </div>
                                    )}
                                </div>

                                <div className="grid grid-cols-2 gap-2">
                                    <span className="text-muted-foreground">Priority</span>
                                    {isAdmin ? (
                                        <Select value={priority} onValueChange={setPriority}>
                                            <SelectTrigger className="h-8">
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
                                        <span className="font-medium text-right capitalize">{ticket.priority}</span>
                                    )}
                                </div>
                                <div className="grid grid-cols-2 gap-2">
                                    <span className="text-muted-foreground">State</span>
                                    {isAdmin ? (
                                        <Select value={state} onValueChange={setState}>
                                            <SelectTrigger className="h-8">
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="open">Open</SelectItem>
                                                <SelectItem value="pending">Pending</SelectItem>
                                                <SelectItem value="resolved">Resolved</SelectItem>
                                                <SelectItem value="closed">Closed</SelectItem>
                                                <SelectItem value="cancelled">Cancelled</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    ) : (
                                        <span className="font-medium text-right capitalize">{ticket.state}</span>
                                    )}
                                </div>
                            </CardContent>
                        </Card>

                        <Card>
                            <CardHeader><CardTitle className="text-base">Actions</CardTitle></CardHeader>
                            <CardContent className="space-y-3">
                                <Button
                                    onClick={handleUpdate}
                                    disabled={!hasChanges || updateTicket.isPending}
                                    className="w-full"
                                >
                                    {updateTicket.isPending ? (
                                        <>
                                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                            Updating...
                                        </>
                                    ) : (
                                        "Update Ticket"
                                    )}
                                </Button>

                                {canCancel && (
                                    <Dialog open={cancelDialogOpen} onOpenChange={setCancelDialogOpen}>
                                        <DialogTrigger asChild>
                                            <Button variant="destructive" className="w-full">
                                                Cancel Ticket
                                            </Button>
                                        </DialogTrigger>
                                        <DialogContent>
                                            <DialogHeader>
                                                <DialogTitle>Cancel Ticket</DialogTitle>
                                                <DialogDescription>
                                                    Are you sure you want to cancel this ticket? This action cannot be undone.
                                                </DialogDescription>
                                            </DialogHeader>
                                            <DialogFooter>
                                                <Button
                                                    variant="outline"
                                                    onClick={() => setCancelDialogOpen(false)}
                                                >
                                                    No, go back
                                                </Button>
                                                <Button
                                                    variant="destructive"
                                                    onClick={handleCancelTicket}
                                                    disabled={updateTicket.isPending}
                                                >
                                                    {updateTicket.isPending ? (
                                                        <>
                                                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                                            Cancelling...
                                                        </>
                                                    ) : (
                                                        "Yes, cancel ticket"
                                                    )}
                                                </Button>
                                            </DialogFooter>
                                        </DialogContent>
                                    </Dialog>
                                )}
                            </CardContent>
                        </Card>
                    </div>
                </div>
            </div>
        </AppShell>
    );
}
