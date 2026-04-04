import { Link } from "@tanstack/react-router";
import { LayoutDashboard, Ticket, Layers, UserCheck, Settings } from "lucide-react";
import { useUser } from "@/app/user-context";

export function Sidebar() {
    const { user } = useUser();
    const isAdmin = user?.role === "admin";
    const isAgent = user?.role === "agent";
    const canSeeDashboard = isAdmin || isAgent;

    return (
        <div className="flex h-full flex-col gap-4 py-4">
            <div className="px-4 py-2">
                <h2 className="text-lg font-semibold tracking-tight">Navigation</h2>
                <div className="mt-2 space-y-1">
                    {canSeeDashboard && (
                        <Link
                            to="/"
                            className="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium hover:bg-accent hover:text-accent-foreground"
                            activeProps={{ className: "bg-accent text-accent-foreground" }}
                        >
                            <LayoutDashboard className="h-4 w-4" />
                            Dashboard
                        </Link>
                    )}
                    <Link
                        to="/tickets"
                        className="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium hover:bg-accent hover:text-accent-foreground"
                        activeProps={{ className: "bg-accent text-accent-foreground" }}
                    >
                        <Ticket className="h-4 w-4" />
                        My Tickets
                    </Link>
                    {isAdmin && (
                        <Link
                            to="/tickets/all"
                            className="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium hover:bg-accent hover:text-accent-foreground"
                            activeProps={{ className: "bg-accent text-accent-foreground" }}
                        >
                            <Layers className="h-4 w-4" />
                            All Tickets
                        </Link>
                    )}
                    {isAgent && (
                        <Link
                            to="/tickets/assigned"
                            className="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium hover:bg-accent hover:text-accent-foreground"
                            activeProps={{ className: "bg-accent text-accent-foreground" }}
                        >
                            <UserCheck className="h-4 w-4" />
                            My Assigned Tickets
                        </Link>
                    )}
                    {isAdmin && (
                        <Link
                            to="/admin"
                            className="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium hover:bg-accent hover:text-accent-foreground"
                            activeProps={{ className: "bg-accent text-accent-foreground" }}
                        >
                            <Settings className="h-4 w-4" />
                            Admin
                        </Link>
                    )}
                </div>
            </div>
        </div>
    );
}
