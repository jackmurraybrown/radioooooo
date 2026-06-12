<script setup lang="ts">
// ⊹ ࣪ ˖ sidebar nav — logo + links
import { ref, onMounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const auth = useAuthStore()

const logoUrl = ref<string | null>(null)
const stationName = ref('')

async function loadStation() {
  if (!auth.stationId) return
  try {
    const res = await api(`/stations/${auth.stationId}`).get()
    if (!res.ok) return
    const data = await res.json()
    logoUrl.value   = data.logoUrl ?? null
    stationName.value = data.name ?? ''
  } catch {}
}

const links = [
  { name: 'schedule',  label: 'schedule',  to: '/schedule' },
  { name: 'media',     label: 'media',     to: '/media' },
  { name: 'playlists', label: 'playlists', to: '/playlists' },
  { name: 'settings',  label: 'settings',  to: '/settings' },
]

onMounted(loadStation)
</script>

<template>
  <nav class="sidebar">
    <RouterLink to="/settings" class="sidebar-brand">
      <img
        v-if="logoUrl"
        :src="logoUrl"
        :alt="stationName"
        class="station-logo"
      />
      <span v-else class="station-name">{{ stationName || 'radiooo' }}</span>
    </RouterLink>

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

.sidebar-brand {
  display: flex;
  align-items: center;
  margin-bottom: 2rem;
  padding: 0 0.5rem;
  text-decoration: none;
  min-height: 40px;
}

.station-logo {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  object-fit: cover;
}

.station-name {
  font-weight: 600;
  font-size: 1.1rem;
  color: #111827;
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
