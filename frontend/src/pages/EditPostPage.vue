<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import NoticeBanner from "../components/NoticeBanner.vue";
import PageLayout from "../components/PageLayout.vue";
import { useAuth } from "../composables/useAuth";
import { useToast } from "../composables/useToast";
import { forumApi } from "../services/api/forumApi";

const route = useRoute();
const router = useRouter();
const auth = useAuth();
const { showToast } = useToast();

const parsedPostId = Number(route.params.postId);

const title = ref("");
const body = ref("");
const loading = ref(true);
const submitting = ref(false);
const error = ref("");
const ownerId = ref<number | null>(null);

const isOwner = computed(
  () => ownerId.value !== null && auth.state.currentUser?.id === ownerId.value,
);

onMounted(async () => {
  if (!parsedPostId) {
    error.value = "Invalid post ID.";
    loading.value = false;
    return;
  }

  try {
    const post = await forumApi.getPost(parsedPostId);
    title.value = post.title;
    body.value = post.body;
    ownerId.value = post.createdBy;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Failed to load post.";
  } finally {
    loading.value = false;
  }
});

const handleSubmit = async () => {
  if (!parsedPostId || !title.value.trim() || !body.value.trim()) {
    error.value = "Title and body are required.";
    return;
  }

  submitting.value = true;

  try {
    await forumApi.updatePost(parsedPostId, {
      title: title.value.trim(),
      body: body.value.trim(),
    });
    showToast("Changes saved", "success");
    await router.push(`/posts/${parsedPostId}`);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Failed to update post.";
  } finally {
    submitting.value = false;
  }
};
</script>

<template>
  <PageLayout title="Edit Post" subtitle="Revise the title and body for this post.">
    <NoticeBanner v-if="error" tone="error">{{ error }}</NoticeBanner>
    <NoticeBanner v-if="ownerId !== null && !isOwner" tone="error">
      You cannot edit a post you do not own.
    </NoticeBanner>

    <p v-if="loading" class="empty-state">Loading post...</p>

    <form v-else class="form-grid" @submit.prevent="handleSubmit">
      <div class="field">
        <label for="title">Title</label>
        <input id="title" v-model="title" :disabled="submitting || !isOwner" />
      </div>

      <div class="field">
        <label for="body">Body</label>
        <textarea id="body" v-model="body" rows="8" :disabled="submitting || !isOwner" />
      </div>

      <div class="form-actions">
        <button class="button button--primary" type="submit" :disabled="submitting || !isOwner">
          {{ submitting ? "Saving..." : "Save Changes" }}
        </button>
        <button
          class="button button--secondary"
          type="button"
          :disabled="submitting"
          @click="router.back()"
        >
          Cancel
        </button>
      </div>
    </form>
  </PageLayout>
</template>
