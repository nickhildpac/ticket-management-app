import { AppShell } from "@/app/shell";
import { useTicketStats, useTickets } from "../tickets/queries";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Link } from "@tanstack/react-router";
import {
    Ticket as TicketIcon,
    CheckCircle2,
    Clock,
    AlertCircle,
    User as UserIcon,
    ChevronRight
} from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { getTicketStateString, getTicketPriorityString } from "@/lib/types";

export function Dashboard() {
    const { data: stats, isLoading: statsLoading } = useTicketStats();
    const { data: recentTickets, isLoading: ticketsLoading } = useTickets({ page: 1 });

    const recent = recentTickets?.items.slice(0, 5) || [];

    const statCards = [
        { title: "Total Tickets", value: stats?.total, icon: TicketIcon, color: "text-blue-500", bg: "bg-blue-50 dark:bg-blue-900/20" },
        { title: "Open", value: stats?.open, icon: AlertCircle, color: "text-red-500", bg: "bg-red-50 dark:bg-red-900/20" },
        { title: "Pending", value: stats?.pending, icon: Clock, color: "text-yellow-500", bg: "bg-yellow-50 dark:bg-yellow-900/20" },
        { title: "Resolved", value: stats?.resolved, icon: CheckCircle2, color: "text-green-500", bg: "bg-green-50 dark:bg-green-900/20" },
        { title: "My Tickets", value: stats?.mine, icon: UserIcon, color: "text-purple-500", bg: "bg-purple-50 dark:bg-purple-900/20" },
    ];

    return (
        <AppShell>
            <div className="space-y-8 pb-10">
                <div className="flex items-center justify-between">
                    <div>
                        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
                        <p className="text-muted-foreground text-lg">System overview and recent activity.</p>
                    </div>
                </div>

                <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-5">
                    {statCards.map((card) => (
                        <Card key={card.title} className="border-none shadow-sm overflow-hidden relative">
                            <div className={`absolute top-0 left-0 w-1 h-full ${card.color.replace('text', 'bg')}`} />
                            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                                <CardTitle className="text-sm font-medium text-muted-foreground uppercase tracking-wider">{card.title}</CardTitle>
                                <div className={`p-2 rounded-full ${card.bg}`}>
                                    <card.icon className={`h-4 w-4 ${card.color}`} />
                                </div>
                            </CardHeader>
                            <CardContent>
                                <div className="text-3xl font-bold">
                                    {statsLoading ? <Skeleton className="h-8 w-12" /> : card.value}
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>

                <div className="grid gap-6 grid-cols-1 md:grid-cols-1">
                    <Card className="shadow-sm">
                        <CardHeader className="flex flex-row items-center justify-between">
                            <div>
                                <CardTitle>Recent Tickets</CardTitle>
                                <p className="text-sm text-muted-foreground">Latest issues reported in the system.</p>
                            </div>
                            <Button variant="outline" size="sm" asChild>
                                <Link to="/tickets" className="flex items-center gap-1 text-xs">
                                    View All <ChevronRight className="h-3 w-3" />
                                </Link>
                            </Button>
                        </CardHeader>
                        <CardContent>
                            <div className="rounded-md border">
                                <Table>
                                    <TableHeader className="bg-muted/50">
                                        <TableRow>
                                            <TableHead className="py-3">Title</TableHead>
                                            <TableHead>Status</TableHead>
                                            <TableHead>Priority</TableHead>
                                            <TableHead className="text-right">Date</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {ticketsLoading ? (
                                            Array.from({ length: 5 }).map((_, i) => (
                                                <TableRow key={i}>
                                                    <TableCell><Skeleton className="h-4 w-40" /></TableCell>
                                                    <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                                                    <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                                                    <TableCell><Skeleton className="h-4 w-16 ml-auto" /></TableCell>
                                                </TableRow>
                                            ))
                                        ) : recent.length === 0 ? (
                                            <TableRow>
                                                <TableCell colSpan={4} className="text-center py-10 text-muted-foreground">No recent tickets</TableCell>
                                            </TableRow>
                                        ) : (
                                            recent.map((ticket) => {
                                                const stateStr = typeof ticket.state === 'number' ? getTicketStateString(ticket.state) : ticket.state;
                                                const priorityStr = typeof ticket.priority === 'number' ? getTicketPriorityString(ticket.priority) : ticket.priority;

                                                return (
                                                    <TableRow key={ticket.id} className="hover:bg-muted/30 transition-colors">
                                                        <TableCell className="py-4">
                                                            <Link to="/tickets/$id" params={{ id: ticket.id }} className="block">
                                                                <div className="font-semibold text-sm">{ticket.title}</div>
                                                                <div className="text-xs text-muted-foreground mt-1">#{ticket.id.slice(0, 8)}</div>
                                                            </Link>
                                                        </TableCell>
                                                        <TableCell>
                                                            <Badge variant={stateStr === 'open' ? 'default' : 'secondary'} className="capitalize">
                                                                {stateStr}
                                                            </Badge>
                                                        </TableCell>
                                                        <TableCell>
                                                            <span className={`text-xs font-medium px-2 py-1 rounded-full border ${
                                                                priorityStr === 'critical' ? 'bg-red-50 text-red-700 border-red-200' :
                                                                priorityStr === 'high' ? 'bg-orange-50 text-orange-700 border-orange-200' :
                                                                priorityStr === 'medium' ? 'bg-yellow-50 text-yellow-700 border-yellow-200' :
                                                                'bg-slate-50 text-slate-700 border-slate-200'
                                                            }`}>
                                                                {priorityStr}
                                                            </span>
                                                        </TableCell>
                                                        <TableCell className="text-right text-xs text-muted-foreground">
                                                            {new Date(ticket.created_at).toLocaleDateString()}
                                                        </TableCell>
                                                    </TableRow>
                                                );
                                            })
                                        )}
                                    </TableBody>
                                </Table>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </AppShell>
    );
}
