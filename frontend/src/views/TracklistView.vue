<script setup lang="ts">
// ⊹ ࣪ ˖ public tracklist submission form — accessed via token link
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import TracklistEditor from '@/components/TracklistEditor.vue'
import { secondsToTime, timeToSeconds, type Track } from '@/utils/tracklist'

const BASE = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
const route = useRoute()
const token = route.params.token as string

interface EpisodeInfo {
  id:        string
  title:     string
  startTime: string
  endTime:   string
}

const episode  = ref<EpisodeInfo | null>(null)
const tracks   = ref<Track[]>([])
const loading  = ref(true)
const saving   = ref(false)
const saved    = ref(false)
const error    = ref<string | null>(null)

const editorRef = ref<InstanceType<typeof TracklistEditor>>()

const episodeDuration = computed(() => {
  if (!episode.value) return 0
  return Math.floor(
    (new Date(episode.value.endTime).getTime() - new Date(episode.value.startTime).getTime()) / 1000
  )
})

// ⋆˙⟡ must have at least one titled track before saving
const canSave = computed(() =>
  tracks.value.some(t => t.title.trim()) &&
  (editorRef.value?.isValid ?? true) &&
  !saving.value
)

onMounted(async () => {
  try {
    const res = await fetch(`${BASE}/tracklists/${token}`)
    if (!res.ok) { error.value = 'invalid or expired link'; return }
    const data = await res.json()
    episode.value = data.episode
    tracks.value  = (data.tracks ?? []).map((t: any) => ({
      title:   t.title,
      artist:  t.artist ?? '',
      time:    secondsToTime(t.startedAt ?? null),
      endTime: secondsToTime(t.endedAt ?? null),
    }))
    if (tracks.value.length === 0) tracks.value.push({ title: '', artist: '', time: '', endTime: '' })
  } finally {
    loading.value = false
  }
})

async function save() {
  if (!canSave.value) return
  saving.value = true
  saved.value  = false
  error.value  = null
  try {
    const res = await fetch(`${BASE}/tracklists/${token}`, {
      method:  'PUT',
      headers: { 'Content-Type': 'application/json' },
      body:    JSON.stringify({
        tracks: tracks.value
          .filter(t => t.title.trim())
          .map(t => ({
            title:     t.title.trim(),
            artist:    t.artist.trim() || undefined,
            startedAt: timeToSeconds(t.time),
            endedAt:   timeToSeconds(t.endTime) ?? undefined,
          })),
      }),
    })
    if (!res.ok) { error.value = 'failed to save'; return }
    saved.value = true
  } finally {
    saving.value = false
  }
}

function formatTime(iso: string) {
  return new Date(iso).toLocaleString()
}
</script>

<template>
  <div class="max-w-180 my-8 mx-auto px-4 text-foreground">
    <div v-if="loading" class="text-center py-12 text-muted-foreground">loading...</div>

    <div v-else-if="error && !episode" class="text-center py-12 text-destructive">
      <p>{{ error }}</p>
    </div>

    <div v-else-if="episode">
      <header class="mb-6">
        <h1 class="text-[1.4rem] mb-1">{{ episode.title }}</h1>
        <p class="text-muted-foreground text-[0.85rem]">{{ formatTime(episode.startTime) }} — {{ formatTime(episode.endTime) }}</p>
      </header>

      <TracklistEditor ref="editorRef" v-model="tracks" :episodeDuration="episodeDuration" />

      <div class="flex justify-end mt-4">
        <button
          @click="save"
          :disabled="!canSave"
          class="px-5 py-2 border border-primary bg-primary text-primary-foreground cursor-pointer text-[0.85rem] font-sans hover:opacity-85 disabled:opacity-40 disabled:cursor-default"
        >
          {{ saving ? 'saving...' : 'save tracklist' }}
        </button>
      </div>

      <p v-if="saved" class="text-success mt-3 text-[0.85rem]">saved</p>
      <p v-if="error" class="text-destructive mt-3 text-[0.85rem]">{{ error }}</p>
    </div>
  </div>
</template>
