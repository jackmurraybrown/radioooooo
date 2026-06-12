<script setup lang="ts">
// ⊹ ࣪ ˖ schedule view — channel picker + calendar + episode dialog
import { ref, onMounted } from 'vue'
import ScheduleCalendar from '@/components/ScheduleCalendar.vue'
import EpisodeDialog from '@/components/EpisodeDialog.vue'
import { useSchedule } from '@/composables/useSchedule'
import { api } from '@/api/client'
import type { Channel } from '@/api/types'
import type { EpisodeBody } from '@/api/types'

const channels = ref<Channel[]>([])
const activeChannelId = ref('')
const dialogEl = ref<InstanceType<typeof EpisodeDialog>>()

const { episodes, events, createEpisode, updateEpisode, deleteEpisode } = useSchedule(activeChannelId)

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

async function onEventDrop(id: string, start: Date, end: Date) {
  await updateEpisode(id, { startTime: start.toISOString(), endTime: end.toISOString() })
}

async function onCreate(body: Omit<EpisodeBody, '$schema'>) {
  await createEpisode(body)
}

async function onUpdate(id: string, body: Omit<EpisodeBody, '$schema'>) {
  await updateEpisode(id, body)
}

async function onDelete(id: string) {
  await deleteEpisode(id)
}

onMounted(loadChannels)
</script>

<template>
  <div class="schedule-page">
    <div class="toolbar">
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
    </div>

    <ScheduleCalendar
      :events="events"
      @date-select="onDateSelect"
      @event-click="onEventClick"
      @event-drop="onEventDrop"
    />

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
  gap: 1rem;
  height: 100%;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.channel-select {
  font-size: 0.9rem;
  padding: 0.35rem 0.6rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  background: #fff;
  font-family: inherit;
  cursor: pointer;
  outline: none;
}

.channel-select:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99,102,241,0.15);
}

.channel-name {
  font-size: 0.9rem;
  font-weight: 500;
  color: #374151;
}
</style>
