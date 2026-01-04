import { getAccessToken, tryRefresh } from "@/app/auth";

export async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
    const token = getAccessToken();

    const headers = new Headers(init.headers);
    if (token) {
        headers.set("Authorization", `Bearer ${token}`);
    }
    if (!headers.has("Content-Type")) {
        headers.set("Content-Type", "application/json");
    }

    const config = {
        ...init,
        headers,
        credentials: "include" as RequestCredentials, // For httpOnly cookies
    };

    let res = await fetch(`${import.meta.env.VITE_API_URL}${path}`, config);

    if (res.status === 401) {
        const refreshed = await tryRefresh();
        if (refreshed) {
            // Retry with new token
            const newToken = getAccessToken();
            headers.set("Authorization", `Bearer ${newToken}`);
            res = await fetch(`${import.meta.env.VITE_API_URL}${path}`, {
                ...config,
                headers,
            });
        }
    }

    if (!res.ok) {
        const errorText = await res.text();
        throw new Error(errorText || `API Error ${res.status}`);
    }

    // Handle 204 No Content
    if (res.status === 204) {
        return {} as T;
    }

    return res.json() as Promise<T>;
}
