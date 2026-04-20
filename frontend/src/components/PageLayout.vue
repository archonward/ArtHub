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

<style scoped>
.page-shell {
  min-height: 100vh;
  padding: 32px 16px;
}

.card {
  width: 100%;
  max-width: 760px;
  margin: 0 auto;
  background: #fff;
  border: 1px solid var(--border);
  border-radius: 16px;
  box-shadow: 0 12px 30px rgba(16, 24, 40, 0.08);
  padding: 28px;
}

.card--narrow {
  max-width: 520px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  margin-bottom: 24px;
}

.page-header h1,
:slotted(.section-title) {
  margin: 0;
  font-size: 1.75rem;
}

.page-header p {
  color: var(--muted);
}

.page-header__actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

@media (max-width: 640px) {
  .card {
    padding: 20px;
  }

  .page-header {
    flex-direction: column;
  }
}
</style>
