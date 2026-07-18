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
  <div class="flex flex-col gap-2">
    <table class="w-full border-collapse">
      <thead>
        <tr>
          <th class="w-7 text-center text-[0.72rem] uppercase tracking-wider text-muted-foreground px-1 py-[0.3rem] border-b border-border">#</th>
          <th class="text-left text-[0.72rem] uppercase tracking-wider text-muted-foreground px-1 py-[0.3rem] border-b border-border">title</th>
          <th class="text-left text-[0.72rem] uppercase tracking-wider text-muted-foreground px-1 py-[0.3rem] border-b border-border">artist</th>
          <th class="w-20 text-left text-[0.72rem] uppercase tracking-wider text-muted-foreground px-1 py-[0.3rem] border-b border-border">start</th>
          <th class="w-20 text-left text-[0.72rem] uppercase tracking-wider text-muted-foreground px-1 py-[0.3rem] border-b border-border">end</th>
          <th class="w-20 text-left text-[0.72rem] uppercase tracking-wider text-muted-foreground px-1 py-[0.3rem] border-b border-border"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(track, i) in tracks" :key="i" class="hover:bg-muted">
          <td class="w-7 text-center text-muted-foreground text-[0.8rem] px-1 py-[0.2rem] align-middle">{{ i + 1 }}</td>
          <td class="px-1 py-[0.2rem] align-middle">
            <input v-model="track.title" placeholder="title" class="w-full px-[0.4rem] py-[0.3rem] border border-border text-[0.85rem] font-sans bg-input text-foreground box-border outline-none focus:border-ring placeholder:text-muted-foreground placeholder:opacity-60" />
          </td>
          <td class="px-1 py-[0.2rem] align-middle">
            <input v-model="track.artist" placeholder="artist" class="w-full px-[0.4rem] py-[0.3rem] border border-border text-[0.85rem] font-sans bg-input text-foreground box-border outline-none focus:border-ring placeholder:text-muted-foreground placeholder:opacity-60" />
          </td>
          <td class="w-20 px-1 py-[0.2rem] align-middle">
            <input v-model="track.time" placeholder="mm:ss" class="w-full px-[0.4rem] py-[0.3rem] border border-border text-[0.85rem] font-sans bg-input text-foreground box-border outline-none focus:border-ring placeholder:text-muted-foreground placeholder:opacity-60 text-center" />
          </td>
          <td class="w-20 px-1 py-[0.2rem] align-middle">
            <input v-model="track.endTime" placeholder="mm:ss" class="w-full px-[0.4rem] py-[0.3rem] border border-border text-[0.85rem] font-sans bg-input text-foreground box-border outline-none focus:border-ring placeholder:text-muted-foreground placeholder:opacity-60 text-center" />
          </td>
          <td class="w-20 whitespace-nowrap text-right px-1 py-[0.2rem] align-middle">
            <button
              type="button"
              @click="moveUp(i)"
              :disabled="i === 0"
              class="bg-transparent border border-border px-[0.35rem] py-[0.15rem] cursor-pointer text-[0.8rem] text-muted-foreground leading-none hover:not-disabled:bg-muted hover:not-disabled:text-foreground disabled:opacity-25 disabled:cursor-default"
            >↑</button>
            <button
              type="button"
              @click="moveDown(i)"
              :disabled="i === tracks.length - 1"
              class="bg-transparent border border-border px-[0.35rem] py-[0.15rem] cursor-pointer text-[0.8rem] text-muted-foreground leading-none hover:not-disabled:bg-muted hover:not-disabled:text-foreground disabled:opacity-25 disabled:cursor-default"
            >↓</button>
            <button
              type="button"
              @click="removeTrack(i)"
              class="bg-transparent border border-border px-[0.35rem] py-[0.15rem] cursor-pointer text-[0.8rem] text-muted-foreground leading-none hover:not-disabled:bg-muted hover:not-disabled:text-destructive hover:not-disabled:border-destructive disabled:opacity-25 disabled:cursor-default"
            >×</button>
          </td>
        </tr>
      </tbody>
    </table>

    <div class="flex">
      <button type="button" @click="addTrack" class="px-3 py-[0.35rem] border border-border bg-background text-[0.8rem] text-muted-foreground cursor-pointer font-sans hover:bg-muted hover:text-foreground">+ add track</button>
    </div>

    <!-- ⋆˙⟡ ⋆.˚ validation -->
    <ul v-if="validationErrors.length" class="m-0 pl-[1.1rem] text-[0.78rem] text-destructive">
      <li v-for="err in validationErrors" :key="err" class="mb-[0.15rem]">{{ err }}</li>
    </ul>
  </div>
</template>
