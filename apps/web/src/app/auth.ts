/**
 * Keycloak authentication for the SPA: Authorization Code + PKCE (S256).
 *
 * There is no password form here any more — the browser is redirected to
 * Keycloak, comes back with a `code`, and this module exchanges it for tokens
 * directly against the realm's `/token` endpoint. The ticket API is a pure
 * resource server that only ever sees the resulting access token.
 *
 * Implicit flow is deliberately not used: it returns tokens in the URL
 * fragment, where they land in history and referrers. PKCE removes the need for
 * it, and the realm has implicit disabled on the client.
 *
 * Token storage is in-memory only. localStorage would make any XSS a permanent
 * account takeover; a page reload instead re-runs the redirect, which Keycloak
 * answers silently from its SSO session cookie.
 */
import { queryClient } from "@/lib/query-client";
import type { UserInfo } from "./user-types";

// ---- configuration ------------------------------------------------------

type OidcConfig = {
    issuer: string;
    clientId: string;
};

/**
 * Discovery endpoints, resolved once and reused.
 * Fetched from the API rather than baked into the bundle so the same build runs
 * against any environment; the Vite vars are an override for local work.
 */
type Endpoints = {
    authorization: string;
    token: string;
    endSession: string;
};

let configPromise: Promise<OidcConfig> | null = null;
let endpointsPromise: Promise<Endpoints> | null = null;

async function loadConfig(): Promise<OidcConfig> {
    const envIssuer = import.meta.env.VITE_KEYCLOAK_ISSUER as string | undefined;
    const envClientId = import.meta.env.VITE_KEYCLOAK_CLIENT_ID as string | undefined;
    if (envIssuer && envClientId) {
        return { issuer: envIssuer.replace(/\/$/, ""), clientId: envClientId };
    }

    const res = await fetch(`${import.meta.env.VITE_API_URL}/api/v1/auth/config`);
    if (!res.ok) throw new Error("Could not load authentication configuration");
    const body = (await res.json()) as { issuer: string; client_id: string };
    return { issuer: body.issuer.replace(/\/$/, ""), clientId: body.client_id };
}

export function getOidcConfig(): Promise<OidcConfig> {
    configPromise ??= loadConfig().catch((err) => {
        // Don't cache a failure, or a transient blip locks the app out of login.
        configPromise = null;
        throw err;
    });
    return configPromise;
}

async function loadEndpoints(): Promise<Endpoints> {
    const { issuer } = await getOidcConfig();
    try {
        const res = await fetch(`${issuer}/.well-known/openid-configuration`);
        if (res.ok) {
            const doc = (await res.json()) as {
                authorization_endpoint: string;
                token_endpoint: string;
                end_session_endpoint: string;
            };
            return {
                authorization: doc.authorization_endpoint,
                token: doc.token_endpoint,
                endSession: doc.end_session_endpoint,
            };
        }
    } catch {
        // fall through to the conventional paths
    }
    // Keycloak's paths are stable, so a discovery hiccup shouldn't break login.
    return {
        authorization: `${issuer}/protocol/openid-connect/auth`,
        token: `${issuer}/protocol/openid-connect/token`,
        endSession: `${issuer}/protocol/openid-connect/logout`,
    };
}

function getEndpoints(): Promise<Endpoints> {
    endpointsPromise ??= loadEndpoints().catch((err) => {
        endpointsPromise = null;
        throw err;
    });
    return endpointsPromise;
}

export const REDIRECT_PATH = "/auth/callback";

function redirectUri(): string {
    return `${window.location.origin}${REDIRECT_PATH}`;
}

// ---- session state ------------------------------------------------------

let accessToken: string | null = null;
let refreshToken: string | null = null;
let idToken: string | null = null;
let refreshTimer: ReturnType<typeof setTimeout> | null = null;

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

export const getAccessToken = () => accessToken;
export const setAccessToken = (token: string | null) => {
    accessToken = token;
};

export const isAuthenticated = () => !!accessToken;

/** Routes where a failed refresh must not hard-redirect (avoids loops; router guards handle protected routes). */
export const AUTH_PUBLIC_PATHS = ["/login", "/signup", REDIRECT_PATH] as const;

export function isAuthPublicPath(pathname: string): boolean {
    return (AUTH_PUBLIC_PATHS as readonly string[]).includes(pathname);
}

function clearAuthSession(): void {
    accessToken = null;
    refreshToken = null;
    idToken = null;
    authState = { user: null };
    if (refreshTimer) {
        clearTimeout(refreshTimer);
        refreshTimer = null;
    }
}

// ---- PKCE ---------------------------------------------------------------

const VERIFIER_KEY = "kc_pkce_verifier";
const STATE_KEY = "kc_oauth_state";
const RETURN_TO_KEY = "kc_return_to";

function base64UrlEncode(bytes: Uint8Array): string {
    let binary = "";
    for (const byte of bytes) binary += String.fromCharCode(byte);
    return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

function randomUrlSafeString(byteLength: number): string {
    const bytes = new Uint8Array(byteLength);
    crypto.getRandomValues(bytes);
    return base64UrlEncode(bytes);
}

/** S256 challenge. Keycloak's client requires this method; `plain` is rejected. */
async function codeChallenge(verifier: string): Promise<string> {
    const digest = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(verifier));
    return base64UrlEncode(new Uint8Array(digest));
}

// ---- login / callback ---------------------------------------------------

type AuthorizeOptions = {
    /** Path to return to once authenticated. Defaults to the current location. */
    returnTo?: string;
    /** Land on Keycloak's registration form instead of its login form. */
    register?: boolean;
};

/**
 * Sends the browser to Keycloak. This navigates away, so it never resolves.
 */
export async function login(options: AuthorizeOptions = {}): Promise<void> {
    const [{ clientId }, endpoints] = await Promise.all([getOidcConfig(), getEndpoints()]);

    const verifier = randomUrlSafeString(32);
    const state = randomUrlSafeString(16);

    // sessionStorage, not localStorage: the verifier is single-use and must not
    // outlive the tab or leak to other tabs mid-flow.
    sessionStorage.setItem(VERIFIER_KEY, verifier);
    sessionStorage.setItem(STATE_KEY, state);

    const returnTo = options.returnTo ?? window.location.pathname + window.location.search;
    sessionStorage.setItem(RETURN_TO_KEY, isAuthPublicPath(returnTo) ? "/" : returnTo);

    const params = new URLSearchParams({
        client_id: clientId,
        redirect_uri: redirectUri(),
        response_type: "code",
        scope: "openid profile email",
        state,
        code_challenge: await codeChallenge(verifier),
        code_challenge_method: "S256",
    });

    const base = options.register
        ? endpoints.authorization.replace(/\/auth$/, "/registrations")
        : endpoints.authorization;

    window.location.assign(`${base}?${params.toString()}`);
}

/** Starts registration in Keycloak (the app no longer stores credentials). */
export function register(): Promise<void> {
    return login({ register: true, returnTo: "/" });
}

export type CallbackResult = {
    /** Where the user was headed before being sent to Keycloak. */
    returnTo: string;
};

/**
 * Completes the flow: validates `state`, exchanges the code for tokens.
 * Call this from the redirect route.
 */
export async function handleAuthCallback(search: URLSearchParams): Promise<CallbackResult> {
    const expectedState = sessionStorage.getItem(STATE_KEY);
    const verifier = sessionStorage.getItem(VERIFIER_KEY);
    const returnTo = sessionStorage.getItem(RETURN_TO_KEY) ?? "/";

    sessionStorage.removeItem(STATE_KEY);
    sessionStorage.removeItem(VERIFIER_KEY);
    sessionStorage.removeItem(RETURN_TO_KEY);

    const error = search.get("error");
    if (error) {
        throw new Error(search.get("error_description") || error);
    }

    const code = search.get("code");
    const state = search.get("state");
    if (!code) throw new Error("Authorization response contained no code");

    // The CSRF check for the authorization code flow: without it an attacker
    // could feed us their own code and log the victim into the attacker account.
    if (!expectedState || state !== expectedState) {
        throw new Error("Authorization state mismatch — please sign in again");
    }
    if (!verifier) {
        throw new Error("Missing PKCE verifier — please sign in again");
    }

    const [{ clientId }, endpoints] = await Promise.all([getOidcConfig(), getEndpoints()]);

    const res = await fetch(endpoints.token, {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: new URLSearchParams({
            grant_type: "authorization_code",
            client_id: clientId,
            code,
            redirect_uri: redirectUri(),
            code_verifier: verifier,
        }),
    });

    if (!res.ok) {
        throw new Error("Could not complete sign-in");
    }

    applyTokenResponse((await res.json()) as TokenResponse);
    // The profile comes from `GET /me`, which also provisions the local user row.
    queryClient.removeQueries({ queryKey: ["me"] });

    return { returnTo };
}

// ---- token handling -----------------------------------------------------

type TokenResponse = {
    access_token: string;
    refresh_token?: string;
    id_token?: string;
    expires_in: number;
};

function applyTokenResponse(tokens: TokenResponse): void {
    accessToken = tokens.access_token;
    if (tokens.refresh_token) refreshToken = tokens.refresh_token;
    if (tokens.id_token) idToken = tokens.id_token;
    setAuthUser(userFromAccessToken(tokens.access_token));
    scheduleRefresh(tokens.expires_in);
}

/**
 * Refreshes shortly before the access token expires, so requests don't have to
 * fail with a 401 first. `api()` still retries on 401 as a backstop for a
 * suspended tab or a clock skew.
 */
function scheduleRefresh(expiresInSeconds: number): void {
    if (refreshTimer) clearTimeout(refreshTimer);

    // Refresh at 75% of the lifetime, and never later than 30s before expiry.
    const leadMs = Math.min(expiresInSeconds * 0.25, 30) * 1000;
    const delay = Math.max(expiresInSeconds * 1000 - leadMs, 5_000);

    refreshTimer = setTimeout(() => {
        void tryRefresh();
    }, delay);
}

/** Decodes the access token's payload without verifying it. */
function decodeJwtPayload(token: string): Record<string, unknown> | null {
    const parts = token.split(".");
    if (parts.length !== 3) return null;
    try {
        const padded = parts[1].replace(/-/g, "+").replace(/_/g, "/");
        return JSON.parse(atob(padded + "=".repeat((4 - (padded.length % 4)) % 4))) as Record<string, unknown>;
    } catch {
        return null;
    }
}

/**
 * Builds a provisional user from the token so route guards have a role before
 * `GET /me` resolves.
 *
 * The signature is *not* checked here, and nothing security-relevant may depend
 * on it: the API re-derives the role from the verified token on every request.
 * This only decides which nav links render a moment earlier.
 */
function userFromAccessToken(token: string): UserInfo | null {
    const payload = decodeJwtPayload(token);
    if (!payload) return null;

    const realmAccess = payload.realm_access as { roles?: string[] } | undefined;
    const roles = realmAccess?.roles ?? [];
    const role = roles.includes("admin") ? "admin" : roles.includes("agent") ? "agent" : "user";

    return {
        // Placeholder until `GET /me` returns the local id that owns tickets;
        // the Keycloak subject is a different identifier.
        id: "",
        first_name: (payload.given_name as string) ?? "",
        last_name: (payload.family_name as string) ?? "",
        email: (payload.email as string) ?? (payload.preferred_username as string) ?? "",
        role,
        created_at: "",
    };
}

/**
 * Exchanges the refresh token for a new access token.
 *
 * Returns false when there is no usable session. Callers on protected routes
 * send the user back to Keycloak; `AuthProvider` does that on cold load, where
 * an existing SSO cookie makes it invisible.
 */
export async function tryRefresh(): Promise<boolean> {
    if (!refreshToken) {
        clearAuthSession();
        return false;
    }

    try {
        const [{ clientId }, endpoints] = await Promise.all([getOidcConfig(), getEndpoints()]);
        const res = await fetch(endpoints.token, {
            method: "POST",
            headers: { "Content-Type": "application/x-www-form-urlencoded" },
            body: new URLSearchParams({
                grant_type: "refresh_token",
                client_id: clientId,
                refresh_token: refreshToken,
            }),
        });

        if (!res.ok) throw new Error("Refresh failed");

        applyTokenResponse((await res.json()) as TokenResponse);
        return true;
    } catch {
        clearAuthSession();
        queryClient.removeQueries({ queryKey: ["me"] });
        return false;
    }
}

/**
 * Reports whether there is a live session.
 *
 * In-memory tokens don't survive a page reload, so after one this is false and
 * the caller redirects to Keycloak. That round-trip is invisible when the SSO
 * session cookie is still valid — Keycloak redirects straight back with a fresh
 * code instead of showing a login form. Persisting tokens to survive reloads
 * without the round-trip would mean putting them in localStorage, which is the
 * trade this module deliberately refuses.
 */
export async function restoreSession(): Promise<boolean> {
    if (accessToken) return true;
    return tryRefresh();
}

// ---- logout -------------------------------------------------------------

/**
 * RP-initiated logout. Ending the session at Keycloak is what actually
 * invalidates the refresh token and the SSO cookie — dropping local state alone
 * would leave the user silently signed back in on the next redirect.
 */
export async function logout(): Promise<void> {
    const currentIdToken = idToken;
    clearAuthSession();
    queryClient.removeQueries({ queryKey: ["me"] });
    queryClient.clear();

    try {
        const [{ clientId }, endpoints] = await Promise.all([getOidcConfig(), getEndpoints()]);
        const params = new URLSearchParams({
            client_id: clientId,
            post_logout_redirect_uri: `${window.location.origin}/login`,
        });
        // Keycloak requires either an id_token_hint or client_id + a registered
        // post-logout URI; sending the hint avoids its "are you sure?" page.
        if (currentIdToken) params.set("id_token_hint", currentIdToken);

        window.location.assign(`${endpoints.endSession}?${params.toString()}`);
    } catch {
        // If the realm can't be reached, at least don't leave the user looking
        // signed in.
        window.location.assign("/login");
    }
}
