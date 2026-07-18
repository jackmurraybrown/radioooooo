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
  <nav class="w-[180px] h-screen flex flex-col py-5 border-r border-border bg-sidebar-background shrink-0">
    <RouterLink to="/settings" class="flex items-center px-4 pb-6 no-underline min-h-[40px] border-b border-border mb-3">
      <img v-if="logoUrl" :src="logoUrl" :alt="stationName" class="w-8 h-8 object-cover" />
      <span v-else class="text-[0.85rem] font-semibold text-foreground tracking-[0.05em] lowercase">{{ stationName || 'radiooo' }}</span>
    </RouterLink>

    <ul class="list-none m-0 p-0 flex flex-col">
      <li v-for="link in links" :key="link.name">
        <RouterLink
          :to="link.to"
          class="block px-4 py-[0.45rem] no-underline text-[0.8rem] border-l-2 border-transparent transition-colors duration-100"
          :class="route.name === link.name ? 'text-foreground border-l-foreground' : 'text-muted-foreground hover:text-foreground'"
        >
          {{ link.label }}
        </RouterLink>
      </li>
    </ul>

    <button
      class="mt-auto block w-full px-4 py-[0.45rem] text-left bg-transparent border-0 border-t border-border text-muted-foreground text-[0.8rem] cursor-pointer hover:text-foreground"
      @click="logout"
    >logout</button>
  </nav>
</template>
