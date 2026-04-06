import type { Comment, Post, Topic, User } from "./models";

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
}

export interface CommentDto {
  id: number;
  post_id: number;
  body: string;
  created_by: number;
  created_at: string;
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
});

export const mapComment = (dto: CommentDto): Comment => ({
  id: dto.id,
  postId: dto.post_id,
  body: dto.body,
  createdBy: dto.created_by,
  createdAt: dto.created_at,
});
