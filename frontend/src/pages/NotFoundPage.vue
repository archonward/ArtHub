<script setup lang="ts">
import { useRouter } from "vue-router";
import PageLayout from "../components/PageLayout.vue";

const router = useRouter();
</script>

<template>
  <PageLayout title="Page Not Found" subtitle="The page you requested does not exist.">
    <p class="empty-state">
      The address you entered does not match any page in ArtHub.
    </p>
    <div class="form-actions">
      <button class="button button--primary" type="button" @click="router.push('/companies')">
        Back to Companies
      </button>
    </div>
  </PageLayout>
</template>
