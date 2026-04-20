import { computed, reactive, readonly } from "vue";
import { forumApi } from "../services/api/forumApi";
import { AUTH_SESSION_EXPIRED_EVENT, ApiError } from "../services/api/client";
import type { User } from "../types";

type AuthState = {
  currentUser: User | null;
  isBootstrapping: boolean;
  authNotice: string | null;
  bootstrapped: boolean;
};

const state = reactive<AuthState>({
  currentUser: null,
  isBootstrapping: true,
  authNotice: null,
  bootstrapped: false,
});

let bootstrapPromise: Promise<User | null> | null = null;
let listenersRegistered = false;

const setCurrentUser = (user: User | null) => {
  state.currentUser = user;
};

const bootstrap = async (): Promise<User | null> => {
  state.isBootstrapping = true;

  try {
    const user = await forumApi.getCurrentUser();
    setCurrentUser(user);
    return user;
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      setCurrentUser(null);
      return null;
    }

    console.error("Failed to bootstrap auth session:", error);
    setCurrentUser(null);
    return null;
  } finally {
    state.bootstrapped = true;
    state.isBootstrapping = false;
  }
};

const registerListeners = () => {
  if (listenersRegistered || typeof window === "undefined") {
    return;
  }

  window.addEventListener(AUTH_SESSION_EXPIRED_EVENT, (event: Event) => {
    const detail = (event as CustomEvent<{ message?: string }>).detail;
    setCurrentUser(null);
    state.authNotice =
      detail?.message || "Your session expired. Please log in again.";
  });

  listenersRegistered = true;
};

export const useAuth = () => {
  registerListeners();

  const ensureBootstrapped = async () => {
    if (state.bootstrapped) {
      return state.currentUser;
    }

    if (!bootstrapPromise) {
      bootstrapPromise = bootstrap().finally(() => {
        bootstrapPromise = null;
      });
    }

    return bootstrapPromise;
  };

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

  const login = async (username: string, password: string) => {
    state.authNotice = null;
    const user = await forumApi.login(username, password);
    setCurrentUser(user);
    state.bootstrapped = true;
    return user;
  };

  const signup = async (username: string, password: string) => {
    state.authNotice = null;
    const user = await forumApi.signup(username, password);
    setCurrentUser(user);
    state.bootstrapped = true;
    return user;
  };

  const logout = async () => {
    await forumApi.logout();
    setCurrentUser(null);
    state.authNotice = null;
    state.bootstrapped = true;
  };

  return {
    state: readonly(state),
    isAuthenticated: computed(() => state.currentUser !== null),
    ensureBootstrapped,
    refreshCurrentUser,
    login,
    signup,
    logout,
  };
};
