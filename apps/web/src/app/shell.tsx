import { Link, useRouterState } from "@tanstack/react-router";
import { Sidebar } from "@/components/layout/sidebar";
import { Topbar } from "@/components/layout/topbar";
import { MaterialSymbol } from "@/components/material-symbol";

export function AppShell({ children }: { children: React.ReactNode }) {
    const pathname = useRouterState({ select: (s) => s.location.pathname });
    const showFab = pathname !== "/tickets/new";

    return (
        <div className="min-h-screen bg-background text-foreground">
            <aside className="fixed left-0 top-0 z-50 hidden h-screen w-64 flex-col border-r border-outline-variant bg-surface-container-lowest lg:flex">
                <Sidebar />
            </aside>
            <Topbar />
            <main className="min-h-screen bg-background pt-16 lg:ml-64">
                <div className="mx-auto max-w-7xl p-4 sm:p-8">{children}</div>
            </main>
            {showFab && (
                <Link
                    to="/tickets/new"
                    className="fixed bottom-8 right-8 z-50 flex h-14 w-14 items-center justify-center rounded-full bg-primary text-on-primary shadow-2xl transition-transform hover:scale-110 active:scale-95 lg:hidden"
                    aria-label="New ticket"
                >
                    <MaterialSymbol name="add" size="lg" filled />
                </Link>
            )}
        </div>
    );
}
