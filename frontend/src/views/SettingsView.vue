<script setup lang="ts">
// ✮⋆‧°—°‧⋆✮ station settings — profile + channels
import { ref, reactive, watch, onMounted } from 'vue'
import { useCalendarPrefs } from '@/composables/useCalendarPrefs'
import * as v from 'valibot'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
import { useValidation } from '@/composables/useValidation'
import type { Station, Channel } from '@/api/types'

const auth = useAuthStore()
const toast = useToast()
const { prefs: calPrefs } = useCalendarPrefs()

// ⊹ ₊ ⟡ station profile
const station = ref<Station | null>(null)
const loading = ref(true)
const saving = ref(false)
const fetchError = ref<string | null>(null)

const profileSchema = v.object({
  name: v.pipe(v.string(), v.minLength(1, 'required'), v.maxLength(100)),
  slug: v.pipe(v.string(), v.minLength(1, 'required'), v.maxLength(50), v.regex(/^[a-z0-9-]+$/, 'lowercase letters, numbers and hyphens only')),
  logoUrl: v.optional(v.string()),
  tracklistWebhookUrl: v.optional(v.string()),
})

interface ProfileForm { name: string; slug: string; logoUrl: string; tracklistWebhookUrl: string }

const form = reactive<ProfileForm>({ name: '', slug: '', logoUrl: '', tracklistWebhookUrl: '' })
const { errors, validate } = useValidation(profileSchema)

function populate(st: Station) {
  form.name = st.name
  form.slug = st.slug
  form.logoUrl = st.logoUrl ?? ''
  form.tracklistWebhookUrl = st.tracklistWebhookUrl ?? ''
}

async function fetchStation() {
  if (!auth.stationId) return
  loading.value = true
  fetchError.value = null
  try {
    const res = await api(`/stations/${auth.stationId}`).get()
    if (!res.ok) throw new Error(`${res.status}`)
    station.value = await res.json()
    populate(station.value!)
  } catch {
    fetchError.value = 'failed to load station'
  } finally {
    loading.value = false
  }
}

async function save() {
  if (!validate({ ...form })) return
  saving.value = true
  try {
    const res = await api(`/stations/${auth.stationId}`).put({
      name: form.name,
      slug: form.slug,
      logoUrl: form.logoUrl || undefined,
      tracklistWebhookUrl: form.tracklistWebhookUrl || undefined,
    })
    if (!res.ok) throw new Error(`${res.status}`)
    station.value = await res.json()
    populate(station.value!)
    toast.success('saved')
  } catch {
    toast.error('failed to save changes')
  } finally {
    saving.value = false
  }
}

// ✶. ݁ channels
const channels = ref<Channel[]>([])
const channelsLoading = ref(false)
const channelCreating = ref(false)
const confirmDialogEl = ref<HTMLDialogElement>()
const pendingDelete = ref<{ id: string; name: string } | null>(null)

const channelSchema = v.object({
  name: v.pipe(v.string(), v.minLength(1, 'required'), v.maxLength(100)),
  slug: v.pipe(v.string(), v.minLength(1, 'required'), v.maxLength(50), v.regex(/^[a-z0-9-]+$/, 'lowercase letters, numbers and hyphens only')),
})

const newChannel = reactive({ name: '', slug: '' })
const { errors: channelErrors, validate: validateChannel } = useValidation(channelSchema)

// ⋆˙⟡ auto-derive slug from name
watch(() => newChannel.name, (val) => {
  newChannel.slug = val.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')
})

async function fetchChannels() {
  channelsLoading.value = true
  try {
    const res = await api('/channels').get()
    if (!res.ok) throw new Error(`${res.status}`)
    const data = await res.json()
    channels.value = data.channels ?? []
  } catch {
    toast.error('failed to load channels')
  } finally {
    channelsLoading.value = false
  }
}

async function createChannel() {
  if (!validateChannel({ ...newChannel })) return
  channelCreating.value = true
  try {
    const res = await api('/channels').post({ name: newChannel.name, slug: newChannel.slug })
    if (!res.ok) {
      const body = await res.json().catch(() => ({}))
      throw new Error(body?.detail ?? `${res.status}`)
    }
    const ch: Channel = await res.json()
    channels.value.push(ch)
    newChannel.name = ''
    newChannel.slug = ''
    toast.success(`channel "${ch.name}" created`)
  } catch (e) {
    toast.error(e instanceof Error ? e.message : 'failed to create channel')
  } finally {
    channelCreating.value = false
  }
}

function promptDelete(id: string, name: string) {
  pendingDelete.value = { id, name }
  confirmDialogEl.value?.showModal()
}

async function confirmDelete() {
  if (!pendingDelete.value) return
  const { id, name } = pendingDelete.value
  confirmDialogEl.value?.close()
  try {
    const res = await api(`/channels/${id}`).delete()
    if (!res.ok) {
      const body = await res.json().catch(() => ({}))
      throw new Error(body?.detail ?? `${res.status}`)
    }
    channels.value = channels.value.filter(ch => ch.id !== id)
    toast.success(`channel "${name}" deleted`)
  } catch (e) {
    toast.error(e instanceof Error ? e.message : 'failed to delete channel')
  } finally {
    pendingDelete.value = null
  }
}

onMounted(() => {
  fetchStation()
  fetchChannels()
})
</script>

<template>
  <div class="flex flex-col gap-6 p-8 max-w-130">
    <h1>settings</h1>

    <p v-if="fetchError" class="text-[0.9rem] text-destructive">{{ fetchError }}</p>
    <div v-if="loading" class="text-[0.9rem] text-muted-foreground">loading…</div>

    <template v-else>
      <!-- ✮ ⋆ ˚｡𖦹 station profile -->
      <form @submit.prevent="save" novalidate class="flex flex-col gap-4">
        <section class="flex flex-col gap-4 p-5 border border-border">
          <h2 class="mb-4">station profile</h2>

          <div class="flex flex-col gap-[0.3rem]">
            <label for="s-logo" class="text-[0.8rem] text-muted-foreground font-medium flex items-baseline gap-2">logo url</label>
            <!-- ⊹ ₊ ⟡ TODO: support uploading logo directly, not just a url -->
            <div class="flex items-center gap-3">
              <img v-if="form.logoUrl" :src="form.logoUrl" class="w-12 h-12 object-cover border border-border shrink-0" alt="station logo" />
              <div v-else class="w-12 h-12 border border-dashed border-border flex items-center justify-center text-[0.65rem] text-muted-foreground shrink-0">no logo</div>
              <input
                id="s-logo"
                v-model="form.logoUrl"
                placeholder="https://…"
                class="flex-1 text-[0.9rem] px-[0.6rem] py-[0.45rem] border border-border outline-none font-sans text-foreground bg-input focus:border-ring"
              />
            </div>
          </div>

          <div class="flex flex-col gap-[0.3rem]">
            <label for="s-name" class="text-[0.8rem] text-muted-foreground font-medium flex items-baseline gap-2">name</label>
            <input
              id="s-name"
              v-model="form.name"
              maxlength="100"
              class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border outline-none font-sans text-foreground bg-input focus:border-ring"
              :class="errors.name ? 'border-destructive' : 'border-border'"
            />
            <span v-if="errors.name" class="text-xs text-destructive">{{ errors.name }}</span>
          </div>

          <div class="flex flex-col gap-[0.3rem]">
            <label for="s-slug" class="text-[0.8rem] text-muted-foreground font-medium flex items-baseline gap-2">
              slug
              <span class="text-[0.73rem] text-muted-foreground font-normal opacity-70">used in public urls — lowercase, hyphens only</span>
            </label>
            <input
              id="s-slug"
              v-model="form.slug"
              maxlength="50"
              class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border outline-none font-sans text-foreground bg-input focus:border-ring"
              :class="errors.slug ? 'border-destructive' : 'border-border'"
            />
            <span v-if="errors.slug" class="text-xs text-destructive">{{ errors.slug }}</span>
          </div>

          <div class="flex flex-col gap-[0.3rem]">
            <label for="s-webhook" class="text-[0.8rem] text-muted-foreground font-medium flex items-baseline gap-2">
              tracklist webhook url
              <span class="text-[0.73rem] text-muted-foreground font-normal opacity-70">receives tracklist data after each show</span>
            </label>
            <input
              id="s-webhook"
              v-model="form.tracklistWebhookUrl"
              placeholder="https://…"
              class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border border-border outline-none font-sans text-foreground bg-input focus:border-ring"
            />
          </div>
        </section>

        <div class="flex items-center justify-end">
          <button
            type="submit"
            :disabled="saving"
            class="px-5 py-[0.45rem] border-0 bg-primary text-primary-foreground text-[0.85rem] cursor-pointer font-sans whitespace-nowrap hover:opacity-85 disabled:opacity-50 disabled:cursor-default"
          >
            {{ saving ? 'saving…' : 'save changes' }}
          </button>
        </div>
      </form>

      <!-- ⊹ ₊ ⟡ channels -->
      <section class="flex flex-col gap-4 p-5 border border-border">
        <h2 class="mb-4">channels</h2>

        <div v-if="channelsLoading" class="text-[0.9rem] text-muted-foreground">loading…</div>

        <ul v-else-if="channels.length > 0" class="list-none m-0 p-0 flex flex-col gap-1">
          <li v-for="ch in channels" :key="ch.id" class="flex items-center gap-3 px-[0.6rem] py-2 bg-muted text-[0.88rem]">
            <span class="font-medium text-foreground flex-1">{{ ch.name }}</span>
            <span class="text-[0.78rem] text-muted-foreground">{{ ch.slug }}</span>
            <button
              v-if="channels.length > 1"
              class="bg-transparent border-0 cursor-pointer text-muted-foreground text-xs p-[0.15rem] leading-none shrink-0 opacity-50 hover:text-destructive hover:opacity-100"
              @click="promptDelete(ch.id, ch.name)"
              aria-label="delete channel"
            >✕</button>
          </li>
        </ul>
        <p v-else class="text-[0.85rem] text-muted-foreground m-0">no channels yet</p>

        <!-- ˎˊ˗ inline create form -->
        <form @submit.prevent="createChannel" novalidate class="flex items-start gap-2 pt-1 border-t border-border">
          <div class="flex gap-2 flex-1">
            <div class="flex flex-col gap-[0.3rem] flex-1">
              <input
                v-model="newChannel.name"
                placeholder="channel name"
                maxlength="100"
                class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border outline-none font-sans text-foreground bg-input focus:border-ring"
                :class="channelErrors.name ? 'border-destructive' : 'border-border'"
              />
              <span v-if="channelErrors.name" class="text-xs text-destructive">{{ channelErrors.name }}</span>
            </div>
            <div class="flex flex-col gap-[0.3rem] flex-1">
              <input
                v-model="newChannel.slug"
                placeholder="slug"
                maxlength="50"
                class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border outline-none font-sans text-foreground bg-input focus:border-ring"
                :class="channelErrors.slug ? 'border-destructive' : 'border-border'"
              />
              <span v-if="channelErrors.slug" class="text-xs text-destructive">{{ channelErrors.slug }}</span>
            </div>
          </div>
          <button
            type="submit"
            :disabled="channelCreating"
            class="px-5 py-[0.45rem] border-0 bg-primary text-primary-foreground text-[0.85rem] cursor-pointer font-sans whitespace-nowrap hover:opacity-85 disabled:opacity-50 disabled:cursor-default"
          >
            {{ channelCreating ? 'creating…' : 'add channel' }}
          </button>
        </form>
      </section>

      <!-- calendar display preferences — stored locally -->
      <section class="flex flex-col gap-4 p-5 border border-border">
        <h2 class="mb-4">calendar</h2>
        <div class="flex flex-col gap-[0.3rem]">
          <label class="text-[0.8rem] text-muted-foreground font-medium flex items-baseline gap-2">start of week</label>
          <select v-model.number="calPrefs.firstDay">
            <option :value="1">monday</option>
            <option :value="0">sunday</option>
          </select>
        </div>
        <div class="flex flex-col gap-[0.3rem]">
          <label class="text-[0.8rem] text-muted-foreground font-medium flex items-baseline gap-2">scroll to time</label>
          <input
            type="time"
            v-model="calPrefs.scrollTime"
            step="1800"
            class="border border-border bg-input text-foreground font-sans text-[0.85rem] px-2 py-[0.3rem] scheme-dark"
          />
        </div>
        <div class="flex flex-col gap-[0.3rem]">
          <label class="text-[0.8rem] text-muted-foreground font-medium flex items-baseline gap-2">slot duration</label>
          <select v-model="calPrefs.slotDuration">
            <option value="00:15:00">15 minutes</option>
            <option value="00:30:00">30 minutes</option>
            <option value="01:00:00">1 hour</option>
          </select>
        </div>
        <p class="text-xs text-muted-foreground m-0">changes apply immediately on the schedule page.</p>
      </section>
    </template>
  </div>

  <!-- ⊹ ₊ ⟡ channel delete confirmation -->
  <dialog
    ref="confirmDialogEl"
    class="border border-border p-6 w-[min(360px,90vw)] bg-background text-foreground backdrop:bg-backdrop"
  >
    <p class="mb-5 text-[0.9rem]">delete <strong>{{ pendingDelete?.name }}</strong>? all its episodes will be removed.</p>
    <div class="flex justify-end gap-2">
      <button
        class="px-4 py-[0.45rem] border border-border bg-background text-foreground text-[0.85rem] cursor-pointer font-sans hover:bg-muted"
        @click="confirmDialogEl?.close(); pendingDelete = null"
      >cancel</button>
      <button
        class="px-5 py-[0.45rem] border-0 bg-destructive text-foreground text-[0.85rem] cursor-pointer font-sans hover:opacity-85"
        @click="confirmDelete"
      >delete</button>
    </div>
  </dialog>
</template>
