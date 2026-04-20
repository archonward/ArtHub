<script setup lang="ts">
import { TransitionGroup } from "vue";
import { useToast } from "../composables/useToast";

const { toasts, dismissToast } = useToast();
</script>

<template>
  <div class="toast-container">
    <TransitionGroup name="toast" tag="div" class="toast-container__stack">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        :class="['toast-card', `toast-card--${toast.tone}`]"
      >
        <p class="toast-card__message">{{ toast.message }}</p>
        <button class="toast-card__dismiss" type="button" @click="dismissToast(toast.id)">
          Dismiss
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-container {
  position: fixed;
  right: 1.5rem;
  bottom: 1.5rem;
  z-index: 30;
}

.toast-container__stack {
  display: grid;
  gap: 0.75rem;
}

.toast-card {
  width: min(22rem, calc(100vw - 3rem));
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 0.9rem 1rem;
  box-shadow: 0 14px 30px rgba(16, 24, 40, 0.14);
}

.toast-card--success {
  background: var(--info-soft);
  color: var(--primary-dark);
}

.toast-card--error {
  background: var(--danger-soft);
  color: var(--danger);
}

.toast-card__message {
  margin: 0;
  line-height: 1.4;
}

.toast-card__dismiss {
  border: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  font-weight: 600;
}

.toast-enter-active,
.toast-leave-active {
  transition:
    opacity 180ms ease,
    transform 180ms ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(10px);
}
</style>
