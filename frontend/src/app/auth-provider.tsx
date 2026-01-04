import { useState, useEffect } from "react";
import { tryRefresh } from "./auth";

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [initializing, setInitializing] = useState(true);

    useEffect(() => {
        async function initAuth() {
            try {
                await tryRefresh();
            } catch (err) {
                console.error("Failed to refresh token on init", err);
            } finally {
                setInitializing(false);
            }
        }
        initAuth();
    }, []);

    if (initializing) {
        return (
            <div className="flex h-screen w-screen items-center justify-center">
                <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
            </div>
        );
    }

    return <>{children}</>;
}
