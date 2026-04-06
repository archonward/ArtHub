import type {
  Comment,
  CommentDto,
  Company,
  CompanyDetails,
  CompanyDto,
  CompanyPostsPageDto,
  CreateCommentInput,
  CreateCompanyInput,
  CreatePostInput,
  DeleteResultDto,
  Post,
  PostDto,
  PostSort,
  UpdateCompanyInput,
  UpdatePostInput,
  User,
  UserDto,
  VoteInput,
} from "../../types";
import {
  mapComment,
  mapCompany,
  mapPagination,
  mapPost,
  mapUser,
} from "../../types";
import { request } from "./client";

export const forumApi = {
  signup: async (username: string, password: string): Promise<User> => {
    const user = await request<UserDto>("/auth/signup", {
      method: "POST",
      body: { username, password },
      notifyOnUnauthorized: false,
    });
    return mapUser(user);
  },

  login: async (username: string, password: string): Promise<User> => {
    const user = await request<UserDto>("/auth/login", {
      method: "POST",
      body: { username, password },
      notifyOnUnauthorized: false,
    });
    return mapUser(user);
  },

  logout: (): Promise<{ logged_out: boolean }> =>
    request<{ logged_out: boolean }>("/auth/logout", {
      method: "POST",
    }),

  getCurrentUser: async (): Promise<User> => {
    const user = await request<UserDto>("/auth/me", {
      notifyOnUnauthorized: false,
    });
    return mapUser(user);
  },

  getCompanies: async (): Promise<Company[]> => {
    const companies = await request<CompanyDto[]>("/companies");
    return companies.map(mapCompany);
  },

  getCompany: async (id: number): Promise<Company> => {
    const company = await request<CompanyDto>(`/companies/${id}`);
    return mapCompany(company);
  },

  getCompanyPosts: async (
    id: number,
    sort: PostSort = "top",
    page = 1,
    pageSize = 10,
  ): Promise<{ posts: Post[]; pagination: CompanyDetails["pagination"] }> => {
    const companyPostsPage = await request<CompanyPostsPageDto>(
      `/companies/${id}/posts?sort=${encodeURIComponent(sort)}&page=${page}&pageSize=${pageSize}`,
    );
    return {
      posts: companyPostsPage.posts.map(mapPost),
      pagination: mapPagination(companyPostsPage.pagination),
    };
  },

  getCompanyDetails: async (
    id: number,
    sort: PostSort = "top",
    page = 1,
    pageSize = 10,
  ): Promise<CompanyDetails> => {
    const [company, companyPostsPage] = await Promise.all([
      forumApi.getCompany(id),
      forumApi.getCompanyPosts(id, sort, page, pageSize),
    ]);

    return {
      company,
      posts: companyPostsPage.posts,
      pagination: companyPostsPage.pagination,
    };
  },

  createCompany: async (input: CreateCompanyInput): Promise<Company> => {
    const company = await request<CompanyDto>("/companies", {
      method: "POST",
      body: {
        ticker: input.ticker,
        name: input.name,
        description: input.description,
      },
    });
    return mapCompany(company);
  },

  updateCompany: async (
    id: number,
    input: UpdateCompanyInput,
  ): Promise<Company> => {
    const company = await request<CompanyDto>(`/companies/${id}`, {
      method: "PUT",
      body: input,
    });
    return mapCompany(company);
  },

  deleteCompany: (id: number): Promise<void> =>
    request<DeleteResultDto>(`/companies/${id}`, { method: "DELETE" }).then(
      () => undefined,
    ),

  getPost: async (id: number): Promise<Post> => {
    const post = await request<PostDto>(`/posts/${id}`);
    return mapPost(post);
  },

  createPost: async (
    companyId: number,
    input: CreatePostInput,
  ): Promise<Post> => {
    const post = await request<PostDto>(`/companies/${companyId}/posts`, {
      method: "POST",
      body: {
        title: input.title,
        body: input.body,
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
    request<DeleteResultDto>(`/posts/${id}`, { method: "DELETE" }).then(
      () => undefined,
    ),

  voteOnPost: async (id: number, input: VoteInput): Promise<Post> => {
    const post = await request<PostDto>(`/posts/${id}/vote`, {
      method: "POST",
      body: input,
    });
    return mapPost(post);
  },

  removePostVote: async (id: number): Promise<Post> => {
    const post = await request<PostDto>(`/posts/${id}/vote`, {
      method: "DELETE",
    });
    return mapPost(post);
  },

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
      },
    });
    return mapComment(comment);
  },
};
