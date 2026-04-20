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
