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
  <dialog
    ref="dialogEl"
    class="border border-border p-0 w-[min(560px,90vw)] min-h-175 bg-background text-foreground backdrop:bg-backdrop"
    @click.self="close"
  >
    <TabsRoot v-model="activeTab">
      <header class="border-b border-border">
        <div class="flex items-center justify-between px-6 pt-5 pb-4">
          <h2 class="text-base font-semibold m-0 normal-case tracking-normal text-foreground">{{ mode === 'create' ? 'new episode' : 'edit episode' }}</h2>
          <button type="button" class="bg-transparent border-0 cursor-pointer text-muted-foreground text-base p-1 leading-none hover:text-foreground" @click="close" aria-label="close">✕</button>
        </div>
        <!-- ⋆˙⟡ reka-ui tabs — classes applied straight to TabsTrigger so they land on the rendered button regardless of scoped-css quirks -->
        <TabsList v-if="mode === 'edit'" class="flex px-6 border-t border-border">
          <TabsTrigger
            value="details"
            class="bg-transparent border-0 border-b-2 border-transparent px-3 py-2 text-[0.8rem] font-medium text-muted-foreground cursor-pointer font-sans -mb-px hover:text-foreground data-[state=active]:text-foreground data-[state=active]:border-b-foreground"
          >details</TabsTrigger>
          <TabsTrigger
            value="tracklist"
            class="bg-transparent border-0 border-b-2 border-transparent px-3 py-2 text-[0.8rem] font-medium text-muted-foreground cursor-pointer font-sans -mb-px hover:text-foreground data-[state=active]:text-foreground data-[state=active]:border-b-foreground"
          >tracklist</TabsTrigger>
        </TabsList>
      </header>

      <!-- ⊹ ₊ details tab -->
      <TabsContent value="details" as-child>
        <form class="flex flex-col" @submit.prevent="submit" novalidate>
          <div class="flex flex-col gap-4 px-6 py-5">
            <div class="flex flex-col gap-[0.3rem]">
              <label for="ep-title" class="text-[0.8rem] text-muted-foreground font-medium flex items-baseline gap-[0.4rem] flex-wrap">title</label>
              <input
                id="ep-title"
                v-model="form.title"
                maxlength="200"
                class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border outline-none font-sans text-foreground bg-input focus:border-ring"
                :class="errors.title ? 'border-destructive' : 'border-border'"
              />
              <span v-if="errors.title" class="text-xs text-destructive">{{ errors.title }}</span>
            </div>

            <div class="flex flex-col gap-[0.3rem]">
              <label for="ep-type" class="text-[0.8rem] text-muted-foreground font-medium flex items-baseline gap-[0.4rem] flex-wrap">type</label>
              <select
                id="ep-type"
                v-model="form.type"
                class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border border-border outline-none font-sans text-foreground bg-input focus:border-ring"
              >
                <option value="recorded">recorded</option>
                <option value="live">live</option>
                <option value="playlist">playlist</option>
                <option value="external">external</option>
              </select>
            </div>

            <div class="flex flex-col gap-[0.3rem]">
              <label class="text-[0.8rem] text-muted-foreground font-medium flex items-baseline gap-[0.4rem] flex-wrap">{{ refLabel[form.type] }}</label>

              <!-- ⊹ ₊ recorded — searchable picker with upload -->
              <MediaPicker
                v-if="form.type === 'recorded'"
                v-model="form.sourceRef"
                :media="media"
                :class="errors.sourceRef ? 'border-destructive' : ''"
                @media-added="onMediaAdded"
              />

              <!-- ⋆˙⟡ playlist — pick from playlists -->
              <select
                v-else-if="form.type === 'playlist'"
                id="ep-ref"
                v-model="form.sourceRef"
                class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border outline-none font-sans text-foreground bg-input focus:border-ring"
                :class="errors.sourceRef ? 'border-destructive' : 'border-border'"
              >
                <option value="" disabled>select a playlist…</option>
                <option v-for="pl in playlists" :key="pl.id" :value="pl.id">{{ pl.name }}</option>
              </select>

              <!-- live / external — plain text -->
              <input
                v-else
                id="ep-ref"
                v-model="form.sourceRef"
                :placeholder="form.type === 'live' ? 'e.g. main' : 'https://…'"
                class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border outline-none font-sans text-foreground bg-input focus:border-ring"
                :class="errors.sourceRef ? 'border-destructive' : 'border-border'"
              />
              <span v-if="errors.sourceRef" class="text-xs text-destructive">{{ errors.sourceRef }}</span>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div class="flex flex-col gap-[0.3rem]">
                <label for="ep-start" class="text-[0.8rem] text-muted-foreground font-medium flex items-baseline gap-[0.4rem] flex-wrap">start</label>
                <input
                  id="ep-start"
                  type="datetime-local"
                  v-model="form.startTime"
                  class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border outline-none font-sans text-foreground bg-input focus:border-ring"
                  :class="errors.startTime ? 'border-destructive' : 'border-border'"
                />
                <span v-if="errors.startTime" class="text-xs text-destructive">{{ errors.startTime }}</span>
              </div>
              <div class="flex flex-col gap-[0.3rem]">
                <label for="ep-end" class="text-[0.8rem] text-muted-foreground font-medium flex items-baseline gap-[0.4rem] flex-wrap">end</label>
                <input
                  id="ep-end"
                  type="datetime-local"
                  v-model="form.endTime"
                  class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border outline-none font-sans text-foreground bg-input focus:border-ring"
                  :class="(errors.endTime || errors._form) ? 'border-destructive' : 'border-border'"
                />
                <span v-if="errors.endTime || errors._form" class="text-xs text-destructive">{{ errors.endTime ?? errors._form }}</span>
              </div>
            </div>

            <div class="flex flex-col gap-[0.3rem]">
              <label for="ep-desc" class="text-[0.8rem] text-muted-foreground font-medium flex items-baseline gap-[0.4rem] flex-wrap">description</label>
              <textarea
                id="ep-desc"
                v-model="form.description"
                rows="3"
                maxlength="2000"
                class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border border-border outline-none font-sans text-foreground bg-input focus:border-ring resize-y"
              />
            </div>

            <div class="flex flex-col gap-[0.3rem]">
              <label for="ep-email" class="text-[0.8rem] text-muted-foreground font-medium flex items-baseline gap-[0.4rem] flex-wrap">
                contact email
                <span class="text-[0.72rem] text-muted-foreground font-normal opacity-70">for tracklist submission link after show</span>
              </label>
              <input
                id="ep-email"
                type="email"
                v-model="form.contactEmail"
                placeholder="dj@example.com"
                class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border border-border outline-none font-sans text-foreground bg-input focus:border-ring"
              />
            </div>
          </div>

          <footer class="flex items-center justify-between px-6 py-4 border-t border-border">
            <button
              v-if="mode === 'edit'"
              type="button"
              class="px-4 py-[0.45rem] border border-destructive bg-background text-destructive text-[0.85rem] cursor-pointer font-sans hover:bg-destructive/10 disabled:opacity-40 disabled:cursor-default"
              @click="remove"
            >delete</button>
            <div class="flex gap-2 ml-auto">
              <button
                type="button"
                class="px-4 py-[0.45rem] border border-border bg-background text-foreground text-[0.85rem] cursor-pointer font-sans hover:bg-muted disabled:opacity-40 disabled:cursor-default"
                @click="close"
              >cancel</button>
              <button
                type="submit"
                class="px-4 py-[0.45rem] border border-primary bg-primary text-primary-foreground text-[0.85rem] cursor-pointer font-sans hover:opacity-85 disabled:opacity-40 disabled:cursor-default"
              >{{ mode === 'create' ? 'create' : 'save' }}</button>
            </div>
          </footer>
        </form>
      </TabsContent>

      <!-- ✮ ⋆ ˚｡𖦹 tracklist tab -->
      <TabsContent v-if="mode === 'edit'" value="tracklist">
        <div class="flex flex-col">
          <div class="px-6 py-4 overflow-y-auto max-h-85">
            <div v-if="tracksLoading" class="text-[0.85rem] text-muted-foreground py-4">loading…</div>
            <div v-else-if="tracksError && tracks.length === 0" class="text-[0.85rem] text-destructive py-4">{{ tracksError }}</div>
            <template v-else>
              <TracklistEditor ref="editorRef" v-model="tracks" :episodeDuration="episodeDuration" />
              <p v-if="tracksError" class="text-[0.8rem] text-destructive mt-2">{{ tracksError }}</p>
              <p v-if="tracksSaved" class="text-[0.8rem] text-success mt-2">saved</p>
            </template>
          </div>

          <footer class="flex items-center justify-between px-6 py-4 border-t border-border">
            <div class="flex gap-2 ml-auto">
              <!-- ⊹ ₊ ⟡ copy link for sharing with the dj -->
              <button
                type="button"
                class="px-4 py-[0.45rem] border border-border bg-background text-muted-foreground text-[0.8rem] cursor-pointer font-sans hover:bg-muted hover:text-foreground disabled:opacity-40 disabled:cursor-default"
                :disabled="linkLoading"
                @click="copySubmissionLink"
              >
                {{ linkCopied ? 'copied!' : linkLoading ? '…' : 'copy link' }}
              </button>
              <button
                type="button"
                class="px-4 py-[0.45rem] border border-border bg-background text-foreground text-[0.85rem] cursor-pointer font-sans hover:bg-muted disabled:opacity-40 disabled:cursor-default"
                @click="close"
              >cancel</button>
              <button
                type="button"
                class="px-4 py-[0.45rem] border border-primary bg-primary text-primary-foreground text-[0.85rem] cursor-pointer font-sans hover:opacity-85 disabled:opacity-40 disabled:cursor-default"
                :disabled="!canSaveTracks"
                @click="saveTracks"
              >
                {{ tracksSaving ? 'saving…' : 'save tracklist' }}
              </button>
            </div>
          </footer>
        </div>
      </TabsContent>
    </TabsRoot>
  </dialog>
</template>
