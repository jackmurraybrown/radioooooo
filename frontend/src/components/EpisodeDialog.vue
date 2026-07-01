<script setup lang="ts">
// ⋆˙⟡ ⋆.˚ episode create / edit — native <dialog> for free tab trapping + esc
import { ref, reactive, onMounted } from 'vue'
import * as v from 'valibot'
import type { Episode, EpisodeBody } from '@/api/types'
import { useMedia } from '@/composables/useMedia'
import { usePlaylists } from '@/composables/usePlaylists'
import { api } from '@/api/client'
import { useToast } from '@/composables/useToast'

type EpisodeType = EpisodeBody['type']

const emit = defineEmits<{
  create: [body: Omit<EpisodeBody, '$schema'>]
  update: [id: string, body: Omit<EpisodeBody, '$schema'>]
  delete: [id: string]
}>()

const toast = useToast()
const dialogEl = ref<HTMLDialogElement>()
const mode = ref<'create' | 'edit'>('create')
const currentId = ref<string>('')
const currentChannelId = ref<string>('')
const linkCopied = ref(false)
const linkLoading = ref(false)

// ✮ ⋆ ˚｡𖦹 prefetch media + playlists for dropdowns
const { media, fetchMedia } = useMedia()
const { playlists, fetchPlaylists } = usePlaylists()
onMounted(() => { fetchMedia(); fetchPlaylists() })

// ⊹ ࣪ ˖ sourceAdapter is fully determined by type
const adapterForType: Record<EpisodeType, string> = {
  live:     'icecast',
  recorded: 'media',
  external: 'external',
  playlist: 'playlist',
}

const refLabel: Record<EpisodeType, string> = {
  live:     'mount point',
  recorded: 'media id',
  external: 'stream url',
  playlist: 'playlist id',
}

// ✮ ⋆ ˚｡𖦹 schema
const schema = v.pipe(
  v.object({
    title:       v.pipe(v.string(), v.minLength(1, 'required'), v.maxLength(200)),
    type:        v.picklist(['recorded', 'live', 'playlist', 'external'] as const),
    sourceRef:   v.pipe(v.string(), v.minLength(1, 'required')),
    startTime:   v.pipe(v.string(), v.minLength(1, 'required')),
    endTime:     v.pipe(v.string(), v.minLength(1, 'required')),
    description: v.optional(v.string()),
  }),
  v.check(({ startTime, endTime }) => !startTime || !endTime || endTime > startTime, 'end must be after start'),
)

type FormState = {
  title:        string
  type:         EpisodeType
  description:  string
  startTime:    string
  endTime:      string
  sourceRef:    string
  contactEmail: string
}

type FormErrors = Partial<Record<keyof FormState | '_form', string>>

const form = reactive<FormState>({
  title:        '',
  type:         'recorded',
  description:  '',
  startTime:    '',
  endTime:      '',
  sourceRef:    '',
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
  form.title        = partial.title        ?? ''
  form.type         = partial.type         ?? 'recorded'
  form.description  = partial.description  ?? ''
  form.startTime    = partial.startTime    ?? ''
  form.endTime      = partial.endTime      ?? ''
  form.sourceRef    = partial.sourceRef    ?? ''
  form.contactEmail = partial.contactEmail ?? ''
  linkCopied.value  = false
  clearErrors()
}

function openCreate(start: string, end: string, channelId?: string) {
  mode.value             = 'create'
  currentId.value        = ''
  currentChannelId.value = channelId ?? ''
  reset({ startTime: toLocal(start), endTime: toLocal(end) })
  dialogEl.value?.showModal()
}

function openEdit(episode: Episode, channelId?: string) {
  mode.value             = 'edit'
  currentId.value        = episode.id
  currentChannelId.value = channelId ?? episode.channelId
  reset({
    title:        episode.title,
    type:         episode.type as EpisodeType,
    description:  episode.description,
    startTime:    toLocal(episode.startTime),
    endTime:      toLocal(episode.endTime),
    sourceRef:    episode.sourceRef,
    contactEmail: episode.contactEmail ?? '',
  })
  dialogEl.value?.showModal()
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

function close() {
  dialogEl.value?.close()
}

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
    title:         form.title,
    type:          form.type,
    startTime:     new Date(form.startTime).toISOString(),
    endTime:       new Date(form.endTime).toISOString(),
    sourceAdapter: adapterForType[form.type],
    sourceRef:     form.sourceRef,
    ...(form.description  ? { description:  form.description }  : {}),
    ...(form.contactEmail ? { contactEmail: form.contactEmail } : {}),
  }

  if (mode.value === 'create') {
    emit('create', body)
  } else {
    emit('update', currentId.value, body)
  }
  close()
}

function remove() {
  emit('delete', currentId.value)
  close()
}

defineExpose({ openCreate, openEdit, close })
</script>

<template>
  <dialog ref="dialogEl">
    <form @submit.prevent="submit" novalidate>
      <header>
        <h2>{{ mode === 'create' ? 'new episode' : 'edit episode' }}</h2>
        <button type="button" class="close-btn" @click="close" aria-label="close">✕</button>
      </header>

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
          <label for="ep-ref">{{ refLabel[form.type] }}</label>

          <!-- ⊹ ₊ recorded — pick from media library -->
          <select
            v-if="form.type === 'recorded'"
            id="ep-ref"
            v-model="form.sourceRef"
            :class="{ error: errors.sourceRef }"
          >
            <option value="" disabled>select a track…</option>
            <option v-for="m in media" :key="m.id" :value="m.id">
              {{ m.title }}{{ m.artist ? ` — ${m.artist}` : '' }}
            </option>
          </select>

          <!-- ⋆˙⟡ playlist — pick from playlists -->
          <select
            v-else-if="form.type === 'playlist'"
            id="ep-ref"
            v-model="form.sourceRef"
            :class="{ error: errors.sourceRef }"
          >
            <option value="" disabled>select a playlist…</option>
            <option v-for="pl in playlists" :key="pl.id" :value="pl.id">
              {{ pl.name }}
            </option>
          </select>

          <!-- live / external — plain text -->
          <input
            v-else
            id="ep-ref"
            v-model="form.sourceRef"
            :class="{ error: errors.sourceRef }"
            :placeholder="form.type === 'live' ? 'e.g. main' : 'https://…'"
          />

          <span v-if="errors.sourceRef" class="err">{{ errors.sourceRef }}</span>
        </div>

        <div class="row">
          <div class="field">
            <label for="ep-start">start</label>
            <input id="ep-start" type="datetime-local" v-model="form.startTime" :class="{ error: errors.startTime }" />
            <span v-if="errors.startTime" class="err">{{ errors.startTime }}</span>
          </div>
          <div class="field">
            <label for="ep-end">end</label>
            <input id="ep-end" type="datetime-local" v-model="form.endTime" :class="{ error: errors.endTime || errors._form }" />
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
          <!-- ⊹ ₊ ⟡ copy tracklist link for sharing with dj -->
          <button
            v-if="mode === 'edit'"
            type="button"
            class="link-btn"
            :disabled="linkLoading"
            @click="copySubmissionLink"
          >{{ linkCopied ? 'copied!' : linkLoading ? '…' : 'copy tracklist link' }}</button>
          <button type="button" @click="close">cancel</button>
          <button type="submit" class="primary">{{ mode === 'create' ? 'create' : 'save' }}</button>
        </div>
      </footer>
    </form>
  </dialog>
</template>

<style scoped>
dialog {
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  padding: 0;
  width: min(480px, 90vw);
  box-shadow: 0 8px 32px rgba(0,0,0,0.12);
}

dialog::backdrop {
  background: rgba(0,0,0,0.35);
}

form {
  display: flex;
  flex-direction: column;
}

header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.25rem 1.5rem 1rem;
  border-bottom: 1px solid #f3f4f6;
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
  color: #9ca3af;
  font-size: 1rem;
  padding: 0.25rem;
  line-height: 1;
}

.close-btn:hover { color: #111827; }

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
  color: #6b7280;
  font-weight: 500;
  display: flex;
  align-items: baseline;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.label-hint { font-size: 0.72rem; color: #9ca3af; font-weight: 400; }

input,
select,
textarea {
  font-size: 0.9rem;
  padding: 0.45rem 0.6rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  outline: none;
  font-family: inherit;
  color: #111827;
  background: #fff;
}

input:focus,
select:focus,
textarea:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99,102,241,0.15);
}

input.error,
select.error {
  border-color: #dc2626;
}

.err {
  font-size: 0.75rem;
  color: #dc2626;
}

textarea { resize: vertical; }

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
  border-top: 1px solid #f3f4f6;
}

.actions {
  display: flex;
  gap: 0.5rem;
  margin-left: auto;
}

button {
  padding: 0.45rem 1rem;
  border-radius: 6px;
  border: 1px solid #d1d5db;
  background: #fff;
  font-size: 0.85rem;
  cursor: pointer;
  font-family: inherit;
}

button:hover { background: #f9fafb; }

button.primary {
  background: #111827;
  color: #fff;
  border-color: #111827;
}

button.primary:hover { background: #374151; }

button.delete-btn {
  color: #dc2626;
  border-color: #fca5a5;
}

button.delete-btn:hover { background: #fef2f2; }

button.link-btn {
  color: #6366f1;
  border-color: #c7d2fe;
  font-size: 0.8rem;
}

button.link-btn:hover:not(:disabled) { background: #eef2ff; }
button.link-btn:disabled { opacity: 0.5; cursor: default; }
</style>
