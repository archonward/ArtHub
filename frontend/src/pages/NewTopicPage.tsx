import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import Notice from "../components/Notice";
import PageLayout from "../components/PageLayout";
import { useCurrentUser } from "../hooks/useCurrentUser";
import { forumApi } from "../services/api/forumApi";

export default function NewTopicPage() {
  const navigate = useNavigate();
  const currentUser = useCurrentUser();
  const [form, setForm] = useState({ title: "", description: "" });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setError("");

    if (!currentUser) {
      setError("You must be logged in.");
      return;
    }

    if (!form.title.trim()) {
      setError("Title is required.");
      return;
    }

    setLoading(true);
    try {
      const topic = await forumApi.createTopic({
        title: form.title.trim(),
        description: form.description.trim(),
        createdBy: currentUser.id,
      });
      navigate(`/topics/${topic.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create topic.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <PageLayout title="Create Topic" subtitle="Set up a discussion area for related posts.">
      {error ? <Notice tone="error">{error}</Notice> : null}

      <form className="form-grid" onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="title">Title</label>
          <input
            id="title"
            name="title"
            value={form.title}
            onChange={(event) =>
              setForm((current) => ({ ...current, title: event.target.value }))
            }
            disabled={loading}
          />
        </div>

        <div className="field">
          <label htmlFor="description">Description</label>
          <textarea
            id="description"
            name="description"
            value={form.description}
            onChange={(event) =>
              setForm((current) => ({ ...current, description: event.target.value }))
            }
            rows={5}
            disabled={loading}
          />
        </div>

        <div className="form-actions">
          <button className="button button--primary" type="submit" disabled={loading}>
            {loading ? "Creating..." : "Create Topic"}
          </button>
          <button
            className="button button--secondary"
            type="button"
            onClick={() => navigate("/topics")}
            disabled={loading}
          >
            Cancel
          </button>
        </div>
      </form>
    </PageLayout>
  );
}
