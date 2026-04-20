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
