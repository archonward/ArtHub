<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";
import NoticeBanner from "../components/NoticeBanner.vue";
import PageLayout from "../components/PageLayout.vue";
import { forumApi } from "../services/api/forumApi";

const router = useRouter();

const form = reactive({
  ticker: "",
  name: "",
  description: "",
});

const loading = ref(false);
const error = ref("");

const handleSubmit = async () => {
  error.value = "";

  if (!form.ticker.trim()) {
    error.value = "Ticker is required.";
    return;
  }

  if (!form.name.trim()) {
    error.value = "Company name is required.";
    return;
  }

  loading.value = true;

  try {
    const company = await forumApi.createCompany({
      ticker: form.ticker.trim().toUpperCase(),
      name: form.name.trim(),
      description: form.description.trim(),
    });
    await router.push(`/companies/${company.id}`);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Failed to create company.";
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <PageLayout title="Add Company" subtitle="Create a company page anchored to a ticker.">
    <NoticeBanner v-if="error" tone="error">{{ error }}</NoticeBanner>

    <form class="form-grid" @submit.prevent="handleSubmit">
      <div class="field">
        <label for="ticker">Ticker</label>
        <input
          id="ticker"
          v-model="form.ticker"
          :disabled="loading"
          @input="form.ticker = form.ticker.toUpperCase()"
        />
      </div>

      <div class="field">
        <label for="name">Company Name</label>
        <input id="name" v-model="form.name" :disabled="loading" />
      </div>

      <div class="field">
        <label for="description">Description</label>
        <textarea id="description" v-model="form.description" rows="5" :disabled="loading" />
      </div>

      <div class="form-actions">
        <button class="button button--primary" type="submit" :disabled="loading">
          {{ loading ? "Creating..." : "Create Company" }}
        </button>
        <button
          class="button button--secondary"
          type="button"
          :disabled="loading"
          @click="router.push('/companies')"
        >
          Cancel
        </button>
      </div>
    </form>
  </PageLayout>
</template>
