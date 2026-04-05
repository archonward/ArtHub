import { forumApi } from "./forumApi";
import { ApiError } from "./client";

describe("forumApi", () => {
  beforeEach(() => {
    global.fetch = jest.fn();
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  it("maps login responses into the frontend user model", async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ id: 7, username: "arthur" }),
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

  it("surfaces JSON API errors", async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: false,
      status: 403,
      json: async () => ({ error: "forbidden" }),
      text: async () => "",
    });

    await expect(forumApi.deletePost(9)).rejects.toEqual(new ApiError("forbidden", 403));
  });
});
