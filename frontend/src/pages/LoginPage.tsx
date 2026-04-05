import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import Notice from "../components/Notice";
import PageLayout from "../components/PageLayout";
import { CURRENT_USER_STORAGE_KEY } from "../constants/storage";
import { forumApi } from "../services/api/forumApi";

const LoginPage: React.FC = () => {
  const [username, setUsername] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (username.trim() === "") {
      setError("Username is required");
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const user = await forumApi.login(username.trim());
      localStorage.setItem(CURRENT_USER_STORAGE_KEY, JSON.stringify(user));
      navigate("/topics");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <PageLayout
      title="ArtHub"
      subtitle="Enter a username to create or resume a lightweight local session."
      narrow
    >
      {error ? <Notice tone="error">{error}</Notice> : null}
      <form className="form-grid" onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="username">Username</label>
          <input
            id="username"
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            disabled={loading}
            autoComplete="username"
          />
        </div>

        <button className="button button--primary button--full" type="submit" disabled={loading}>
          {loading ? "Logging in..." : "Log In"}
        </button>
      </form>
    </PageLayout>
  );
};

export default LoginPage;
