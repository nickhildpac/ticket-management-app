import { getAccessToken, tryRefresh } from "@/app/auth";

type ApiErrorEnvelope = {
    code?: string;
    message?: string;
    details?: unknown;
    error?: string;
};

async function errorMessageFromResponse(res: Response): Promise<string> {
    const fallback = `API Error ${res.status}`;
    const text = await res.text();
    if (!text) return fallback;

    try {
        const body = JSON.parse(text) as ApiErrorEnvelope;
        return body.message || body.error || fallback;
    } catch {
        return text;
    }
}

export async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
    const token = getAccessToken();

    const headers = new Headers(init.headers);
    if (token) {
        headers.set("Authorization", `Bearer ${token}`);
    }
    // For FormData the browser must set `Content-Type` itself (with the multipart
    // boundary); forcing JSON here would corrupt the request. Only default to JSON.
    if (!headers.has("Content-Type") && !(init.body instanceof FormData)) {
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
        throw new Error(await errorMessageFromResponse(res));
    }

    // Handle 204 No Content
    if (res.status === 204) {
        return {} as T;
    }

    return res.json() as Promise<T>;
}
