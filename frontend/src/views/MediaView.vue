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

// ✶. ݁ ˖ status badge classes — tinted dark bg so it reads without being loud
function statusClasses(status?: string | null): string {
  const base = 'text-xs px-2 py-[0.2rem]'
  if (status === 'ready') return `${base} bg-[oklch(0.15_0.04_145)] text-[oklch(0.7_0.18_145)]`
  if (status === 'pending') return `${base} bg-[oklch(0.15_0.04_75)] text-[oklch(0.75_0.12_75)]`
  if (status === 'error') return `${base} bg-[oklch(0.15_0.04_27)] text-destructive`
  return `${base} bg-muted text-muted-foreground`
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
      { class: statusClasses(row.original.downloadStatus) },
      row.original.downloadStatus ?? '—',
    ),
  },
  {
    id: 'actions',
    enableSorting: false,
    cell: ({ row }) => h(
      'button',
      {
        class: 'px-[0.6rem] py-1 border border-border bg-transparent text-[0.8rem] cursor-pointer text-muted-foreground font-sans hover:bg-muted hover:text-foreground',
        onClick: () => dialogEl.value?.openEdit(row.original),
      },
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
  <div class="flex flex-col gap-5 p-8">
    <header class="flex items-center justify-between">
      <h1>media</h1>
      <button
        class="px-4 py-[0.45rem] border-0 bg-primary text-primary-foreground text-[0.85rem] cursor-pointer font-sans hover:opacity-85"
        @click="dialogEl?.openCreate()"
      >add media</button>
    </header>

    <p v-if="error" class="text-destructive text-[0.85rem]">{{ error }}</p>
    <div v-if="loading" class="text-muted-foreground text-[0.9rem] py-8">loading…</div>

    <table v-else-if="media.length > 0" class="w-full border-collapse text-[0.88rem]">
      <thead>
        <tr v-for="headerGroup in table.getHeaderGroups()" :key="headerGroup.id">
          <th
            v-for="header in headerGroup.headers"
            :key="header.id"
            class="text-left px-3 py-2 text-xs font-medium text-muted-foreground border-b border-border whitespace-nowrap select-none"
            :class="header.column.getCanSort() ? 'cursor-pointer hover:text-foreground' : ''"
            @click="header.column.getToggleSortingHandler()?.($event)"
          >
            <FlexRender :render="header.column.columnDef.header" :props="header.getContext()" />
            <span v-if="header.column.getIsSorted() === 'asc'" class="ml-1 opacity-60">↑</span>
            <span v-else-if="header.column.getIsSorted() === 'desc'" class="ml-1 opacity-60">↓</span>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in table.getRowModel().rows" :key="row.id">
          <td
            v-for="cell in row.getVisibleCells()"
            :key="cell.id"
            class="px-3 py-[0.65rem] border-b border-border align-middle"
            :class="cell.column.id === 'title' ? 'font-medium text-foreground' : 'text-muted-foreground'"
          >
            <FlexRender :render="cell.column.columnDef.cell" :props="cell.getContext()" />
          </td>
        </tr>
      </tbody>
    </table>

    <div v-else-if="!loading" class="text-muted-foreground text-[0.9rem] py-8">no media yet — add something to get started</div>

    <MediaDialog
      ref="dialogEl"
      @create="onCreate"
      @update="onUpdate"
      @delete="onDelete"
    />
  </div>
</template>
