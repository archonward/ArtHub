<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import NoticeBanner from "../components/NoticeBanner.vue";
import PageLayout from "../components/PageLayout.vue";
import { useToast } from "../composables/useToast";
import { forumApi } from "../services/api/forumApi";

const route = useRoute();
const router = useRouter();
const { showToast } = useToast();

const companyId = computed(() => Number(route.params.id));

const title = ref("");
const body = ref("");
const loading = ref(false);
const error = ref("");

const handleSubmit = async () => {
  if (!companyId.value) {
    error.value = "Company ID is missing.";
    return;
  }

  if (!title.value.trim() || !body.value.trim()) {
    error.value = "Title and body are required.";
    return;
  }

  loading.value = true;
  error.value = "";

  try {
    await forumApi.createPost(companyId.value, {
      title: title.value.trim(),
      body: body.value.trim(),
    });
    showToast("Post created", "success");
    await router.push(`/companies/${companyId.value}`);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Failed to create post.";
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <PageLayout title="Create Post" subtitle="Add a new thread for this company.">
    <NoticeBanner v-if="error" tone="error">{{ error }}</NoticeBanner>

    <form class="form-grid" @submit.prevent="handleSubmit">
      <div class="field">
        <label for="title">Title</label>
        <input id="title" v-model="title" :disabled="loading" />
      </div>

      <div class="field">
        <label for="body">Body</label>
        <textarea id="body" v-model="body" rows="8" :disabled="loading" />
      </div>

      <div class="form-actions">
        <button class="button button--primary" type="submit" :disabled="loading">
          {{ loading ? "Creating..." : "Create Post" }}
        </button>
        <button
          class="button button--secondary"
          type="button"
          :disabled="loading"
          @click="router.push(companyId ? `/companies/${companyId}` : '/companies')"
        >
          Cancel
        </button>
      </div>
    </form>
  </PageLayout>
</template>
