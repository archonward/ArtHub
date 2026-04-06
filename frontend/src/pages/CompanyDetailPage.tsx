import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import Notice from "../components/Notice";
import PageLayout from "../components/PageLayout";
import PostVoteControls from "../components/PostVoteControls";
import { useAuth } from "../context/AuthContext";
import { forumApi } from "../services/api/forumApi";
import type { Company, Pagination, Post, PostSort } from "../types";

const COMPANY_POSTS_PAGE_SIZE = 10;

const CompanyDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { currentUser, isAuthenticated } = useAuth();
  const [company, setCompany] = useState<Company | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [sort, setSort] = useState<PostSort>("top");
  const [page, setPage] = useState(1);
  const [pagination, setPagination] = useState<Pagination | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const handlePostChange = (updatedPost: Post) => {
    setPosts((currentPosts) =>
      currentPosts.map((post) => (post.id === updatedPost.id ? updatedPost : post)),
    );
  };

  useEffect(() => {
    const companyId = Number(id);
    if (!companyId) {
      setError("Invalid company ID.");
      setLoading(false);
      return;
    }

    const fetchCompany = async () => {
      setLoading(true);
      setError(null);

      try {
        const details = await forumApi.getCompanyDetails(
          companyId,
          sort,
          page,
          COMPANY_POSTS_PAGE_SIZE,
        );
        setCompany(details.company);
        setPosts(details.posts);
        setPagination(details.pagination);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load company.");
      } finally {
        setLoading(false);
      }
    };

    void fetchCompany();
  }, [id, sort, page]);

  if (loading) {
    return (
      <PageLayout title="Company" subtitle="Loading company discussion...">
        <p className="empty-state">Loading company...</p>
      </PageLayout>
    );
  }

  if (error || !company) {
    return (
      <PageLayout title="Company" subtitle="Discussion area unavailable.">
        <Notice tone="error">{error || "Company not found."}</Notice>
      </PageLayout>
    );
  }

  return (
    <PageLayout
      title={company.ticker}
      subtitle={company.name}
      actions={
        <div className="action-row">
          <button className="button button--secondary" onClick={() => navigate("/companies")}>
            Back to Companies
          </button>
          {isAuthenticated ? (
            <button
              className="button button--secondary"
              onClick={() => navigate(`/companies/${company.id}/posts/new`)}
            >
              New Post
            </button>
          ) : null}
          {currentUser?.id === company.createdBy ? (
            <button
              className="button button--ghost"
              onClick={() => navigate(`/companies/${company.id}/edit`)}
            >
              Edit Company
            </button>
          ) : null}
        </div>
      }
    >
      <p className="meta">{company.description || "No company summary provided."}</p>
      {!isAuthenticated ? (
        <Notice tone="info">Log in if you want to create a post for this company.</Notice>
      ) : null}
      {currentUser && currentUser.id !== company.createdBy ? (
        <Notice tone="info">Only the company owner can edit this company profile.</Notice>
      ) : null}
      <p className="meta">
        Added by user {company.createdBy} on {new Date(company.createdAt).toLocaleString()}
      </p>

      <div className="stack">
        <div className="section-header">
          <h2 className="section-title">Posts ({posts.length})</h2>
          <label className="sort-control" htmlFor="post-sort">
            <span>Sort</span>
            <select
              id="post-sort"
              value={sort}
              onChange={(event) => {
                setSort(event.target.value as PostSort);
                setPage(1);
              }}
            >
              <option value="top">Top</option>
              <option value="new">New</option>
            </select>
          </label>
        </div>

        {posts.length === 0 ? (
          <p className="empty-state">No posts yet. Create the first post for this company.</p>
        ) : (
          <ul className="list">
            {posts.map((post) => (
              <li
                key={post.id}
                className="list-item list-item--interactive"
                onClick={() => navigate(`/posts/${post.id}`)}
              >
                <div className="list-item__header">
                  <h3>{post.title}</h3>
                  <PostVoteControls post={post} onPostChange={handlePostChange} compact />
                </div>
                <p className="content-body">{post.body}</p>
                <p className="meta">
                  Created by user {post.createdBy} on {new Date(post.createdAt).toLocaleString()}
                </p>
              </li>
            ))}
          </ul>
        )}

        {pagination && pagination.totalPages > 0 ? (
          <div className="pagination">
            <button
              className="button button--secondary"
              type="button"
              disabled={!pagination.hasPrev}
              onClick={() => setPage((currentPage) => Math.max(1, currentPage - 1))}
            >
              Previous
            </button>
            <p className="meta pagination__label">
              Page {pagination.page} of {pagination.totalPages}
            </p>
            <button
              className="button button--secondary"
              type="button"
              disabled={!pagination.hasNext}
              onClick={() =>
                setPage((currentPage) =>
                  pagination.totalPages > 0
                    ? Math.min(pagination.totalPages, currentPage + 1)
                    : currentPage,
                )
              }
            >
              Next
            </button>
          </div>
        ) : null}
      </div>
    </PageLayout>
  );
};

export default CompanyDetailPage;
