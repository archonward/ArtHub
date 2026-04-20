import { ref, unref, watch, type Ref } from "vue";
import { forumApi } from "../services/api/forumApi";
import type { Company } from "../types";

export const useCompany = (companyId: Ref<number> | number) => {
  const company = ref<Company | null>(null);
  const loading = ref(false);
  const error = ref("");

  const fetchCompany = async () => {
    const resolvedCompanyId = unref(companyId);

    if (!resolvedCompanyId) {
      company.value = null;
      error.value = "Invalid company ID.";
      loading.value = false;
      return;
    }

    loading.value = true;
    error.value = "";

    try {
      company.value = await forumApi.getCompany(resolvedCompanyId);
    } catch (err) {
      company.value = null;
      error.value = err instanceof Error ? err.message : "Failed to load company.";
    } finally {
      loading.value = false;
    }
  };

  watch(() => unref(companyId), fetchCompany, { immediate: true });

  return {
    company,
    loading,
    error,
    fetchCompany,
  };
};
