import React, { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import Notice from "../components/Notice";
import PageLayout from "../components/PageLayout";
import { forumApi } from "../services/api/forumApi";

const NewPostPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const companyId = Number(id);

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();

    if (!companyId) {
      setError("Company ID is missing.");
      return;
    }

    if (!title.trim() || !body.trim()) {
      setError("Title and body are required.");
      return;
    }

    setLoading(true);
    setError(null);

    try {
      await forumApi.createPost(companyId, {
        title: title.trim(),
        body: body.trim(),
      });
      navigate(`/companies/${companyId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create post.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <PageLayout title="Create Post" subtitle="Add a new thread for this company.">
      {error ? <Notice tone="error">{error}</Notice> : null}

      <form className="form-grid" onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="title">Title</label>
          <input
            id="title"
            value={title}
            onChange={(event) => setTitle(event.target.value)}
            disabled={loading}
          />
        </div>

        <div className="field">
          <label htmlFor="body">Body</label>
          <textarea
            id="body"
            value={body}
            onChange={(event) => setBody(event.target.value)}
            rows={8}
            disabled={loading}
          />
        </div>

        <div className="form-actions">
          <button className="button button--primary" type="submit" disabled={loading}>
            {loading ? "Creating..." : "Create Post"}
          </button>
          <button
            className="button button--secondary"
            type="button"
            onClick={() => navigate(companyId ? `/companies/${companyId}` : "/companies")}
            disabled={loading}
          >
            Cancel
          </button>
        </div>
      </form>
    </PageLayout>
  );
};

export default NewPostPage;
