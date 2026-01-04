import { AppShell } from "@/app/shell";
import { useUsers, useUpdateUserRole, useDeleteUser } from "./queries";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Trash2, UserCog, ShieldCheck, Mail, Calendar } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import type { Role } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";

export function AdminPanel() {
    const { data: users, isLoading } = useUsers();
    const updateUserRole = useUpdateUserRole();
    const deleteUser = useDeleteUser();

    const handleRoleChange = (userId: string, newRole: Role) => {
        updateUserRole.mutate({ id: userId, role: newRole });
    };

    const handleDeleteUser = (userId: string) => {
        if (confirm("Are you sure you want to delete this user? This action cannot be undone.")) {
            deleteUser.mutate(userId);
        }
    };

    return (
        <AppShell>
            <div className="space-y-8 pb-10">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Admin Console</h1>
                    <p className="text-muted-foreground text-lg">Manage system users, permissions, and entity settings.</p>
                </div>

                <Card className="border-none shadow-sm">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <UserCog className="h-5 w-5 text-primary" />
                            User Management
                        </CardTitle>
                        <CardDescription>
                            Configure user roles and access levels across the platform.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="rounded-lg border bg-white dark:bg-slate-950 overflow-hidden">
                            <Table>
                                <TableHeader className="bg-muted/40">
                                    <TableRow>
                                        <TableHead className="w-1/4">User</TableHead>
                                        <TableHead className="w-1/4">Contact</TableHead>
                                        <TableHead className="w-1/6">Role</TableHead>
                                        <TableHead className="w-1/6">Joined</TableHead>
                                        <TableHead className="w-16 text-right"></TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {isLoading ? (
                                        Array.from({ length: 5 }).map((_, i) => (
                                            <TableRow key={i}>
                                                <TableCell><Skeleton className="h-5 w-32" /></TableCell>
                                                <TableCell><Skeleton className="h-5 w-48" /></TableCell>
                                                <TableCell><Skeleton className="h-9 w-28" /></TableCell>
                                                <TableCell><Skeleton className="h-5 w-24" /></TableCell>
                                                <TableCell className="text-right"><Skeleton className="h-9 w-9 ml-auto rounded-full" /></TableCell>
                                            </TableRow>
                                        ))
                                    ) : (
                                        users?.map((user) => (
                                            <TableRow key={user.id} className="group hover:bg-muted/20 transition-colors">
                                                <TableCell className="py-4">
                                                    <div className="flex items-center gap-3">
                                                        <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center text-primary font-bold">
                                                            {user.first_name[0]}{user.last_name[0]}
                                                        </div>
                                                        <div>
                                                            <div className="font-semibold">{user.first_name} {user.last_name}</div>
                                                            <div className="text-xs text-muted-foreground font-mono">#{user.id.slice(0, 8)}</div>
                                                        </div>
                                                    </div>
                                                </TableCell>
                                                <TableCell>
                                                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                                        <Mail className="h-3 w-3" />
                                                        {user.email}
                                                    </div>
                                                </TableCell>
                                                <TableCell>
                                                    <Select
                                                        defaultValue={user.role}
                                                        onValueChange={(val) => handleRoleChange(user.id, val as Role)}
                                                        disabled={updateUserRole.isPending}
                                                    >
                                                        <SelectTrigger className="w-[130px] h-9 shadow-none border-dashed hover:border-solid transition-all">
                                                            <SelectValue />
                                                        </SelectTrigger>
                                                        <SelectContent>
                                                            <SelectItem value="user">
                                                                <div className="flex items-center gap-2">
                                                                    <div className="h-2 w-2 rounded-full bg-slate-400" />
                                                                    <span>Standard</span>
                                                                </div>
                                                            </SelectItem>
                                                            <SelectItem value="agent">
                                                                <div className="flex items-center gap-2">
                                                                    <div className="h-2 w-2 rounded-full bg-blue-500" />
                                                                    <span>Agent</span>
                                                                </div>
                                                            </SelectItem>
                                                            <SelectItem value="admin">
                                                                <div className="flex items-center gap-2">
                                                                    <ShieldCheck className="h-3 w-3 text-red-500" />
                                                                    <span>Administrator</span>
                                                                </div>
                                                            </SelectItem>
                                                        </SelectContent>
                                                    </Select>
                                                </TableCell>
                                                <TableCell>
                                                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                                        <Calendar className="h-3 w-3" />
                                                        {user.created_at ? new Date(user.created_at).toLocaleDateString() : '—'}
                                                    </div>
                                                </TableCell>
                                                <TableCell className="text-right">
                                                    <Button
                                                        variant="ghost"
                                                        size="icon"
                                                        onClick={() => handleDeleteUser(user.id)}
                                                        className="opacity-0 group-hover:opacity-100 transition-opacity text-red-500 hover:text-red-700 hover:bg-red-50 rounded-full"
                                                        disabled={deleteUser.isPending}
                                                    >
                                                        <Trash2 className="h-4 w-4" />
                                                    </Button>
                                                </TableCell>
                                            </TableRow>
                                        ))
                                    )}
                                </TableBody>
                            </Table>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </AppShell>
    );
}
