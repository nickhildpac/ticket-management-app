import { createContext, useState, useContext } from 'react';
import type { ReactNode } from 'react';

interface AuthContextType {
  isAuthenticated: boolean;
  token: string;
  user: string;
  login: (token: string, user: string) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [token, setToken] = useState("");
  const [user, setUser] = useState("");

  const login = (token: string, user: string) => {
    // In a real app, this would validate credentials and set tokens
    setIsAuthenticated(true);
    setUser(user)
    setToken("Bearer " + token);

  };

  const logout = () => {
    // In a real app, this would clear tokens
    fetch(`${import.meta.env.VITE_SERVER_URL}/v1/logout`,{
      method:"GET",
      credentials:"include"
    })
      .then((res) => {
        console.log(res)
        if (res.ok) {
          setIsAuthenticated(false);
          setToken("")
          setUser("")
        }
      })

  };

  return (
    <AuthContext.Provider value={{ isAuthenticated,user, token, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
