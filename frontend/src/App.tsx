import { BrowserRouter as Router, Navigate, Route, Routes } from "react-router-dom";
import PageLayout from "./components/PageLayout";
import ProtectedRoute from "./components/ProtectedRoute";
import { AuthProvider, useAuth } from "./context/AuthContext";
import EditPostPage from "./pages/EditPostPage";
import EditTopicPage from "./pages/EditTopicPage";
import LoginPage from "./pages/LoginPage";
import NewPostPage from "./pages/NewPostPage";
import NewTopicPage from "./pages/NewTopicPage";
import PostDetailPage from "./pages/PostDetailPage";
import SignupPage from "./pages/SignupPage";
import TopicDetailPage from "./pages/TopicDetailPage";
import TopicListPage from "./pages/TopicListPage";

function RootRedirect() {
  const { isAuthenticated, isBootstrapping } = useAuth();

  if (isBootstrapping) {
    return (
      <PageLayout title="Loading" subtitle="Restoring your session...">
        <p className="empty-state">Loading session...</p>
      </PageLayout>
    );
  }

  return <Navigate to={isAuthenticated ? "/topics" : "/login"} replace />;
}

function App() {
  return (
    <AuthProvider>
      <Router>
        <Routes>
          <Route path="/" element={<RootRedirect />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/signup" element={<SignupPage />} />
          <Route path="/topics" element={<TopicListPage />} />
          <Route path="/topics/:id" element={<TopicDetailPage />} />
          <Route path="/posts/:postId" element={<PostDetailPage />} />

          <Route element={<ProtectedRoute />}>
            <Route path="/topics/new" element={<NewTopicPage />} />
            <Route path="/topics/:id/posts/new" element={<NewPostPage />} />
            <Route path="/topics/:id/edit" element={<EditTopicPage />} />
            <Route path="/posts/:postId/edit" element={<EditPostPage />} />
          </Route>
        </Routes>
      </Router>
    </AuthProvider>
  );
}

export default App;
