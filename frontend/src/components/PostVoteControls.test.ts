import { mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import PostVoteControls from "./PostVoteControls.vue";
import type { Post } from "../types";

const { voteOnPost, removePostVote, isAuthenticated } = vi.hoisted(() => ({
  voteOnPost: vi.fn(),
  removePostVote: vi.fn(),
  isAuthenticated: { value: true },
}));

vi.mock("../services/api/forumApi", () => ({
  forumApi: {
    voteOnPost,
    removePostVote,
  },
}));

vi.mock("../composables/useAuth", () => ({
  useAuth: () => ({
    isAuthenticated,
  }),
}));

const createPost = (overrides: Partial<Post> = {}): Post => ({
  id: 10,
  companyId: 2,
  title: "Title",
  body: "Body",
  createdBy: 7,
  createdAt: "2026-04-20T00:00:00Z",
  voteScore: 3,
  currentUserVote: null,
  ...overrides,
});

describe("PostVoteControls", () => {
  beforeEach(() => {
    voteOnPost.mockReset();
    removePostVote.mockReset();
    isAuthenticated.value = true;
  });

  it("renders upvote and downvote buttons", () => {
    const wrapper = mount(PostVoteControls, {
      props: {
        post: createPost(),
      },
    });

    const buttons = wrapper.findAll("button");
    expect(buttons).toHaveLength(2);
    expect(buttons[0].attributes("aria-label")).toBe("Upvote post");
    expect(buttons[1].attributes("aria-label")).toBe("Downvote post");
  });

  it("clicking upvote calls forumApi.voteOnPost with the correct value", async () => {
    voteOnPost.mockResolvedValue(createPost({ currentUserVote: 1, voteScore: 4 }));

    const wrapper = mount(PostVoteControls, {
      props: {
        post: createPost(),
      },
    });

    await wrapper.get('button[aria-label="Upvote post"]').trigger("click");

    expect(voteOnPost).toHaveBeenCalledWith(10, { value: 1 });
    expect(removePostVote).not.toHaveBeenCalled();
  });

  it("clicking an already selected vote removes the vote", async () => {
    removePostVote.mockResolvedValue(createPost({ currentUserVote: null, voteScore: 2 }));

    const wrapper = mount(PostVoteControls, {
      props: {
        post: createPost({ currentUserVote: 1 }),
      },
    });

    await wrapper.get('button[aria-label="Upvote post"]').trigger("click");

    expect(removePostVote).toHaveBeenCalledWith(10);
    expect(voteOnPost).not.toHaveBeenCalled();
  });

  it("disables buttons when the user is not authenticated", () => {
    isAuthenticated.value = false;

    const wrapper = mount(PostVoteControls, {
      props: {
        post: createPost(),
      },
    });

    const buttons = wrapper.findAll("button");
    expect(buttons[0].attributes("disabled")).toBeDefined();
    expect(buttons[1].attributes("disabled")).toBeDefined();
  });
});
