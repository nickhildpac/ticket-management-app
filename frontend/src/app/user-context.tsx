import { createContext, useCallback, useContext, useLayoutEffect, useMemo, type ReactNode } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { getAccessToken, getAuthUser, setAuthUser } from "@/app/auth";
import type { UserInfo } from "@/app/user-types";
import { useMe } from "@/features/users/queries";

export type { UserInfo } from "@/app/user-types";

interface UserContextType {
    /** Current user: `GET /me` when available, otherwise JWT/bootstrap user while authenticated. */
    user: UserInfo | null;
    /** Seed the `me` query cache after login (auth module is already updated by `login()`). */
    primeMeCache: (user: UserInfo) => void;
    /** Clear profile cache and auth user (e.g. future use); prefer `logout()` for session teardown. */
    clearUser: () => void;
}

const UserContext = createContext<UserContextType | undefined>(undefined);

export function UserProvider({ children }: { children: ReactNode }) {
    const qc = useQueryClient();
    const { data: me } = useMe();

    const hasToken = !!getAccessToken();
    const user = useMemo(() => {
        if (me) return me;
        if (hasToken) {
            const fromAuth = getAuthUser();
            if (fromAuth) return fromAuth;
        }
        return null;
    }, [me, hasToken]);

    useLayoutEffect(() => {
        setAuthUser(user);
    }, [user]);

    const primeMeCache = useCallback(
        (info: UserInfo) => {
            qc.setQueryData(["me"], info);
        },
        [qc]
    );

    const clearUser = useCallback(() => {
        qc.removeQueries({ queryKey: ["me"] });
        setAuthUser(null);
    }, [qc]);

    const value = useMemo(
        () => ({
            user,
            primeMeCache,
            clearUser,
        }),
        [user, primeMeCache, clearUser]
    );

    return <UserContext.Provider value={value}>{children}</UserContext.Provider>;
}

export function useUser() {
    const context = useContext(UserContext);
    if (context === undefined) {
        throw new Error("useUser must be used within a UserProvider");
    }
    return context;
}
