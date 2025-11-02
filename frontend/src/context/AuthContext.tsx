import { createContext, useState } from 'react';
import type { ReactNode } from 'react';

export interface AuthContextType {
  isAuthenticated: boolean;
  token: string;
  user: string | null;
  login: (token: string, user: string) => void;
  logout: () => void;
}

export const AuthContext = createContext<AuthContextType | undefined>(
  undefined
);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [token, setToken] = useState("");
  const [user, setUser] = useState<string | null>(null);
  const isAuthenticated = token !== "";

  const login = (token: string, user: string) => {
    // In a real app, this would validate credentials and set tokens
    setUser(user)
    setToken("Bearer " + token);

  };

  const logout = () => {
    // In a real app, this would clear tokens
    fetch(`${import.meta.env.VITE_SERVER_URL}/logout`, {
      method: "GET",
      credentials: "include"
    })
      .then((res) => {
        console.log(res)
        if (res.ok) {
          setToken("")
          setUser(null)
        }
      })

  };

  return (
    <AuthContext.Provider value={{ isAuthenticated, user, token, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};
