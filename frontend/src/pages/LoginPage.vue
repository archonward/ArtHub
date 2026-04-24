<script setup lang="ts">
import { ref } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import arthubLogo from "../Arthub_logo.jpg";
import NoticeBanner from "../components/NoticeBanner.vue";
import PageLayout from "../components/PageLayout.vue";
import { useAuth } from "../composables/useAuth";

const auth = useAuth();
const router = useRouter();
const route = useRoute();

const username = ref("");
const password = ref("");
const loading = ref(false);
const error = ref("");

const handleSubmit = async () => {
  if (!username.value.trim()) {
    error.value = "Username is required.";
    return;
  }

  if (!password.value) {
    error.value = "Password is required.";
    return;
  }

  loading.value = true;
  error.value = "";

  try {
    await auth.login(username.value.trim(), password.value);
    const redirectTo =
      typeof route.query.redirect === "string"
        ? route.query.redirect
        : "/companies";
    await router.replace(redirectTo);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Login failed.";
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <div class="login-page">
    <PageLayout
      title="ArtHub"
      subtitle="Log in to manage companies, posts, and comments."
      narrow
    >
      <img class="login-logo" :src="arthubLogo" alt="ArtHub logo" />

      <NoticeBanner v-if="auth.state.authNotice">
        {{ auth.state.authNotice }}
      </NoticeBanner>
      <NoticeBanner v-if="error" tone="error">
        {{ error }}
      </NoticeBanner>

      <form class="form-grid" @submit.prevent="handleSubmit">
        <div class="field">
          <label for="username">Username</label>
          <input
            id="username"
            v-model="username"
            type="text"
            :disabled="loading"
            autocomplete="username"
          />
        </div>

        <div class="field">
          <label for="password">Password</label>
          <input
            id="password"
            v-model="password"
            type="password"
            :disabled="loading"
            autocomplete="current-password"
          />
        </div>

        <button class="button button--primary button--full" type="submit" :disabled="loading">
          {{ loading ? "Logging in..." : "Log In" }}
        </button>
      </form>

      <p class="meta">
        Need an account? <RouterLink to="/signup">Sign up</RouterLink>.
      </p>
    </PageLayout>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #edf6ef 0%, #f7faf7 100%);
}

.login-page :deep(.card) {
  background: rgba(255, 255, 255, 0.96);
  border-color: #d7e7d9;
}

.login-logo {
  display: block;
  width: 108px;
  max-width: 100%;
  margin: 0 auto 20px;
}
</style>
