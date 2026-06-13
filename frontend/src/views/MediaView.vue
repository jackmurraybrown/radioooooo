<script setup lang="ts">
// . ݁₊ ✶. ݁ ˖ˎˊ˗ media library
import { ref, onMounted } from 'vue'
import MediaDialog from '@/components/MediaDialog.vue'
import { useMedia } from '@/composables/useMedia'
import { useToast } from '@/composables/useToast'
import type { MediaCreateBody, MediaUpdateBody } from '@/api/types'

const dialogEl = ref<InstanceType<typeof MediaDialog>>()
const { media, loading, error, fetchMedia, createMedia, updateMedia, deleteMedia } = useMedia()
const toast = useToast()

function formatBytes(bytes?: number | null): string {
  if (!bytes) return '—'
  if (bytes < 1_000_000) return `${(bytes / 1000).toFixed(0)} kb`
  return `${(bytes / 1_000_000).toFixed(1)} mb`
}

function formatDuration(seconds?: number | null): string {
  if (!seconds) return '—'
  const m = Math.floor(seconds / 60)
  const s = String(seconds % 60).padStart(2, '0')
  return `${m}:${s}`
}

async function onCreate(body: Omit<MediaCreateBody, '$schema'>) {
  try { await createMedia(body) } catch (e) { toast.error(e instanceof Error ? e.message : 'failed to create media') }
}

async function onUpdate(id: string, body: Omit<MediaUpdateBody, '$schema'>) {
  try { await updateMedia(id, body) } catch (e) { toast.error(e instanceof Error ? e.message : 'failed to update media') }
}

async function onDelete(id: string) {
  try { await deleteMedia(id) } catch (e) { toast.error(e instanceof Error ? e.message : 'failed to delete media') }
}

onMounted(fetchMedia)
</script>

<template>
  <div class="media-page">
    <div class="toolbar">
      <h1>media</h1>
      <button class="primary" @click="dialogEl?.openCreate()">add media</button>
    </div>

    <p v-if="error" class="error-msg">{{ error }}</p>

    <div v-if="loading" class="empty">loading…</div>

    <table v-else-if="media.length > 0">
      <thead>
        <tr>
          <th>title</th>
          <th>artist</th>
          <th>format</th>
          <th>size</th>
          <th>duration</th>
          <th>status</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in media" :key="item.id">
          <td class="title-cell">{{ item.title }}</td>
          <td>{{ item.artist ?? '—' }}</td>
          <td>{{ item.fileFormat ?? '—' }}</td>
          <td>{{ formatBytes(item.fileSizeBytes) }}</td>
          <td>{{ formatDuration(item.duration) }}</td>
          <td>
            <span class="status" :class="item.downloadStatus">{{ item.downloadStatus }}</span>
          </td>
          <td>
            <button class="edit-btn" @click="dialogEl?.openEdit(item)">edit</button>
          </td>
        </tr>
      </tbody>
    </table>

    <div v-else-if="!loading" class="empty">no media yet — add something to get started</div>

    <MediaDialog
      ref="dialogEl"
      @create="onCreate"
      @update="onUpdate"
      @delete="onDelete"
    />
  </div>
</template>

<style scoped>
.media-page {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

h1 {
  font-size: 1.1rem;
  font-weight: 600;
  margin: 0;
}

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

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.88rem;
}

th {
  text-align: left;
  padding: 0.5rem 0.75rem;
  font-size: 0.75rem;
  font-weight: 500;
  color: #6b7280;
  border-bottom: 1px solid #e5e7eb;
}

td {
  padding: 0.65rem 0.75rem;
  border-bottom: 1px solid #f3f4f6;
  color: #374151;
  vertical-align: middle;
}

.title-cell {
  font-weight: 500;
  color: #111827;
}

.status {
  font-size: 0.75rem;
  padding: 0.2rem 0.5rem;
  border-radius: 99px;
  background: #f3f4f6;
  color: #6b7280;
}

.status.ready   { background: #dcfce7; color: #166534; }
.status.pending { background: #fef9c3; color: #854d0e; }
.status.error   { background: #fee2e2; color: #991b1b; }

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

.edit-btn:hover {
  background: #f3f4f6;
  color: #111827;
}

.empty {
  color: #9ca3af;
  font-size: 0.9rem;
  padding: 2rem 0;
}

.error-msg {
  color: #dc2626;
  font-size: 0.85rem;
}
</style>
