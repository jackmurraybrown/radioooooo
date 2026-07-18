<script setup lang="ts">
// ✮ ⋆ ˚｡𖦹 authenticated — sidebar + page content
import AppSidebar from './AppSidebar.vue'
import ToastContainer from './ToastContainer.vue'
</script>

<template>
  <div class="layout">
    <AppSidebar />
    <main class="content">
      <RouterView />
    </main>
    <ToastContainer />
  </div>
</template>

<style scoped>
.layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: var(--background);
}

.content {
  flex: 1;
  overflow-y: auto;
}
</style>
