<script setup lang="ts">
// ✮ ⋆ ˚｡𖦹 authenticated — sidebar + page content
import AppSidebar from './AppSidebar.vue'
import ToastContainer from './ToastContainer.vue'
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-background">
    <AppSidebar />
    <main class="flex-1 overflow-y-auto">
      <RouterView />
    </main>
    <ToastContainer />
  </div>
</template>
