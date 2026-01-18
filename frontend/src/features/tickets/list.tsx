import { AppShell } from "@/app/shell";
import { Link } from "@tanstack/react-router";
import { useTickets } from "./queries";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useState, useEffect } from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { getTicketStateString, getTicketPriorityString } from "@/lib/types";

export function TicketList() {
    const [page, setPage] = useState(1);
    const [state, setState] = useState<string>("all");
    const [q, setQ] = useState("");
    const [searchQuery, setSearchQuery] = useState("");

    // Debounce search input
    useEffect(() => {
        const timer = setTimeout(() => {
            setSearchQuery(q);
            setPage(1); // Reset page when search changes
        }, 500);
        return () => clearTimeout(timer);
    }, [q]);

    // Reset page when state changes
    const handleStateChange = (newState: string) => {
        setState(newState);
        setPage(1);
    };

    const { data, isLoading, isError } = useTickets({ page, state, q: searchQuery });

    return (
        <AppShell>
            <div className="flex items-center justify-between mb-6">
                <h1 className="text-2xl font-bold tracking-tight">Tickets</h1>
                <Button asChild>
                    <Link to="/tickets/new">New Ticket</Link>
                </Button>
            </div>

            <div className="flex items-center gap-4 mb-6">
                <Input
                    placeholder="Search tickets..."
                    className="max-w-xs"
                    value={q}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setQ(e.target.value)}
                />
                <Select value={state} onValueChange={handleStateChange}>
                    <SelectTrigger className="w-[180px]">
                        <SelectValue placeholder="All Statuses" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="all">All Statuses</SelectItem>
                        <SelectItem value="open">Open</SelectItem>
                        <SelectItem value="pending">Pending</SelectItem>
                        <SelectItem value="resolved">Resolved</SelectItem>
                        <SelectItem value="closed">Closed</SelectItem>
                        <SelectItem value="cancelled">Cancelled</SelectItem>
                    </SelectContent>
                </Select>
            </div>

            <div className="rounded-md border bg-card">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead className="w-[120px]">Ticket No.</TableHead>
                            <TableHead>Short Description</TableHead>
                            <TableHead>Description</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead>Priority</TableHead>
                            <TableHead className="text-right">Created</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {isLoading ? (
                            Array.from({ length: 5 }).map((_, i) => (
                                <TableRow key={i}>
                                    <TableCell><Skeleton className="h-4 w-16" /></TableCell>
                                    <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                                    <TableCell><Skeleton className="h-4 w-48" /></TableCell>
                                    <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                                    <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                                    <TableCell><Skeleton className="h-4 w-24 ml-auto" /></TableCell>
                                </TableRow>
                            ))
                        ) : isError ? (
                            <TableRow>
                                <TableCell colSpan={6} className="text-center text-red-500 py-4">Failed to load tickets</TableCell>
                            </TableRow>
                        ) : data?.items?.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={6} className="text-center py-4">No tickets found</TableCell>
                            </TableRow>
                        ) : (
                            data?.items?.map((ticket) => {
                                const stateStr = typeof ticket.state === 'number' ? getTicketStateString(ticket.state) : ticket.state;
                                const priorityStr = typeof ticket.priority === 'number' ? getTicketPriorityString(ticket.priority) : ticket.priority;

                                return (
                                    <TableRow key={ticket.id}>
                                        <TableCell className="font-mono text-xs">
                                            <Link to="/tickets/$id" params={{ id: ticket.id }} className="hover:underline font-bold text-primary">
                                                #{ticket.ticket_number}
                                            </Link>
                                        </TableCell>
                                        <TableCell>
                                            {ticket.title}
                                        </TableCell>
                                        <TableCell className="text-muted-foreground">
                                            {ticket.description?.substring(0, 100)}
                                            {ticket.description && ticket.description.length > 100 ? '...' : ''}
                                        </TableCell>
                                        <TableCell>
                                            <Badge variant={stateStr === 'open' ? 'default' : 'secondary'} className="capitalize">
                                                {stateStr}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>
                                            <Badge
                                                variant="outline"
                                                className={`capitalize ${priorityStr === 'critical' ? 'bg-red-50 text-red-700 border-red-200' :
                                                    priorityStr === 'high' ? 'bg-orange-50 text-orange-700 border-orange-200' :
                                                        priorityStr === 'medium' ? 'bg-yellow-50 text-yellow-700 border-yellow-200' :
                                                            'bg-slate-50 text-slate-700 border-slate-200'
                                                    }`}
                                            >
                                                {priorityStr}
                                            </Badge>
                                        </TableCell>
                                        <TableCell className="text-right text-muted-foreground">
                                            {new Date(ticket.created_at).toLocaleString()}
                                        </TableCell>
                                    </TableRow>
                                );
                            })
                        )}
                    </TableBody>
                </Table>
            </div>

            <div className="flex items-center justify-end space-x-2 py-4">
                <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                    disabled={page === 1 || isLoading}
                >
                    Previous
                </Button>
                <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage((p) => p + 1)}
                    disabled={!data || !data.items || data.items.length < (data.pageSize || 10) || isLoading}
                >
                    Next
                </Button>
            </div>
        </AppShell>
    );
}
