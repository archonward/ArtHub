<script setup lang="ts">
withDefaults(
  defineProps<{
    tone?: "error" | "info";
  }>(),
  {
    tone: "info",
  },
);
</script>

<template>
  <div :class="['notice', `notice--${tone}`]">
    <slot />
  </div>
</template>

<style scoped>
.notice {
  padding: 12px 14px;
  border-radius: 12px;
  margin-bottom: 20px;
}

.notice--error {
  background: var(--danger-soft);
  color: var(--danger);
}

.notice--info {
  background: var(--info-soft);
  color: var(--primary-dark);
}
</style>
