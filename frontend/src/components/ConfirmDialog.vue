<script setup lang="ts">
withDefaults(
  defineProps<{
    open: boolean;
    title: string;
    message: string;
    confirmLabel?: string;
    cancelLabel?: string;
  }>(),
  {
    confirmLabel: "Confirm",
    cancelLabel: "Cancel",
  },
);

defineEmits<{
  confirm: [];
  cancel: [];
}>();
</script>

<template>
  <div v-if="open" class="confirm-dialog-root">
    <div class="confirm-dialog__overlay">
      <div
        class="confirm-dialog__card"
        role="dialog"
        aria-modal="true"
        :aria-label="title"
      >
        <h2 class="confirm-dialog__title">{{ title }}</h2>
        <p class="confirm-dialog__message">{{ message }}</p>
        <div class="confirm-dialog__actions">
          <button class="button button--secondary" type="button" @click="$emit('cancel')">
            {{ cancelLabel }}
          </button>
          <button class="button button--danger" type="button" @click="$emit('confirm')">
            {{ confirmLabel }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.confirm-dialog-root {
  position: relative;
  min-height: 0;
}

.confirm-dialog__overlay {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  padding: 16px;
  background: rgba(24, 33, 43, 0.14);
  border-radius: 16px;
}

.confirm-dialog__card {
  width: min(100%, 28rem);
  border: 1px solid var(--border);
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 16px 40px rgba(16, 24, 40, 0.12);
  padding: 20px;
}

.confirm-dialog__title {
  margin: 0 0 8px;
  color: var(--text);
  font-size: 1.25rem;
}

.confirm-dialog__message {
  margin: 0;
  color: var(--muted);
  line-height: 1.5;
}

.confirm-dialog__actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
  flex-wrap: wrap;
}
</style>
