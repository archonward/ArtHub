import { beforeEach, describe, expect, it, vi } from "vitest";

const login = vi.fn();
const signup = vi.fn();
const logout = vi.fn();
const getCurrentUser = vi.fn();

vi.mock("../services/api/forumApi", () => ({
  forumApi: {
    login,
    signup,
    logout,
    getCurrentUser,
  },
}));

describe("useAuth", () => {
  beforeEach(() => {
    vi.resetModules();
    vi.clearAllMocks();
  });

  it("isAuthenticated is false when currentUser is null", async () => {
    const { useAuth } = await import("./useAuth");

    const auth = useAuth();

    expect(auth.isAuthenticated.value).toBe(false);
  });

  it("login sets the current user and isAuthenticated becomes true", async () => {
    const user = { id: 1, username: "arthur" };
    login.mockResolvedValue(user);

    const { useAuth } = await import("./useAuth");
    const auth = useAuth();

    await auth.login("arthur", "secret");

    expect(auth.state.currentUser).toEqual(user);
    expect(auth.isAuthenticated.value).toBe(true);
  });

  it("logout clears the current user", async () => {
    const user = { id: 2, username: "bea" };
    login.mockResolvedValue(user);
    logout.mockResolvedValue({ logged_out: true });

    const { useAuth } = await import("./useAuth");
    const auth = useAuth();

    await auth.login("bea", "secret");
    await auth.logout();

    expect(auth.state.currentUser).toBeNull();
    expect(auth.isAuthenticated.value).toBe(false);
  });

  it("session expired event clears the user and sets authNotice", async () => {
    const user = { id: 3, username: "casey" };
    login.mockResolvedValue(user);

    const { useAuth } = await import("./useAuth");
    const { AUTH_SESSION_EXPIRED_EVENT } = await import("../services/api/client");
    const auth = useAuth();

    await auth.login("casey", "secret");

    window.dispatchEvent(
      new CustomEvent(AUTH_SESSION_EXPIRED_EVENT, {
        detail: { message: "Session expired." },
      }),
    );

    expect(auth.state.currentUser).toBeNull();
    expect(auth.state.authNotice).toBe("Session expired.");
  });
});
