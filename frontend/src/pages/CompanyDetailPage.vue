<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import NoticeBanner from "../components/NoticeBanner.vue";
import PageLayout from "../components/PageLayout.vue";
import PostVoteControls from "../components/PostVoteControls.vue";
import { useAuth } from "../composables/useAuth";
import { forumApi } from "../services/api/forumApi";
import type { Company, Pagination, Post, PostSort } from "../types";

const COMPANY_POSTS_PAGE_SIZE = 10;

const route = useRoute();
const router = useRouter();
const auth = useAuth();

const company = ref<Company | null>(null);
const posts = ref<Post[]>([]);
const sort = ref<PostSort>("top");
const page = ref(1);
const pagination = ref<Pagination | null>(null);
const loading = ref(true);
const error = ref("");

const companyId = computed(() => Number(route.params.id));

const handlePostChange = (updatedPost: Post) => {
  posts.value = posts.value.map((post) => (post.id === updatedPost.id ? updatedPost : post));
};

const fetchCompany = async () => {
  if (!companyId.value) {
    error.value = "Invalid company ID.";
    loading.value = false;
    return;
  }

  loading.value = true;
  error.value = "";

  try {
    const details = await forumApi.getCompanyDetails(
      companyId.value,
      sort.value,
      page.value,
      COMPANY_POSTS_PAGE_SIZE,
    );
    company.value = details.company;
    posts.value = details.posts;
    pagination.value = details.pagination;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Failed to load company.";
  } finally {
    loading.value = false;
  }
};

watch([companyId, sort, page], fetchCompany, { immediate: true });
</script>

<template>
  <PageLayout
    v-if="!loading && company"
    :title="company.ticker"
    :subtitle="company.name"
  >
    <template #actions>
      <div class="action-row">
        <button class="button button--secondary" @click="router.push('/companies')">
          Back to Companies
        </button>
        <button
          v-if="auth.isAuthenticated.value"
          class="button button--secondary"
          @click="router.push(`/companies/${company.id}/posts/new`)"
        >
          New Post
        </button>
        <button
          v-if="auth.state.currentUser?.id === company.createdBy"
          class="button button--ghost"
          @click="router.push(`/companies/${company.id}/edit`)"
        >
          Edit Company
        </button>
      </div>
    </template>

    <p class="meta">{{ company.description || "No company summary provided." }}</p>
    <NoticeBanner v-if="!auth.isAuthenticated.value">
      Log in if you want to create a post for this company.
    </NoticeBanner>
    <NoticeBanner
      v-if="auth.state.currentUser && auth.state.currentUser.id !== company.createdBy"
    >
      Only the company owner can edit this company profile.
    </NoticeBanner>
    <p class="meta">
      Added by user {{ company.createdBy }} on
      {{ new Date(company.createdAt).toLocaleString() }}
    </p>

    <div class="stack">
      <div class="section-header">
        <h2 class="section-title">Posts ({{ posts.length }})</h2>
        <label class="sort-control" for="post-sort">
          <span>Sort</span>
          <select id="post-sort" v-model="sort" @change="page = 1">
            <option value="top">Top</option>
            <option value="new">New</option>
          </select>
        </label>
      </div>

      <p v-if="posts.length === 0" class="empty-state">
        No posts yet. Create the first post for this company.
      </p>
      <ul v-else class="list">
        <li
          v-for="post in posts"
          :key="post.id"
          class="list-item list-item--interactive"
          @click="router.push(`/posts/${post.id}`)"
        >
          <div class="list-item__header">
            <h3>{{ post.title }}</h3>
            <PostVoteControls :post="post" compact @changed="handlePostChange" />
          </div>
          <p class="content-body">{{ post.body }}</p>
          <p class="meta">
            Created by user {{ post.createdBy }} on
            {{ new Date(post.createdAt).toLocaleString() }}
          </p>
        </li>
      </ul>

      <div v-if="pagination && pagination.totalPages > 0" class="pagination">
        <button
          class="button button--secondary"
          type="button"
          :disabled="!pagination.hasPrev"
          @click="page = Math.max(1, page - 1)"
        >
          Previous
        </button>
        <p class="meta pagination__label">
          Page {{ pagination.page }} of {{ pagination.totalPages }}
        </p>
        <button
          class="button button--secondary"
          type="button"
          :disabled="!pagination.hasNext"
          @click="page = pagination.totalPages > 0 ? Math.min(pagination.totalPages, page + 1) : page"
        >
          Next
        </button>
      </div>
    </div>
  </PageLayout>

  <PageLayout v-else-if="loading" title="Company" subtitle="Loading company discussion...">
    <p class="empty-state">Loading company...</p>
  </PageLayout>

  <PageLayout v-else title="Company" subtitle="Discussion area unavailable.">
    <NoticeBanner tone="error">{{ error || "Company not found." }}</NoticeBanner>
  </PageLayout>
</template>
