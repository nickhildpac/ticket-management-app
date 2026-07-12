import { Link, useRouterState } from "@tanstack/react-router";
import { MaterialSymbol } from "@/components/material-symbol";
import { useUser } from "@/app/user-context";
import { cn } from "@/lib/utils";

const navLinkClass =
    "flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-on-surface-variant transition-colors hover:bg-surface-container";
const activeNavClass =
    "border-r-2 border-primary bg-primary/10 font-bold text-primary";

/** Primary ticket queue at /tickets; not active on /tickets/all, /tickets/assigned, or /tickets/new. */
function isTicketsQueueNavActive(pathname: string) {
    return (
        pathname === "/tickets" ||
        (pathname.startsWith("/tickets/") &&
            !pathname.startsWith("/tickets/all") &&
            !pathname.startsWith("/tickets/assigned") &&
            !pathname.startsWith("/tickets/new"))
    );
}

type SidebarNavProps = {
    onNavigate?: () => void;
};

export function SidebarNavLinks({ onNavigate }: SidebarNavProps) {
    const { user } = useUser();
    const pathname = useRouterState({ select: (s) => s.location.pathname });
    const isAdmin = user?.role === "admin";
    const isAgent = user?.role === "agent";
    const canSeeDashboard = isAdmin || isAgent;

    const dashboardActive = pathname === "/";
    const ticketsQueueActive = isTicketsQueueNavActive(pathname);
    const assignedActive = pathname.startsWith("/tickets/assigned");
    const adminActive = pathname === "/admin";
    const documentsActive = pathname.startsWith("/admin/documents");

    return (
        <>
            {canSeeDashboard && (
                <Link
                    to="/"
                    onClick={onNavigate}
                    className={cn(navLinkClass, dashboardActive && activeNavClass)}
                >
                    <MaterialSymbol name="dashboard" />
                    Dashboard
                </Link>
            )}
            <Link
                to="/tickets"
                onClick={onNavigate}
                className={cn(navLinkClass, ticketsQueueActive && activeNavClass)}
            >
                <MaterialSymbol name="confirmation_number" />
                All tickets
            </Link>
            {isAgent && (
                <Link
                    to="/tickets/assigned"
                    onClick={onNavigate}
                    className={cn(navLinkClass, assignedActive && activeNavClass)}
                >
                    <MaterialSymbol name="queue_play_next" />
                    My Assigned Tickets
                </Link>
            )}
            {isAdmin && (
                <Link
                    to="/admin"
                    onClick={onNavigate}
                    className={cn(navLinkClass, adminActive && activeNavClass)}
                >
                    <MaterialSymbol name="settings" />
                    Admin
                </Link>
            )}
            {isAdmin && (
                <Link
                    to="/admin/documents"
                    onClick={onNavigate}
                    className={cn(navLinkClass, documentsActive && activeNavClass)}
                >
                    <MaterialSymbol name="upload_file" />
                    Documents
                </Link>
            )}
        </>
    );
}
