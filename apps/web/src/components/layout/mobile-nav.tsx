import { useState } from "react";
import { Link } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { MaterialSymbol } from "@/components/material-symbol";
import { SidebarNavLinks } from "@/components/layout/sidebar-nav";
import { useUser } from "@/app/user-context";

function getUserInitials(user: { first_name?: string; last_name?: string } | null | undefined) {
    if (!user) return "?";
    const firstInitial = user.first_name?.charAt(0).toUpperCase() || "";
    const lastInitial = user.last_name?.charAt(0).toUpperCase() || "";
    return `${firstInitial}${lastInitial}` || "?";
}

export function MobileNavTrigger() {
    const [open, setOpen] = useState(false);
    const { user } = useUser();

    return (
        <>
            <Button
                type="button"
                variant="ghost"
                size="icon"
                className="shrink-0 text-on-surface-variant hover:bg-primary/10 hover:text-primary lg:hidden"
                aria-label="Open menu"
                onClick={() => setOpen(true)}
            >
                <MaterialSymbol name="menu" className="!text-2xl" />
            </Button>
            <Dialog open={open} onOpenChange={setOpen}>
                <DialogContent
                    className="fixed left-0 top-0 z-50 flex h-full max-h-none w-64 max-w-[min(100vw,16rem)] translate-x-0 translate-y-0 flex-col gap-0 rounded-none border-y-0 border-l-0 border-r border-outline-variant bg-surface-container-lowest p-0 shadow-xl !duration-200 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:slide-out-to-left-full data-[state=open]:slide-in-from-left-full sm:rounded-none [&>button]:hidden"
                >
                    <DialogHeader className="sr-only">
                        <DialogTitle>Navigation</DialogTitle>
                    </DialogHeader>
                    <div className="flex flex-1 flex-col overflow-y-auto py-6">
                        <div className="mb-8 flex items-center gap-3 px-6">
                            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-on-primary">
                                <MaterialSymbol name="security" size="sm" />
                            </div>
                            <div>
                                <p className="text-xl font-bold leading-none tracking-tight text-on-surface font-headline">
                                    Data Sanctuary
                                </p>
                                <p className="mt-1 text-[10px] font-medium uppercase tracking-widest text-on-surface-variant">
                                    Enterprise Support
                                </p>
                            </div>
                        </div>
                        <div className="mb-6 px-4">
                            <Button
                                asChild
                                className="h-auto w-full rounded-xl bg-primary py-2.5 font-headline text-sm font-semibold text-on-primary shadow-primary-soft hover:bg-primary/90"
                            >
                                <Link
                                    to="/tickets/new"
                                    onClick={() => setOpen(false)}
                                    className="flex items-center justify-center gap-2"
                                >
                                    <MaterialSymbol name="add" />
                                    New Ticket
                                </Link>
                            </Button>
                        </div>
                        <nav className="no-scrollbar flex-1 space-y-1 overflow-y-auto px-3">
                            <SidebarNavLinks onNavigate={() => setOpen(false)} />
                        </nav>
                        <div className="border-t border-outline-variant px-3 pt-4">
                            <Link
                                to="/profile"
                                onClick={() => setOpen(false)}
                                className="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-on-surface-variant transition-colors hover:bg-surface-container"
                            >
                                <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-outline-variant bg-surface-container text-xs font-semibold text-primary">
                                    {getUserInitials(user)}
                                </span>
                                Profile
                            </Link>
                        </div>
                    </div>
                </DialogContent>
            </Dialog>
        </>
    );
}
