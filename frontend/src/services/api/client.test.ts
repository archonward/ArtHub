import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ApiError, AUTH_SESSION_EXPIRED_EVENT, request } from "./client";

describe("request", () => {
  const fetchMock = vi.fn();

  beforeEach(() => {
    fetchMock.mockReset();
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("returns the data field from a successful JSON envelope", async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      status: 200,
      json: vi.fn().mockResolvedValue({ data: { id: 1, name: "ArtHub" } }),
    });

    await expect(request("/companies")).resolves.toEqual({ id: 1, name: "ArtHub" });
  });

  it("throws an ApiError with the correct status and message for non-OK responses", async () => {
    fetchMock.mockResolvedValue({
      ok: false,
      status: 404,
      json: vi.fn().mockResolvedValue({
        error: {
          message: "Not found.",
          code: "not_found",
        },
      }),
      text: vi.fn().mockResolvedValue(""),
    });

    await expect(request("/missing")).rejects.toMatchObject<ApiError>({
      name: "ApiError",
      message: "Not found.",
      status: 404,
      code: "not_found",
    });
  });

  it("dispatches the auth session expired event on 401 responses", async () => {
    fetchMock.mockResolvedValue({
      ok: false,
      status: 401,
      json: vi.fn().mockResolvedValue({
        error: {
          message: "Unauthorized.",
          code: "unauthorized",
        },
      }),
      text: vi.fn().mockResolvedValue(""),
    });

    const listener = vi.fn();
    window.addEventListener(AUTH_SESSION_EXPIRED_EVENT, listener);

    await expect(request("/auth/me")).rejects.toBeInstanceOf(ApiError);

    expect(listener).toHaveBeenCalledTimes(1);
    window.removeEventListener(AUTH_SESSION_EXPIRED_EVENT, listener);
  });
});
