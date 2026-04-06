import {
  createContext,
  type PropsWithChildren,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { forumApi } from "../services/api/forumApi";
import { AUTH_SESSION_EXPIRED_EVENT, ApiError } from "../services/api/client";
import type { User } from "../types";

type AuthContextValue = {
  currentUser: User | null;
  isAuthenticated: boolean;
  isBootstrapping: boolean;
  authNotice: string | null;
  login: (username: string, password: string) => Promise<User>;
  signup: (username: string, password: string) => Promise<User>;
  logout: () => Promise<void>;
  refreshCurrentUser: () => Promise<User | null>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: PropsWithChildren) {
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [isBootstrapping, setIsBootstrapping] = useState(true);
  const [authNotice, setAuthNotice] = useState<string | null>(null);

  const refreshCurrentUser = async (): Promise<User | null> => {
    try {
      const user = await forumApi.getCurrentUser();
      setCurrentUser(user);
      return user;
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        setCurrentUser(null);
        return null;
      }
      throw error;
    }
  };

  useEffect(() => {
    let active = true;

    const bootstrap = async () => {
      try {
        const user = await forumApi.getCurrentUser();
        if (active) {
          setCurrentUser(user);
        }
      } catch (error) {
        if (error instanceof ApiError && error.status === 401) {
          if (active) {
            setCurrentUser(null);
          }
        } else {
          console.error("Failed to bootstrap auth session:", error);
          if (active) {
            setCurrentUser(null);
          }
        }
      } finally {
        if (active) {
          setIsBootstrapping(false);
        }
      }
    };

    void bootstrap();

    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    const handleSessionExpired = (event: Event) => {
      const detail = (event as CustomEvent<{ message?: string }>).detail;
      setCurrentUser(null);
      setAuthNotice(detail?.message || "Your session expired. Please log in again.");
    };

    window.addEventListener(AUTH_SESSION_EXPIRED_EVENT, handleSessionExpired);
    return () => {
      window.removeEventListener(AUTH_SESSION_EXPIRED_EVENT, handleSessionExpired);
    };
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      currentUser,
      isAuthenticated: currentUser !== null,
      isBootstrapping,
      authNotice,
      login: async (username: string, password: string) => {
        setAuthNotice(null);
        const user = await forumApi.login(username, password);
        setCurrentUser(user);
        return user;
      },
      signup: async (username: string, password: string) => {
        setAuthNotice(null);
        const user = await forumApi.signup(username, password);
        setCurrentUser(user);
        return user;
      },
      logout: async () => {
        await forumApi.logout();
        setCurrentUser(null);
        setAuthNotice(null);
      },
      refreshCurrentUser,
    }),
    [authNotice, currentUser, isBootstrapping],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return value;
}
