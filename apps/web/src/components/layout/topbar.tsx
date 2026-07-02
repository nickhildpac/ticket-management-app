import { Link, useRouterState } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ThemeToggle } from "@/components/theme-toggle";
import { logout } from "@/app/auth";
import { useUser } from "@/app/user-context";
import { MaterialSymbol } from "@/components/material-symbol";
import { MobileNavTrigger } from "@/components/layout/mobile-nav";
import { cn } from "@/lib/utils";

function tabClass(active: boolean) {
    return cn(
        "flex h-16 items-center border-b-2 pb-1 text-sm transition-all",
        active
            ? "border-primary font-semibold text-primary"
            : "border-transparent font-medium text-on-surface-variant hover:text-on-surface"
    );
}

export function Topbar() {
    const { user } = useUser();
    const pathname = useRouterState({ select: (s) => s.location.pathname });
    const isAdmin = user?.role === "admin";
    const isAgent = user?.role === "agent";

    const getInitials = () => {
        if (!user) return "?";
        const firstInitial = user.first_name?.charAt(0).toUpperCase() || "";
        const lastInitial = user.last_name?.charAt(0).toUpperCase() || "";
        return `${firstInitial}${lastInitial}` || "?";
    };

    const ticketsActive =
        pathname === "/tickets" || (pathname.startsWith("/tickets/") && !pathname.startsWith("/tickets/all") && !pathname.startsWith("/tickets/assigned") && !pathname.startsWith("/tickets/new"));
    const assignedActive = pathname.startsWith("/tickets/assigned");

    return (
        <header className="fixed left-0 right-0 top-0 z-40 h-16 border-b border-outline-variant bg-surface-container-lowest/80 shadow-sm backdrop-blur-xl lg:left-64">
            <div className="flex h-full w-full items-center justify-between gap-4 px-4 sm:px-8">
                <div className="flex min-w-0 flex-1 items-center gap-4 lg:gap-8">
                    <MobileNavTrigger />
                    <div className="relative hidden min-w-0 flex-1 group md:block lg:max-w-md">
                        <MaterialSymbol
                            name="search"
                            className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-lg text-on-surface-variant group-focus-within:text-primary"
                        />
                        <Input
                            type="search"
                            placeholder="Search tickets, users, or IDs..."
                            className="h-9 w-full min-w-0 rounded-full border-0 bg-surface-container pl-10 pr-4 text-on-surface shadow-none placeholder:text-on-surface-variant/50 focus-visible:ring-2 focus-visible:ring-primary/20 md:w-80"
                        />
                    </div>
                    <nav className="hidden items-center gap-6 lg:flex">
                        <Link to="/tickets" className={tabClass(ticketsActive)}>
                            All tickets
                        </Link>
                        {(isAgent || isAdmin) && (
                            <Link to="/tickets/assigned" className={tabClass(assignedActive)}>
                                Assigned
                            </Link>
                        )}
                    </nav>
                </div>

                <div className="flex shrink-0 items-center gap-2 sm:gap-4">
                    <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="hidden rounded-full text-on-surface-variant hover:bg-primary/10 hover:text-primary sm:inline-flex"
                        aria-label="Notifications"
                    >
                        <MaterialSymbol name="notifications" />
                    </Button>
                    <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="hidden rounded-full text-on-surface-variant hover:bg-primary/10 hover:text-primary sm:inline-flex"
                        aria-label="History"
                    >
                        <MaterialSymbol name="history" />
                    </Button>
                    <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="hidden rounded-full text-on-surface-variant hover:bg-primary/10 hover:text-primary md:inline-flex"
                        aria-label="Help"
                    >
                        <MaterialSymbol name="help_outline" />
                    </Button>
                    <ThemeToggle />
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button
                                variant="ghost"
                                size="icon"
                                className="ml-1 h-8 w-8 shrink-0 rounded-full border border-outline-variant bg-surface-container p-0 hover:bg-surface-container-high"
                            >
                                <span className="text-xs font-semibold text-primary">{getInitials()}</span>
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="w-[200px]">
                            <DropdownMenuItem asChild>
                                <Link to="/profile" className="cursor-pointer">
                                    <MaterialSymbol name="person" className="mr-2 !text-lg" />
                                    View profile
                                </Link>
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => logout()} className="cursor-pointer">
                                <MaterialSymbol name="logout" className="mr-2 !text-lg" />
                                Logout
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                </div>
            </div>
        </header>
    );
}
