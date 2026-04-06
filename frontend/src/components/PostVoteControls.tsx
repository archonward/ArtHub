import React, { useState } from "react";
import { useAuth } from "../context/AuthContext";
import { forumApi } from "../services/api/forumApi";
import type { Post } from "../types";

type PostVoteControlsProps = {
  post: Post;
  onPostChange: (post: Post) => void;
  compact?: boolean;
};

const PostVoteControls: React.FC<PostVoteControlsProps> = ({
  post,
  onPostChange,
  compact = false,
}) => {
  const { isAuthenticated } = useAuth();
  const [pending, setPending] = useState<null | -1 | 1 | 0>(null);
  const [error, setError] = useState<string | null>(null);

  const handleVoteClick =
    (value: -1 | 1) => async (event: React.MouseEvent<HTMLButtonElement>) => {
      event.stopPropagation();

      if (!isAuthenticated) {
        return;
      }

      try {
        setError(null);
        setPending(value);

        const updatedPost =
          post.currentUserVote === value
            ? await forumApi.removePostVote(post.id)
            : await forumApi.voteOnPost(post.id, { value });

        onPostChange(updatedPost);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to save vote.");
      } finally {
        setPending(null);
      }
    };

  return (
    <div className={`vote-controls${compact ? " vote-controls--compact" : ""}`}>
      <div className="vote-controls__row">
        <button
          type="button"
          className={`vote-button${post.currentUserVote === 1 ? " vote-button--active" : ""}`}
          onClick={handleVoteClick(1)}
          disabled={!isAuthenticated || pending !== null}
          aria-label="Upvote post"
        >
          +
        </button>
        <span className="vote-score" aria-label="Vote score">
          {post.voteScore}
        </span>
        <button
          type="button"
          className={`vote-button${post.currentUserVote === -1 ? " vote-button--active vote-button--down" : " vote-button--down"}`}
          onClick={handleVoteClick(-1)}
          disabled={!isAuthenticated || pending !== null}
          aria-label="Downvote post"
        >
          -
        </button>
      </div>
      {error ? <p className="vote-error">{error}</p> : null}
    </div>
  );
};

export default PostVoteControls;
