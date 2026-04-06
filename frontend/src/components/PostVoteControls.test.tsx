import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React, { useState } from "react";
import PostVoteControls from "./PostVoteControls";
import { useAuth } from "../context/AuthContext";
import { forumApi } from "../services/api/forumApi";
import type { Post } from "../types";

jest.mock("../context/AuthContext", () => ({
  useAuth: jest.fn(),
}));

jest.mock("../services/api/forumApi", () => ({
  forumApi: {
    voteOnPost: jest.fn(),
    removePostVote: jest.fn(),
  },
}));

const mockedUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;
const mockedForumApi = forumApi as jest.Mocked<typeof forumApi>;

const basePost: Post = {
  id: 4,
  companyId: 2,
  title: "Studio critique",
  body: "Need feedback",
  createdBy: 8,
  createdAt: "2026-04-06T00:00:00Z",
  voteScore: 0,
  currentUserVote: null,
};

function StatefulVoteControls({
  initialPost = basePost,
}: {
  initialPost?: Post;
}) {
  const [post, setPost] = useState(initialPost);
  return <PostVoteControls post={post} onPostChange={setPost} />;
}

describe("PostVoteControls", () => {
  beforeEach(() => {
    mockedUseAuth.mockReturnValue({
      currentUser: { id: 12, username: "arthur" },
      isAuthenticated: true,
      isBootstrapping: false,
      authNotice: null,
      login: jest.fn(),
      signup: jest.fn(),
      logout: jest.fn(),
      refreshCurrentUser: jest.fn(),
    });
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  it("upvotes a post and updates the score", async () => {
    mockedForumApi.voteOnPost.mockResolvedValue({
      ...basePost,
      voteScore: 1,
      currentUserVote: 1,
    });

    render(<StatefulVoteControls />);

    await userEvent.click(screen.getByRole("button", { name: "Upvote post" }));

    await waitFor(() => {
      expect(screen.getByLabelText("Vote score")).toHaveTextContent("1");
    });

    expect(mockedForumApi.voteOnPost).toHaveBeenCalledWith(4, { value: 1 });
    expect(screen.getByRole("button", { name: "Upvote post" })).toHaveClass(
      "vote-button--active",
    );
  });

  it("removes the current vote when clicking the same direction again", async () => {
    mockedForumApi.removePostVote.mockResolvedValue({
      ...basePost,
      voteScore: 0,
      currentUserVote: null,
    });

    render(
      <StatefulVoteControls
        initialPost={{
          ...basePost,
          voteScore: 1,
          currentUserVote: 1,
        }}
      />,
    );

    await userEvent.click(screen.getByRole("button", { name: "Upvote post" }));

    await waitFor(() => {
      expect(screen.getByLabelText("Vote score")).toHaveTextContent("0");
    });

    expect(mockedForumApi.removePostVote).toHaveBeenCalledWith(4);
    expect(screen.getByRole("button", { name: "Upvote post" })).not.toHaveClass(
      "vote-button--active",
    );
  });
});
