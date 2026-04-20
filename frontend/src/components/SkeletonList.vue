<script setup lang="ts">
withDefaults(
  defineProps<{
    rows?: number;
  }>(),
  {
    rows: 4,
  },
);
</script>

<template>
  <ul class="skeleton-list" aria-hidden="true">
    <li v-for="row in rows" :key="row" class="skeleton-list__item">
      <div class="skeleton-list__title"></div>
      <div class="skeleton-list__meta"></div>
    </li>
  </ul>
</template>

<style scoped>
.skeleton-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 12px;
}

.skeleton-list__item {
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 16px;
  background: #fbfcfd;
}

.skeleton-list__title,
.skeleton-list__meta {
  border-radius: 999px;
  animation: pulse 1.4s ease-in-out infinite;
}

.skeleton-list__title {
  width: 60%;
  height: 16px;
  background: var(--border);
}

.skeleton-list__meta {
  width: 36%;
  height: 12px;
  background: #eef2f6;
  margin-top: 12px;
}

@keyframes pulse {
  0%,
  100% {
    opacity: 0.65;
  }

  50% {
    opacity: 1;
  }
}
</style>
