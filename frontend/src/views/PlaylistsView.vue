<script setup lang="ts">
// ⊹ ࣪ ˖ playlists — master list + item detail panel
import { ref, onMounted } from 'vue'
import PlaylistDialog from '@/components/PlaylistDialog.vue'
import MediaPicker from '@/components/MediaPicker.vue'
import { usePlaylists } from '@/composables/usePlaylists'
import { useMedia } from '@/composables/useMedia'
import { useToast } from '@/composables/useToast'
import type { Playlist, PlaylistCreateBody, PlaylistUpdateBody } from '@/api/types'

const dialogEl = ref<InstanceType<typeof PlaylistDialog>>()
const active = ref<Playlist | null>(null)
const addMediaId = ref('')

const {
  playlists, activeItems, loading, itemsLoading, error,
  fetchPlaylists, fetchItems, createPlaylist, updatePlaylist, deletePlaylist,
  addItem, removeItem,
} = usePlaylists()

const { media, fetchMedia } = useMedia()
const toast = useToast()

async function selectPlaylist(pl: Playlist) {
  active.value = pl
  addMediaId.value = ''
  await fetchItems(pl.id)
}

async function onCreate(body: Omit<PlaylistCreateBody, '$schema'>) {
  try {
    const pl = await createPlaylist(body)
    await selectPlaylist(pl)
  } catch (e) { toast.error(e instanceof Error ? e.message : 'failed to create playlist') }
}

async function onUpdate(id: string, body: Omit<PlaylistUpdateBody, '$schema'>) {
  try {
    await updatePlaylist(id, body)
    if (active.value?.id === id) {
      active.value = playlists.value.find(p => p.id === id) ?? null
    }
  } catch (e) { toast.error(e instanceof Error ? e.message : 'failed to update playlist') }
}

async function onDelete(id: string) {
  try {
    await deletePlaylist(id)
    if (active.value?.id === id) active.value = null
  } catch (e) { toast.error(e instanceof Error ? e.message : 'failed to delete playlist') }
}

async function onAddItem() {
  if (!active.value || !addMediaId.value) return
  try {
    await addItem(active.value.id, addMediaId.value)
    addMediaId.value = ''
  } catch (e) { toast.error(e instanceof Error ? e.message : 'failed to add track') }
}

// ⊹ ₊ ⟡ refresh media list after upload so the picker shows the new track
async function onMediaUploaded(id: string) {
  await fetchMedia()
  addMediaId.value = id
}

async function onRemoveItem(itemId: string) {
  if (!active.value) return
  try {
    await removeItem(active.value.id, itemId)
  } catch (e) { toast.error(e instanceof Error ? e.message : 'failed to remove track') }
}

function formatDuration(seconds?: number | null): string {
  if (!seconds) return ''
  const m = Math.floor(seconds / 60)
  const s = String(seconds % 60).padStart(2, '0')
  return `${m}:${s}`
}

onMounted(() => {
  fetchPlaylists()
  fetchMedia()
})
</script>

<template>
  <div class="flex flex-col gap-5 p-8 h-full">
    <header class="flex items-center justify-between shrink-0">
      <h1>playlists</h1>
      <button
        class="px-4 py-[0.45rem] border-0 bg-primary text-primary-foreground text-[0.85rem] cursor-pointer font-sans hover:opacity-85 disabled:opacity-40 disabled:cursor-default"
        @click="dialogEl?.openCreate()"
      >new playlist</button>
    </header>

    <p v-if="error" class="text-destructive text-[0.85rem]">{{ error }}</p>

    <div class="grid grid-cols-[220px_1fr] gap-4 flex-1 min-h-0">
      <!-- ✮ ⋆ ˚｡𖦹 left: playlist list -->
      <aside class="border border-border overflow-y-auto">
        <div v-if="loading" class="text-muted-foreground text-[0.88rem] p-8 text-center">loading…</div>
        <ul v-else-if="playlists.length > 0" class="list-none m-0 p-2">
          <li
            v-for="pl in playlists"
            :key="pl.id"
            class="flex flex-col gap-[0.15rem] px-3 py-[0.6rem] cursor-pointer text-[0.88rem] hover:bg-muted"
            :class="active?.id === pl.id ? 'bg-muted border-l-2 border-l-foreground' : ''"
            @click="selectPlaylist(pl)"
          >
            <span class="font-medium text-foreground">{{ pl.name }}</span>
            <span class="flex gap-[0.4rem] text-[0.73rem] text-muted-foreground">
              <span v-if="pl.shuffle">shuffle</span>
              <span v-if="pl.loop">loop</span>
            </span>
          </li>
        </ul>
        <div v-else class="text-muted-foreground text-[0.88rem] p-8 text-center">no playlists yet</div>
      </aside>

      <!-- ⋆˙⟡ right: items for selected playlist -->
      <section class="border border-border flex flex-col overflow-hidden" v-if="active">
        <div class="flex items-center justify-between px-4 py-[0.85rem] border-b border-border shrink-0">
          <span class="font-semibold text-[0.95rem]">{{ active.name }}</span>
          <button
            class="px-[0.6rem] py-1 border border-border bg-transparent text-[0.8rem] cursor-pointer text-muted-foreground font-sans hover:bg-muted hover:text-foreground"
            @click="dialogEl?.openEdit(active)"
          >edit</button>
        </div>

        <div v-if="itemsLoading" class="text-muted-foreground text-[0.88rem] p-8 text-center">loading…</div>

        <ol v-else-if="activeItems.length > 0" class="list-none m-0 p-2 overflow-y-auto flex-1">
          <li v-for="item in activeItems" :key="item.id" class="flex items-center gap-3 px-2 py-2 text-[0.87rem] hover:bg-muted">
            <span class="text-muted-foreground text-[0.78rem] min-w-6 text-right">{{ item.position }}</span>
            <span class="flex-1 flex flex-col gap-[0.1rem] overflow-hidden">
              <span class="font-medium text-foreground whitespace-nowrap overflow-hidden text-ellipsis">{{ item.title }}</span>
              <span v-if="item.artist" class="text-[0.78rem] text-muted-foreground">{{ item.artist }}</span>
            </span>
            <span class="text-[0.78rem] text-muted-foreground whitespace-nowrap">{{ formatDuration(item.duration) }}</span>
            <button
              class="bg-transparent border-0 cursor-pointer text-muted-foreground text-[0.8rem] p-[0.2rem] leading-none shrink-0 opacity-40 hover:text-destructive hover:opacity-100"
              @click="onRemoveItem(item.id)"
              aria-label="remove"
            >✕</button>
          </li>
        </ol>

        <div v-else class="text-muted-foreground text-[0.88rem] p-8 text-center">no tracks yet</div>

        <!-- ✮ ⋆ ˚｡𖦹 add track — searchable picker + upload -->
        <div class="flex flex-col gap-2 px-4 py-3 border-t border-border shrink-0">
          <MediaPicker
            v-model="addMediaId"
            :media="media"
            @media-added="onMediaUploaded"
          />
          <button
            class="px-4 py-[0.45rem] border-0 bg-primary text-primary-foreground text-[0.85rem] cursor-pointer font-sans hover:opacity-85 disabled:opacity-40 disabled:cursor-default"
            :disabled="!addMediaId"
            @click="onAddItem"
          >add</button>
        </div>
      </section>

      <section class="border border-border flex flex-col overflow-hidden items-center justify-center" v-else>
        select a playlist to view its tracks
      </section>
    </div>

    <PlaylistDialog
      ref="dialogEl"
      @create="onCreate"
      @update="onUpdate"
      @delete="onDelete"
    />
  </div>
</template>
