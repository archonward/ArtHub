import React, { useEffect, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import Notice from "../components/Notice";
import PageLayout from "../components/PageLayout";
import { useAuth } from "../context/AuthContext";

type LoginLocationState = {
  from?: {
    pathname?: string;
  };
};

const LoginPage: React.FC = () => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const navigate = useNavigate();
  const location = useLocation();
  const { authNotice, isAuthenticated, isBootstrapping, login } = useAuth();

  useEffect(() => {
    if (!isBootstrapping && isAuthenticated) {
      navigate("/companies", { replace: true });
    }
  }, [isAuthenticated, isBootstrapping, navigate]);

  if (isBootstrapping) {
    return (
      <PageLayout title="ArtHub" subtitle="Restoring your session..." narrow>
        <p className="empty-state">Loading session...</p>
      </PageLayout>
    );
  }

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (username.trim() === "") {
      setError("Username is required.");
      return;
    }
    if (!password) {
      setError("Password is required.");
      return;
    }

    setLoading(true);
    setError(null);

    try {
      await login(username.trim(), password);
      const redirectTo = (location.state as LoginLocationState | null)?.from?.pathname || "/companies";
      navigate(redirectTo, { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <PageLayout
      title="ArtHub"
      subtitle="Log in to manage companies, posts, and comments."
      narrow
    >
      {authNotice ? <Notice tone="info">{authNotice}</Notice> : null}
      {error ? <Notice tone="error">{error}</Notice> : null}
      <form className="form-grid" onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="username">Username</label>
          <input
            id="username"
            type="text"
            value={username}
            onChange={(event) => setUsername(event.target.value)}
            disabled={loading}
            autoComplete="username"
          />
        </div>

        <div className="field">
          <label htmlFor="password">Password</label>
          <input
            id="password"
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            disabled={loading}
            autoComplete="current-password"
          />
        </div>

        <button className="button button--primary button--full" type="submit" disabled={loading}>
          {loading ? "Logging in..." : "Log In"}
        </button>
      </form>

      <p className="meta">
        Need an account? <Link to="/signup">Sign up</Link>.
      </p>
    </PageLayout>
  );
};

export default LoginPage;
