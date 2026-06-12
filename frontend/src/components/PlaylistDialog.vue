<script setup lang="ts">
// ✮⋆‧°—°‧⋆✮ playlist create / edit
import { ref, reactive } from 'vue'
import * as v from 'valibot'
import type { Playlist, PlaylistCreateBody, PlaylistUpdateBody } from '@/api/types'

const emit = defineEmits<{
  create: [body: Omit<PlaylistCreateBody, '$schema'>]
  update: [id: string, body: Omit<PlaylistUpdateBody, '$schema'>]
  delete: [id: string]
}>()

const dialogEl = ref<HTMLDialogElement>()
const mode = ref<'create' | 'edit'>('create')
const currentId = ref('')

const schema = v.object({
  name: v.pipe(v.string(), v.minLength(1, 'required'), v.maxLength(200)),
})

interface FormState {
  name:    string
  shuffle: boolean
  loop:    boolean
}

const form = reactive<FormState>({ name: '', shuffle: false, loop: false })
const nameError = ref('')

function reset(partial: Partial<FormState>) {
  form.name    = partial.name    ?? ''
  form.shuffle = partial.shuffle ?? false
  form.loop    = partial.loop    ?? false
  nameError.value = ''
}

function openCreate() {
  mode.value      = 'create'
  currentId.value = ''
  reset({})
  dialogEl.value?.showModal()
}

function openEdit(pl: Playlist) {
  mode.value      = 'edit'
  currentId.value = pl.id
  reset({ name: pl.name, shuffle: pl.shuffle, loop: pl.loop })
  dialogEl.value?.showModal()
}

function close() {
  dialogEl.value?.close()
}

function submit() {
  nameError.value = ''
  const result = v.safeParse(schema, { name: form.name })
  if (!result.success) {
    nameError.value = result.issues[0].message
    return
  }
  const body = { name: form.name, shuffle: form.shuffle, loop: form.loop }
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
        <h2>{{ mode === 'create' ? 'new playlist' : 'edit playlist' }}</h2>
        <button type="button" class="close-btn" @click="close" aria-label="close">✕</button>
      </header>

      <div class="fields">
        <div class="field">
          <label for="pl-name">name</label>
          <input id="pl-name" v-model="form.name" :class="{ error: nameError }" maxlength="200" />
          <span v-if="nameError" class="err">{{ nameError }}</span>
        </div>

        <div class="toggles">
          <label class="toggle">
            <input type="checkbox" v-model="form.shuffle" />
            shuffle
          </label>
          <label class="toggle">
            <input type="checkbox" v-model="form.loop" />
            loop
          </label>
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
  </dialog>
</template>

<style scoped>
dialog {
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  padding: 0;
  width: min(380px, 90vw);
  box-shadow: 0 8px 32px rgba(0,0,0,0.12);
}

dialog::backdrop { background: rgba(0,0,0,0.35); }

form { display: flex; flex-direction: column; }

header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.25rem 1.5rem 1rem;
  border-bottom: 1px solid #f3f4f6;
}

h2 { font-size: 1rem; font-weight: 600; margin: 0; }

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

.field { display: flex; flex-direction: column; gap: 0.3rem; }

label {
  font-size: 0.8rem;
  color: #6b7280;
  font-weight: 500;
}

input[type="text"],
input:not([type="checkbox"]) {
  font-size: 0.9rem;
  padding: 0.45rem 0.6rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  outline: none;
  font-family: inherit;
  color: #111827;
  background: #fff;
}

input:not([type="checkbox"]):focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99,102,241,0.15);
}

input.error { border-color: #dc2626; }

.err { font-size: 0.75rem; color: #dc2626; }

.toggles {
  display: flex;
  gap: 1.5rem;
}

.toggle {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-direction: row;
  cursor: pointer;
  color: #374151;
  font-size: 0.88rem;
}

footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  border-top: 1px solid #f3f4f6;
}

.actions { display: flex; gap: 0.5rem; margin-left: auto; }

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

button.primary { background: #111827; color: #fff; border-color: #111827; }
button.primary:hover { background: #374151; }
button.delete-btn { color: #dc2626; border-color: #fca5a5; }
button.delete-btn:hover { background: #fef2f2; }
</style>
