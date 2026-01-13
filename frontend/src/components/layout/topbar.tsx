import { Link } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Search, Plus, Ticket } from "lucide-react";
import { ThemeToggle } from "@/components/theme-toggle";
import { logout } from "@/app/auth";
import { useUser } from "@/app/user-context";

export function Topbar() {
    const { user } = useUser();

    // Get user initials (first letter of first name and last name)
    const getInitials = () => {
        if (!user) return "?";
        const firstInitial = user.first_name?.charAt(0).toUpperCase() || "";
        const lastInitial = user.last_name?.charAt(0).toUpperCase() || "";
        return `${firstInitial}${lastInitial}` || "?";
    };

    return (
        <header className="sticky top-0 z-40 border-b bg-background/80 backdrop-blur">
            <div className="flex h-14 items-center gap-4 px-6">
                <Link to="/" className="flex items-center gap-2 font-semibold text-lg">
                    <Ticket className="h-6 w-6" />
                    <span>Ticket management system</span>
                </Link>
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
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" className="rounded-full h-9 w-9 bg-primary text-primary-foreground hover:bg-primary/90">
                                <span className="text-sm font-semibold">{getInitials()}</span>
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="w-[200px]">
                            <DropdownMenuItem asChild>
                                <Link to="/profile" className="cursor-pointer">
                                    <Ticket className="mr-2 h-4 w-4" />
                                    View profile
                                </Link>
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={logout} className="cursor-pointer">
                                <Ticket className="mr-2 h-4 w-4" />
                                Logout
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                </div>
            </div>
        </header>
    );
}
