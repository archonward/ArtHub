<script setup lang="ts">
defineProps<{
  title: string;
  subtitle?: string;
  narrow?: boolean;
}>();
</script>

<template>
  <main class="page-shell">
    <section :class="['card', { 'card--narrow': narrow }]">
      <header class="page-header">
        <div>
          <h1>{{ title }}</h1>
          <p v-if="subtitle">{{ subtitle }}</p>
        </div>
        <div v-if="$slots.actions" class="page-header__actions">
          <slot name="actions" />
        </div>
      </header>
      <slot />
    </section>
  </main>
</template>
