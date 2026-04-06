import React, { useEffect, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import Notice from "../components/Notice";
import PageLayout from "../components/PageLayout";
import { useAuth } from "../context/AuthContext";

type SignupLocationState = {
  from?: {
    pathname?: string;
  };
};

const SignupPage: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { isAuthenticated, isBootstrapping, signup } = useAuth();

  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isBootstrapping && isAuthenticated) {
      navigate("/topics", { replace: true });
    }
  }, [isAuthenticated, isBootstrapping, navigate]);

  if (isBootstrapping) {
    return (
      <PageLayout title="Create ArtHub Account" subtitle="Restoring your session..." narrow>
        <p className="empty-state">Loading session...</p>
      </PageLayout>
    );
  }

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();

    if (!username.trim()) {
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
      await signup(username.trim(), password);
      const redirectTo = (location.state as SignupLocationState | null)?.from?.pathname || "/topics";
      navigate(redirectTo, { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to sign up.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <PageLayout
      title="Create ArtHub Account"
      subtitle="Set up an account to create topics, posts, and comments."
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
            autoComplete="new-password"
          />
        </div>

        <button className="button button--primary button--full" type="submit" disabled={loading}>
          {loading ? "Creating account..." : "Sign Up"}
        </button>
      </form>

      <p className="meta">
        Already have an account? <Link to="/login">Log in</Link>.
      </p>
    </PageLayout>
  );
};

export default SignupPage;
