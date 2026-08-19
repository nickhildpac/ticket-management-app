import { useState, useEffect } from "react";
import { isAuthPublicPath, login, restoreSession } from "./auth";

/**
 * Establishes the session before the router renders, so route guards never see
 * a half-initialised auth state.
 *
 * Tokens live in memory, so a page reload starts with no session. On a
 * protected path that means redirecting to Keycloak — invisible to the user
 * when their SSO session is still valid, since Keycloak redirects straight
 * back. Public paths (including the OAuth callback) are left alone so this
 * can't fight with the code exchange or loop on the login page.
 */
export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [initializing, setInitializing] = useState(true);

    useEffect(() => {
        let cancelled = false;

        async function initAuth() {
            try {
                const restored = await restoreSession();
                if (cancelled) return;

                if (!restored && !isAuthPublicPath(window.location.pathname)) {
                    // Navigates away; don't clear the spinner first or the app
                    // flashes an unauthenticated frame.
                    await login();
                    return;
                }
            } catch (err) {
                console.error("Failed to restore session on init", err);
            }
            if (!cancelled) setInitializing(false);
        }

        void initAuth();
        return () => {
            cancelled = true;
        };
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
