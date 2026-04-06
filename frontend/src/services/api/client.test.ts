import { AUTH_SESSION_EXPIRED_EVENT, ApiError, request } from "./client";

describe("request", () => {
  beforeEach(() => {
    global.fetch = jest.fn();
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  it("dispatches a session-expired event on unauthorized protected requests", async () => {
    const listener = jest.fn();
    window.addEventListener(AUTH_SESSION_EXPIRED_EVENT, listener);

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: async () => ({
        error: { message: "authentication required", code: "not_authenticated" },
      }),
      text: async () => "",
    });

    await expect(request("/topics", { method: "POST" })).rejects.toEqual(
      new ApiError("authentication required", 401, "not_authenticated"),
    );

    expect(listener).toHaveBeenCalledTimes(1);
    window.removeEventListener(AUTH_SESSION_EXPIRED_EVENT, listener);
  });

  it("can suppress the session-expired event for auth bootstrap requests", async () => {
    const listener = jest.fn();
    window.addEventListener(AUTH_SESSION_EXPIRED_EVENT, listener);

    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: async () => ({
        error: { message: "authentication required", code: "not_authenticated" },
      }),
      text: async () => "",
    });

    await expect(request("/auth/me", { notifyOnUnauthorized: false })).rejects.toEqual(
      new ApiError("authentication required", 401, "not_authenticated"),
    );

    expect(listener).not.toHaveBeenCalled();
    window.removeEventListener(AUTH_SESSION_EXPIRED_EVENT, listener);
  });
});
