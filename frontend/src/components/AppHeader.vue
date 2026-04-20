<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink, useRouter } from "vue-router";
import { useAuth } from "../composables/useAuth";

const auth = useAuth();
const router = useRouter();
const loggingOut = ref(false);

const showLogin = computed(
  () => !auth.state.isBootstrapping && !auth.isAuthenticated.value,
);

const showLogout = computed(
  () => !auth.state.isBootstrapping && auth.isAuthenticated.value,
);

const handleLogout = async () => {
  try {
    loggingOut.value = true;
    await auth.logout();
    await router.push("/login");
  } finally {
    loggingOut.value = false;
  }
};
</script>

<template>
  <header class="app-header">
    <div class="app-header__inner">
      <RouterLink class="app-header__brand" to="/companies">ArtHub</RouterLink>
      <div class="app-header__actions">
        <RouterLink v-if="showLogin" class="button button--secondary" to="/login">
          Log In
        </RouterLink>
        <button
          v-if="showLogout"
          class="button button--ghost"
          type="button"
          :disabled="loggingOut"
          @click="handleLogout"
        >
          {{ loggingOut ? "Logging out..." : "Log Out" }}
        </button>
      </div>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  width: 100%;
  border-bottom: 1px solid var(--border);
  background: #fff;
  box-shadow: 0 8px 24px rgba(16, 24, 40, 0.04);
}

.app-header__inner {
  width: 100%;
  max-width: 960px;
  margin: 0 auto;
  padding: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.app-header__brand {
  color: var(--text);
  text-decoration: none;
  font-size: 1.25rem;
  font-weight: 700;
}

.app-header__actions {
  display: flex;
  align-items: center;
  gap: 12px;
}
</style>
