import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import Notice from "../components/Notice";
import PageLayout from "../components/PageLayout";
import { useAuth } from "../context/AuthContext";
import { forumApi } from "../services/api/forumApi";
import type { Topic } from "../types";

const TopicListPage: React.FC = () => {
  const navigate = useNavigate();
  const { currentUser, isAuthenticated, logout, isBootstrapping } = useAuth();
  const [topics, setTopics] = useState<Topic[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [deletingTopicId, setDeletingTopicId] = useState<number | null>(null);
  const [loggingOut, setLoggingOut] = useState(false);

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

  const handleLogout = async () => {
    try {
      setLoggingOut(true);
      setActionError(null);
      await logout();
      navigate("/login");
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "Failed to log out.");
    } finally {
      setLoggingOut(false);
    }
  };

  return (
    <PageLayout
      title="Topics"
      subtitle="Browse the forum structure or start a new discussion area."
      actions={
        <div className="action-row">
          {isAuthenticated ? (
            <button className="button button--secondary" onClick={() => navigate("/topics/new")}>
              New Topic
            </button>
          ) : null}
          {isBootstrapping ? null : isAuthenticated ? (
            <button className="button button--ghost" onClick={handleLogout} disabled={loggingOut}>
              {loggingOut ? "Logging out..." : "Log Out"}
            </button>
          ) : (
            <button className="button button--secondary" onClick={() => navigate("/login")}>
              Log In
            </button>
          )}
        </div>
      }
    >
      {!isAuthenticated && !isBootstrapping ? (
        <Notice tone="info">You can browse topics publicly. Log in to create or manage content.</Notice>
      ) : null}
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
