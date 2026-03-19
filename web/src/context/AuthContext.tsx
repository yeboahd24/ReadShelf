import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import type { User } from '../types';
import * as authApi from '../api/auth';
import { setAccessToken } from '../api/client';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    authApi.refresh()
      .then(() => {
        // Token refreshed but we don't have user info from refresh endpoint.
        // We'll set a minimal user marker; pages fetch data via authenticated API.
        setUser({ id: '', email: '', plan: '', created_at: '' });
      })
      .catch(() => {
        // No valid refresh token.
      })
      .finally(() => setLoading(false));
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const res = await authApi.login(email, password);
    setUser(res.user);
  }, []);

  const register = useCallback(async (email: string, password: string) => {
    const res = await authApi.register(email, password);
    setUser(res.user);
  }, []);

  const logout = useCallback(() => {
    setAccessToken(null);
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
