import { ApiError } from "./client";
import { forumApi } from "./forumApi";

describe("forumApi", () => {
  beforeEach(() => {
    global.fetch = jest.fn();
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  it("maps login responses from the auth API envelope", async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ data: { id: 7, username: "arthur", created_at: "2026-04-06T00:00:00Z" } }),
    });

    await expect(forumApi.login("arthur", "secret123")).resolves.toEqual({
      id: 7,
      username: "arthur",
    });

    expect(global.fetch).toHaveBeenCalledWith(
      "http://localhost:8080/auth/login",
      expect.objectContaining({
        method: "POST",
        credentials: "include",
      }),
    );
  });

  it("uses the session bootstrap endpoint", async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ data: { id: 3, username: "owner", created_at: "2026-04-06T00:00:00Z" } }),
    });

    await expect(forumApi.getCurrentUser()).resolves.toEqual({
      id: 3,
      username: "owner",
    });

    expect(global.fetch).toHaveBeenCalledWith(
      "http://localhost:8080/auth/me",
      expect.objectContaining({
        credentials: "include",
      }),
    );
  });

  it("surfaces structured JSON API errors", async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: false,
      status: 403,
      json: async () => ({
        error: { message: "you are not allowed to modify this resource", code: "forbidden" },
      }),
      text: async () => "",
    });

    await expect(forumApi.deletePost(9)).rejects.toEqual(
      new ApiError("you are not allowed to modify this resource", 403, "forbidden"),
    );
  });
});
