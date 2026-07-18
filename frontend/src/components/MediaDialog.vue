<script setup lang="ts">
// ⊹ ࣪ ˖ media create / edit dialog
import { ref, reactive } from 'vue'
import * as v from 'valibot'
import type { Media, MediaCreateBody, MediaUpdateBody } from '@/api/types'

const emit = defineEmits<{
  create: [body: Omit<MediaCreateBody, '$schema'>]
  update: [id: string, body: Omit<MediaUpdateBody, '$schema'>]
  delete: [id: string]
}>()

const dialogEl = ref<HTMLDialogElement>()
const mode = ref<'create' | 'edit'>('create')
const currentId = ref('')

const createSchema = v.object({
  title:         v.pipe(v.string(), v.minLength(1, 'required'), v.maxLength(200)),
  artist:        v.optional(v.pipe(v.string(), v.maxLength(200))),
  sourceAdapter: v.pipe(v.string(), v.minLength(1, 'required')),
  sourceRef:     v.pipe(v.string(), v.minLength(1, 'required')),
  fileFormat:    v.optional(v.picklist(['mp3', 'aac', 'm4a'] as const)),
})

const updateSchema = v.object({
  title:      v.pipe(v.string(), v.minLength(1, 'required'), v.maxLength(200)),
  artist:     v.optional(v.pipe(v.string(), v.maxLength(200))),
  fileFormat: v.optional(v.picklist(['mp3', 'aac', 'm4a'] as const)),
})

interface FormState {
  title:         string
  artist:        string
  sourceAdapter: string
  sourceRef:     string
  fileFormat:    '' | 'mp3' | 'aac' | 'm4a'
}

type FormErrors = Partial<Record<keyof FormState, string>>

const form = reactive<FormState>({
  title:         '',
  artist:        '',
  sourceAdapter: '',
  sourceRef:     '',
  fileFormat:    '',
})

const errors = reactive<FormErrors>({})

function clearErrors() {
  Object.keys(errors).forEach(k => delete (errors as Record<string, string>)[k])
}

function reset(partial: Partial<FormState>) {
  form.title         = partial.title         ?? ''
  form.artist        = partial.artist        ?? ''
  form.sourceAdapter = partial.sourceAdapter ?? ''
  form.sourceRef     = partial.sourceRef     ?? ''
  form.fileFormat    = partial.fileFormat    ?? ''
  clearErrors()
}

function openCreate() {
  mode.value      = 'create'
  currentId.value = ''
  reset({})
  dialogEl.value?.showModal()
}

function openEdit(item: Media) {
  mode.value      = 'edit'
  currentId.value = item.id
  reset({
    title:         item.title,
    artist:        item.artist,
    sourceAdapter: item.sourceAdapter,
    sourceRef:     item.sourceRef,
    fileFormat:    (item.fileFormat as FormState['fileFormat']) ?? '',
  })
  dialogEl.value?.showModal()
}

function close() {
  dialogEl.value?.close()
}

function validate<T>(schema: v.BaseSchema<unknown, T, v.BaseIssue<unknown>>): T | null {
  clearErrors()
  const result = v.safeParse(schema, { ...form })
  if (!result.success) {
    for (const issue of result.issues) {
      const key = issue.path?.[0]?.key as keyof FormState
      if (key && !errors[key]) errors[key] = issue.message
    }
    return null
  }
  return result.output
}

function submit() {
  if (mode.value === 'create') {
    const data = validate(createSchema)
    if (!data) return
    emit('create', {
      title:         data.title,
      sourceAdapter: data.sourceAdapter,
      sourceRef:     data.sourceRef,
      ...(data.artist     ? { artist:     data.artist }     : {}),
      ...(data.fileFormat ? { fileFormat: data.fileFormat } : {}),
    })
  } else {
    const data = validate(updateSchema)
    if (!data) return
    emit('update', currentId.value, {
      title:      data.title,
      ...(data.artist     ? { artist:     data.artist }     : {}),
      ...(data.fileFormat ? { fileFormat: data.fileFormat } : {}),
    })
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
        <h2>{{ mode === 'create' ? 'add media' : 'edit media' }}</h2>
        <button type="button" class="close-btn" @click="close" aria-label="close">✕</button>
      </header>

      <div class="fields">
        <div class="field">
          <label for="m-title">title</label>
          <input id="m-title" v-model="form.title" :class="{ error: errors.title }" maxlength="200" />
          <span v-if="errors.title" class="err">{{ errors.title }}</span>
        </div>

        <div class="field">
          <label for="m-artist">artist</label>
          <input id="m-artist" v-model="form.artist" :class="{ error: errors.artist }" maxlength="200" />
          <span v-if="errors.artist" class="err">{{ errors.artist }}</span>
        </div>

        <div class="field">
          <label for="m-format">format</label>
          <select id="m-format" v-model="form.fileFormat">
            <option value="">unknown</option>
            <option value="mp3">mp3</option>
            <option value="aac">aac</option>
            <option value="m4a">m4a</option>
          </select>
        </div>

        <template v-if="mode === 'create'">
          <div class="field">
            <label for="m-adapter">source adapter</label>
            <input id="m-adapter" v-model="form.sourceAdapter" :class="{ error: errors.sourceAdapter }" placeholder="e.g. url" />
            <span v-if="errors.sourceAdapter" class="err">{{ errors.sourceAdapter }}</span>
          </div>

          <div class="field">
            <label for="m-ref">source ref</label>
            <input id="m-ref" v-model="form.sourceRef" :class="{ error: errors.sourceRef }" placeholder="e.g. https://..." />
            <span v-if="errors.sourceRef" class="err">{{ errors.sourceRef }}</span>
          </div>
        </template>
      </div>

      <footer>
        <button v-if="mode === 'edit'" type="button" class="delete-btn" @click="remove">delete</button>
        <div class="actions">
          <button type="button" @click="close">cancel</button>
          <button type="submit" class="primary">{{ mode === 'create' ? 'add' : 'save' }}</button>
        </div>
      </footer>
    </form>
  </dialog>
</template>

<style scoped>
dialog {
  border: 1px solid var(--border);
  padding: 0;
  width: min(440px, 90vw);
  background: var(--background);
  color: var(--foreground);
}

dialog::backdrop { background: oklch(0 0 0 / 0.65); }

form { display: flex; flex-direction: column; }

header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.25rem 1.5rem 1rem;
  border-bottom: 1px solid var(--border);
}

h2 { font-size: 1rem; font-weight: 600; margin: 0; }

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--muted-foreground);
  font-size: 1rem;
  padding: 0.25rem;
  line-height: 1;
}

.close-btn:hover { color: var(--foreground); }

.fields {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
}

.field { display: flex; flex-direction: column; gap: 0.3rem; }

label {
  font-size: 0.8rem;
  color: var(--muted-foreground);
  font-weight: 500;
}

input,
select {
  font-size: 0.9rem;
  padding: 0.45rem 0.6rem;
  border: 1px solid var(--border);
  outline: none;
  font-family: inherit;
  color: var(--foreground);
  background: var(--input);
}

input:focus,
select:focus { border-color: var(--ring); }

input.error { border-color: var(--destructive); }

.err { font-size: 0.75rem; color: var(--destructive); }

footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border);
}

.actions { display: flex; gap: 0.5rem; margin-left: auto; }

button {
  padding: 0.45rem 1rem;
  border: 1px solid var(--border);
  background: var(--background);
  color: var(--foreground);
  font-size: 0.85rem;
  cursor: pointer;
  font-family: inherit;
}

button:hover { background: var(--muted); }

button.primary { background: var(--primary); color: var(--primary-foreground); border-color: var(--primary); }
button.primary:hover { opacity: 0.85; background: var(--primary); }
button.delete-btn { color: var(--destructive); border-color: var(--destructive); }
button.delete-btn:hover { background: color-mix(in oklch, var(--destructive) 10%, transparent); }
</style>
