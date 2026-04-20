<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import NoticeBanner from "../components/NoticeBanner.vue";
import PageLayout from "../components/PageLayout.vue";
import { useAuth } from "../composables/useAuth";
import { forumApi } from "../services/api/forumApi";

const route = useRoute();
const router = useRouter();
const auth = useAuth();

const companyId = Number(route.params.id);

const ticker = ref("");
const name = ref("");
const description = ref("");
const loading = ref(true);
const submitting = ref(false);
const error = ref("");
const ownerId = ref<number | null>(null);

const isOwner = computed(
  () => ownerId.value !== null && auth.state.currentUser?.id === ownerId.value,
);

onMounted(async () => {
  if (!companyId) {
    error.value = "Invalid company ID.";
    loading.value = false;
    return;
  }

  try {
    const company = await forumApi.getCompany(companyId);
    ticker.value = company.ticker;
    name.value = company.name;
    description.value = company.description;
    ownerId.value = company.createdBy;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Failed to load company.";
  } finally {
    loading.value = false;
  }
});

const handleSubmit = async () => {
  if (!companyId || !ticker.value.trim() || !name.value.trim()) {
    error.value = "Ticker and company name are required.";
    return;
  }

  submitting.value = true;

  try {
    await forumApi.updateCompany(companyId, {
      ticker: ticker.value.trim().toUpperCase(),
      name: name.value.trim(),
      description: description.value.trim(),
    });
    await router.push(`/companies/${companyId}`);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Failed to update company.";
  } finally {
    submitting.value = false;
  }
};
</script>

<template>
  <PageLayout title="Edit Company" subtitle="Update the ticker, name, or summary.">
    <NoticeBanner v-if="error" tone="error">{{ error }}</NoticeBanner>
    <NoticeBanner v-if="ownerId !== null && !isOwner" tone="error">
      You cannot edit a company you do not own.
    </NoticeBanner>

    <p v-if="loading" class="empty-state">Loading company...</p>

    <form v-else class="form-grid" @submit.prevent="handleSubmit">
      <div class="field">
        <label for="ticker">Ticker</label>
        <input
          id="ticker"
          v-model="ticker"
          :disabled="submitting || !isOwner"
          @input="ticker = ticker.toUpperCase()"
        />
      </div>

      <div class="field">
        <label for="name">Company Name</label>
        <input id="name" v-model="name" :disabled="submitting || !isOwner" />
      </div>

      <div class="field">
        <label for="description">Description</label>
        <textarea
          id="description"
          v-model="description"
          rows="5"
          :disabled="submitting || !isOwner"
        />
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
