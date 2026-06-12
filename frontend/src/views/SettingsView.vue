<script setup lang="ts">
// ✮⋆‧°—°‧⋆✮ station settings — profile + logo
import { ref, reactive, onMounted } from 'vue'
import * as v from 'valibot'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import type { Station } from '@/api/types'

const auth = useAuthStore()
const station = ref<Station | null>(null)
const loading = ref(true)
const saving = ref(false)
const saved = ref(false)
const fetchError = ref<string | null>(null)

const schema = v.object({
  name:    v.pipe(v.string(), v.minLength(1, 'required'), v.maxLength(100)),
  slug:    v.pipe(v.string(), v.minLength(1, 'required'), v.maxLength(50), v.regex(/^[a-z0-9-]+$/, 'lowercase letters, numbers and hyphens only')),
  logoUrl: v.optional(v.string()),
})

interface FormState { name: string; slug: string; logoUrl: string }
type FormErrors = Partial<Record<keyof FormState, string>>

const form = reactive<FormState>({ name: '', slug: '', logoUrl: '' })
const errors = reactive<FormErrors>({})

function clearErrors() {
  Object.keys(errors).forEach(k => delete (errors as Record<string, string>)[k])
}

function populate(st: Station) {
  form.name    = st.name
  form.slug    = st.slug
  form.logoUrl = st.logoUrl ?? ''
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
  } catch (e) {
    fetchError.value = 'failed to load station'
  } finally {
    loading.value = false
  }
}

async function save() {
  clearErrors()
  const result = v.safeParse(schema, { ...form })
  if (!result.success) {
    for (const issue of result.issues) {
      const key = issue.path?.[0]?.key as keyof FormState
      if (key && !errors[key]) errors[key] = issue.message
    }
    return
  }
  saving.value = true
  saved.value  = false
  try {
    const res = await api(`/stations/${auth.stationId}`).put({
      name:    form.name,
      slug:    form.slug,
      logoUrl: form.logoUrl || undefined,
    })
    if (!res.ok) throw new Error(`${res.status}`)
    station.value = await res.json()
    populate(station.value!)
    saved.value = true
    setTimeout(() => { saved.value = false }, 2500)
  } catch {
    fetchError.value = 'failed to save changes'
  } finally {
    saving.value = false
  }
}

onMounted(fetchStation)
</script>

<template>
  <div class="settings-page">
    <h1>settings</h1>

    <p v-if="fetchError" class="error-msg">{{ fetchError }}</p>
    <div v-if="loading" class="empty">loading…</div>

    <form v-else @submit.prevent="save" novalidate class="settings-form">
      <section>
        <h2>station profile</h2>

        <div class="field">
          <label for="s-logo">logo url</label>
          <div class="logo-row">
            <img
              v-if="form.logoUrl"
              :src="form.logoUrl"
              class="logo-preview"
              alt="station logo"
            />
            <div v-else class="logo-placeholder">no logo</div>
            <input
              id="s-logo"
              v-model="form.logoUrl"
              placeholder="https://…"
              class="logo-input"
            />
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
      </section>

      <div class="form-footer">
        <span v-if="saved" class="saved-msg">saved</span>
        <button type="submit" class="primary" :disabled="saving">
          {{ saving ? 'saving…' : 'save changes' }}
        </button>
      </div>
    </form>
  </div>
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
  color: #9ca3af;
  margin: 0 0 1rem;
}

section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.25rem;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.field { display: flex; flex-direction: column; gap: 0.3rem; }

label {
  font-size: 0.8rem;
  color: #6b7280;
  font-weight: 500;
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
}

.label-hint { font-size: 0.73rem; color: #9ca3af; font-weight: 400; }

input {
  font-size: 0.9rem;
  padding: 0.45rem 0.6rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  outline: none;
  font-family: inherit;
  color: #111827;
  background: #fff;
}

input:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99,102,241,0.15);
}

input.error { border-color: #dc2626; }
.err { font-size: 0.75rem; color: #dc2626; }

.logo-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.logo-preview {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  object-fit: cover;
  border: 1px solid #e5e7eb;
  flex-shrink: 0;
}

.logo-placeholder {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  border: 1px dashed #d1d5db;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.65rem;
  color: #9ca3af;
  flex-shrink: 0;
}

.logo-input { flex: 1; }

.form-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
}

.saved-msg { font-size: 0.85rem; color: #16a34a; }

button.primary {
  padding: 0.45rem 1.25rem;
  border-radius: 6px;
  border: none;
  background: #111827;
  color: #fff;
  font-size: 0.85rem;
  cursor: pointer;
  font-family: inherit;
}

button.primary:hover { background: #374151; }
button.primary:disabled { opacity: 0.5; cursor: default; }

.empty { font-size: 0.9rem; color: #9ca3af; }
.error-msg { font-size: 0.9rem; color: #dc2626; }
</style>
