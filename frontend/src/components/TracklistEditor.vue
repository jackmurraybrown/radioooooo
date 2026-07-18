<script setup lang="ts">
// ✮⋆‧°—°‧⋆✮ shared tracklist editor — used in the public form and the admin episode dialog
import { computed } from 'vue'
import { type Track, timeToSeconds } from '@/utils/tracklist'

const tracks = defineModel<Track[]>({ required: true })

const props = defineProps<{
  episodeDuration?: number // seconds — 0 or undefined means no duration check
}>()

// ✶. ݁ ˖ client-side validation — all four fields required on touched rows
const validationErrors = computed(() => {
  const errs: string[] = []
  // a row is "touched" if any field has content
  const active = tracks.value.filter(t =>
    t.title.trim() || t.artist.trim() || t.time.trim() || t.endTime.trim()
  )
  if (active.length === 0) return errs

  for (let i = 0; i < tracks.value.length; i++) {
    const t = tracks.value[i]
    const anyFilled = t.title.trim() || t.artist.trim() || t.time.trim() || t.endTime.trim()
    if (!anyFilled) continue

    if (!t.title.trim())   errs.push(`track ${i + 1}: title is required`)
    if (!t.artist.trim())  errs.push(`track ${i + 1}: artist is required`)
    if (!t.time.trim())    errs.push(`track ${i + 1}: start time is required`)
    if (!t.endTime.trim()) errs.push(`track ${i + 1}: end time is required`)

    const secs = timeToSeconds(t.time)
    const endSecs = timeToSeconds(t.endTime)
    if (t.time.trim() && secs == null)
      errs.push(`track ${i + 1}: invalid start time (use mm:ss or h:mm:ss)`)
    if (t.endTime.trim() && endSecs == null)
      errs.push(`track ${i + 1}: invalid end time (use mm:ss or h:mm:ss)`)
    if (secs != null && secs < 0)
      errs.push(`track ${i + 1}: start time can't be negative`)
    if (secs != null && endSecs != null && endSecs <= secs)
      errs.push(`track ${i + 1}: end time must be after start time`)
    if (secs != null && props.episodeDuration && secs > props.episodeDuration)
      errs.push(`track ${i + 1}: start time exceeds episode duration`)
  }

  // ⊹ ₊ ⟡ start times must be ascending across tracks
  const times = tracks.value
    .map((t, i) => ({ i, secs: timeToSeconds(t.time) }))
    .filter(x => x.secs != null)
  for (let j = 1; j < times.length; j++) {
    if (times[j].secs! < times[j - 1].secs!)
      errs.push(`track ${times[j].i + 1}: start time is earlier than previous track`)
  }
  return errs
})

const isValid = computed(() => validationErrors.value.length === 0)

function addTrack() {
  tracks.value.push({ title: '', artist: '', time: '', endTime: '' })
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
          <th class="col-title">title</th>
          <th class="col-artist">artist</th>
          <th class="col-time">start</th>
          <th class="col-time">end</th>
          <th class="col-actions"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(track, i) in tracks" :key="i">
          <td class="col-num">{{ i + 1 }}</td>
          <td class="col-title">
            <input v-model="track.title" placeholder="title" />
          </td>
          <td class="col-artist">
            <input v-model="track.artist" placeholder="artist" />
          </td>
          <td class="col-time">
            <input v-model="track.time" placeholder="mm:ss" />
          </td>
          <td class="col-time">
            <input v-model="track.endTime" placeholder="mm:ss" />
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
  color: var(--muted-foreground);
  padding: 0.3rem 0.25rem;
  border-bottom: 1px solid var(--border);
}

.track-table td {
  padding: 0.2rem 0.25rem;
  vertical-align: middle;
}

.track-table tbody tr:hover { background: var(--muted); }

.col-num    { width: 1.75rem; text-align: center; color: var(--muted-foreground); font-size: 0.8rem; }
.col-time   { width: 5rem; }
.col-time input { text-align: center; }
.col-actions { width: 5rem; white-space: nowrap; text-align: right; }

.track-table input {
  width: 100%;
  padding: 0.3rem 0.4rem;
  border: 1px solid var(--border);
  font-size: 0.85rem;
  font-family: inherit;
  background: var(--input);
  color: var(--foreground);
  box-sizing: border-box;
  outline: none;
}

.track-table input:focus { border-color: var(--ring); }
.track-table input::placeholder { color: var(--muted-foreground); opacity: 0.6; }

.btn-icon {
  background: none;
  border: 1px solid var(--border);
  padding: 0.15rem 0.35rem;
  cursor: pointer;
  font-size: 0.8rem;
  color: var(--muted-foreground);
  line-height: 1;
}

.btn-icon:hover:not(:disabled) { background: var(--muted); color: var(--foreground); }
.btn-icon:disabled { opacity: 0.25; cursor: default; }
.btn-remove:hover:not(:disabled) { color: var(--destructive); border-color: var(--destructive); }

.editor-footer { display: flex; }

.add-btn {
  padding: 0.35rem 0.75rem;
  border: 1px solid var(--border);
  background: var(--background);
  font-size: 0.8rem;
  color: var(--muted-foreground);
  cursor: pointer;
  font-family: inherit;
}

.add-btn:hover { background: var(--muted); color: var(--foreground); }

.errors {
  margin: 0;
  padding-left: 1.1rem;
  font-size: 0.78rem;
  color: var(--destructive);
}

.errors li { margin-bottom: 0.15rem; }
</style>
