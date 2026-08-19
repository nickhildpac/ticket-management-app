import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardContent, CardTitle, CardFooter } from "@/components/ui/card";
import { Link } from "@tanstack/react-router";
import { register } from "@/app/auth";

/**
 * Registration is handled by Keycloak's own form; this route just forwards to
 * it. The app no longer accepts or stores passwords.
 */
export function Signup() {
    const [error, setError] = useState<string | null>(null);
    const [redirecting, setRedirecting] = useState(true);

    const start = () => {
        setRedirecting(true);
        setError(null);
        register().catch((err: unknown) => {
            setRedirecting(false);
            setError(err instanceof Error ? err.message : "Could not reach the sign-up service");
        });
    };

    useEffect(() => {
        const timer = setTimeout(start, 0);
        return () => clearTimeout(timer);
    }, []);

    return (
        <div className="flex min-h-screen items-center justify-center bg-background px-4 py-8 font-body">
            <Card className="w-full max-w-sm border-outline-variant/40 bg-surface-container-low">
                <CardHeader>
                    <CardTitle className="font-headline text-xl text-on-surface">Create an account</CardTitle>
                </CardHeader>
                <CardContent className="grid gap-4">
                    {error && (
                        <div className="rounded-lg border border-error/30 bg-error/10 p-2 text-center text-sm font-medium text-error">
                            {error}
                        </div>
                    )}
                    <p className="text-sm text-muted-foreground">
                        {redirecting ? "Taking you to the sign-up page…" : "You'll be redirected to sign up securely."}
                    </p>
                    <Button className="w-full" disabled={redirecting} onClick={start}>
                        {redirecting ? "Redirecting…" : "Continue to sign up"}
                    </Button>
                </CardContent>
                <CardFooter className="flex justify-center text-sm text-muted-foreground">
                    <span>
                        Already have an account?{" "}
                        <Link to="/login" className="text-primary hover:underline">
                            Sign in
                        </Link>
                    </span>
                </CardFooter>
            </Card>
        </div>
    );
}
