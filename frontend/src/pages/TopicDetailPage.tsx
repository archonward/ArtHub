import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import Notice from "../components/Notice";
import PageLayout from "../components/PageLayout";
import { useCurrentUser } from "../hooks/useCurrentUser";
import { forumApi } from "../services/api/forumApi";
import type { Post, Topic } from "../types";

const TopicDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const currentUser = useCurrentUser();
  const [topic, setTopic] = useState<Topic | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const topicId = Number(id);
    if (!topicId) {
      setError("Invalid topic ID.");
      setLoading(false);
      return;
    }

    const fetchTopic = async () => {
      try {
        const details = await forumApi.getTopicDetails(topicId);
        setTopic(details.topic);
        setPosts(details.posts);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load topic.");
      } finally {
        setLoading(false);
      }
    };

    void fetchTopic();
  }, [id]);

  if (loading) {
    return (
      <PageLayout title="Topic" subtitle="Loading discussion area...">
        <p className="empty-state">Loading topic...</p>
      </PageLayout>
    );
  }

  if (error || !topic) {
    return (
      <PageLayout title="Topic" subtitle="Discussion area unavailable.">
        <Notice tone="error">{error || "Topic not found."}</Notice>
      </PageLayout>
    );
  }

  return (
    <PageLayout
      title={topic.title}
      subtitle={topic.description || "No description provided."}
      actions={
        <div className="action-row">
          <button className="button button--secondary" onClick={() => navigate("/topics")}>
            Back to Topics
          </button>
          <button
            className="button button--secondary"
            onClick={() => navigate(`/topics/${topic.id}/posts/new`)}
          >
            New Post
          </button>
          {currentUser?.id === topic.createdBy ? (
            <button
              className="button button--ghost"
              onClick={() => navigate(`/topics/${topic.id}/edit`)}
            >
              Edit Topic
            </button>
          ) : null}
        </div>
      }
    >
      {currentUser && currentUser.id !== topic.createdBy ? (
        <Notice tone="info">Only the topic owner can edit this topic.</Notice>
      ) : null}
      <p className="meta">
        Created by user {topic.createdBy} on {new Date(topic.createdAt).toLocaleString()}
      </p>

      <div className="stack">
        <h2 className="section-title">Posts ({posts.length})</h2>

        {posts.length === 0 ? (
          <p className="empty-state">No posts yet. Create the first post in this topic.</p>
        ) : (
          <ul className="list">
            {posts.map((post) => (
              <li
                key={post.id}
                className="list-item list-item--interactive"
                onClick={() => navigate(`/posts/${post.id}`)}
              >
                <h3>{post.title}</h3>
                <p className="content-body">{post.body}</p>
                <p className="meta">
                  Created by user {post.createdBy} on{" "}
                  {new Date(post.createdAt).toLocaleString()}
                </p>
              </li>
            ))}
          </ul>
        )}
      </div>
    </PageLayout>
  );
};

export default TopicDetailPage;
