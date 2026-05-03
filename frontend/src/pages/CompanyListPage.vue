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
const companyCountLabel = computed(() =>
  companies.value.length === 1 ? "1 listed company" : `${companies.value.length} listed companies`,
);

const formatDate = (value: string) => new Date(value).toLocaleDateString();

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

    <section v-if="!loading && companies.length > 0" class="company-overview">
      <div>
        <p class="company-overview__eyebrow">Market Directory</p>
        <h2 class="company-overview__title">{{ companyCountLabel }}</h2>
        <p class="company-overview__text">
          Discover active company pages and jump straight into each ticker's discussion stream.
        </p>
      </div>
      <div class="company-overview__badge">Public browse</div>
    </section>

    <SkeletonList v-if="loading" />
    <p v-else-if="companies.length === 0" class="empty-state">
      No companies yet. Add the first ticker to get started.
    </p>

    <ul v-else class="list company-list">
      <li
        v-for="company in companies"
        :key="company.id"
        class="list-item list-item--interactive company-card"
        @click="router.push(`/companies/${company.id}`)"
      >
        <div class="action-row action-row--spread company-card__layout">
          <div class="company-card__content">
            <div class="company-card__header">
              <span class="company-card__ticker-tag">{{ company.ticker }}</span>
              <span class="company-card__arrow" aria-hidden="true">-&gt;</span>
            </div>
            <h2 class="content-title company-card__title">{{ company.name }}</h2>
            <p class="content-body company-card__description">
              {{ company.description || "No company summary provided." }}
            </p>
            <div class="company-card__meta">
              <span class="company-card__meta-pill">Ticker {{ company.ticker }}</span>
              <span class="company-card__meta-pill">Added by user {{ company.createdBy }}</span>
              <span class="company-card__meta-pill">{{ formatDate(company.createdAt) }}</span>
            </div>
          </div>
          <button
            v-if="currentUserId === company.createdBy"
            class="button button--danger company-card__delete"
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

<style scoped>
.company-overview {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
  padding: 18px 20px;
  margin-bottom: 20px;
  border: 1px solid rgba(15, 92, 192, 0.12);
  border-radius: 14px;
  background:
    radial-gradient(circle at top left, rgba(15, 92, 192, 0.1), transparent 36%),
    linear-gradient(135deg, #f9fbff 0%, #eef5ff 100%);
}

.company-overview__eyebrow {
  margin: 0 0 6px;
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #0f5cc0;
}

.company-overview__title {
  margin: 0;
  font-size: 1.4rem;
}

.company-overview__text {
  margin: 8px 0 0;
  max-width: 44ch;
  color: var(--muted);
}

.company-overview__badge {
  flex-shrink: 0;
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.85);
  border: 1px solid rgba(15, 92, 192, 0.16);
  color: #0f5cc0;
  font-size: 0.88rem;
  font-weight: 700;
}

.company-list {
  gap: 16px;
}

.company-card {
  position: relative;
  overflow: hidden;
  padding: 20px;
  border-color: rgba(15, 92, 192, 0.12);
  background: #fff;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
  transition:
    transform 160ms ease,
    box-shadow 160ms ease,
    border-color 160ms ease;
}

.company-card::before {
  content: "";
  position: absolute;
  inset: 0 auto 0 0;
  width: 3px;
  background: linear-gradient(180deg, #0f5cc0 0%, #7ab3ff 100%);
}

.company-card:hover {
  transform: translateY(-2px);
  border-color: rgba(15, 92, 192, 0.28);
  box-shadow: 0 16px 32px rgba(15, 23, 42, 0.08);
}

.company-card__layout {
  align-items: flex-start;
}

.company-card__content {
  min-width: 0;
}

.company-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.company-card__ticker-tag {
  display: inline-flex;
  align-items: center;
  padding: 7px 11px;
  border-radius: 999px;
  background: rgba(15, 92, 192, 0.1);
  color: #0a458f;
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.company-card__arrow {
  color: #7b8a9a;
  font-size: 1rem;
  transition: transform 160ms ease, color 160ms ease;
}

.company-card:hover .company-card__arrow {
  color: #0f5cc0;
  transform: translateX(2px);
}

.company-card__title {
  margin-bottom: 6px;
  font-size: 1.2rem;
}

.company-card__description {
  margin-bottom: 16px;
  color: #314154;
}

.company-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.company-card__meta-pill {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 6px 10px;
  border-radius: 8px;
  background: #f7f9fb;
  border: 1px solid rgba(21, 35, 56, 0.08);
  color: var(--muted);
  font-size: 0.88rem;
}

.company-card__delete {
  align-self: center;
}

@media (max-width: 640px) {
  .company-overview {
    align-items: flex-start;
    padding: 18px;
  }

  .company-overview__badge {
    display: none;
  }

  .company-card {
    padding: 18px;
  }

  .company-card__delete {
    width: 100%;
  }
}
</style>
