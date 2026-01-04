import { Sidebar } from "@/components/layout/sidebar";
import { Topbar } from "@/components/layout/topbar";

export function AppShell({ children }: { children: React.ReactNode }) {
    return (
        <div className="grid min-h-screen grid-rows-[auto,1fr] bg-background text-foreground">
            <Topbar />
            <div className="grid grid-cols-12 h-full">
                <aside className="col-span-2 hidden border-r lg:block">
                    <Sidebar />
                </aside>
                <main className="col-span-12 lg:col-span-10 p-6">
                    {children}
                </main>
            </div>
        </div>
    );
}
