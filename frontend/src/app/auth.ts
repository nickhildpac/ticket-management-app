import { queryClient } from "@/lib/query-client";
import type { UserInfo } from "./user-types";

let accessToken: string | null = null;

export type AuthState = {
    user: UserInfo | null;
};

let authState: AuthState = { user: null };

export function getAuthState(): Readonly<AuthState> {
    return authState;
}

export function getAuthUser(): UserInfo | null {
    return authState.user;
}

/** Keeps route guards and React user context aligned with login, refresh, and /me. */
export function setAuthUser(user: UserInfo | null): void {
    authState = { user };
}

function normalizeUserFromAuthPayload(user: {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    role: string;
}): UserInfo {
    return {
        id: typeof user.id === "string" ? user.id : String(user.id),
        first_name: user.first_name,
        last_name: user.last_name,
        email: user.email,
        role: typeof user.role === "string" ? user.role : String(user.role),
        created_at: "",
    };
}

function clearAuthSession(): void {
    accessToken = null;
    authState = { user: null };
}

export const getAccessToken = () => accessToken;
export const setAccessToken = (token: string | null) => {
    accessToken = token;
};

export const isAuthenticated = () => !!accessToken;

/** Routes where a failed refresh must not hard-redirect (avoids loops; router guards handle protected routes). */
export const AUTH_PUBLIC_PATHS = ["/login", "/signup"] as const;

export function isAuthPublicPath(pathname: string): boolean {
    return (AUTH_PUBLIC_PATHS as readonly string[]).includes(pathname);
}

export async function tryRefresh(): Promise<boolean> {
    try {
        // We assume the refresh token is in an httpOnly cookie
        const res = await fetch(`${import.meta.env.VITE_API_URL}/api/v1/auth/refresh`, {
            method: "GET",
            credentials: "include",
        });

        if (!res.ok) throw new Error("Refresh failed");

        const data = await res.json() as { access_token: string; user: UserInfo };

        setAccessToken(data.access_token);
        setAuthUser(normalizeUserFromAuthPayload(data.user));
        // Drop stale `me` so UI uses JWT user until /me refetches (avoids wrong profile after account switch).
        queryClient.removeQueries({ queryKey: ["me"] });

        return true;
    } catch {
        clearAuthSession();
        queryClient.removeQueries({ queryKey: ["me"] });
        const path = window.location.pathname;
        if (!isAuthPublicPath(path)) {
            window.location.href = "/login";
        }
        return false;
    }
}

export async function login(email: string, password: string): Promise<{ success: boolean; user?: UserInfo }> {
    try {
        const res = await fetch(`${import.meta.env.VITE_API_URL}/api/v1/auth/login`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            credentials: "include",
            body: JSON.stringify({ email, password }),
        });

        if (!res.ok) {
            return { success: false };
        }

        const data = await res.json() as { access_token: string; user: UserInfo };

        setAccessToken(data.access_token);
        const user = normalizeUserFromAuthPayload(data.user);
        setAuthUser(user);

        return { success: true, user };
    } catch {
        return { success: false };
    }
}

export async function logout() {
    try {
        await fetch(`${import.meta.env.VITE_API_URL}/api/v1/auth/logout`, {
            method: "GET",
            credentials: "include",
        });
    } catch {
        // Ignore error
    }
    clearAuthSession();
    queryClient.removeQueries({ queryKey: ["me"] });
    const path = window.location.pathname;
    if (!isAuthPublicPath(path)) {
        window.location.href = "/login";
    }
}
