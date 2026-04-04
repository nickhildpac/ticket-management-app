import { createContext, useContext, useState, useEffect } from "react";
import type { ReactNode } from "react";
import { setAuthUser } from "@/app/auth";
import { useMe } from "@/features/users/queries";

export interface UserInfo {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    role: string;
    skills?: string[];
    created_at: string;
    updated_at?: string;
}

interface UserContextType {
    user: UserInfo | null;
    setUser: (user: UserInfo | null) => void;
    clearUser: () => void;
}

const UserContext = createContext<UserContextType | undefined>(undefined);

export function UserProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<UserInfo | null>(null);
    const { data: me } = useMe();

    useEffect(() => {
        if (me) {
            setUser(me);
            setAuthUser(me);
        }
    }, [me]);

    const clearUser = () => setUser(null);

    return (
        <UserContext.Provider value={{ user, setUser, clearUser }}>
            {children}
        </UserContext.Provider>
    );
}

export function useUser() {
    const context = useContext(UserContext);
    if (context === undefined) {
        throw new Error("useUser must be used within a UserProvider");
    }
    return context;
}
