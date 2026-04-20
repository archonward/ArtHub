<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import NoticeBanner from "../components/NoticeBanner.vue";
import PageLayout from "../components/PageLayout.vue";
import { useAuth } from "../composables/useAuth";
import { forumApi } from "../services/api/forumApi";
import type { Company } from "../types";

const router = useRouter();
const auth = useAuth();

const companies = ref<Company[]>([]);
const loading = ref(true);
const error = ref("");
const actionError = ref("");
const deletingCompanyId = ref<number | null>(null);
const loggingOut = ref(false);

const currentUserId = computed(() => auth.state.currentUser?.id ?? null);

const loadCompanies = async () => {
  try {
    companies.value = await forumApi.getCompanies();
  } catch (err) {
    error.value =
      err instanceof Error ? err.message : "Failed to load companies.";
  } finally {
    loading.value = false;
  }
};

const handleDelete = async (companyId: number) => {
  if (!window.confirm("Delete this company? This also removes its posts and comments.")) {
    return;
  }

  try {
    actionError.value = "";
    deletingCompanyId.value = companyId;
    await forumApi.deleteCompany(companyId);
    companies.value = companies.value.filter((company) => company.id !== companyId);
  } catch (err) {
    actionError.value =
      err instanceof Error ? err.message : "Failed to delete company.";
  } finally {
    deletingCompanyId.value = null;
  }
};

const handleLogout = async () => {
  try {
    loggingOut.value = true;
    actionError.value = "";
    await auth.logout();
    await router.push("/login");
  } catch (err) {
    actionError.value =
      err instanceof Error ? err.message : "Failed to log out.";
  } finally {
    loggingOut.value = false;
  }
};

onMounted(loadCompanies);
</script>

<template>
  <PageLayout
    title="Companies"
    subtitle="Browse companies and open the discussion stream for each ticker."
  >
    <template #actions>
      <div class="action-row">
        <button
          v-if="auth.isAuthenticated.value"
          class="button button--secondary"
          @click="router.push('/companies/new')"
        >
          New Company
        </button>
        <button
          v-if="!auth.state.isBootstrapping && auth.isAuthenticated.value"
          class="button button--ghost"
          :disabled="loggingOut"
          @click="handleLogout"
        >
          {{ loggingOut ? "Logging out..." : "Log Out" }}
        </button>
        <button
          v-else-if="!auth.state.isBootstrapping"
          class="button button--secondary"
          @click="router.push('/login')"
        >
          Log In
        </button>
      </div>
    </template>

    <NoticeBanner v-if="!auth.isAuthenticated.value && !auth.state.isBootstrapping">
      You can browse companies publicly. Log in to create or manage content.
    </NoticeBanner>
    <NoticeBanner v-if="error" tone="error">{{ error }}</NoticeBanner>
    <NoticeBanner v-if="actionError" tone="error">{{ actionError }}</NoticeBanner>
    <p v-if="loading" class="empty-state">Loading companies...</p>
    <p v-else-if="companies.length === 0" class="empty-state">
      No companies yet. Add the first ticker to get started.
    </p>

    <ul v-else class="list">
      <li
        v-for="company in companies"
        :key="company.id"
        class="list-item list-item--interactive"
        @click="router.push(`/companies/${company.id}`)"
      >
        <div class="action-row action-row--spread">
          <div>
            <h2 class="content-title">{{ company.ticker }}</h2>
            <p class="content-body">{{ company.name }}</p>
            <p class="meta">{{ company.description || "No company summary provided." }}</p>
            <p class="meta">
              Added by user {{ company.createdBy }} on
              {{ new Date(company.createdAt).toLocaleString() }}
            </p>
          </div>
          <button
            v-if="currentUserId === company.createdBy"
            class="button button--danger"
            :disabled="deletingCompanyId === company.id"
            @click.stop="handleDelete(company.id)"
          >
            {{ deletingCompanyId === company.id ? "Deleting..." : "Delete" }}
          </button>
        </div>
      </li>
    </ul>
  </PageLayout>
</template>
