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
      title:  t.title,
      artist: t.artist ?? '',
      time:   secondsToTime(t.startedAt ?? null),
    }))
    if (tracks.value.length === 0) tracks.value.push({ title: '', artist: '', time: '' })
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
  <div class="tracklist-page">
    <div v-if="loading" class="status">loading...</div>

    <div v-else-if="error && !episode" class="status status-err">
      <p>{{ error }}</p>
    </div>

    <div v-else-if="episode">
      <header class="episode-header">
        <h1>{{ episode.title }}</h1>
        <p class="episode-time">{{ formatTime(episode.startTime) }} — {{ formatTime(episode.endTime) }}</p>
      </header>

      <TracklistEditor ref="editorRef" v-model="tracks" :episodeDuration="episodeDuration" />

      <div class="actions">
        <button @click="save" :disabled="!canSave" class="btn btn-primary">
          {{ saving ? 'saving...' : 'save tracklist' }}
        </button>
      </div>

      <p v-if="saved"  class="msg-ok">saved</p>
      <p v-if="error"  class="msg-err">{{ error }}</p>
    </div>
  </div>
</template>

<style scoped>
.tracklist-page {
  max-width: 720px;
  margin: 2rem auto;
  padding: 0 1rem;
  color: var(--foreground);
}

.status {
  text-align: center;
  padding: 3rem 0;
  color: var(--muted-foreground);
}

.status-err { color: var(--destructive); }

.episode-header { margin-bottom: 1.5rem; }

.episode-header h1 {
  font-size: 1.4rem;
  margin: 0 0 0.25rem;
}

.episode-time {
  color: var(--muted-foreground);
  font-size: 0.85rem;
  margin: 0;
}

.actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 1rem;
}

.btn {
  padding: 0.5rem 1.25rem;
  border: 1px solid var(--border);
  background: var(--background);
  color: var(--foreground);
  cursor: pointer;
  font-size: 0.85rem;
  font-family: inherit;
}

.btn-primary {
  background: var(--primary);
  color: var(--primary-foreground);
  border-color: var(--primary);
}

.btn-primary:hover:not(:disabled) { opacity: 0.85; }
.btn:disabled { opacity: 0.4; cursor: default; }

.msg-ok  { color: oklch(0.7 0.18 145); margin-top: 0.75rem; font-size: 0.85rem; }
.msg-err { color: var(--destructive); margin-top: 0.75rem; font-size: 0.85rem; }
</style>
