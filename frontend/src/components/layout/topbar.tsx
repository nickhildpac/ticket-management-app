import { Link } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Search, Plus } from "lucide-react";
import { ThemeToggle } from "@/components/theme-toggle";
import { logout } from "@/app/auth";

export function Topbar() {
    return (
        <header className="sticky top-0 z-40 border-b bg-background/80 backdrop-blur">
            <div className="flex h-14 items-center gap-4 px-6">
                <Link to="/" className="font-semibold text-lg">Ticketing</Link>
                <div className="ml-auto flex items-center gap-4">
                    <div className="relative hidden md:block">
                        <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                        <Input
                            type="search"
                            placeholder="Search..."
                            className="w-64 pl-8"
                        />
                    </div>
                    <Button asChild size="sm">
                        <Link to="/tickets/new">
                            <Plus className="mr-2 h-4 w-4" /> New Ticket
                        </Link>
                    </Button>
                    <ThemeToggle />
                    <Button variant="ghost" size="sm" onClick={logout}>Logout</Button>
                </div>
            </div>
        </header>
    );
}
