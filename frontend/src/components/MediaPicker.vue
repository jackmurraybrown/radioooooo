<script setup lang="ts">
// ✮ ⋆ ˚｡𖦹 ⋆｡°✩ searchable media picker — combobox + drag-drop upload
import { ref, computed, watch } from 'vue'
import {
  ComboboxRoot, ComboboxAnchor, ComboboxInput, ComboboxTrigger,
  ComboboxContent, ComboboxViewport,
  ComboboxItem, ComboboxItemIndicator, ComboboxEmpty,
} from 'reka-ui'
import { apiUpload, api } from '@/api/client'
import type { Media } from '@/api/types'

const props = defineProps<{
  modelValue: string
  media: Media[]
}>()

const emit = defineEmits<{
  'update:modelValue': [id: string]
  mediaAdded: [id: string, title: string]
}>()

const fileInputEl   = ref<HTMLInputElement>()
const titleInputEl  = ref<HTMLInputElement>()
const uploading     = ref(false)
const uploadError   = ref<string | null>(null)
const isDragOver    = ref(false)

// ⊹ ₊ ⟡ only shown right after an upload — clears when user picks from list
const justUploadedId     = ref<string | null>(null)
const uploadedTitle      = ref('')
const uploadedTitleSaved = ref(false)
const renameSaving       = ref(false)
const renameError        = ref<string | null>(null)

// ⋆˙⟡ bridge v-model to combobox
const selected = computed({
  get: () => props.modelValue,
  set: (val) => {
    const id = String(val ?? '')
    if (id !== justUploadedId.value) justUploadedId.value = null
    emit('update:modelValue', id)
  },
})

watch(() => props.modelValue, (id) => {
  if (id !== justUploadedId.value) justUploadedId.value = null
})

function displayValue(val: string): string {
  const m = props.media.find(x => x.id === val)
  return m ? `${m.title}${m.artist ? ` — ${m.artist}` : ''}` : ''
}

const renameDirty = computed(() =>
  uploadedTitle.value.trim() !== '' &&
  uploadedTitle.value.trim() !== props.media.find(m => m.id === justUploadedId.value)?.title
)

async function saveUploadedTitle() {
  if (!renameDirty.value || !justUploadedId.value) return
  renameSaving.value = true
  renameError.value  = null
  try {
    const res = await api(`/media/${justUploadedId.value}`).put({ title: uploadedTitle.value.trim() })
    if (!res.ok) throw new Error(`${res.status}`)
    uploadedTitleSaved.value = true
    setTimeout(() => { uploadedTitleSaved.value = false }, 2000)
    emit('mediaAdded', justUploadedId.value, uploadedTitle.value.trim())
  } catch {
    renameError.value = 'rename failed'
  } finally {
    renameSaving.value = false
  }
}

function triggerUpload() { fileInputEl.value?.click() }

async function uploadFiles(files: FileList | null) {
  if (!files || files.length === 0) return
  uploadError.value = null
  uploading.value   = true
  const fd = new FormData()
  for (const f of Array.from(files)) fd.append('files', f)
  try {
    const res = await apiUpload('/media/upload', fd)
    if (!res.ok) throw new Error(`${res.status}`)
    const data = await res.json()
    const first = (data.uploads as Array<{ id: string; title: string; error?: string }>).find(u => !u.error)
    if (first) {
      justUploadedId.value    = first.id
      uploadedTitle.value     = first.title
      uploadedTitleSaved.value = false
      emit('mediaAdded', first.id, first.title)
      emit('update:modelValue', first.id)
      // ✶. ₊ focus rename so user can fix the filename-derived title straight away
      setTimeout(() => { titleInputEl.value?.focus(); titleInputEl.value?.select() }, 50)
    } else {
      uploadError.value = (data.uploads[0] as any)?.error ?? 'upload failed'
    }
  } catch {
    uploadError.value = 'upload failed'
  } finally {
    uploading.value = false
    if (fileInputEl.value) fileInputEl.value.value = ''
  }
}

function onFileSelect(e: Event) { uploadFiles((e.target as HTMLInputElement).files) }
function onDragOver(e: DragEvent) { e.preventDefault(); isDragOver.value = true }
function onDragLeave(e: DragEvent) {
  if (!(e.currentTarget as HTMLElement).contains(e.relatedTarget as Node)) isDragOver.value = false
}
function onDrop(e: DragEvent) {
  e.preventDefault(); isDragOver.value = false; uploadFiles(e.dataTransfer?.files ?? null)
}
</script>

<template>
  <div
    class="flex flex-col gap-1.5 w-full relative"
    @dragover="onDragOver"
    @dragleave="onDragLeave"
    @drop="onDrop"
  >
    <!-- ⋆˙⟡ combobox + upload button -->
    <div class="flex w-full gap-1.5 items-stretch">
      <ComboboxRoot v-model="selected" class="combobox-root flex-1 min-w-0">
        <ComboboxAnchor
          class="flex items-center w-full border border-border bg-input transition-colors"
          :class="isDragOver ? 'border-ring' : 'focus-within:border-ring'"
        >
          <ComboboxInput
            class="flex-1 px-2.5 py-[0.45rem] text-sm text-foreground bg-transparent border-none outline-none min-w-0 font-sans"
            :display-value="displayValue"
            placeholder="search or select a track…"
          />
          <ComboboxTrigger
            class="self-stretch px-2 border-0 border-l border-border text-muted-foreground hover:text-foreground bg-transparent cursor-pointer text-xs flex items-center"
          >▾</ComboboxTrigger>
        </ComboboxAnchor>

        <!-- ⊹ ₊ ⟡ no portal — stays inside native dialog top layer
             positionStrategy=fixed escapes overflow-y:auto on .fields -->
        <ComboboxContent
          class="mp-content bg-background border border-border z-9999 overflow-hidden"
          :side-offset="2"
          position="popper"
          position-strategy="fixed"
          style="box-shadow: 0 4px 16px oklch(0 0 0 / 0.35)"
        >
          <ComboboxViewport class="p-1 overflow-y-auto max-h-50">
            <ComboboxEmpty class="px-3 py-2 text-sm text-muted-foreground">no media found</ComboboxEmpty>
            <ComboboxItem
              v-for="m in props.media"
              :key="m.id"
              :value="m.id"
              class="mp-item flex items-center gap-2 px-3 py-[0.45rem] text-sm cursor-pointer select-none text-foreground"
            >
              <span class="font-medium flex-1 truncate">{{ m.title }}</span>
              <span v-if="m.artist" class="text-xs text-muted-foreground truncate max-w-32">{{ m.artist }}</span>
              <ComboboxItemIndicator class="text-xs text-foreground ml-auto shrink-0">✓</ComboboxItemIndicator>
            </ComboboxItem>
          </ComboboxViewport>
        </ComboboxContent>
      </ComboboxRoot>

      <button
        type="button"
        class="px-3 py-[0.45rem] text-sm font-sans border border-border bg-transparent text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50 disabled:cursor-default cursor-pointer whitespace-nowrap shrink-0"
        :disabled="uploading"
        @click="triggerUpload"
      >{{ uploading ? '…' : 'upload' }}</button>
    </div>

    <span v-if="uploadError" class="text-xs text-destructive">{{ uploadError }}</span>

    <!-- ✮ ⋆ upload edit panel — only shown right after uploading a file -->
    <div v-if="justUploadedId" class="flex flex-col gap-1.5 p-2.5 border border-dashed border-border bg-muted">
      <span class="text-xs text-muted-foreground">just uploaded — edit name if needed</span>
      <div class="flex gap-1.5">
        <input
          ref="titleInputEl"
          v-model="uploadedTitle"
          class="flex-1 px-2.5 py-1.5 text-sm text-foreground bg-background border border-border outline-none min-w-0 focus:border-ring font-sans"
          placeholder="track name"
          maxlength="200"
          @keydown.enter.prevent="saveUploadedTitle"
        />
        <button
          type="button"
          class="px-2.5 py-1.5 text-sm font-sans border border-border bg-transparent text-muted-foreground hover:bg-foreground hover:text-background disabled:opacity-[0.35] disabled:cursor-default cursor-pointer shrink-0"
          :disabled="!renameDirty || renameSaving"
          @click="saveUploadedTitle"
        >{{ renameSaving ? '…' : uploadedTitleSaved ? 'saved' : 'save' }}</button>
      </div>
      <span v-if="renameError" class="text-xs text-destructive">{{ renameError }}</span>
    </div>

    <span v-if="isDragOver" class="text-xs text-ring">drop to upload (mp3 · aac · m4a)</span>

    <input ref="fileInputEl" type="file" accept=".mp3,.aac,.m4a" multiple class="hidden" @change="onFileSelect" />
  </div>
</template>

<style scoped>
/* ⋆˙⟡ reka injects PopperRoot > ListboxRoot between ComboboxRoot and ComboboxAnchor
   — force both wrapper divs to fill available width so the anchor stretches */
.combobox-root :deep(> *),
.combobox-root :deep(> * > *) {
  width: 100%;
}

/* ✶. ₊ dropdown width matches the anchor */
.mp-content {
  width: var(--reka-combobox-trigger-width);
  min-width: 220px;
}

.mp-item[data-highlighted] { background: var(--muted); }
.mp-item[data-state="checked"] { background: var(--muted); }
</style>
