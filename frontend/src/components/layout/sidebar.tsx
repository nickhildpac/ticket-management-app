import { Link } from "@tanstack/react-router";
import { LayoutDashboard, Ticket, Settings } from "lucide-react";

export function Sidebar() {
    return (
        <div className="flex h-full flex-col gap-4 py-4">
            <div className="px-4 py-2">
                <h2 className="text-lg font-semibold tracking-tight">Navigation</h2>
                <div className="mt-2 space-y-1">
                    <Link
                        to="/"
                        className="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium hover:bg-accent hover:text-accent-foreground"
                        activeProps={{ className: "bg-accent text-accent-foreground" }}
                    >
                        <LayoutDashboard className="h-4 w-4" />
                        Dashboard
                    </Link>
                    <Link
                        to="/tickets"
                        className="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium hover:bg-accent hover:text-accent-foreground"
                        activeProps={{ className: "bg-accent text-accent-foreground" }}
                    >
                        <Ticket className="h-4 w-4" />
                        Tickets
                    </Link>
                    <Link
                        to="/admin"
                        className="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium hover:bg-accent hover:text-accent-foreground"
                        activeProps={{ className: "bg-accent text-accent-foreground" }}
                    >
                        <Settings className="h-4 w-4" />
                        Admin
                    </Link>
                </div>
            </div>
        </div>
    );
}
