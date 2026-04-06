import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import Notice from "../components/Notice";
import PageLayout from "../components/PageLayout";
import { useAuth } from "../context/AuthContext";
import { forumApi } from "../services/api/forumApi";

const EditCompanyPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { currentUser } = useAuth();
  const companyId = Number(id);

  const [ticker, setTicker] = useState("");
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [ownerId, setOwnerId] = useState<number | null>(null);

  useEffect(() => {
    if (!companyId) {
      setError("Invalid company ID.");
      setLoading(false);
      return;
    }

    const fetchCompany = async () => {
      try {
        const company = await forumApi.getCompany(companyId);
        setTicker(company.ticker);
        setName(company.name);
        setDescription(company.description);
        setOwnerId(company.createdBy);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load company.");
      } finally {
        setLoading(false);
      }
    };

    void fetchCompany();
  }, [companyId]);

  const isOwner = ownerId !== null && currentUser?.id === ownerId;

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!companyId || !ticker.trim() || !name.trim()) {
      setError("Ticker and company name are required.");
      return;
    }

    setSubmitting(true);
    try {
      await forumApi.updateCompany(companyId, {
        ticker: ticker.trim().toUpperCase(),
        name: name.trim(),
        description: description.trim(),
      });
      navigate(`/companies/${companyId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update company.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <PageLayout title="Edit Company" subtitle="Update the ticker, name, or summary.">
      {error ? <Notice tone="error">{error}</Notice> : null}
      {ownerId !== null && !isOwner ? (
        <Notice tone="error">You cannot edit a company you do not own.</Notice>
      ) : null}
      {loading ? (
        <p className="empty-state">Loading company...</p>
      ) : (
        <form className="form-grid" onSubmit={handleSubmit}>
          <div className="field">
            <label htmlFor="ticker">Ticker</label>
            <input
              id="ticker"
              value={ticker}
              onChange={(event) => setTicker(event.target.value.toUpperCase())}
              disabled={submitting || !isOwner}
            />
          </div>

          <div className="field">
            <label htmlFor="name">Company Name</label>
            <input
              id="name"
              value={name}
              onChange={(event) => setName(event.target.value)}
              disabled={submitting || !isOwner}
            />
          </div>

          <div className="field">
            <label htmlFor="description">Description</label>
            <textarea
              id="description"
              value={description}
              onChange={(event) => setDescription(event.target.value)}
              rows={5}
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

export default EditCompanyPage;
