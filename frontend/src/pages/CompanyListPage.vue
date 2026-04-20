<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import ConfirmDialog from "../components/ConfirmDialog.vue";
import NoticeBanner from "../components/NoticeBanner.vue";
import PageLayout from "../components/PageLayout.vue";
import SkeletonList from "../components/SkeletonList.vue";
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
const dialogOpen = ref(false);
const pendingCompanyId = ref<number | null>(null);

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

const openDeleteDialog = (companyId: number) => {
  pendingCompanyId.value = companyId;
  dialogOpen.value = true;
};

const closeDeleteDialog = () => {
  dialogOpen.value = false;
  pendingCompanyId.value = null;
};

const confirmDelete = async () => {
  if (!pendingCompanyId.value) {
    return;
  }

  try {
    actionError.value = "";
    deletingCompanyId.value = pendingCompanyId.value;
    await forumApi.deleteCompany(pendingCompanyId.value);
    companies.value = companies.value.filter(
      (company) => company.id !== pendingCompanyId.value,
    );
    closeDeleteDialog();
  } catch (err) {
    actionError.value =
      err instanceof Error ? err.message : "Failed to delete company.";
  } finally {
    deletingCompanyId.value = null;
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
      </div>
    </template>

    <NoticeBanner v-if="!auth.isAuthenticated.value && !auth.state.isBootstrapping">
      You can browse companies publicly. Log in to create or manage content.
    </NoticeBanner>
    <NoticeBanner v-if="error" tone="error">{{ error }}</NoticeBanner>
    <NoticeBanner v-if="actionError" tone="error">{{ actionError }}</NoticeBanner>
    <SkeletonList v-if="loading" />
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
            @click.stop="openDeleteDialog(company.id)"
          >
            {{ deletingCompanyId === company.id ? "Deleting..." : "Delete" }}
          </button>
        </div>
      </li>
    </ul>

    <ConfirmDialog
      :open="dialogOpen"
      title="Delete Company"
      message="Delete this company? This also removes its posts and comments."
      confirm-label="Delete Company"
      @cancel="closeDeleteDialog"
      @confirm="confirmDelete"
    />
  </PageLayout>
</template>
