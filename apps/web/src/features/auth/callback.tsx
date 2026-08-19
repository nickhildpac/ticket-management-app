import { useEffect, useRef, useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { handleAuthCallback, login } from "@/app/auth";

/**
 * Landing route for Keycloak's redirect. Exchanges the authorization code for
 * tokens, then continues to wherever the user was originally headed.
 */
export function AuthCallback() {
    const navigate = useNavigate();
    const [error, setError] = useState<string | null>(null);
    // React 18 StrictMode double-invokes effects in development; the code is
    // single-use, so a second exchange would fail against a valid session.
    const exchanged = useRef(false);

    useEffect(() => {
        if (exchanged.current) return;
        exchanged.current = true;

        handleAuthCallback(new URLSearchParams(window.location.search))
            .then(({ returnTo }) => {
                // replace: the callback URL carries a spent code, so it must not
                // stay in history where Back would re-trigger it.
                navigate({ to: returnTo, replace: true });
            })
            .catch((err: unknown) => {
                setError(err instanceof Error ? err.message : "Sign-in failed");
            });
    }, [navigate]);

    if (error) {
        return (
            <div className="flex min-h-screen items-center justify-center bg-background px-4 py-8 font-body">
                <Card className="w-full max-w-sm border-outline-variant/40 bg-surface-container-low">
                    <CardHeader>
                        <CardTitle className="font-headline text-xl text-on-surface">Sign-in failed</CardTitle>
                    </CardHeader>
                    <CardContent className="grid gap-4">
                        <p className="text-sm text-error">{error}</p>
                        <Button className="w-full" onClick={() => void login({ returnTo: "/" })}>
                            Try again
                        </Button>
                    </CardContent>
                </Card>
            </div>
        );
    }

    return (
        <div className="flex h-screen w-screen flex-col items-center justify-center gap-4">
            <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
            <p className="text-sm text-muted-foreground">Completing sign-in…</p>
        </div>
    );
}
