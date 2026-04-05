import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import LoginPage from './pages/LoginPage';
import TopicListPage from './pages/TopicListPage';
import NewTopicPage from './pages/NewTopicPage';
import TopicDetailPage from './pages/TopicDetailPage';
import NewPostPage from './pages/NewPostPage';
import PostDetailPage from './pages/PostDetailPage';
import EditTopicPage from "./pages/EditTopicPage";
import EditPostPage from "./pages/EditPostPage";

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<LoginPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/topics" element={<TopicListPage />} />
        <Route path="/topics/new" element={<NewTopicPage />} />
        <Route path="/topics/:id" element={<TopicDetailPage />} />
        <Route path="/topics/:id/posts/new" element={<NewPostPage />} />
        <Route path="/posts/:postId" element={<PostDetailPage />} />
        <Route path="/topics/:id/edit" element={<EditTopicPage />} />
        <Route path="/posts/:postId/edit" element={<EditPostPage />} />
      </Routes>
    </Router>
  );
}


export default App;
