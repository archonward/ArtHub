<script setup lang="ts">
import { ref } from "vue";
import { useAuth } from "../composables/useAuth";
import { forumApi } from "../services/api/forumApi";
import type { Post } from "../types";

const props = withDefaults(
  defineProps<{
    post: Post;
    compact?: boolean;
  }>(),
  {
    compact: false,
  },
);

const emit = defineEmits<{
  changed: [post: Post];
}>();

const auth = useAuth();
const pending = ref<null | -1 | 1 | 0>(null);
const error = ref("");

const handleVoteClick = async (value: -1 | 1) => {
  if (!auth.isAuthenticated.value) {
    return;
  }

  try {
    error.value = "";
    pending.value = value;

    const updatedPost =
      props.post.currentUserVote === value
        ? await forumApi.removePostVote(props.post.id)
        : await forumApi.voteOnPost(props.post.id, { value });

    emit("changed", updatedPost);
  } catch (err) {
    error.value =
      err instanceof Error ? err.message : "Failed to save vote.";
  } finally {
    pending.value = null;
  }
};
</script>

<template>
  <div :class="['vote-controls', { 'vote-controls--compact': compact }]">
    <div class="vote-controls__row">
      <button
        type="button"
        :class="['vote-button', { 'vote-button--active': post.currentUserVote === 1 }]"
        :disabled="!auth.isAuthenticated.value || pending !== null"
        aria-label="Upvote post"
        @click.stop="handleVoteClick(1)"
      >
        +
      </button>
      <span class="vote-score" aria-label="Vote score">{{ post.voteScore }}</span>
      <button
        type="button"
        :class="[
          'vote-button',
          'vote-button--down',
          { 'vote-button--active': post.currentUserVote === -1 },
        ]"
        :disabled="!auth.isAuthenticated.value || pending !== null"
        aria-label="Downvote post"
        @click.stop="handleVoteClick(-1)"
      >
        -
      </button>
    </div>
    <p v-if="error" class="vote-error">{{ error }}</p>
  </div>
</template>

<style scoped>
.vote-controls {
  display: inline-grid;
  gap: 6px;
  margin-bottom: 12px;
}

.vote-controls--compact {
  margin-bottom: 0;
}

.vote-controls__row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.vote-button {
  width: 36px;
  height: 36px;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: #fff;
  color: var(--muted);
  cursor: pointer;
  font-weight: 700;
}

.vote-button:disabled {
  cursor: not-allowed;
}

.vote-button--active {
  border-color: var(--primary);
  color: var(--primary);
  background: #edf4ff;
}

.vote-button--down.vote-button--active {
  border-color: var(--danger);
  color: var(--danger);
  background: #fff1f0;
}

.vote-score {
  min-width: 24px;
  text-align: center;
  font-weight: 700;
}

.vote-error {
  margin: 0;
  color: var(--danger);
  font-size: 0.9rem;
}
</style>
