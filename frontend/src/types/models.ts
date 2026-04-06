export interface User {
  id: number;
  username: string;
}

export interface Topic {
  id: number;
  title: string;
  description: string;
  createdBy: number;
  createdAt: string;
}

export interface Post {
  id: number;
  topicId: number;
  title: string;
  body: string;
  createdBy: number;
  createdAt: string;
  voteScore: number;
  currentUserVote: -1 | 1 | null;
}

export interface Comment {
  id: number;
  postId: number;
  body: string;
  createdBy: number;
  createdAt: string;
}

export interface TopicDetails {
  topic: Topic;
  posts: Post[];
  pagination: Pagination;
}

export type PostSort = "top" | "new";

export interface Pagination {
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
  hasPrev: boolean;
  hasNext: boolean;
}
