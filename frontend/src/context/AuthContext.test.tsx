import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { AuthProvider, useAuth } from "./AuthContext";

const jsonResponse = (data: unknown, status = 200) => ({
  ok: status >= 200 && status < 300,
  status,
  json: async () => data,
  text: async () => "",
});

function AuthConsumer() {
  const {
    authNotice,
    currentUser,
    isAuthenticated,
    isBootstrapping,
    login,
    signup,
    logout,
  } = useAuth();

  return (
    <div>
      <div data-testid="bootstrapping">{String(isBootstrapping)}</div>
      <div data-testid="authenticated">{String(isAuthenticated)}</div>
      <div data-testid="username">{currentUser?.username || "anonymous"}</div>
      <div data-testid="auth-notice">{authNotice || "none"}</div>
      <button onClick={() => void login("arthur", "secret123")}>login</button>
      <button onClick={() => void signup("newuser", "secret123")}>signup</button>
      <button onClick={() => void logout()}>logout</button>
    </div>
  );
}

describe("AuthContext", () => {
  beforeEach(() => {
    global.fetch = jest.fn();
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  it("bootstraps the current session from /auth/me", async () => {
    (global.fetch as jest.Mock).mockResolvedValueOnce(
      jsonResponse({ data: { id: 1, username: "arthur", created_at: "2026-04-06T00:00:00Z" } }),
    );

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("authenticated")).toHaveTextContent("true");
    });
    expect(screen.getByTestId("username")).toHaveTextContent("arthur");
  });

  it("handles unauthenticated bootstrap cleanly", async () => {
    (global.fetch as jest.Mock).mockResolvedValueOnce(
      jsonResponse({ error: { message: "authentication required", code: "not_authenticated" } }, 401),
    );

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("bootstrapping")).toHaveTextContent("false");
    });
    expect(screen.getByTestId("authenticated")).toHaveTextContent("false");
    expect(screen.getByTestId("username")).toHaveTextContent("anonymous");
  });

  it("logs in through the auth endpoint", async () => {
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce(
        jsonResponse({ error: { message: "authentication required", code: "not_authenticated" } }, 401),
      )
      .mockResolvedValueOnce(
        jsonResponse({ data: { id: 2, username: "arthur", created_at: "2026-04-06T00:00:00Z" } }),
      );

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("bootstrapping")).toHaveTextContent("false");
    });

    fireEvent.click(screen.getByRole("button", { name: "login" }));

    await waitFor(() => {
      expect(screen.getByTestId("username")).toHaveTextContent("arthur");
    });
  });

  it("signs up through the auth endpoint", async () => {
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce(
        jsonResponse({ error: { message: "authentication required", code: "not_authenticated" } }, 401),
      )
      .mockResolvedValueOnce(
        jsonResponse({ data: { id: 3, username: "newuser", created_at: "2026-04-06T00:00:00Z" } }, 201),
      );

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("bootstrapping")).toHaveTextContent("false");
    });

    fireEvent.click(screen.getByRole("button", { name: "signup" }));

    await waitFor(() => {
      expect(screen.getByTestId("username")).toHaveTextContent("newuser");
    });
  });

  it("logs out and clears the current user", async () => {
    (global.fetch as jest.Mock)
      .mockResolvedValueOnce(
        jsonResponse({ data: { id: 1, username: "arthur", created_at: "2026-04-06T00:00:00Z" } }),
      )
      .mockResolvedValueOnce(jsonResponse({ data: { logged_out: true } }));

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("authenticated")).toHaveTextContent("true");
    });

    fireEvent.click(screen.getByRole("button", { name: "logout" }));

    await waitFor(() => {
      expect(screen.getByTestId("authenticated")).toHaveTextContent("false");
    });
    expect(screen.getByTestId("username")).toHaveTextContent("anonymous");
  });

  it("clears the session and shows a notice when a protected request reports expiration", async () => {
    (global.fetch as jest.Mock).mockResolvedValueOnce(
      jsonResponse({ data: { id: 1, username: "arthur", created_at: "2026-04-06T00:00:00Z" } }),
    );

    render(
      <AuthProvider>
        <AuthConsumer />
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("authenticated")).toHaveTextContent("true");
    });

    await act(async () => {
      window.dispatchEvent(
        new CustomEvent("arthub:auth-session-expired", {
          detail: { message: "Your session expired. Please log in again." },
        }),
      );
    });

    await waitFor(() => {
      expect(screen.getByTestId("authenticated")).toHaveTextContent("false");
    });
    expect(screen.getByTestId("auth-notice")).toHaveTextContent(
      "Your session expired. Please log in again.",
    );
  });
});
