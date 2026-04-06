import { Navigate, Outlet, useLocation } from "react-router-dom";
import PageLayout from "./PageLayout";
import { useAuth } from "../context/AuthContext";

export default function ProtectedRoute() {
  const { isAuthenticated, isBootstrapping } = useAuth();
  const location = useLocation();

  if (isBootstrapping) {
    return (
      <PageLayout title="Loading" subtitle="Checking your session...">
        <p className="empty-state">Loading session...</p>
      </PageLayout>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  return <Outlet />;
}
