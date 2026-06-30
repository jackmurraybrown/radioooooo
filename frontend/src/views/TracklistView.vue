<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'

const BASE = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
const route = useRoute()
const token = route.params.token as string

interface Track {
  title: string
  artist: string
  time: string
}

interface EpisodeInfo {
  id: string
  title: string
  startTime: string
  endTime: string
}

const episode = ref<EpisodeInfo | null>(null)
const tracks = ref<Track[]>([])
const loading = ref(true)
const saving = ref(false)
const saved = ref(false)
const error = ref<string | null>(null)

const episodeDuration = computed(() => {
  if (!episode.value) return 0
  const start = new Date(episode.value.startTime).getTime()
  const end = new Date(episode.value.endTime).getTime()
  return Math.floor((end - start) / 1000)
})

function secondsToTime(s: number | null): string {
  if (s == null) return ''
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const sec = s % 60
  if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`
  return `${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`
}

function timeToSeconds(t: string): number | null {
  if (!t.trim()) return null
  const parts = t.split(':').map(Number)
  if (parts.some(isNaN)) return null
  if (parts.length === 3) return parts[0] * 3600 + parts[1] * 60 + parts[2]
  if (parts.length === 2) return parts[0] * 60 + parts[1]
  return null
}

const validationErrors = computed(() => {
  const errs: string[] = []
  const filled = tracks.value.filter(t => t.title.trim())
  if (filled.length === 0) {
    errs.push('add at least one track')
    return errs
  }
  for (let i = 0; i < tracks.value.length; i++) {
    const t = tracks.value[i]
    if (!t.title.trim() && t.artist.trim()) {
      errs.push(`track ${i + 1}: title is required`)
    }
    const secs = timeToSeconds(t.time)
    if (t.time.trim() && secs == null) {
      errs.push(`track ${i + 1}: invalid time format (use mm:ss or h:mm:ss)`)
    }
    if (secs != null && secs < 0) {
      errs.push(`track ${i + 1}: time can't be negative`)
    }
    if (secs != null && episodeDuration.value > 0 && secs > episodeDuration.value) {
      errs.push(`track ${i + 1}: time exceeds episode duration`)
    }
  }
  const times = tracks.value
    .map((t, i) => ({ i, secs: timeToSeconds(t.time) }))
    .filter(x => x.secs != null)
  for (let j = 1; j < times.length; j++) {
    if (times[j].secs! < times[j - 1].secs!) {
      errs.push(`track ${times[j].i + 1}: time is earlier than previous track`)
    }
  }
  return errs
})

const canSave = computed(() => validationErrors.value.length === 0 && !saving.value)

onMounted(async () => {
  try {
    const res = await fetch(`${BASE}/tracklists/${token}`)
    if (!res.ok) {
      error.value = 'invalid or expired link'
      return
    }
    const data = await res.json()
    episode.value = data.episode
    tracks.value = (data.tracks ?? []).map((t: any) => ({
      title: t.title,
      artist: t.artist ?? '',
      time: secondsToTime(t.startedAt ?? null),
    }))
    if (tracks.value.length === 0) addTrack()
  } finally {
    loading.value = false
  }
})

function addTrack() {
  tracks.value.push({ title: '', artist: '', time: '' })
}

function removeTrack(i: number) {
  tracks.value.splice(i, 1)
}

function moveUp(i: number) {
  if (i === 0) return
  const tmp = tracks.value[i]
  tracks.value[i] = tracks.value[i - 1]
  tracks.value[i - 1] = tmp
}

function moveDown(i: number) {
  if (i >= tracks.value.length - 1) return
  const tmp = tracks.value[i]
  tracks.value[i] = tracks.value[i + 1]
  tracks.value[i + 1] = tmp
}

async function save() {
  if (!canSave.value) return
  saving.value = true
  saved.value = false
  error.value = null
  try {
    const body = {
      tracks: tracks.value
        .filter(t => t.title.trim())
        .map(t => ({
          title: t.title.trim(),
          artist: t.artist.trim() || undefined,
          startedAt: timeToSeconds(t.time),
        })),
    }
    const res = await fetch(`${BASE}/tracklists/${token}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    if (!res.ok) {
      error.value = 'failed to save'
      return
    }
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
    <div v-if="loading" class="loading">loading...</div>

    <div v-else-if="error && !episode" class="error-page">
      <p>{{ error }}</p>
    </div>

    <div v-else-if="episode">
      <header class="episode-header">
        <h1>{{ episode.title }}</h1>
        <p class="episode-time">
          {{ formatTime(episode.startTime) }} - {{ formatTime(episode.endTime) }}
        </p>
      </header>

      <table class="track-table">
        <thead>
          <tr>
            <th class="col-num">#</th>
            <th class="col-time">time</th>
            <th class="col-artist">artist</th>
            <th class="col-title">title</th>
            <th class="col-actions"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(track, i) in tracks" :key="i">
            <td class="col-num">{{ i + 1 }}</td>
            <td class="col-time">
              <input v-model="track.time" placeholder="mm:ss" />
            </td>
            <td class="col-artist">
              <input v-model="track.artist" placeholder="artist" />
            </td>
            <td class="col-title">
              <input v-model="track.title" placeholder="title *" />
            </td>
            <td class="col-actions">
              <button @click="moveUp(i)" :disabled="i === 0" title="move up" class="btn-icon">↑</button>
              <button @click="moveDown(i)" :disabled="i === tracks.length - 1" title="move down" class="btn-icon">↓</button>
              <button @click="removeTrack(i)" title="remove" class="btn-icon btn-remove">×</button>
            </td>
          </tr>
        </tbody>
      </table>

      <div class="actions">
        <button @click="addTrack" class="btn">+ add track</button>
        <button @click="save" :disabled="!canSave" class="btn btn-primary">
          {{ saving ? 'saving...' : 'save tracklist' }}
        </button>
      </div>

      <ul v-if="validationErrors.length" class="validation-errors">
        <li v-for="err in validationErrors" :key="err">{{ err }}</li>
      </ul>
      <p v-if="saved" class="msg-success">saved</p>
      <p v-if="error" class="msg-error">{{ error }}</p>
    </div>
  </div>
</template>

<style scoped>
.tracklist-page {
  max-width: 720px;
  margin: 2rem auto;
  padding: 0 1rem;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  color: #1a1a1a;
}

.loading,
.error-page {
  text-align: center;
  padding: 3rem 0;
  color: #666;
}

.episode-header {
  margin-bottom: 1.5rem;
}

.episode-header h1 {
  font-size: 1.4rem;
  margin: 0 0 0.25rem;
}

.episode-time {
  color: #666;
  font-size: 0.85rem;
  margin: 0;
}

.track-table {
  width: 100%;
  border-collapse: collapse;
}

.track-table th {
  text-align: left;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #888;
  padding: 0.4rem 0.3rem;
  border-bottom: 1px solid #ddd;
}

.track-table td {
  padding: 0.25rem 0.3rem;
  vertical-align: middle;
}

.track-table tbody tr:hover {
  background: #f8f8f8;
}

.col-num {
  width: 2rem;
  text-align: center;
  color: #999;
  font-size: 0.85rem;
}

.col-time {
  width: 5.5rem;
}

.col-time input {
  text-align: center;
  width: 100%;
}

.col-artist,
.col-title {
  width: auto;
}

.col-actions {
  width: 5.5rem;
  white-space: nowrap;
  text-align: right;
}

.track-table input {
  width: 100%;
  padding: 0.35rem 0.5rem;
  border: 1px solid #ddd;
  border-radius: 3px;
  font-size: 0.9rem;
  font-family: inherit;
  background: #fff;
  box-sizing: border-box;
}

.track-table input:focus {
  outline: none;
  border-color: #888;
}

.track-table input::placeholder {
  color: #bbb;
}

.btn-icon {
  background: none;
  border: 1px solid #ddd;
  border-radius: 3px;
  padding: 0.2rem 0.4rem;
  cursor: pointer;
  font-size: 0.85rem;
  color: #666;
  line-height: 1;
}

.btn-icon:hover:not(:disabled) {
  background: #eee;
  color: #333;
}

.btn-icon:disabled {
  opacity: 0.3;
  cursor: default;
}

.btn-remove:hover:not(:disabled) {
  background: #fee;
  color: #c00;
  border-color: #fcc;
}

.actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 1rem;
}

.btn {
  padding: 0.5rem 1rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: #fff;
  cursor: pointer;
  font-size: 0.85rem;
  font-family: inherit;
}

.btn:hover:not(:disabled) {
  background: #f5f5f5;
}

.btn-primary {
  background: #1a1a1a;
  color: #fff;
  border-color: #1a1a1a;
}

.btn-primary:hover:not(:disabled) {
  background: #333;
}

.btn:disabled {
  opacity: 0.4;
  cursor: default;
}

.validation-errors {
  margin-top: 0.75rem;
  padding-left: 1.2rem;
  color: #c00;
  font-size: 0.85rem;
}

.validation-errors li {
  margin-bottom: 0.2rem;
}

.msg-success {
  color: #2a7;
  margin-top: 0.75rem;
  font-size: 0.85rem;
}

.msg-error {
  color: #c00;
  margin-top: 0.75rem;
  font-size: 0.85rem;
}
</style>
