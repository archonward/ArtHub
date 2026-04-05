import type {
  Comment,
  CommentDto,
  CreateCommentInput,
  CreatePostInput,
  CreateTopicInput,
  Post,
  PostDto,
  Topic,
  TopicDetails,
  TopicDto,
  UpdatePostInput,
  UpdateTopicInput,
  User,
  UserDto,
} from "../../types";
import { mapComment, mapPost, mapTopic, mapUser } from "../../types";
import { request } from "./client";

export const forumApi = {
  login: async (username: string): Promise<User> => {
    const user = await request<UserDto>("/login", {
      method: "POST",
      body: { username },
    });
    return mapUser(user);
  },

  getTopics: async (): Promise<Topic[]> => {
    const topics = await request<TopicDto[]>("/topics");
    return topics.map(mapTopic);
  },

  getTopic: async (id: number): Promise<Topic> => {
    const topic = await request<TopicDto>(`/topics/${id}`);
    return mapTopic(topic);
  },

  getTopicPosts: async (id: number): Promise<Post[]> => {
    const posts = await request<PostDto[]>(`/topics/${id}/posts`);
    return posts.map(mapPost);
  },

  getTopicDetails: async (id: number): Promise<TopicDetails> => {
    const [topic, posts] = await Promise.all([
      forumApi.getTopic(id),
      forumApi.getTopicPosts(id),
    ]);

    return { topic, posts };
  },

  createTopic: async (input: CreateTopicInput): Promise<Topic> => {
    const topic = await request<TopicDto>("/topics", {
      method: "POST",
      body: {
        title: input.title,
        description: input.description,
        created_by: input.createdBy,
      },
    });
    return mapTopic(topic);
  },

  updateTopic: async (id: number, input: UpdateTopicInput): Promise<Topic> => {
    const topic = await request<TopicDto>(`/topics/${id}`, {
      method: "PUT",
      body: input,
    });
    return mapTopic(topic);
  },

  deleteTopic: (id: number): Promise<void> =>
    request<void>(`/topics/${id}`, { method: "DELETE" }),

  getPost: async (id: number): Promise<Post> => {
    const post = await request<PostDto>(`/posts/${id}`);
    return mapPost(post);
  },

  createPost: async (topicId: number, input: CreatePostInput): Promise<Post> => {
    const post = await request<PostDto>(`/topics/${topicId}/posts`, {
      method: "POST",
      body: {
        title: input.title,
        body: input.body,
        created_by: input.createdBy,
      },
    });
    return mapPost(post);
  },

  updatePost: async (id: number, input: UpdatePostInput): Promise<Post> => {
    const post = await request<PostDto>(`/posts/${id}`, {
      method: "PUT",
      body: input,
    });
    return mapPost(post);
  },

  deletePost: (id: number): Promise<void> =>
    request<void>(`/posts/${id}`, { method: "DELETE" }),

  getPostComments: async (id: number): Promise<Comment[]> => {
    const comments = await request<CommentDto[]>(`/posts/${id}/comments`);
    return comments.map(mapComment);
  },

  createComment: async (
    postId: number,
    input: CreateCommentInput,
  ): Promise<Comment> => {
    const comment = await request<CommentDto>(`/posts/${postId}/comments`, {
      method: "POST",
      body: {
        body: input.body,
        created_by: input.createdBy,
      },
    });
    return mapComment(comment);
  },
};
