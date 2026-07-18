<script setup lang="ts">
// ⋆˙⟡ ⋆.˚ episode create / edit
import { ref, reactive, computed, onMounted, watch } from 'vue'
import * as v from 'valibot'
import type { Episode, EpisodeBody } from '@/api/types'
import { useMedia } from '@/composables/useMedia'
import { usePlaylists } from '@/composables/usePlaylists'
import { api } from '@/api/client'
import { useToast } from '@/composables/useToast'
import TracklistEditor from '@/components/TracklistEditor.vue'
import MediaPicker from '@/components/MediaPicker.vue'
import { secondsToTime, timeToSeconds, type Track } from '@/utils/tracklist'
import { TabsRoot, TabsList, TabsTrigger, TabsContent } from 'reka-ui'

type EpisodeType = EpisodeBody['type']

const emit = defineEmits<{
  create: [body: Omit<EpisodeBody, '$schema'>]
  update: [id: string, body: Omit<EpisodeBody, '$schema'>]
  delete: [id: string]
}>()

const toast = useToast()
const dialogEl = ref<HTMLDialogElement>()
const mode = ref<'create' | 'edit'>('create')
const activeTab = ref<'details' | 'tracklist'>('details')
const currentId = ref<string>('')
const currentChannelId = ref<string>('')

// ✮ ⋆ ˚｡𖦹 prefetch media + playlists for dropdowns
const { media, fetchMedia } = useMedia()
const { playlists, fetchPlaylists } = usePlaylists()
onMounted(() => { fetchMedia(); fetchPlaylists() })

// ⊹ ₊ ⟡ after upload: refresh full list so combobox displayValue resolves the title
async function onMediaAdded(id: string, _title: string) {
  await fetchMedia()
  form.sourceRef = id
}

// ⊹ ࣪ ˖ sourceAdapter is fully determined by type
const adapterForType: Record<EpisodeType, string> = {
  live: 'icecast',
  recorded: 'media',
  external: 'external',
  playlist: 'playlist',
}

const refLabel: Record<EpisodeType, string> = {
  live: 'mount point',
  recorded: 'audio track',
  external: 'stream url',
  playlist: 'playlist id',
}

// ✮ ⋆ ˚｡𖦹 details form schema
const schema = v.pipe(
  v.object({
    title: v.pipe(v.string(), v.minLength(1, 'required'), v.maxLength(200)),
    type: v.picklist(['recorded', 'live', 'playlist', 'external'] as const),
    sourceRef: v.pipe(v.string(), v.minLength(1, 'required')),
    startTime: v.pipe(v.string(), v.minLength(1, 'required')),
    endTime: v.pipe(v.string(), v.minLength(1, 'required')),
    description: v.optional(v.string()),
  }),
  v.check(({ startTime, endTime }) => !startTime || !endTime || endTime > startTime, 'end must be after start'),
)

type FormState = {
  title: string
  type: EpisodeType
  description: string
  startTime: string
  endTime: string
  sourceRef: string
  contactEmail: string
}

type FormErrors = Partial<Record<keyof FormState | '_form', string>>

const form = reactive<FormState>({
  title: '',
  type: 'recorded',
  description: '',
  startTime: '',
  endTime: '',
  sourceRef: '',
  contactEmail: '',
})

const errors = reactive<FormErrors>({})

function clearErrors() {
  Object.keys(errors).forEach(k => delete (errors as Record<string, string>)[k])
}

function toLocal(iso: string): string {
  return iso ? iso.slice(0, 16) : ''
}

function reset(partial: Partial<FormState>) {
  form.title = partial.title ?? ''
  form.type = partial.type ?? 'recorded'
  form.description = partial.description ?? ''
  form.startTime = partial.startTime ?? ''
  form.endTime = partial.endTime ?? ''
  form.sourceRef = partial.sourceRef ?? ''
  form.contactEmail = partial.contactEmail ?? ''
  clearErrors()
}

// ⊹ ₊ ⟡ tracklist tab state
const editorRef = ref<InstanceType<typeof TracklistEditor>>()

const tracks = ref<Track[]>([])
const tracksLoading = ref(false)
const tracksSaving = ref(false)
const tracksSaved = ref(false)
const tracksError = ref<string | null>(null)
const linkCopied = ref(false)
const linkLoading = ref(false)

const episodeDuration = computed(() => {
  if (!form.startTime || !form.endTime) return 0
  return Math.floor((new Date(form.endTime).getTime() - new Date(form.startTime).getTime()) / 1000)
})

const canSaveTracks = computed(() => (editorRef.value?.isValid ?? true) && !tracksSaving.value)

async function fetchTracks() {
  if (!currentId.value || !currentChannelId.value) return
  tracksLoading.value = true
  tracksError.value = null
  try {
    const res = await api(`/channels/${currentChannelId.value}/episodes/${currentId.value}/tracks`).get()
    if (!res.ok) throw new Error(`${res.status}`)
    const data = await res.json()
    tracks.value = (data.tracks ?? []).map((t: any) => ({
      title:   t.title,
      artist:  t.artist ?? '',
      time:    secondsToTime(t.startedAt ?? null),
      endTime: secondsToTime(t.endedAt ?? null),
    }))
    if (tracks.value.length === 0) tracks.value.push({ title: '', artist: '', time: '', endTime: '' })
  } catch {
    tracksError.value = 'failed to load tracks'
  } finally {
    tracksLoading.value = false
  }
}

async function saveTracks() {
  if (!canSaveTracks.value) return
  tracksSaving.value = true
  tracksSaved.value = false
  tracksError.value = null
  try {
    const res = await api(`/channels/${currentChannelId.value}/episodes/${currentId.value}/tracks`).put({
      tracks: tracks.value
        .filter(t => t.title.trim())
        .map(t => ({
          title:     t.title.trim(),
          artist:    t.artist.trim() || undefined,
          startedAt: timeToSeconds(t.time),
          endedAt:   timeToSeconds(t.endTime) ?? undefined,
        })),
    })
    if (!res.ok) throw new Error(`${res.status}`)
    tracksSaved.value = true
    setTimeout(() => { tracksSaved.value = false }, 2000)
  } catch {
    tracksError.value = 'failed to save'
  } finally {
    tracksSaving.value = false
  }
}



async function copySubmissionLink() {
  if (!currentId.value || !currentChannelId.value) return
  linkLoading.value = true
  try {
    const res = await api(`/channels/${currentChannelId.value}/episodes/${currentId.value}/submission-link`).post({})
    if (!res.ok) throw new Error(`${res.status}`)
    const data = await res.json()
    await navigator.clipboard.writeText(data.url)
    linkCopied.value = true
    setTimeout(() => { linkCopied.value = false }, 2000)
  } catch {
    toast.error('failed to generate link')
  } finally {
    linkLoading.value = false
  }
}

// ✶. ݁ ˖ lazy-load tracks when switching to the tab
watch(activeTab, (tab) => {
  if (tab === 'tracklist' && tracks.value.length === 0) fetchTracks()
})

function openCreate(start: string, end: string, channelId?: string) {
  mode.value = 'create'
  activeTab.value = 'details'
  currentId.value = ''
  currentChannelId.value = channelId ?? ''
  tracks.value = []
  reset({ startTime: toLocal(start), endTime: toLocal(end) })
  dialogEl.value?.showModal()
}

function openEdit(episode: Episode, channelId?: string) {
  mode.value = 'edit'
  activeTab.value = 'details'
  currentId.value = episode.id
  currentChannelId.value = channelId ?? episode.channelId
  tracks.value = []
  linkCopied.value = false
  reset({
    title: episode.title,
    type: episode.type as EpisodeType,
    description: episode.description,
    startTime: toLocal(episode.startTime),
    endTime: toLocal(episode.endTime),
    sourceRef: episode.sourceRef,
    contactEmail: episode.contactEmail ?? '',
  })
  dialogEl.value?.showModal()
}

function close() { dialogEl.value?.close() }

function submit() {
  clearErrors()
  const result = v.safeParse(schema, { ...form })
  if (!result.success) {
    for (const issue of result.issues) {
      const key = (issue.path?.[0]?.key as keyof FormState) ?? '_form'
      if (!errors[key]) errors[key] = issue.message
    }
    return
  }
  const body: Omit<EpisodeBody, '$schema'> = {
    title: form.title,
    type: form.type,
    startTime: new Date(form.startTime).toISOString(),
    endTime: new Date(form.endTime).toISOString(),
    sourceAdapter: adapterForType[form.type],
    sourceRef: form.sourceRef,
    ...(form.description ? { description: form.description } : {}),
    ...(form.contactEmail ? { contactEmail: form.contactEmail } : {}),
  }
  if (mode.value === 'create') emit('create', body)
  else emit('update', currentId.value, body)
  close()
}

function remove() { emit('delete', currentId.value); close() }

defineExpose({ openCreate, openEdit, close })
</script>

<template>
  <dialog ref="dialogEl" @click.self="close">
    <TabsRoot v-model="activeTab">
      <header>
        <div class="header-main">
          <h2>{{ mode === 'create' ? 'new episode' : 'edit episode' }}</h2>
          <button type="button" class="close-btn" @click="close" aria-label="close">✕</button>
        </div>
        <!-- ⋆˙⟡ reka-ui tabs — only visible in edit mode -->
        <TabsList v-if="mode === 'edit'" class="tabs">
          <TabsTrigger value="details">details</TabsTrigger>
          <TabsTrigger value="tracklist">tracklist</TabsTrigger>
        </TabsList>
      </header>

      <!-- ⊹ ₊ details tab -->
      <TabsContent value="details" as-child>
        <form @submit.prevent="submit" novalidate>
          <div class="fields">
            <div class="field">
              <label for="ep-title">title</label>
              <input id="ep-title" v-model="form.title" :class="{ error: errors.title }" maxlength="200" />
              <span v-if="errors.title" class="err">{{ errors.title }}</span>
            </div>

            <div class="field">
              <label for="ep-type">type</label>
              <select id="ep-type" v-model="form.type">
                <option value="recorded">recorded</option>
                <option value="live">live</option>
                <option value="playlist">playlist</option>
                <option value="external">external</option>
              </select>
            </div>

            <div class="field">
              <label>{{ refLabel[form.type] }}</label>

              <!-- ⊹ ₊ recorded — searchable picker with upload -->
              <MediaPicker
                v-if="form.type === 'recorded'"
                v-model="form.sourceRef"
                :media="media"
                :class="{ error: errors.sourceRef }"
                @media-added="onMediaAdded"
              />

              <!-- ⋆˙⟡ playlist — pick from playlists -->
              <select v-else-if="form.type === 'playlist'" id="ep-ref" v-model="form.sourceRef"
                :class="{ error: errors.sourceRef }">
                <option value="" disabled>select a playlist…</option>
                <option v-for="pl in playlists" :key="pl.id" :value="pl.id">{{ pl.name }}</option>
              </select>

              <!-- live / external — plain text -->
              <input v-else id="ep-ref" v-model="form.sourceRef" :class="{ error: errors.sourceRef }"
                :placeholder="form.type === 'live' ? 'e.g. main' : 'https://…'" />
              <span v-if="errors.sourceRef" class="err">{{ errors.sourceRef }}</span>
            </div>

            <div class="row">
              <div class="field">
                <label for="ep-start">start</label>
                <input id="ep-start" type="datetime-local" v-model="form.startTime"
                  :class="{ error: errors.startTime }" />
                <span v-if="errors.startTime" class="err">{{ errors.startTime }}</span>
              </div>
              <div class="field">
                <label for="ep-end">end</label>
                <input id="ep-end" type="datetime-local" v-model="form.endTime"
                  :class="{ error: errors.endTime || errors._form }" />
                <span v-if="errors.endTime || errors._form" class="err">{{ errors.endTime ?? errors._form }}</span>
              </div>
            </div>

            <div class="field">
              <label for="ep-desc">description</label>
              <textarea id="ep-desc" v-model="form.description" rows="3" maxlength="2000" />
            </div>

            <div class="field">
              <label for="ep-email">
                contact email
                <span class="label-hint">for tracklist submission link after show</span>
              </label>
              <input id="ep-email" type="email" v-model="form.contactEmail" placeholder="dj@example.com" />
            </div>
          </div>

          <footer>
            <button v-if="mode === 'edit'" type="button" class="delete-btn" @click="remove">delete</button>
            <div class="actions">
              <button type="button" @click="close">cancel</button>
              <button type="submit" class="primary">{{ mode === 'create' ? 'create' : 'save' }}</button>
            </div>
          </footer>
        </form>
      </TabsContent>

      <!-- ✮ ⋆ ˚｡𖦹 tracklist tab -->
      <TabsContent v-if="mode === 'edit'" value="tracklist">
        <div class="tracklist-panel">
          <div class="track-content">
            <div v-if="tracksLoading" class="track-empty">loading…</div>
            <div v-else-if="tracksError && tracks.length === 0" class="track-empty track-err">{{ tracksError }}</div>
            <template v-else>
              <TracklistEditor ref="editorRef" v-model="tracks" :episodeDuration="episodeDuration" />
              <p v-if="tracksError" class="track-err-msg">{{ tracksError }}</p>
              <p v-if="tracksSaved" class="track-ok-msg">saved</p>
            </template>
          </div>

          <footer>
            <div class="actions">
              <!-- ⊹ ₊ ⟡ copy link for sharing with the dj -->
              <button type="button" class="link-btn" :disabled="linkLoading" @click="copySubmissionLink">
                {{ linkCopied ? 'copied!' : linkLoading ? '…' : 'copy link' }}
              </button>
              <button type="button" @click="close">cancel</button>
              <button type="button" class="primary" :disabled="!canSaveTracks" @click="saveTracks">
                {{ tracksSaving ? 'saving…' : 'save tracklist' }}
              </button>
            </div>
          </footer>
        </div>
      </TabsContent>
    </TabsRoot>
  </dialog>
</template>

<style scoped>
dialog {
  border: 1px solid var(--border);
  padding: 0;
  width: min(560px, 90vw);
  min-height: 700px;
  background: var(--background);
  color: var(--foreground);
}

dialog::backdrop {
  background: oklch(0 0 0 / 0.65);
}

header {
  border-bottom: 1px solid var(--border);
}

.header-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.25rem 1.5rem 1rem;
}

h2 {
  font-size: 1rem;
  font-weight: 600;
  margin: 0;
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--muted-foreground);
  font-size: 1rem;
  padding: 0.25rem;
  line-height: 1;
}

.close-btn:hover {
  color: var(--foreground);
}

/* ⋆˙⟡ tabs — TabsList gets .tabs class; triggers render internally so need :deep() */
.tabs {
  display: flex;
  gap: 0;
  padding: 0 1.5rem;
  border-top: 1px solid var(--border);
}

.tabs :deep(button) {
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  padding: 0.5rem 0.75rem;
  font-size: 0.8rem;
  font-weight: 500;
  color: var(--muted-foreground);
  cursor: pointer;
  font-family: inherit;
  margin-bottom: -1px;
}

.tabs :deep(button:hover) {
  color: var(--foreground);
}

.tabs :deep(button[data-state="active"]) {
  color: var(--foreground);
  border-bottom-color: var(--foreground);
}

/* details form */
form {
  display: flex;
  flex-direction: column;
}

.fields {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

label {
  font-size: 0.8rem;
  color: var(--muted-foreground);
  font-weight: 500;
  display: flex;
  align-items: baseline;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.label-hint {
  font-size: 0.72rem;
  color: var(--muted-foreground);
  font-weight: 400;
  opacity: 0.7;
}

input,
select,
textarea {
  font-size: 0.9rem;
  padding: 0.45rem 0.6rem;
  border: 1px solid var(--border);
  outline: none;
  font-family: inherit;
  color: var(--foreground);
  background: var(--input);
}

input:focus,
select:focus,
textarea:focus {
  border-color: var(--ring);
}

input.error,
select.error {
  border-color: var(--destructive);
}

.err {
  font-size: 0.75rem;
  color: var(--destructive);
}

textarea {
  resize: vertical;
}

.row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border);
}

.actions {
  display: flex;
  gap: 0.5rem;
  margin-left: auto;
}

button {
  padding: 0.45rem 1rem;
  border: 1px solid var(--border);
  background: var(--background);
  color: var(--foreground);
  font-size: 0.85rem;
  cursor: pointer;
  font-family: inherit;
}

button:hover:not(:disabled) {
  background: var(--muted);
}

button:disabled {
  opacity: 0.4;
  cursor: default;
}

button.primary {
  background: var(--primary);
  color: var(--primary-foreground);
  border-color: var(--primary);
}

button.primary:hover:not(:disabled) {
  opacity: 0.85;
  background: var(--primary);
}

button.delete-btn {
  color: var(--destructive);
  border-color: var(--destructive);
}

button.delete-btn:hover {
  background: color-mix(in oklch, var(--destructive) 10%, transparent);
}

button.link-btn {
  color: var(--muted-foreground);
  border-color: var(--border);
  font-size: 0.8rem;
}

button.link-btn:hover:not(:disabled) {
  background: var(--muted);
  color: var(--foreground);
}

/* ✮⋆‧° tracklist panel */
.tracklist-panel {
  display: flex;
  flex-direction: column;
}

.track-content {
  padding: 1rem 1.5rem;
  overflow-y: auto;
  max-height: 340px;
}

.track-empty {
  font-size: 0.85rem;
  color: var(--muted-foreground);
  padding: 1rem 0;
}

.track-err {
  color: var(--destructive);
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

.track-table tbody tr:hover {
  background: var(--muted);
}

.col-num {
  width: 1.75rem;
  text-align: center;
  color: var(--muted-foreground);
  font-size: 0.8rem;
}

.col-time {
  width: 5rem;
}

.col-time input {
  text-align: center;
}

.col-actions {
  width: 5rem;
  white-space: nowrap;
  text-align: right;
}

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

.track-table input:focus {
  border-color: var(--ring);
}

.track-table input::placeholder {
  color: var(--muted-foreground);
  opacity: 0.6;
}

.btn-icon {
  background: none;
  border: 1px solid var(--border);
  padding: 0.15rem 0.35rem;
  cursor: pointer;
  font-size: 0.8rem;
  color: var(--muted-foreground);
  line-height: 1;
}

.btn-icon:hover:not(:disabled) {
  background: var(--muted);
  color: var(--foreground);
}

.btn-icon:disabled {
  opacity: 0.25;
  cursor: default;
}

.btn-remove:hover:not(:disabled) {
  color: var(--destructive);
  border-color: var(--destructive);
}

.track-errors {
  margin: 0.5rem 0 0;
  padding-left: 1.1rem;
  font-size: 0.78rem;
  color: var(--destructive);
}

.track-errors li {
  margin-bottom: 0.15rem;
}

.track-err-msg {
  font-size: 0.8rem;
  color: var(--destructive);
  margin: 0.5rem 0 0;
}

.track-ok-msg {
  font-size: 0.8rem;
  color: oklch(0.7 0.18 145);
  margin: 0.5rem 0 0;
}

.add-track-btn {
  font-size: 0.8rem;
  color: var(--muted-foreground);
  border-color: var(--border);
}

.add-track-btn:hover {
  background: var(--muted);
}
</style>
