<script setup lang="ts">
// ⊹ ࣪ ˖ sidebar nav
import { useRoute, RouterLink } from 'vue-router'

const route = useRoute()

const links = [
  { name: 'schedule',  label: 'schedule',  to: '/schedule' },
  { name: 'media',     label: 'media',     to: '/media' },
  { name: 'playlists', label: 'playlists', to: '/playlists' },
  { name: 'settings',  label: 'settings',  to: '/settings' },
]
</script>

<template>
  <nav class="sidebar">
    <div class="sidebar-logo">radiooo</div>
    <ul>
      <li v-for="link in links" :key="link.name">
        <RouterLink
          :to="link.to"
          :class="{ active: route.name === link.name }"
        >
          {{ link.label }}
        </RouterLink>
      </li>
    </ul>
  </nav>
</template>

<style scoped>
.sidebar {
  width: 200px;
  height: 100vh;
  display: flex;
  flex-direction: column;
  padding: 1.5rem 1rem;
  border-right: 1px solid #e5e7eb;
  flex-shrink: 0;
}

.sidebar-logo {
  font-weight: 600;
  font-size: 1.1rem;
  margin-bottom: 2rem;
  padding: 0 0.5rem;
}

ul {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

a {
  display: block;
  padding: 0.5rem;
  border-radius: 6px;
  text-decoration: none;
  color: #6b7280;
  font-size: 0.9rem;
}

a:hover {
  background: #f3f4f6;
  color: #111827;
}

a.active {
  background: #f3f4f6;
  color: #111827;
  font-weight: 500;
}
</style>
