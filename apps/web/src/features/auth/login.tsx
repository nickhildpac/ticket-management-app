import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardContent, CardTitle, CardFooter } from "@/components/ui/card";
import { login, register } from "@/app/auth";

/**
 * Sign-in entry point. There is no credential form: the app never handles
 * passwords, it hands off to Keycloak's hosted login page (Authorization Code +
 * PKCE) and gets back a code.
 */
export function Login() {
    const [error, setError] = useState<string | null>(null);
    const [redirecting, setRedirecting] = useState(false);

    const start = (register_: boolean) => {
        setRedirecting(true);
        setError(null);
        const target = register_ ? register() : login({ returnTo: "/" });
        target.catch((err: unknown) => {
            setRedirecting(false);
            setError(err instanceof Error ? err.message : "Could not reach the sign-in service");
        });
    };

    // A user landing on /login with no session is here to sign in, so go
    // straight there rather than making them press a button. Keycloak returns
    // them immediately if their SSO session is still valid.
    useEffect(() => {
        const timer = setTimeout(() => start(false), 0);
        return () => clearTimeout(timer);
    }, []);

    return (
        <div className="flex min-h-screen items-center justify-center bg-background px-4 py-8 font-body">
            <Card className="w-full max-w-sm border-outline-variant/40 bg-surface-container-low">
                <CardHeader>
                    <CardTitle className="font-headline text-xl text-on-surface">Sign in</CardTitle>
                </CardHeader>
                <CardContent className="grid gap-4">
                    {error && (
                        <div className="rounded-lg border border-error/30 bg-error/10 p-2 text-center text-sm font-medium text-error">
                            {error}
                        </div>
                    )}
                    <p className="text-sm text-muted-foreground">
                        {redirecting
                            ? "Taking you to the sign-in page…"
                            : "You'll be redirected to sign in securely."}
                    </p>
                    <Button className="w-full" disabled={redirecting} onClick={() => start(false)}>
                        {redirecting ? "Redirecting…" : "Continue to sign in"}
                    </Button>
                </CardContent>
                <CardFooter className="flex flex-col items-center gap-2 text-sm text-muted-foreground">
                    <span>Demo: alice@admin.com / password123</span>
                    <span>
                        Don't have an account?{" "}
                        <button
                            type="button"
                            className="text-primary hover:underline"
                            onClick={() => start(true)}
                        >
                            Sign up
                        </button>
                    </span>
                </CardFooter>
            </Card>
        </div>
    );
}
