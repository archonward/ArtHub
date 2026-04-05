import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import Notice from "../components/Notice";
import PageLayout from "../components/PageLayout";
import { CURRENT_USER_STORAGE_KEY } from "../constants/storage";
import { useCurrentUser } from "../hooks/useCurrentUser";
import { forumApi } from "../services/api/forumApi";
import type { Topic } from "../types";

const TopicListPage: React.FC = () => {
  const navigate = useNavigate();
  const currentUser = useCurrentUser();
  const [topics, setTopics] = useState<Topic[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [deletingTopicId, setDeletingTopicId] = useState<number | null>(null);

  useEffect(() => {
    const fetchTopics = async () => {
      try {
        setTopics(await forumApi.getTopics());
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load topics.");
      } finally {
        setLoading(false);
      }
    };

    void fetchTopics();
  }, []);

  const handleDelete = async (topicId: number) => {
    if (!window.confirm("Delete this topic? This also removes its posts and comments.")) {
      return;
    }

    try {
      setActionError(null);
      setDeletingTopicId(topicId);
      await forumApi.deleteTopic(topicId);
      setTopics((current) => current.filter((topic) => topic.id !== topicId));
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "Failed to delete topic.");
    } finally {
      setDeletingTopicId(null);
    }
  };

  return (
    <PageLayout
      title="Topics"
      subtitle="Browse the forum structure or start a new discussion area."
      actions={
        <div className="action-row">
          <button className="button button--secondary" onClick={() => navigate("/topics/new")}>
            New Topic
          </button>
          <button
            className="button button--ghost"
            onClick={() => {
              localStorage.removeItem(CURRENT_USER_STORAGE_KEY);
              navigate("/login");
            }}
          >
            Log Out
          </button>
        </div>
      }
    >
      {error ? <Notice tone="error">{error}</Notice> : null}
      {actionError ? <Notice tone="error">{actionError}</Notice> : null}
      {loading ? <p className="empty-state">Loading topics...</p> : null}

      {!loading && topics.length === 0 ? (
        <p className="empty-state">No topics yet. Create the first one to get started.</p>
      ) : null}

      {!loading && topics.length > 0 ? (
        <ul className="list">
          {topics.map((topic) => (
            <li
              key={topic.id}
              className="list-item list-item--interactive"
              onClick={() => navigate(`/topics/${topic.id}`)}
            >
              <div className="action-row" style={{ justifyContent: "space-between" }}>
                <div>
                  <h2 className="content-title">{topic.title}</h2>
                  <p className="content-body">{topic.description || "No description provided."}</p>
                  <p className="meta">
                    Created by user {topic.createdBy} on{" "}
                    {new Date(topic.createdAt).toLocaleString()}
                  </p>
                </div>
                {currentUser?.id === topic.createdBy ? (
                  <button
                    className="button button--danger"
                    disabled={deletingTopicId === topic.id}
                    onClick={(event) => {
                      event.stopPropagation();
                      void handleDelete(topic.id);
                    }}
                  >
                    {deletingTopicId === topic.id ? "Deleting..." : "Delete"}
                  </button>
                ) : null}
              </div>
            </li>
          ))}
        </ul>
      ) : null}
    </PageLayout>
  );
};

export default TopicListPage;
