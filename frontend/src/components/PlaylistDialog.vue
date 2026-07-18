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
  <dialog ref="dialogEl" class="border border-border p-0 w-[min(380px,90vw)] bg-background text-foreground backdrop:bg-backdrop">
    <form class="flex flex-col" @submit.prevent="submit" novalidate>
      <header class="flex items-center justify-between px-6 pt-5 pb-4 border-b border-border">
        <h2 class="text-base font-semibold m-0 normal-case tracking-normal text-foreground">{{ mode === 'create' ? 'new playlist' : 'edit playlist' }}</h2>
        <button type="button" class="bg-transparent border-0 cursor-pointer text-muted-foreground text-base p-1 leading-none hover:text-foreground" @click="close" aria-label="close">✕</button>
      </header>

      <div class="flex flex-col gap-4 px-6 py-5">
        <div class="flex flex-col gap-[0.3rem]">
          <label for="pl-name" class="text-[0.8rem] text-muted-foreground font-medium">name</label>
          <input
            id="pl-name"
            v-model="form.name"
            maxlength="200"
            class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border outline-none font-sans text-foreground bg-input focus:border-ring"
            :class="nameError ? 'border-destructive' : 'border-border'"
          />
          <span v-if="nameError" class="text-xs text-destructive">{{ nameError }}</span>
        </div>

        <div class="flex gap-6">
          <label class="flex items-center gap-[0.4rem] flex-row cursor-pointer text-foreground text-[0.88rem]">
            <input type="checkbox" v-model="form.shuffle" />
            shuffle
          </label>
          <label class="flex items-center gap-[0.4rem] flex-row cursor-pointer text-foreground text-[0.88rem]">
            <input type="checkbox" v-model="form.loop" />
            loop
          </label>
        </div>
      </div>

      <footer class="flex items-center justify-between px-6 py-4 border-t border-border">
        <button v-if="mode === 'edit'" type="button" class="px-4 py-[0.45rem] border border-destructive bg-background text-destructive text-[0.85rem] cursor-pointer font-sans hover:bg-destructive/10" @click="remove">delete</button>
        <div class="flex gap-2 ml-auto">
          <button type="button" class="px-4 py-[0.45rem] border border-border bg-background text-foreground text-[0.85rem] cursor-pointer font-sans hover:bg-muted" @click="close">cancel</button>
          <button type="submit" class="px-4 py-[0.45rem] border border-primary bg-primary text-primary-foreground text-[0.85rem] cursor-pointer font-sans hover:opacity-85">{{ mode === 'create' ? 'create' : 'save' }}</button>
        </div>
      </footer>
    </form>
  </dialog>
</template>
