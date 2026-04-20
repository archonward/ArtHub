<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import NoticeBanner from "../components/NoticeBanner.vue";
import PageLayout from "../components/PageLayout.vue";
import PostVoteControls from "../components/PostVoteControls.vue";
import { useAuth } from "../composables/useAuth";
import { forumApi } from "../services/api/forumApi";
import type { Comment, Post } from "../types";

const route = useRoute();
const router = useRouter();
const auth = useAuth();

const parsedPostId = computed(() => Number(route.params.postId));

const post = ref<Post | null>(null);
const comments = ref<Comment[]>([]);
const newComment = ref("");
const loading = ref(true);
const error = ref("");
const submitting = ref(false);
const deleting = ref(false);
const actionError = ref("");

const isOwner = computed(() => auth.state.currentUser?.id === post.value?.createdBy);

onMounted(async () => {
  if (!parsedPostId.value) {
    error.value = "Invalid post ID.";
    loading.value = false;
    return;
  }

  try {
    const [postRecord, commentRecords] = await Promise.all([
      forumApi.getPost(parsedPostId.value),
      forumApi.getPostComments(parsedPostId.value),
    ]);
    post.value = postRecord;
    comments.value = commentRecords;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Failed to load post.";
  } finally {
    loading.value = false;
  }
});

const handleDelete = async () => {
  if (!post.value || !window.confirm("Delete this post and all its comments?")) {
    return;
  }

  try {
    actionError.value = "";
    deleting.value = true;
    await forumApi.deletePost(post.value.id);
    await router.push(`/companies/${post.value.companyId}`);
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : "Failed to delete post.";
  } finally {
    deleting.value = false;
  }
};

const handleAddComment = async () => {
  if (!post.value || !auth.state.currentUser || !newComment.value.trim()) {
    return;
  }

  submitting.value = true;

  try {
    actionError.value = "";
    const comment = await forumApi.createComment(post.value.id, {
      body: newComment.value.trim(),
    });
    comments.value = [...comments.value, comment];
    newComment.value = "";
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : "Failed to post comment.";
  } finally {
    submitting.value = false;
  }
};

const updatePost = (updatedPost: Post) => {
  post.value = updatedPost;
};
</script>

<template>
  <PageLayout
    v-if="!loading && post"
    :title="post.title"
    :subtitle="`Created by user ${post.createdBy} on ${new Date(post.createdAt).toLocaleString()}`"
  >
    <template #actions>
      <div class="action-row">
        <button class="button button--secondary" @click="router.back()">Back</button>
        <button
          v-if="isOwner"
          class="button button--ghost"
          :disabled="deleting"
          @click="router.push(`/posts/${post.id}/edit`)"
        >
          Edit
        </button>
        <button
          v-if="isOwner"
          class="button button--danger"
          :disabled="deleting"
          @click="handleDelete"
        >
          {{ deleting ? "Deleting..." : "Delete" }}
        </button>
      </div>
    </template>

    <NoticeBanner v-if="actionError" tone="error">{{ actionError }}</NoticeBanner>
    <NoticeBanner v-if="!auth.isAuthenticated.value">
      Log in to comment on this post.
    </NoticeBanner>
    <NoticeBanner v-if="auth.state.currentUser && !isOwner">
      Only the post owner can edit or delete this post.
    </NoticeBanner>

    <PostVoteControls :post="post" @changed="updatePost" />
    <p class="content-body">{{ post.body }}</p>

    <hr class="divider" />

    <div class="stack">
      <h2 class="section-title">Comments ({{ comments.length }})</h2>

      <p v-if="comments.length === 0" class="empty-state">No comments yet.</p>
      <ul v-else class="list">
        <li v-for="comment in comments" :key="comment.id" class="list-item">
          <p class="content-body">{{ comment.body }}</p>
          <p class="meta">
            By user {{ comment.createdBy }} on {{ new Date(comment.createdAt).toLocaleString() }}
          </p>
        </li>
      </ul>

      <form v-if="auth.isAuthenticated.value" class="form-grid" @submit.prevent="handleAddComment">
        <div class="field">
          <label for="comment">Add a comment</label>
          <textarea id="comment" v-model="newComment" rows="4" :disabled="submitting" />
        </div>
        <button
          class="button button--primary"
          type="submit"
          :disabled="submitting || !newComment.trim()"
        >
          {{ submitting ? "Posting..." : "Post Comment" }}
        </button>
      </form>
    </div>
  </PageLayout>

  <PageLayout v-else-if="loading" title="Post" subtitle="Loading post details...">
    <p class="empty-state">Loading post...</p>
  </PageLayout>

  <PageLayout v-else title="Post" subtitle="Discussion unavailable.">
    <NoticeBanner tone="error">{{ error || "Post not found." }}</NoticeBanner>
  </PageLayout>
</template>
