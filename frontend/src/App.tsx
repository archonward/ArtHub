import { BrowserRouter as Router, Navigate, Route, Routes } from "react-router-dom";
import PageLayout from "./components/PageLayout";
import ProtectedRoute from "./components/ProtectedRoute";
import { AuthProvider, useAuth } from "./context/AuthContext";
import CompanyDetailPage from "./pages/CompanyDetailPage";
import CompanyListPage from "./pages/CompanyListPage";
import EditCompanyPage from "./pages/EditCompanyPage";
import EditPostPage from "./pages/EditPostPage";
import LoginPage from "./pages/LoginPage";
import NewPostPage from "./pages/NewPostPage";
import NewCompanyPage from "./pages/NewCompanyPage";
import PostDetailPage from "./pages/PostDetailPage";
import SignupPage from "./pages/SignupPage";

function RootRedirect() {
  const { isAuthenticated, isBootstrapping } = useAuth();

  if (isBootstrapping) {
    return (
      <PageLayout title="Loading" subtitle="Restoring your session...">
        <p className="empty-state">Loading session...</p>
      </PageLayout>
    );
  }

  return <Navigate to={isAuthenticated ? "/companies" : "/login"} replace />;
}

function App() {
  return (
    <AuthProvider>
      <Router>
        <Routes>
          <Route path="/" element={<RootRedirect />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/signup" element={<SignupPage />} />
          <Route path="/companies" element={<CompanyListPage />} />
          <Route path="/companies/:id" element={<CompanyDetailPage />} />
          <Route path="/posts/:postId" element={<PostDetailPage />} />

          <Route element={<ProtectedRoute />}>
            <Route path="/companies/new" element={<NewCompanyPage />} />
            <Route path="/companies/:id/posts/new" element={<NewPostPage />} />
            <Route path="/companies/:id/edit" element={<EditCompanyPage />} />
            <Route path="/posts/:postId/edit" element={<EditPostPage />} />
          </Route>
        </Routes>
      </Router>
    </AuthProvider>
  );
}

export default App;
