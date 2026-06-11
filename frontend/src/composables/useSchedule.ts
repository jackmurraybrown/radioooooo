import { ref, computed, watch, toValue, type MaybeRefOrGetter } from 'vue'
import type { EventInput } from '@fullcalendar/core'
import { api } from '@/api/client'
import type { Episode, EpisodeBody } from '@/api/types'

// ✮ ⋆ ˚｡𖦹 episode type → calendar colour
const typeColor: Record<string, string> = {
  live:     '#e63946',
  recorded: '#457b9d',
  external: '#2a9d8f',
  playlist: '#e9c46a',
}

function toEvent(ep: Episode): EventInput {
  return {
    id: ep.id,
    title: ep.title,
    start: ep.startTime,
    end: ep.endTime,
    color: typeColor[ep.type] ?? '#888',
    extendedProps: { type: ep.type, description: ep.description },
  }
}

export function useSchedule(channelId: MaybeRefOrGetter<string>) {
  const episodes = ref<Episode[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const events = computed<EventInput[]>(() => episodes.value.map(toEvent))

  async function fetchEpisodes() {
    const id = toValue(channelId)
    if (!id) return
    loading.value = true
    error.value = null
    try {
      const res = await api(`/channels/${id}/episodes`).get()
      if (!res.ok) throw new Error(`${res.status}`)
      const data = await res.json()
      episodes.value = data.episodes ?? []
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'failed to fetch episodes'
    } finally {
      loading.value = false
    }
  }

  async function createEpisode(body: Omit<EpisodeBody, '$schema'>) {
    const id = toValue(channelId)
    const res = await api(`/channels/${id}/episodes`).post(body)
    if (!res.ok) throw new Error(`${res.status}`)
    const ep: Episode = await res.json()
    episodes.value.push(ep)
  }

  async function updateEpisode(episodeId: string, body: Partial<Omit<EpisodeBody, '$schema'>>) {
    const id = toValue(channelId)
    const res = await api(`/channels/${id}/episodes/${episodeId}`).put(body)
    if (!res.ok) throw new Error(`${res.status}`)
    const ep: Episode = await res.json()
    const i = episodes.value.findIndex(e => e.id === episodeId)
    if (i !== -1) episodes.value[i] = ep
  }

  async function deleteEpisode(episodeId: string) {
    const id = toValue(channelId)
    const res = await api(`/channels/${id}/episodes/${episodeId}`).delete()
    if (!res.ok) throw new Error(`${res.status}`)
    episodes.value = episodes.value.filter(e => e.id !== episodeId)
  }

  // ⊹ ࣪ ˖ refetch whenever the channel changes
  watch(() => toValue(channelId), (id) => { if (id) fetchEpisodes() })

  return { events, loading, error, fetchEpisodes, createEpisode, updateEpisode, deleteEpisode }
}
