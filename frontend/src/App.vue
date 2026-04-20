<template>
  <AppHeader />
  <RouterView />
  <ToastContainer />
</template>

<script setup lang="ts">
import { RouterView } from "vue-router";
import AppHeader from "./components/AppHeader.vue";
import ToastContainer from "./components/ToastContainer.vue";
</script>
