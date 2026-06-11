<script setup lang="ts">
// ✮ ⋆ ˚｡𖦹 authenticated — sidebar + page content
import AppSidebar from './AppSidebar.vue'
</script>

<template>
  <div class="layout">
    <AppSidebar />
    <main class="content">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

.content {
  flex: 1;
  overflow-y: auto;
  padding: 2rem;
}
</style>
