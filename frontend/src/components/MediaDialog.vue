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
  <dialog ref="dialogEl" class="border border-border p-0 w-[min(440px,90vw)] bg-background text-foreground backdrop:bg-[oklch(0_0_0/0.65)]">
    <form class="flex flex-col" @submit.prevent="submit" novalidate>
      <header class="flex items-center justify-between px-6 pt-5 pb-4 border-b border-border">
        <h2 class="text-base font-semibold m-0 text-foreground normal-case tracking-normal">{{ mode === 'create' ? 'add media' : 'edit media' }}</h2>
        <button type="button" class="bg-transparent border-0 cursor-pointer text-muted-foreground text-base p-1 leading-none hover:text-foreground" @click="close" aria-label="close">✕</button>
      </header>

      <div class="flex flex-col gap-4 px-6 py-5">
        <div class="flex flex-col gap-[0.3rem]">
          <label for="m-title" class="text-[0.8rem] text-muted-foreground font-medium">title</label>
          <input
            id="m-title"
            v-model="form.title"
            maxlength="200"
            class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border outline-none font-sans text-foreground bg-input focus:border-ring"
            :class="errors.title ? 'border-destructive' : 'border-border'"
          />
          <span v-if="errors.title" class="text-xs text-destructive">{{ errors.title }}</span>
        </div>

        <div class="flex flex-col gap-[0.3rem]">
          <label for="m-artist" class="text-[0.8rem] text-muted-foreground font-medium">artist</label>
          <input
            id="m-artist"
            v-model="form.artist"
            maxlength="200"
            class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border outline-none font-sans text-foreground bg-input focus:border-ring"
            :class="errors.artist ? 'border-destructive' : 'border-border'"
          />
          <span v-if="errors.artist" class="text-xs text-destructive">{{ errors.artist }}</span>
        </div>

        <div class="flex flex-col gap-[0.3rem]">
          <label for="m-format" class="text-[0.8rem] text-muted-foreground font-medium">format</label>
          <select
            id="m-format"
            v-model="form.fileFormat"
            class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border border-border outline-none font-sans text-foreground bg-input focus:border-ring"
          >
            <option value="">unknown</option>
            <option value="mp3">mp3</option>
            <option value="aac">aac</option>
            <option value="m4a">m4a</option>
          </select>
        </div>

        <template v-if="mode === 'create'">
          <div class="flex flex-col gap-[0.3rem]">
            <label for="m-adapter" class="text-[0.8rem] text-muted-foreground font-medium">source adapter</label>
            <input
              id="m-adapter"
              v-model="form.sourceAdapter"
              placeholder="e.g. url"
              class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border outline-none font-sans text-foreground bg-input focus:border-ring"
              :class="errors.sourceAdapter ? 'border-destructive' : 'border-border'"
            />
            <span v-if="errors.sourceAdapter" class="text-xs text-destructive">{{ errors.sourceAdapter }}</span>
          </div>

          <div class="flex flex-col gap-[0.3rem]">
            <label for="m-ref" class="text-[0.8rem] text-muted-foreground font-medium">source ref</label>
            <input
              id="m-ref"
              v-model="form.sourceRef"
              placeholder="e.g. https://..."
              class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border outline-none font-sans text-foreground bg-input focus:border-ring"
              :class="errors.sourceRef ? 'border-destructive' : 'border-border'"
            />
            <span v-if="errors.sourceRef" class="text-xs text-destructive">{{ errors.sourceRef }}</span>
          </div>
        </template>
      </div>

      <footer class="flex items-center justify-between px-6 py-4 border-t border-border">
        <button
          v-if="mode === 'edit'"
          type="button"
          class="px-4 py-[0.45rem] border border-destructive bg-background text-destructive text-[0.85rem] cursor-pointer font-sans hover:bg-destructive/10"
          @click="remove"
        >delete</button>
        <div class="flex gap-2 ml-auto">
          <button
            type="button"
            class="px-4 py-[0.45rem] border border-border bg-background text-foreground text-[0.85rem] cursor-pointer font-sans hover:bg-muted"
            @click="close"
          >cancel</button>
          <button
            type="submit"
            class="px-4 py-[0.45rem] border border-primary bg-primary text-primary-foreground text-[0.85rem] cursor-pointer font-sans hover:opacity-85"
          >{{ mode === 'create' ? 'add' : 'save' }}</button>
        </div>
      </footer>
    </form>
  </dialog>
</template>
