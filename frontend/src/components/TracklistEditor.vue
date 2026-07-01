<script setup lang="ts">
// ✮⋆‧°—°‧⋆✮ shared tracklist editor — used in the public form and the admin episode dialog
import { computed } from 'vue'

export interface Track {
  title:  string
  artist: string
  time:   string
}

const tracks = defineModel<Track[]>({ required: true })

const props = defineProps<{
  episodeDuration?: number // seconds — 0 or undefined means no duration check
}>()

// ⊹ ₊ ⟡ time helpers
export function secondsToTime(s: number | null): string {
  if (s == null) return ''
  const h   = Math.floor(s / 3600)
  const m   = Math.floor((s % 3600) / 60)
  const sec = s % 60
  if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`
  return `${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`
}

export function timeToSeconds(t: string): number | null {
  if (!t.trim()) return null
  const parts = t.split(':').map(Number)
  if (parts.some(isNaN)) return null
  if (parts.length === 3) return parts[0] * 3600 + parts[1] * 60 + parts[2]
  if (parts.length === 2) return parts[0] * 60 + parts[1]
  return null
}

// ✶. ݁ ˖ client-side validation
const validationErrors = computed(() => {
  const errs: string[] = []
  const filled = tracks.value.filter(t => t.title.trim())
  if (filled.length === 0) return errs
  for (let i = 0; i < tracks.value.length; i++) {
    const t    = tracks.value[i]
    const secs = timeToSeconds(t.time)
    if (t.time.trim() && secs == null)
      errs.push(`track ${i + 1}: invalid time format (use mm:ss or h:mm:ss)`)
    if (secs != null && secs < 0)
      errs.push(`track ${i + 1}: time can't be negative`)
    if (secs != null && props.episodeDuration && secs > props.episodeDuration)
      errs.push(`track ${i + 1}: time exceeds episode duration`)
  }
  const times = tracks.value
    .map((t, i) => ({ i, secs: timeToSeconds(t.time) }))
    .filter(x => x.secs != null)
  for (let j = 1; j < times.length; j++) {
    if (times[j].secs! < times[j - 1].secs!)
      errs.push(`track ${times[j].i + 1}: time is earlier than previous track`)
  }
  return errs
})

const isValid = computed(() => validationErrors.value.length === 0)

function addTrack() {
  tracks.value.push({ title: '', artist: '', time: '' })
}

function removeTrack(i: number) {
  tracks.value.splice(i, 1)
}

function moveUp(i: number) {
  if (i === 0) return
  ;[tracks.value[i], tracks.value[i - 1]] = [tracks.value[i - 1], tracks.value[i]]
}

function moveDown(i: number) {
  if (i >= tracks.value.length - 1) return
  ;[tracks.value[i], tracks.value[i + 1]] = [tracks.value[i + 1], tracks.value[i]]
}

defineExpose({ isValid, validationErrors, timeToSeconds })
</script>

<template>
  <div class="tracklist-editor">
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
            <button type="button" @click="moveUp(i)"   :disabled="i === 0"                   class="btn-icon">↑</button>
            <button type="button" @click="moveDown(i)" :disabled="i === tracks.length - 1"   class="btn-icon">↓</button>
            <button type="button" @click="removeTrack(i)"                                     class="btn-icon btn-remove">×</button>
          </td>
        </tr>
      </tbody>
    </table>

    <div class="editor-footer">
      <button type="button" @click="addTrack" class="add-btn">+ add track</button>
    </div>

    <!-- ⋆˙⟡ ⋆.˚ validation -->
    <ul v-if="validationErrors.length" class="errors">
      <li v-for="err in validationErrors" :key="err">{{ err }}</li>
    </ul>
  </div>
</template>

<style scoped>
.tracklist-editor {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.track-table {
  width: 100%;
  border-collapse: collapse;
}

.track-table th {
  text-align: left;
  font-size: 0.72rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #9ca3af;
  padding: 0.3rem 0.25rem;
  border-bottom: 1px solid #e5e7eb;
}

.track-table td {
  padding: 0.2rem 0.25rem;
  vertical-align: middle;
}

.track-table tbody tr:hover { background: #f9fafb; }

.col-num    { width: 1.75rem; text-align: center; color: #9ca3af; font-size: 0.8rem; }
.col-time   { width: 5rem; }
.col-time input { text-align: center; }
.col-actions { width: 5rem; white-space: nowrap; text-align: right; }

.track-table input {
  width: 100%;
  padding: 0.3rem 0.4rem;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  font-size: 0.85rem;
  font-family: inherit;
  background: #fff;
  box-sizing: border-box;
  outline: none;
}

.track-table input:focus    { border-color: #6366f1; }
.track-table input::placeholder { color: #d1d5db; }

.btn-icon {
  background: none;
  border: 1px solid #e5e7eb;
  border-radius: 3px;
  padding: 0.15rem 0.35rem;
  cursor: pointer;
  font-size: 0.8rem;
  color: #6b7280;
  line-height: 1;
}

.btn-icon:hover:not(:disabled) { background: #f3f4f6; }
.btn-icon:disabled { opacity: 0.25; cursor: default; }
.btn-remove:hover:not(:disabled) { background: #fef2f2; color: #dc2626; border-color: #fca5a5; }

.editor-footer { display: flex; }

.add-btn {
  padding: 0.35rem 0.75rem;
  border-radius: 5px;
  border: 1px solid #e5e7eb;
  background: #fff;
  font-size: 0.8rem;
  color: #6b7280;
  cursor: pointer;
  font-family: inherit;
}

.add-btn:hover { background: #f9fafb; }

.errors {
  margin: 0;
  padding-left: 1.1rem;
  font-size: 0.78rem;
  color: #dc2626;
}

.errors li { margin-bottom: 0.15rem; }
</style>
