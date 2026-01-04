

let accessToken: string | null = null;

export const getAccessToken = () => accessToken;
export const setAccessToken = (token: string | null) => {
    accessToken = token;
};

export const isAuthenticated = () => !!accessToken;

export async function tryRefresh(): Promise<boolean> {
    try {
        // We assume the refresh token is in an httpOnly cookie
        const res = await fetch(`${import.meta.env.VITE_API_URL}/api/v1/refresh`, {
            method: "GET",
            credentials: "include",
        });

        if (!res.ok) throw new Error("Refresh failed");

        const data = await res.json() as { access_token: string };
        setAccessToken(data.access_token);
        return true;
    } catch {
        setAccessToken(null);
        // Ideally we should redirect to login here or let the caller handle it
        if (window.location.pathname !== "/login") {
            window.location.href = "/login";
        }
        return false;
    }
}

export async function logout() {
    try {
        await fetch(`${import.meta.env.VITE_API_URL}/api/v1/logout`, {
            method: "GET",
            credentials: "include",
        });
    } catch {
        // Ignore error
    }
    setAccessToken(null);
    if (window.location.pathname !== "/login") {
        window.location.href = "/login";
    }
}
