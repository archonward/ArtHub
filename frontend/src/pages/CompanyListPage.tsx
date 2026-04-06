import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import Notice from "../components/Notice";
import PageLayout from "../components/PageLayout";
import { useAuth } from "../context/AuthContext";
import { forumApi } from "../services/api/forumApi";
import type { Company } from "../types";

const CompanyListPage: React.FC = () => {
  const navigate = useNavigate();
  const { currentUser, isAuthenticated, logout, isBootstrapping } = useAuth();
  const [companies, setCompanies] = useState<Company[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [deletingCompanyId, setDeletingCompanyId] = useState<number | null>(null);
  const [loggingOut, setLoggingOut] = useState(false);

  useEffect(() => {
    const fetchCompanies = async () => {
      try {
        setCompanies(await forumApi.getCompanies());
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load companies.");
      } finally {
        setLoading(false);
      }
    };

    void fetchCompanies();
  }, []);

  const handleDelete = async (companyId: number) => {
    if (!window.confirm("Delete this company? This also removes its posts and comments.")) {
      return;
    }

    try {
      setActionError(null);
      setDeletingCompanyId(companyId);
      await forumApi.deleteCompany(companyId);
      setCompanies((current) => current.filter((company) => company.id !== companyId));
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "Failed to delete company.");
    } finally {
      setDeletingCompanyId(null);
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
      title="Companies"
      subtitle="Browse companies and open the discussion stream for each ticker."
      actions={
        <div className="action-row">
          {isAuthenticated ? (
            <button className="button button--secondary" onClick={() => navigate("/companies/new")}>
              New Company
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
        <Notice tone="info">You can browse companies publicly. Log in to create or manage content.</Notice>
      ) : null}
      {error ? <Notice tone="error">{error}</Notice> : null}
      {actionError ? <Notice tone="error">{actionError}</Notice> : null}
      {loading ? <p className="empty-state">Loading companies...</p> : null}

      {!loading && companies.length === 0 ? (
        <p className="empty-state">No companies yet. Add the first ticker to get started.</p>
      ) : null}

      {!loading && companies.length > 0 ? (
        <ul className="list">
          {companies.map((company) => (
            <li
              key={company.id}
              className="list-item list-item--interactive"
              onClick={() => navigate(`/companies/${company.id}`)}
            >
              <div className="action-row" style={{ justifyContent: "space-between" }}>
                <div>
                  <h2 className="content-title">{company.ticker}</h2>
                  <p className="content-body">{company.name}</p>
                  <p className="meta">{company.description || "No company summary provided."}</p>
                  <p className="meta">
                    Added by user {company.createdBy} on {new Date(company.createdAt).toLocaleString()}
                  </p>
                </div>
                {currentUser?.id === company.createdBy ? (
                  <button
                    className="button button--danger"
                    disabled={deletingCompanyId === company.id}
                    onClick={(event) => {
                      event.stopPropagation();
                      void handleDelete(company.id);
                    }}
                  >
                    {deletingCompanyId === company.id ? "Deleting..." : "Delete"}
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

export default CompanyListPage;
