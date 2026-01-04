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
import { useState } from "react";
import { Skeleton } from "@/components/ui/skeleton";

export function TicketList() {
    const [page, setPage] = useState(1);
    const [state, setState] = useState<string>("");
    const [q, setQ] = useState("");

    const { data, isLoading, isError } = useTickets({ page, state, q });

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
                <Select value={state} onValueChange={setState}>
                    <SelectTrigger className="w-[180px]">
                        <SelectValue placeholder="All Statuses" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="all">All Statuses</SelectItem>
                        <SelectItem value="open">Open</SelectItem>
                        <SelectItem value="pending">Pending</SelectItem>
                        <SelectItem value="resolved">Resolved</SelectItem>
                        <SelectItem value="closed">Closed</SelectItem>
                    </SelectContent>
                </Select>
            </div>

            <div className="rounded-md border bg-card">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead className="w-[100px]">ID</TableHead>
                            <TableHead>Title</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead>Priority</TableHead>
                            <TableHead className="text-right">Created</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {isLoading ? (
                            Array.from({ length: 5 }).map((_, i) => (
                                <TableRow key={i}>
                                    <TableCell><Skeleton className="h-4 w-8" /></TableCell>
                                    <TableCell><Skeleton className="h-4 w-40" /></TableCell>
                                    <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                                    <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                                    <TableCell><Skeleton className="h-4 w-24 ml-auto" /></TableCell>
                                </TableRow>
                            ))
                        ) : isError ? (
                            <TableRow>
                                <TableCell colSpan={5} className="text-center text-red-500 py-4">Failed to load tickets</TableCell>
                            </TableRow>
                        ) : data?.items?.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={5} className="text-center py-4">No tickets found</TableCell>
                            </TableRow>
                        ) : (
                            data?.items?.map((ticket) => (
                                <TableRow key={ticket.id}>
                                    <TableCell className="font-medium">
                                        <Link to="/tickets/$id" params={{ id: ticket.id }} className="hover:underline">
                                            #{ticket.id}
                                        </Link>
                                    </TableCell>
                                    <TableCell>
                                        <Link to="/tickets/$id" params={{ id: ticket.id }} className="hover:underline block font-medium">
                                            {ticket.title}
                                        </Link>
                                    </TableCell>
                                    <TableCell>
                                        <Badge variant={ticket.state === 'open' ? 'default' : 'secondary'}>{ticket.state}</Badge>
                                    </TableCell>
                                    <TableCell>
                                        <Badge variant="outline" className="capitalize">{ticket.priority}</Badge>
                                    </TableCell>
                                    <TableCell className="text-right text-muted-foreground">
                                        {new Date(ticket.created_at).toLocaleDateString()}
                                    </TableCell>
                                </TableRow>
                            ))
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
