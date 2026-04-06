import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import Notice from "../components/Notice";
import PageLayout from "../components/PageLayout";
import { forumApi } from "../services/api/forumApi";

export default function NewCompanyPage() {
  const navigate = useNavigate();
  const [form, setForm] = useState({ ticker: "", name: "", description: "" });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setError("");

    if (!form.ticker.trim()) {
      setError("Ticker is required.");
      return;
    }

    if (!form.name.trim()) {
      setError("Company name is required.");
      return;
    }

    setLoading(true);
    try {
      const company = await forumApi.createCompany({
        ticker: form.ticker.trim().toUpperCase(),
        name: form.name.trim(),
        description: form.description.trim(),
      });
      navigate(`/companies/${company.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create company.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <PageLayout title="Add Company" subtitle="Create a company page anchored to a ticker.">
      {error ? <Notice tone="error">{error}</Notice> : null}

      <form className="form-grid" onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="ticker">Ticker</label>
          <input
            id="ticker"
            name="ticker"
            value={form.ticker}
            onChange={(event) =>
              setForm((current) => ({ ...current, ticker: event.target.value.toUpperCase() }))
            }
            disabled={loading}
          />
        </div>

        <div className="field">
          <label htmlFor="name">Company Name</label>
          <input
            id="name"
            name="name"
            value={form.name}
            onChange={(event) =>
              setForm((current) => ({ ...current, name: event.target.value }))
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
            {loading ? "Creating..." : "Create Company"}
          </button>
          <button
            className="button button--secondary"
            type="button"
            onClick={() => navigate("/companies")}
            disabled={loading}
          >
            Cancel
          </button>
        </div>
      </form>
    </PageLayout>
  );
}
