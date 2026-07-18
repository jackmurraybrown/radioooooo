<script setup lang="ts">
// ⊹ ࣪ ˖ sidebar nav — logo + links
import { ref, onMounted } from 'vue'
import { useRoute, RouterLink, useRouter } from 'vue-router'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

async function logout() {
  await auth.logout()
  router.push('/login')
}

const logoUrl = ref<string | null>(null)
const stationName = ref('')

async function loadStation() {
  if (!auth.stationId) return
  try {
    const res = await api(`/stations/${auth.stationId}`).get()
    if (!res.ok) return
    const data = await res.json()
    logoUrl.value = data.logoUrl ?? null
    stationName.value = data.name ?? ''
  } catch { }
}

const links = [
  { name: 'schedule', label: 'schedule', to: '/schedule' },
  { name: 'media', label: 'media', to: '/media' },
  { name: 'playlists', label: 'playlists', to: '/playlists' },
  { name: 'settings', label: 'settings', to: '/settings' },
]

onMounted(loadStation)
</script>

<template>
  <nav class="sidebar">
    <RouterLink to="/settings" class="brand">
      <img v-if="logoUrl" :src="logoUrl" :alt="stationName" class="brand-logo" />
      <span v-else class="brand-name">{{ stationName || 'radiooo' }}</span>
    </RouterLink>

    <ul>
      <li v-for="link in links" :key="link.name">
        <RouterLink :to="link.to" :class="{ active: route.name === link.name }">
          {{ link.label }}
        </RouterLink>
      </li>
    </ul>

    <button class="logout-btn" @click="logout">logout</button>
  </nav>
</template>

<style scoped>
.sidebar {
  width: 180px;
  height: 100vh;
  display: flex;
  flex-direction: column;
  padding: 1.25rem 0;
  border-right: 1px solid var(--border);
  /* TODO: add --sidebar-background to design tokens */
  background: oklch(0.04 0 0);
  flex-shrink: 0;
}

.brand {
  display: flex;
  align-items: center;
  padding: 0 1rem 1.5rem;
  text-decoration: none;
  min-height: 40px;
  border-bottom: 1px solid var(--border);
  margin-bottom: 0.75rem;
}

.brand-logo {
  width: 32px;
  height: 32px;
  object-fit: cover;
}

.brand-name {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--foreground);
  letter-spacing: 0.05em;
  text-transform: lowercase;
}

ul {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
}

a {
  display: block;
  padding: 0.45rem 1rem;
  text-decoration: none;
  color: var(--muted-foreground);
  font-size: 0.8rem;
  border-left: 2px solid transparent;
  transition: color 0.1s;
}

a:hover {
  color: var(--foreground);
}

a.active {
  color: var(--foreground);
  border-left-color: var(--foreground);
}

.logout-btn {
  margin-top: auto;
  display: block;
  width: 100%;
  padding: 0.45rem 1rem;
  text-align: left;
  background: none;
  border: none;
  border-top: 1px solid var(--border);
  color: var(--muted-foreground);
  font-size: 0.8rem;
  cursor: pointer;
}

.logout-btn:hover {
  color: var(--foreground);
}
</style>
