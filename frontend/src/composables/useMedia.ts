// ✮ ⋆ ˚｡𖦹 ⋆｡°✩ media library — fetch, create, update, delete
import { ref } from 'vue'
import { api } from '@/api/client'
import type { Media, MediaCreateBody, MediaUpdateBody } from '@/api/types'

export function useMedia() {
  const media = ref<Media[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchMedia() {
    loading.value = true
    error.value = null
    try {
      const res = await api('/media').get()
      if (!res.ok) throw new Error(`${res.status}`)
      const data = await res.json()
      media.value = data.media ?? []
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'failed to fetch media'
    } finally {
      loading.value = false
    }
  }

  async function createMedia(body: Omit<MediaCreateBody, '$schema'>) {
    const res = await api('/media').post(body)
    if (!res.ok) throw new Error(`${res.status}`)
    const item: Media = await res.json()
    media.value.unshift(item)
  }

  async function updateMedia(id: string, body: Omit<MediaUpdateBody, '$schema'>) {
    const res = await api(`/media/${id}`).put(body)
    if (!res.ok) throw new Error(`${res.status}`)
    const item: Media = await res.json()
    const i = media.value.findIndex(m => m.id === id)
    if (i !== -1) media.value[i] = item
  }

  async function deleteMedia(id: string) {
    const res = await api(`/media/${id}`).delete()
    if (!res.ok) throw new Error(`${res.status}`)
    media.value = media.value.filter(m => m.id !== id)
  }

  return { media, loading, error, fetchMedia, createMedia, updateMedia, deleteMedia }
}
