import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import Notice from "../components/Notice";
import PageLayout from "../components/PageLayout";
import { useCurrentUser } from "../hooks/useCurrentUser";
import { forumApi } from "../services/api/forumApi";
import type { Comment, Post } from "../types";

const PostDetailPage: React.FC = () => {
  const { postId } = useParams<{ postId: string }>();
  const navigate = useNavigate();
  const currentUser = useCurrentUser();

  const [post, setPost] = useState<Post | null>(null);
  const [comments, setComments] = useState<Comment[]>([]);
  const [newComment, setNewComment] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  useEffect(() => {
    const parsedId = Number(postId);
    if (!parsedId) {
      setError("Invalid post ID.");
      setLoading(false);
      return;
    }

    const fetchPost = async () => {
      try {
        const [postRecord, commentRecords] = await Promise.all([
          forumApi.getPost(parsedId),
          forumApi.getPostComments(parsedId),
        ]);
        setPost(postRecord);
        setComments(commentRecords);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load post.");
      } finally {
        setLoading(false);
      }
    };

    void fetchPost();
  }, [postId]);

  const isOwner = currentUser?.id === post?.createdBy;

  const handleDelete = async () => {
    if (!post || !window.confirm("Delete this post and all its comments?")) {
      return;
    }

    try {
      setActionError(null);
      setDeleting(true);
      await forumApi.deletePost(post.id);
      navigate(`/topics/${post.topicId}`);
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "Failed to delete post.");
    } finally {
      setDeleting(false);
    }
  };

  const handleAddComment = async (event: React.FormEvent) => {
    event.preventDefault();

    if (!post || !currentUser || !newComment.trim()) {
      return;
    }

    setSubmitting(true);
    try {
      setActionError(null);
      const comment = await forumApi.createComment(post.id, {
        body: newComment.trim(),
        createdBy: currentUser.id,
      });
      setComments((current) => [...current, comment]);
      setNewComment("");
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "Failed to post comment.");
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <PageLayout title="Post" subtitle="Loading post details...">
        <p className="empty-state">Loading post...</p>
      </PageLayout>
    );
  }

  if (error || !post) {
    return (
      <PageLayout title="Post" subtitle="Discussion unavailable.">
        <Notice tone="error">{error || "Post not found."}</Notice>
      </PageLayout>
    );
  }

  return (
    <PageLayout
      title={post.title}
      subtitle={`Created by user ${post.createdBy} on ${new Date(post.createdAt).toLocaleString()}`}
      actions={
        <div className="action-row">
          <button className="button button--secondary" onClick={() => navigate(-1)}>
            Back
          </button>
          {isOwner ? (
            <>
              <button
                className="button button--ghost"
                disabled={deleting}
                onClick={() => navigate(`/posts/${post.id}/edit`)}
              >
                Edit
              </button>
              <button className="button button--danger" disabled={deleting} onClick={handleDelete}>
                {deleting ? "Deleting..." : "Delete"}
              </button>
            </>
          ) : null}
        </div>
      }
    >
      {actionError ? <Notice tone="error">{actionError}</Notice> : null}
      {currentUser && !isOwner ? (
        <Notice tone="info">Only the post owner can edit or delete this post.</Notice>
      ) : null}
      <p className="content-body">{post.body}</p>

      <hr className="divider" />

      <div className="stack">
        <h2 className="section-title">Comments ({comments.length})</h2>
        {comments.length === 0 ? (
          <p className="empty-state">No comments yet.</p>
        ) : (
          <ul className="list">
            {comments.map((comment) => (
              <li key={comment.id} className="list-item">
                <p className="content-body">{comment.body}</p>
                <p className="meta">
                  By user {comment.createdBy} on{" "}
                  {new Date(comment.createdAt).toLocaleString()}
                </p>
              </li>
            ))}
          </ul>
        )}

        {currentUser ? (
          <form className="form-grid" onSubmit={handleAddComment}>
            <div className="field">
              <label htmlFor="comment">Add a comment</label>
              <textarea
                id="comment"
                value={newComment}
                onChange={(event) => setNewComment(event.target.value)}
                rows={4}
                disabled={submitting}
              />
            </div>
            <button
              className="button button--primary"
              type="submit"
              disabled={submitting || !newComment.trim()}
            >
              {submitting ? "Posting..." : "Post Comment"}
            </button>
          </form>
        ) : (
          <Notice tone="info">Log in to join the discussion.</Notice>
        )}
      </div>
    </PageLayout>
  );
};

export default PostDetailPage;
