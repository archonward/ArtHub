import type { Comment, Pagination, Post, Topic, User } from "./models";

export interface ApiResponseEnvelope<T> {
  data: T;
}

export interface ApiErrorEnvelope {
  error: {
    message: string;
    code: string;
  };
}

export interface UserDto {
  id: number;
  username: string;
}

export interface TopicDto {
  id: number;
  title: string;
  description: string;
  created_by: number;
  created_at: string;
}

export interface PostDto {
  id: number;
  topic_id: number;
  title: string;
  body: string;
  created_by: number;
  created_at: string;
  vote_score: number;
  current_user_vote: -1 | 1 | null;
}

export interface CommentDto {
  id: number;
  post_id: number;
  body: string;
  created_by: number;
  created_at: string;
}

export interface PaginationDto {
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
  has_prev: boolean;
  has_next: boolean;
}

export interface TopicPostsPageDto {
  posts: PostDto[];
  pagination: PaginationDto;
}

export interface CreateTopicInput {
  title: string;
  description: string;
}

export interface UpdateTopicInput {
  title: string;
  description: string;
}

export interface CreatePostInput {
  title: string;
  body: string;
}

export interface UpdatePostInput {
  title: string;
  body: string;
}

export interface CreateCommentInput {
  body: string;
}

export interface VoteInput {
  value: -1 | 1;
}

export interface DeleteResultDto {
  deleted: boolean;
}

export const mapUser = (dto: UserDto): User => ({
  id: dto.id,
  username: dto.username,
});

export const mapTopic = (dto: TopicDto): Topic => ({
  id: dto.id,
  title: dto.title,
  description: dto.description || "",
  createdBy: dto.created_by,
  createdAt: dto.created_at,
});

export const mapPost = (dto: PostDto): Post => ({
  id: dto.id,
  topicId: dto.topic_id,
  title: dto.title,
  body: dto.body,
  createdBy: dto.created_by,
  createdAt: dto.created_at,
  voteScore: dto.vote_score ?? 0,
  currentUserVote: dto.current_user_vote ?? null,
});

export const mapComment = (dto: CommentDto): Comment => ({
  id: dto.id,
  postId: dto.post_id,
  body: dto.body,
  createdBy: dto.created_by,
  createdAt: dto.created_at,
});

export const mapPagination = (dto: PaginationDto): Pagination => ({
  page: dto.page,
  pageSize: dto.page_size,
  totalItems: dto.total_items,
  totalPages: dto.total_pages,
  hasPrev: dto.has_prev,
  hasNext: dto.has_next,
});
