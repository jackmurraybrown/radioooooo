// ⋆˙⟡ ⋆.˚ playlists — list + active playlist items
import { ref } from 'vue'
import { api } from '@/api/client'
import type { Playlist, PlaylistItem, PlaylistCreateBody, PlaylistUpdateBody } from '@/api/types'

export function usePlaylists() {
  const playlists = ref<Playlist[]>([])
  const activeItems = ref<PlaylistItem[]>([])
  const loading = ref(false)
  const itemsLoading = ref(false)
  const error = ref<string | null>(null)

  async function fetchPlaylists() {
    loading.value = true
    error.value = null
    try {
      const res = await api('/playlists').get()
      if (!res.ok) throw new Error(`${res.status}`)
      const data = await res.json()
      playlists.value = data.playlists ?? []
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'failed to fetch playlists'
    } finally {
      loading.value = false
    }
  }

  async function fetchItems(id: string) {
    itemsLoading.value = true
    try {
      const res = await api(`/playlists/${id}`).get()
      if (!res.ok) throw new Error(`${res.status}`)
      const data = await res.json()
      activeItems.value = data.items ?? []
    } finally {
      itemsLoading.value = false
    }
  }

  async function createPlaylist(body: Omit<PlaylistCreateBody, '$schema'>) {
    const res = await api('/playlists').post(body)
    if (!res.ok) throw new Error(`${res.status}`)
    const pl: Playlist = await res.json()
    playlists.value.unshift(pl)
    return pl
  }

  async function updatePlaylist(id: string, body: Omit<PlaylistUpdateBody, '$schema'>) {
    const res = await api(`/playlists/${id}`).put(body)
    if (!res.ok) throw new Error(`${res.status}`)
    const pl: Playlist = await res.json()
    const i = playlists.value.findIndex(p => p.id === id)
    if (i !== -1) playlists.value[i] = pl
  }

  async function deletePlaylist(id: string) {
    const res = await api(`/playlists/${id}`).delete()
    if (!res.ok) throw new Error(`${res.status}`)
    playlists.value = playlists.value.filter(p => p.id !== id)
    activeItems.value = []
  }

  async function addItem(playlistId: string, mediaId: string) {
    const res = await api(`/playlists/${playlistId}/items`).post({ mediaId })
    if (!res.ok) throw new Error(`${res.status}`)
    const item: PlaylistItem = await res.json()
    activeItems.value.push(item)
  }

  async function removeItem(playlistId: string, itemId: string) {
    const res = await api(`/playlists/${playlistId}/items/${itemId}`).delete()
    if (!res.ok) throw new Error(`${res.status}`)
    activeItems.value = activeItems.value.filter(i => i.id !== itemId)
  }

  return {
    playlists, activeItems, loading, itemsLoading, error,
    fetchPlaylists, fetchItems, createPlaylist, updatePlaylist, deletePlaylist,
    addItem, removeItem,
  }
}
