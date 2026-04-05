import { CURRENT_USER_STORAGE_KEY } from "../../constants/storage";
import { ApiError } from "./client";
import { forumApi } from "./forumApi";

describe("forumApi", () => {
  beforeEach(() => {
    global.fetch = jest.fn();
    localStorage.clear();
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  it("maps login responses from the API envelope", async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ data: { id: 7, username: "arthur" } }),
    });

    await expect(forumApi.login("arthur")).resolves.toEqual({
      id: 7,
      username: "arthur",
    });

    expect(global.fetch).toHaveBeenCalledWith(
      "http://localhost:8080/login",
      expect.objectContaining({
        method: "POST",
      }),
    );
  });

  it("attaches X-User-ID for mutations when a current user exists", async () => {
    localStorage.setItem(CURRENT_USER_STORAGE_KEY, JSON.stringify({ id: 13, username: "owner" }));
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ data: { deleted: true } }),
    });

    await forumApi.deletePost(9);

    expect(global.fetch).toHaveBeenCalledWith(
      "http://localhost:8080/posts/9",
      expect.objectContaining({
        method: "DELETE",
        headers: expect.any(Headers),
      }),
    );

    const [, options] = (global.fetch as jest.Mock).mock.calls[0];
    expect((options.headers as Headers).get("X-User-ID")).toBe("13");
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
