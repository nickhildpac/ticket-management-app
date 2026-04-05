import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardHeader, CardContent, CardTitle, CardFooter } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { useState } from "react";
import { useNavigate, Link } from "@tanstack/react-router";
import { login } from "@/app/auth";
import { useUser } from "@/app/user-context";

export function Login() {
    const navigate = useNavigate();
    const { setUser } = useUser();
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);

    const handleLogin = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError("");
        const result = await login(email, password);
        if (result.success && result.user) {
            setUser(result.user);
            navigate({ to: '/' });
        } else {
            setError("Invalid credentials");
        }
        setLoading(false);
    };

    return (
        <div className="flex min-h-screen items-center justify-center bg-background px-4 py-8 font-body">
            <Card className="w-full max-w-sm border-outline-variant/40 bg-surface-container-low">
                <CardHeader>
                    <CardTitle className="font-headline text-xl text-on-surface">Login</CardTitle>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleLogin} className="grid gap-4">
                        {error && (
                            <div className="rounded-lg border border-error/30 bg-error/10 p-2 text-center text-sm font-medium text-error">
                                {error}
                            </div>
                        )}
                        <div className="grid gap-2">
                            <Label htmlFor="email">Email</Label>
                            <Input
                                id="email"
                                type="email"
                                placeholder="m@example.com"
                                required
                                value={email}
                                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setEmail(e.target.value)}
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="password">Password</Label>
                            <Input
                                id="password"
                                type="password"
                                required
                                value={password}
                                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setPassword(e.target.value)}
                            />
                        </div>
                        <Button type="submit" className="w-full" disabled={loading}>
                            {loading ? "Signing in..." : "Sign in"}
                        </Button>
                    </form>
                </CardContent>
                <CardFooter className="flex flex-col items-center gap-2 text-sm text-muted-foreground">
                    <span>Demo: alice@admin.com / password123</span>
                    <span>Don't have an account? <Link to="/signup" className="text-primary hover:underline">Sign up</Link></span>
                </CardFooter>
            </Card>
        </div>
    );
}
