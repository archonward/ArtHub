import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import Notice from "../components/Notice";
import PageLayout from "../components/PageLayout";
import { useCurrentUser } from "../hooks/useCurrentUser";
import { forumApi } from "../services/api/forumApi";

const EditPostPage: React.FC = () => {
  const { postId } = useParams<{ postId: string }>();
  const navigate = useNavigate();
  const currentUser = useCurrentUser();
  const parsedPostId = Number(postId);

  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [ownerId, setOwnerId] = useState<number | null>(null);

  useEffect(() => {
    if (!parsedPostId) {
      setError("Invalid post ID.");
      setLoading(false);
      return;
    }

    const fetchPost = async () => {
      try {
        const post = await forumApi.getPost(parsedPostId);
        setTitle(post.title);
        setBody(post.body);
        setOwnerId(post.createdBy);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load post.");
      } finally {
        setLoading(false);
      }
    };

    void fetchPost();
  }, [parsedPostId]);

  const isOwner = ownerId !== null && currentUser?.id === ownerId;

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();

    if (!parsedPostId || !title.trim() || !body.trim()) {
      setError("Title and body are required.");
      return;
    }

    setSubmitting(true);
    try {
      await forumApi.updatePost(parsedPostId, {
        title: title.trim(),
        body: body.trim(),
      });
      navigate(`/posts/${parsedPostId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update post.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <PageLayout title="Edit Post" subtitle="Revise the title and body for this post.">
      {error ? <Notice tone="error">{error}</Notice> : null}
      {ownerId !== null && !isOwner ? (
        <Notice tone="error">You cannot edit a post you do not own.</Notice>
      ) : null}
      {loading ? (
        <p className="empty-state">Loading post...</p>
      ) : (
        <form className="form-grid" onSubmit={handleSubmit}>
          <div className="field">
            <label htmlFor="title">Title</label>
            <input
              id="title"
              value={title}
              onChange={(event) => setTitle(event.target.value)}
              disabled={submitting || !isOwner}
            />
          </div>

          <div className="field">
            <label htmlFor="body">Body</label>
            <textarea
              id="body"
              value={body}
              onChange={(event) => setBody(event.target.value)}
              rows={8}
              disabled={submitting || !isOwner}
            />
          </div>

          <div className="form-actions">
            <button className="button button--primary" type="submit" disabled={submitting || !isOwner}>
              {submitting ? "Saving..." : "Save Changes"}
            </button>
            <button
              className="button button--secondary"
              type="button"
              onClick={() => navigate(-1)}
              disabled={submitting}
            >
              Cancel
            </button>
          </div>
        </form>
      )}
    </PageLayout>
  );
};

export default EditPostPage;
