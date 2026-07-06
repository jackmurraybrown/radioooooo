<script setup lang="ts">
// ✮⋆‧°—°‧⋆✮ station settings — profile + channels
import { ref, reactive, watch, onMounted } from 'vue'
import * as v from 'valibot'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
import { useValidation } from '@/composables/useValidation'
import type { Station, Channel } from '@/api/types'

const auth = useAuthStore()
const toast = useToast()

// ⊹ ₊ ⟡ station profile
const station = ref<Station | null>(null)
const loading = ref(true)
const saving = ref(false)
const fetchError = ref<string | null>(null)

const profileSchema = v.object({
  name:                v.pipe(v.string(), v.minLength(1, 'required'), v.maxLength(100)),
  slug:                v.pipe(v.string(), v.minLength(1, 'required'), v.maxLength(50), v.regex(/^[a-z0-9-]+$/, 'lowercase letters, numbers and hyphens only')),
  logoUrl:             v.optional(v.string()),
  tracklistWebhookUrl: v.optional(v.string()),
})

interface ProfileForm { name: string; slug: string; logoUrl: string; tracklistWebhookUrl: string }

const form = reactive<ProfileForm>({ name: '', slug: '', logoUrl: '', tracklistWebhookUrl: '' })
const { errors, validate } = useValidation(profileSchema)

function populate(st: Station) {
  form.name                = st.name
  form.slug                = st.slug
  form.logoUrl             = st.logoUrl ?? ''
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
      name:                form.name,
      slug:                form.slug,
      logoUrl:             form.logoUrl || undefined,
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
  <div class="settings-page">
    <h1>settings</h1>

    <p v-if="fetchError" class="error-msg">{{ fetchError }}</p>
    <div v-if="loading" class="empty">loading…</div>

    <template v-else>
      <!-- ✮ ⋆ ˚｡𖦹 station profile -->
      <form @submit.prevent="save" novalidate class="settings-form">
        <section>
          <h2>station profile</h2>

          <div class="field">
            <label for="s-logo">logo url</label>
            <div class="logo-row">
              <img v-if="form.logoUrl" :src="form.logoUrl" class="logo-preview" alt="station logo" />
              <div v-else class="logo-placeholder">no logo</div>
              <input id="s-logo" v-model="form.logoUrl" placeholder="https://…" class="logo-input" />
            </div>
          </div>

          <div class="field">
            <label for="s-name">name</label>
            <input id="s-name" v-model="form.name" :class="{ error: errors.name }" maxlength="100" />
            <span v-if="errors.name" class="err">{{ errors.name }}</span>
          </div>

          <div class="field">
            <label for="s-slug">
              slug
              <span class="label-hint">used in public urls — lowercase, hyphens only</span>
            </label>
            <input id="s-slug" v-model="form.slug" :class="{ error: errors.slug }" maxlength="50" />
            <span v-if="errors.slug" class="err">{{ errors.slug }}</span>
          </div>

          <div class="field">
            <label for="s-webhook">
              tracklist webhook url
              <span class="label-hint">receives tracklist data after each show</span>
            </label>
            <input id="s-webhook" v-model="form.tracklistWebhookUrl" placeholder="https://…" />
          </div>
        </section>

        <div class="form-footer">
          <button type="submit" class="primary" :disabled="saving">
            {{ saving ? 'saving…' : 'save changes' }}
          </button>
        </div>
      </form>

      <!-- ⊹ ₊ ⟡ channels -->
      <section>
        <h2>channels</h2>

        <div v-if="channelsLoading" class="empty">loading…</div>

        <ul v-else-if="channels.length > 0" class="channel-list">
          <li v-for="ch in channels" :key="ch.id">
            <span class="ch-name">{{ ch.name }}</span>
            <span class="ch-slug">{{ ch.slug }}</span>
            <button
              v-if="channels.length > 1"
              class="delete-btn"
              @click="promptDelete(ch.id, ch.name)"
              aria-label="delete channel"
            >✕</button>
          </li>
        </ul>
        <p v-else class="empty-inline">no channels yet</p>

        <!-- ˎˊ˗ inline create form -->
        <form @submit.prevent="createChannel" novalidate class="new-channel-form">
          <div class="new-channel-fields">
            <div class="field">
              <input
                v-model="newChannel.name"
                placeholder="channel name"
                :class="{ error: channelErrors.name }"
                maxlength="100"
              />
              <span v-if="channelErrors.name" class="err">{{ channelErrors.name }}</span>
            </div>
            <div class="field">
              <input
                v-model="newChannel.slug"
                placeholder="slug"
                :class="{ error: channelErrors.slug }"
                maxlength="50"
              />
              <span v-if="channelErrors.slug" class="err">{{ channelErrors.slug }}</span>
            </div>
          </div>
          <button type="submit" class="primary" :disabled="channelCreating">
            {{ channelCreating ? 'creating…' : 'add channel' }}
          </button>
        </form>
      </section>
    </template>
  </div>

  <!-- ⊹ ₊ ⟡ channel delete confirmation -->
  <dialog ref="confirmDialogEl" class="confirm-dialog">
    <p>delete <strong>{{ pendingDelete?.name }}</strong>? all its episodes will be removed.</p>
    <div class="confirm-actions">
      <button @click="confirmDialogEl?.close(); pendingDelete = null">cancel</button>
      <button class="danger" @click="confirmDelete">delete</button>
    </div>
  </dialog>
</template>

<style scoped>
.settings-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  max-width: 520px;
}

h1 { font-size: 1.1rem; font-weight: 600; margin: 0; }

h2 {
  font-size: 0.8rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--muted-foreground);
  margin: 0 0 1rem;
}

section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.25rem;
  border: 1px solid var(--border);
}

.field { display: flex; flex-direction: column; gap: 0.3rem; }

label {
  font-size: 0.8rem;
  color: var(--muted-foreground);
  font-weight: 500;
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
}

.label-hint { font-size: 0.73rem; color: var(--muted-foreground); font-weight: 400; opacity: 0.7; }

input {
  font-size: 0.9rem;
  padding: 0.45rem 0.6rem;
  border: 1px solid var(--border);
  outline: none;
  font-family: inherit;
  color: var(--foreground);
  background: var(--input);
}

input:focus { border-color: var(--ring); }

input.error { border-color: var(--destructive); }
.err { font-size: 0.75rem; color: var(--destructive); }

.logo-row { display: flex; align-items: center; gap: 0.75rem; }

/* TODO: support uploading logo directly, not just a url */
.logo-preview {
  width: 48px;
  height: 48px;
  object-fit: cover;
  border: 1px solid var(--border);
  flex-shrink: 0;
}

.logo-placeholder {
  width: 48px;
  height: 48px;
  border: 1px dashed var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.65rem;
  color: var(--muted-foreground);
  flex-shrink: 0;
}

.logo-input { flex: 1; }

.form-footer { display: flex; align-items: center; justify-content: flex-end; }

/* ⊹ ₊ ⟡ channel list */
.channel-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.channel-list li {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0.6rem;
  background: var(--muted);
  font-size: 0.88rem;
}

.ch-name { font-weight: 500; color: var(--foreground); flex: 1; }
.ch-slug { font-size: 0.78rem; color: var(--muted-foreground); }

.delete-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--muted-foreground);
  font-size: 0.75rem;
  padding: 0.15rem;
  line-height: 1;
  flex-shrink: 0;
  opacity: 0.5;
}

.delete-btn:hover { color: var(--destructive); opacity: 1; }

.new-channel-form {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  padding-top: 0.25rem;
  border-top: 1px solid var(--border);
}

.new-channel-fields { display: flex; gap: 0.5rem; flex: 1; }
.new-channel-fields .field { flex: 1; }

.empty-inline { font-size: 0.85rem; color: var(--muted-foreground); margin: 0; }

button.primary {
  padding: 0.45rem 1.25rem;
  border: none;
  background: var(--primary);
  color: var(--primary-foreground);
  font-size: 0.85rem;
  cursor: pointer;
  font-family: inherit;
  white-space: nowrap;
}

button.primary:hover { opacity: 0.85; }
button.primary:disabled { opacity: 0.5; cursor: default; }

.empty { font-size: 0.9rem; color: var(--muted-foreground); }
.error-msg { font-size: 0.9rem; color: var(--destructive); }

.confirm-dialog {
  border: 1px solid var(--border);
  padding: 1.5rem;
  width: min(360px, 90vw);
  background: var(--background);
  color: var(--foreground);
}

.confirm-dialog::backdrop { background: oklch(0 0 0 / 0.65); }
.confirm-dialog p { margin: 0 0 1.25rem; font-size: 0.9rem; }

.confirm-actions { display: flex; justify-content: flex-end; gap: 0.5rem; }

button.danger {
  padding: 0.45rem 1.25rem;
  border: none;
  background: var(--destructive);
  color: var(--foreground);
  font-size: 0.85rem;
  cursor: pointer;
  font-family: inherit;
}

button.danger:hover { opacity: 0.85; }
</style>
