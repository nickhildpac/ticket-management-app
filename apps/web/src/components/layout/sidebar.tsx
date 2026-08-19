import { Link } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { MaterialSymbol } from "@/components/material-symbol";
import { SidebarNavLinks } from "@/components/layout/sidebar-nav";
import { logout } from "@/app/auth";

export function Sidebar() {
    return (
        <div className="flex h-full flex-col py-6">
            <div className="mb-8 flex items-center gap-3 px-6">
                <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-on-primary">
                    <MaterialSymbol name="security" size="sm" />
                </div>
                <div>
                    <h1 className="text-xl font-bold leading-none tracking-tight text-on-surface font-headline">
                        Data Sanctuary
                    </h1>
                    <p className="mt-1 text-[10px] font-medium uppercase tracking-widest text-on-surface-variant">
                        Enterprise Support
                    </p>
                </div>
            </div>

            <div className="mb-6 px-4">
                <Button
                    asChild
                    className="h-auto w-full rounded-xl bg-primary py-2.5 font-headline text-sm font-semibold text-on-primary shadow-primary-soft hover:bg-primary/90 hover:opacity-90 active:scale-[0.98]"
                >
                    <Link to="/tickets/new" className="flex items-center justify-center gap-2">
                        <MaterialSymbol name="add" />
                        New Ticket
                    </Link>
                </Button>
            </div>

            <nav className="no-scrollbar flex-1 space-y-1 overflow-y-auto px-3">
                <SidebarNavLinks />
            </nav>

            <div className="space-y-1 border-t border-outline-variant px-3 pt-6">
                <a
                    href="https://example.com/help"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-on-surface-variant transition-colors hover:bg-surface-container"
                >
                    <MaterialSymbol name="help" />
                    Help Center
                </a>
                <button
                    type="button"
                    onClick={() => void logout()}
                    className="flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left text-sm font-medium text-error transition-colors hover:bg-error/10"
                >
                    <MaterialSymbol name="logout" />
                    Logout
                </button>
            </div>
        </div>
    );
}
