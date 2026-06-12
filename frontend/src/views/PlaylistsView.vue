<script setup lang="ts">
// ⊹ ࣪ ˖ playlists — master list + item detail panel
import { ref, onMounted } from 'vue'
import PlaylistDialog from '@/components/PlaylistDialog.vue'
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
  <div class="playlists-page">
    <div class="toolbar">
      <h1>playlists</h1>
      <button class="primary" @click="dialogEl?.openCreate()">new playlist</button>
    </div>

    <p v-if="error" class="error-msg">{{ error }}</p>

    <div class="layout">
      <!-- ✮ ⋆ ˚｡𖦹 left: playlist list -->
      <aside class="list-panel">
        <div v-if="loading" class="empty">loading…</div>
        <ul v-else-if="playlists.length > 0">
          <li
            v-for="pl in playlists"
            :key="pl.id"
            :class="{ active: active?.id === pl.id }"
            @click="selectPlaylist(pl)"
          >
            <span class="pl-name">{{ pl.name }}</span>
            <span class="pl-meta">
              <span v-if="pl.shuffle">shuffle</span>
              <span v-if="pl.loop">loop</span>
            </span>
          </li>
        </ul>
        <div v-else class="empty">no playlists yet</div>
      </aside>

      <!-- ⋆˙⟡ right: items for selected playlist -->
      <section class="detail-panel" v-if="active">
        <div class="detail-header">
          <span class="detail-title">{{ active.name }}</span>
          <button class="edit-btn" @click="dialogEl?.openEdit(active)">edit</button>
        </div>

        <div v-if="itemsLoading" class="empty">loading…</div>

        <ol v-else-if="activeItems.length > 0" class="items-list">
          <li v-for="item in activeItems" :key="item.id">
            <span class="item-pos">{{ item.position }}</span>
            <span class="item-info">
              <span class="item-title">{{ item.title }}</span>
              <span v-if="item.artist" class="item-artist">{{ item.artist }}</span>
            </span>
            <span class="item-dur">{{ formatDuration(item.duration) }}</span>
            <button class="remove-btn" @click="onRemoveItem(item.id)" aria-label="remove">✕</button>
          </li>
        </ol>

        <div v-else class="empty">no tracks yet</div>

        <!-- add track -->
        <div class="add-track" v-if="media.length > 0">
          <select v-model="addMediaId">
            <option value="" disabled>pick a track…</option>
            <option v-for="m in media" :key="m.id" :value="m.id">
              {{ m.title }}{{ m.artist ? ` — ${m.artist}` : '' }}
            </option>
          </select>
          <button class="primary" :disabled="!addMediaId" @click="onAddItem">add</button>
        </div>
      </section>

      <section class="detail-panel empty-detail" v-else>
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

<style scoped>
.playlists-page {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  height: 100%;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}

h1 { font-size: 1.1rem; font-weight: 600; margin: 0; }

button.primary {
  padding: 0.45rem 1rem;
  border-radius: 6px;
  border: none;
  background: #111827;
  color: #fff;
  font-size: 0.85rem;
  cursor: pointer;
  font-family: inherit;
}

button.primary:hover { background: #374151; }
button.primary:disabled { opacity: 0.4; cursor: default; }

.layout {
  display: grid;
  grid-template-columns: 220px 1fr;
  gap: 1rem;
  flex: 1;
  min-height: 0;
}

.list-panel {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow-y: auto;
}

ul { list-style: none; margin: 0; padding: 0.5rem; }

li {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  padding: 0.6rem 0.75rem;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.88rem;
}

li:hover { background: #f9fafb; }
li.active { background: #f3f4f6; }

.pl-name { font-weight: 500; color: #111827; }

.pl-meta {
  display: flex;
  gap: 0.4rem;
  font-size: 0.73rem;
  color: #9ca3af;
}

.detail-panel {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.empty-detail {
  align-items: center;
  justify-content: center;
}

.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.85rem 1rem;
  border-bottom: 1px solid #f3f4f6;
  flex-shrink: 0;
}

.detail-title { font-weight: 600; font-size: 0.95rem; }

.edit-btn {
  padding: 0.25rem 0.6rem;
  border-radius: 5px;
  border: 1px solid #e5e7eb;
  background: transparent;
  font-size: 0.8rem;
  cursor: pointer;
  color: #6b7280;
  font-family: inherit;
}

.edit-btn:hover { background: #f3f4f6; color: #111827; }

.items-list {
  list-style: none;
  margin: 0;
  padding: 0.5rem;
  overflow-y: auto;
  flex: 1;
}

.items-list li {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0.5rem;
  border-radius: 5px;
  font-size: 0.87rem;
}

.items-list li:hover { background: #f9fafb; }

.item-pos { color: #9ca3af; font-size: 0.78rem; min-width: 1.5rem; text-align: right; }

.item-info { flex: 1; display: flex; flex-direction: column; gap: 0.1rem; overflow: hidden; }

.item-title {
  font-weight: 500;
  color: #111827;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-artist { font-size: 0.78rem; color: #6b7280; }

.item-dur { font-size: 0.78rem; color: #9ca3af; white-space: nowrap; }

.remove-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: #d1d5db;
  font-size: 0.8rem;
  padding: 0.2rem;
  line-height: 1;
  flex-shrink: 0;
}

.remove-btn:hover { color: #dc2626; }

.add-track {
  display: flex;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  border-top: 1px solid #f3f4f6;
  flex-shrink: 0;
}

.add-track select {
  flex: 1;
  font-size: 0.88rem;
  padding: 0.4rem 0.6rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  outline: none;
  font-family: inherit;
  background: #fff;
}

.add-track select:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99,102,241,0.15);
}

.empty {
  color: #9ca3af;
  font-size: 0.88rem;
  padding: 2rem;
  text-align: center;
}

.error-msg { color: #dc2626; font-size: 0.85rem; }
</style>
