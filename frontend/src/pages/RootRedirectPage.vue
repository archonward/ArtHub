<script setup lang="ts">
import { computed } from "vue";
import { useAuth } from "../composables/useAuth";
import PageLayout from "../components/PageLayout.vue";

const auth = useAuth();
const message = computed(() =>
  auth.state.isBootstrapping ? "Restoring your session..." : "Redirecting...",
);
</script>

<template>
  <PageLayout title="Loading" :subtitle="message">
    <p class="empty-state">Loading session...</p>
  </PageLayout>
</template>
