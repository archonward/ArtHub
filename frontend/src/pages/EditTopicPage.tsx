import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import Notice from "../components/Notice";
import PageLayout from "../components/PageLayout";
import { forumApi } from "../services/api/forumApi";

const EditTopicPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const topicId = Number(id);

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!topicId) {
      setError("Invalid topic ID.");
      setLoading(false);
      return;
    }

    const fetchTopic = async () => {
      try {
        const topic = await forumApi.getTopic(topicId);
        setTitle(topic.title);
        setDescription(topic.description);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load topic.");
      } finally {
        setLoading(false);
      }
    };

    void fetchTopic();
  }, [topicId]);

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!topicId || !title.trim()) {
      setError("Title is required.");
      return;
    }

    setSubmitting(true);
    try {
      await forumApi.updateTopic(topicId, {
        title: title.trim(),
        description: description.trim(),
      });
      navigate(`/topics/${topicId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update topic.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <PageLayout title="Edit Topic" subtitle="Update the topic title or description.">
      {error ? <Notice tone="error">{error}</Notice> : null}
      {loading ? (
        <p className="empty-state">Loading topic...</p>
      ) : (
        <form className="form-grid" onSubmit={handleSubmit}>
          <div className="field">
            <label htmlFor="title">Title</label>
            <input id="title" value={title} onChange={(event) => setTitle(event.target.value)} />
          </div>

          <div className="field">
            <label htmlFor="description">Description</label>
            <textarea
              id="description"
              value={description}
              onChange={(event) => setDescription(event.target.value)}
              rows={5}
            />
          </div>

          <div className="form-actions">
            <button className="button button--primary" type="submit" disabled={submitting}>
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

export default EditTopicPage;
