<script setup lang="ts">
// . ݁₊ ✶. ݁ ˖ˎˊ˗ media library
import { ref, h, onMounted } from 'vue'
import { useVueTable, FlexRender, getCoreRowModel, getSortedRowModel } from '@tanstack/vue-table'
import type { ColumnDef, SortingState } from '@tanstack/vue-table'
import MediaDialog from '@/components/MediaDialog.vue'
import { useMedia } from '@/composables/useMedia'
import { useToast } from '@/composables/useToast'
import type { Media, MediaCreateBody, MediaUpdateBody } from '@/api/types'

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

// ⋆˙⟡ ⋆.˚ column definitions
const columns: ColumnDef<Media>[] = [
  {
    accessorKey: 'title',
    header: 'title',
    enableSorting: true,
  },
  {
    accessorKey: 'artist',
    header: 'artist',
    enableSorting: true,
    cell: ({ row }) => row.original.artist ?? '—',
  },
  {
    accessorKey: 'fileFormat',
    header: 'format',
    enableSorting: false,
    cell: ({ row }) => row.original.fileFormat ?? '—',
  },
  {
    accessorKey: 'fileSizeBytes',
    header: 'size',
    enableSorting: true,
    cell: ({ row }) => formatBytes(row.original.fileSizeBytes),
  },
  {
    accessorKey: 'duration',
    header: 'duration',
    enableSorting: true,
    cell: ({ row }) => formatDuration(row.original.duration),
  },
  {
    accessorKey: 'downloadStatus',
    header: 'status',
    enableSorting: true,
    cell: ({ row }) => h(
      'span',
      { class: `status ${row.original.downloadStatus ?? ''}`.trim() },
      row.original.downloadStatus ?? '—',
    ),
  },
  {
    id: 'actions',
    enableSorting: false,
    cell: ({ row }) => h(
      'button',
      { class: 'edit-btn', onClick: () => dialogEl.value?.openEdit(row.original) },
      'edit',
    ),
  },
]

const sorting = ref<SortingState>([])

// ⊹ ₊ ⟡ tanstack table
const table = useVueTable({
  get data() { return media.value },
  columns,
  state: { get sorting() { return sorting.value } },
  onSortingChange: updater => {
    sorting.value = typeof updater === 'function' ? updater(sorting.value) : updater
  },
  getCoreRowModel: getCoreRowModel(),
  getSortedRowModel: getSortedRowModel(),
})

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
        <tr v-for="headerGroup in table.getHeaderGroups()" :key="headerGroup.id">
          <th
            v-for="header in headerGroup.headers"
            :key="header.id"
            :class="{ sortable: header.column.getCanSort() }"
            @click="header.column.getToggleSortingHandler()?.($event)"
          >
            <FlexRender :render="header.column.columnDef.header" :props="header.getContext()" />
            <span v-if="header.column.getIsSorted() === 'asc'" class="sort-icon">↑</span>
            <span v-else-if="header.column.getIsSorted() === 'desc'" class="sort-icon">↓</span>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in table.getRowModel().rows" :key="row.id">
          <td
            v-for="cell in row.getVisibleCells()"
            :key="cell.id"
            :class="{ 'title-cell': cell.column.id === 'title' }"
          >
            <FlexRender :render="cell.column.columnDef.cell" :props="cell.getContext()" />
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
  padding: 2rem;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

h1 { font-size: 1.1rem; font-weight: 600; margin: 0; }

button.primary {
  padding: 0.45rem 1rem;
  border: none;
  background: var(--primary);
  color: var(--primary-foreground);
  font-size: 0.85rem;
  cursor: pointer;
  font-family: inherit;
}

button.primary:hover { opacity: 0.85; }

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
  color: var(--muted-foreground);
  border-bottom: 1px solid var(--border);
  white-space: nowrap;
  user-select: none;
}

th.sortable { cursor: pointer; }
th.sortable:hover { color: var(--foreground); }

.sort-icon { margin-left: 0.25rem; opacity: 0.6; }

td {
  padding: 0.65rem 0.75rem;
  border-bottom: 1px solid var(--border);
  color: var(--muted-foreground);
  vertical-align: middle;
}

.title-cell { font-weight: 500; color: var(--foreground); }

/* ✶. ݁ ˖ status badge — tinted dark bg so it reads without being loud */
:deep(.status) {
  font-size: 0.75rem;
  padding: 0.2rem 0.5rem;
  background: var(--muted);
  color: var(--muted-foreground);
}

:deep(.status.ready)   { background: oklch(0.15 0.04 145); color: oklch(0.7 0.18 145); }
:deep(.status.pending) { background: oklch(0.15 0.04 75);  color: oklch(0.75 0.12 75); }
:deep(.status.error)   { background: oklch(0.15 0.04 27);  color: var(--destructive); }

:deep(.edit-btn) {
  padding: 0.25rem 0.6rem;
  border: 1px solid var(--border);
  background: transparent;
  font-size: 0.8rem;
  cursor: pointer;
  color: var(--muted-foreground);
  font-family: inherit;
}

:deep(.edit-btn:hover) { background: var(--muted); color: var(--foreground); }

.empty { color: var(--muted-foreground); font-size: 0.9rem; padding: 2rem 0; }
.error-msg { color: var(--destructive); font-size: 0.85rem; }
</style>
