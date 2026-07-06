<script setup lang="ts">
// ⊹ ࣪ ˖ schedule view — channel picker + calendar + episode dialog
import { ref, onMounted } from 'vue'
import ScheduleCalendar from '@/components/ScheduleCalendar.vue'
import EpisodeDialog from '@/components/EpisodeDialog.vue'
import { useSchedule } from '@/composables/useSchedule'
import { useToast } from '@/composables/useToast'
import { api } from '@/api/client'
import type { Channel } from '@/api/types'
import type { EpisodeBody } from '@/api/types'

const channels = ref<Channel[]>([])
const activeChannelId = ref('')
const dialogEl = ref<InstanceType<typeof EpisodeDialog>>()

const { episodes, events, createEpisode, updateEpisode, deleteEpisode } = useSchedule(activeChannelId)
const toast = useToast()

async function loadChannels() {
  const res = await api('/channels').get()
  if (!res.ok) return
  const data = await res.json()
  channels.value = data.channels ?? []
  if (channels.value.length > 0) {
    activeChannelId.value = channels.value[0].id
  }
}

function onDateSelect(start: Date, end: Date) {
  dialogEl.value?.openCreate(start.toISOString(), end.toISOString())
}

function onEventClick(id: string) {
  const ep = episodes.value.find(e => e.id === id)
  if (ep) dialogEl.value?.openEdit(ep)
}

async function onEventDrop(id: string, start: Date, end: Date, revert: () => void) {
  const ep = episodes.value.find(e => e.id === id)
  if (!ep) { revert(); return }
  try {
    await updateEpisode(id, {
      title:         ep.title,
      description:   ep.description,
      startTime:     start.toISOString(),
      endTime:       end.toISOString(),
      type:          ep.type,
      sourceAdapter: ep.sourceAdapter,
      sourceRef:     ep.sourceRef,
      contactEmail:  ep.contactEmail,
    })
  } catch (e) {
    revert()
    toast.error(e instanceof Error ? e.message : 'failed to move episode')
  }
}

async function onCreate(body: Omit<EpisodeBody, '$schema'>) {
  try { await createEpisode(body) } catch (e) { toast.error(e instanceof Error ? e.message : 'failed to create episode') }
}

async function onUpdate(id: string, body: Omit<EpisodeBody, '$schema'>) {
  try { await updateEpisode(id, body) } catch (e) { toast.error(e instanceof Error ? e.message : 'failed to update episode') }
}

async function onDelete(id: string) {
  try { await deleteEpisode(id) } catch (e) { toast.error(e instanceof Error ? e.message : 'failed to delete episode') }
}

onMounted(loadChannels)
</script>

<template>
  <div class="schedule-page">
    <ScheduleCalendar
      :events="events"
      @date-select="onDateSelect"
      @event-click="onEventClick"
      @event-drop="onEventDrop"
    >
      <template #header-right>
        <span class="channel-label">channel</span>
        <select
          v-if="channels.length > 1"
          v-model="activeChannelId"
          class="channel-select"
        >
          <option v-for="ch in channels" :key="ch.id" :value="ch.id">
            {{ ch.name }}
          </option>
        </select>
        <span v-else-if="channels.length === 1" class="channel-name">
          {{ channels[0].name }}
        </span>
      </template>
    </ScheduleCalendar>

    <EpisodeDialog
      ref="dialogEl"
      @create="onCreate"
      @update="onUpdate"
      @delete="onDelete"
    />
  </div>
</template>

<style scoped>
.schedule-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.channel-label {
  font-size: 0.8rem;
  color: var(--muted-foreground);
}

.channel-select {
  font-size: 0.8rem;
  padding: 0.2rem 0.4rem;
  border: 1px solid var(--border);
  background: var(--input);
  color: var(--foreground);
  font-family: inherit;
  cursor: pointer;
  outline: none;
}

.channel-select:focus { border-color: var(--ring); }

.channel-name {
  font-size: 0.8rem;
  color: var(--foreground);
}
</style>
