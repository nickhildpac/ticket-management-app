import { AppShell } from "@/app/shell";
import { useParams } from "@tanstack/react-router";
import { useTicket, useUpdateTicket } from "./queries";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { CommentsSection } from "./components/comments";

export function TicketDetails() {
    const params = useParams({ strict: false });
    const id = params.id as string;
    const { data: ticket, isLoading } = useTicket(id);
    const updateTicket = useUpdateTicket();

    if (isLoading) return <AppShell><div className="space-y-4"><Skeleton className="h-8 w-1/3" /><Skeleton className="h-[200px] w-full" /></div></AppShell>;
    if (!ticket) return <AppShell><div>Ticket not found</div></AppShell>;

    return (
        <AppShell>
            <div className="flex flex-col gap-6">
                {/* Header */}
                <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
                    <div>
                        <div className="flex items-center gap-3 mb-2">
                            <h1 className="text-2xl font-bold">#{ticket.id} {ticket.title}</h1>
                            <Badge variant={ticket.state === 'open' ? 'default' : 'secondary'}>{ticket.state}</Badge>
                        </div>
                        <div className="text-muted-foreground text-sm">
                            Created by {ticket.creator ? `${ticket.creator.first_name} ${ticket.creator.last_name}` : ticket.created_by} on {new Date(ticket.created_at).toLocaleString()}
                        </div>
                    </div>
                    <div className="flex items-center gap-3">
                        <Select
                            defaultValue={ticket.state}
                            onValueChange={(val: string) => updateTicket.mutate({ id: ticket.id, patch: { state: val as any } })}
                        >
                            <SelectTrigger className="w-[140px]">
                                <SelectValue placeholder="Status" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="open">Open</SelectItem>
                                <SelectItem value="pending">Pending</SelectItem>
                                <SelectItem value="resolved">Resolved</SelectItem>
                                <SelectItem value="closed">Closed</SelectItem>
                            </SelectContent>
                        </Select>
                        <Select
                            defaultValue={ticket.priority}
                            onValueChange={(val: string) => updateTicket.mutate({ id: ticket.id, patch: { priority: val as any } })}
                        >
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
                </div>

                {/* Content */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                    <div className="md:col-span-2 space-y-6">
                        <Card>
                            <CardHeader><CardTitle>Description</CardTitle></CardHeader>
                            <CardContent className="whitespace-pre-wrap text-sm leading-relaxed">
                                {ticket.description}
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
                                <div className="grid grid-cols-2 gap-2">
                                    <span className="text-muted-foreground">Assignee</span>
                                    <span className="font-medium text-right">{(ticket.assigned_to && ticket.assigned_to.length > 0) ? ticket.assigned_to.join(', ') : 'Unassigned'}</span>
                                </div>
                                <div className="grid grid-cols-2 gap-2">
                                    <span className="text-muted-foreground">Category</span>
                                    <span className="font-medium text-right">{ticket.category || 'None'}</span>
                                </div>
                            </CardContent>
                        </Card>
                    </div>
                </div>
            </div>
        </AppShell>
    );
}
