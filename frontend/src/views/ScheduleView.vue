<script setup lang="ts">
// ⊹ ࣪ ˖ schedule view — fetches channels, loads episodes for the active one
import { ref, onMounted } from 'vue'
import ScheduleCalendar from '@/components/ScheduleCalendar.vue'
import { useSchedule } from '@/composables/useSchedule'
import { api } from '@/api/client'

interface Channel {
  id: string
  name: string
}

const channels = ref<Channel[]>([])
const activeChannelId = ref('')

const { events, updateEpisode } = useSchedule(activeChannelId)

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
  // ദ്ദി(°ᵕ°,,) episode create modal goes here
  console.log('create episode', start, end)
}

function onEventClick(id: string) {
  console.log('edit episode', id)
}

async function onEventDrop(id: string, start: Date, end: Date) {
  await updateEpisode(id, {
    startTime: start.toISOString(),
    endTime: end.toISOString(),
  })
}

onMounted(loadChannels)
</script>

<template>
  <main>
    <ScheduleCalendar
      :events="events"
      @date-select="onDateSelect"
      @event-click="onEventClick"
      @event-drop="onEventDrop"
    />
  </main>
</template>
