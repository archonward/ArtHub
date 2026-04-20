<script setup lang="ts">
import { ref } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
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
    await auth.signup(username.value.trim(), password.value);
    const redirectTo =
      typeof route.query.redirect === "string"
        ? route.query.redirect
        : "/companies";
    await router.replace(redirectTo);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Failed to sign up.";
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <PageLayout
    title="Create ArtHub Account"
    subtitle="Set up an account to create companies, posts, and comments."
    narrow
  >
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
          autocomplete="new-password"
        />
      </div>

      <button class="button button--primary button--full" type="submit" :disabled="loading">
        {{ loading ? "Creating account..." : "Sign Up" }}
      </button>
    </form>

    <p class="meta">
      Already have an account? <RouterLink to="/login">Log in</RouterLink>.
    </p>
  </PageLayout>
</template>
