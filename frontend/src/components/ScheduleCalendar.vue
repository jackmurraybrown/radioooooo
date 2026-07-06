<script setup lang="ts">
// ✮⋆‧°—°‧⋆✮ fullcalendar wrapper — timegrid view for radio scheduling
import FullCalendar from '@fullcalendar/vue3'
import dayGridPlugin from '@fullcalendar/daygrid'
import timeGridPlugin from '@fullcalendar/timegrid'
import interactionPlugin from '@fullcalendar/interaction'
import type { CalendarOptions, EventInput } from '@fullcalendar/core'
import { ref, onMounted, onUnmounted } from 'vue'

defineProps<{
  events: EventInput[]
}>()

const emit = defineEmits<{
  eventClick: [id: string]
  dateSelect: [start: Date, end: Date]
  eventDrop: [id: string, start: Date, end: Date, revert: () => void]
}>()

const calendarRef = ref<InstanceType<typeof FullCalendar>>()

// ⋆˙⟡ skip shortcuts when user is typing in an input
function isTyping(e: KeyboardEvent) {
  const tag = (e.target as HTMLElement).tagName
  return tag === 'INPUT' || tag === 'TEXTAREA' || (e.target as HTMLElement).isContentEditable
}

function onKeyDown(e: KeyboardEvent) {
  if (isTyping(e)) return
  const api = calendarRef.value?.getApi()
  if (!api) return
  switch (e.key) {
    case 'ArrowLeft':  case 'h': api.prev(); break
    case 'ArrowRight': case 'l': api.next(); break
    case 't':                    api.today(); break
    case 'd':                    api.changeView('timeGridDay'); break
    case 'w':                    api.changeView('timeGridWeek'); break
    case 'm':                    api.changeView('dayGridMonth'); break
  }
}

onMounted(()  => window.addEventListener('keydown', onKeyDown))
onUnmounted(() => window.removeEventListener('keydown', onKeyDown))

const calendarOptions: CalendarOptions = {
  plugins: [dayGridPlugin, timeGridPlugin, interactionPlugin],
  initialView: 'timeGridWeek',
  headerToolbar: {
    left: 'prev,next today',
    center: 'title',
    right: 'dayGridMonth,timeGridWeek,timeGridDay',
  },
  selectable: true,
  editable: true,
  allDaySlot: false,
  slotMinTime: '00:00:00',
  slotMaxTime: '24:00:00',
  height: 'auto',
  select(info) {
    emit('dateSelect', info.start, info.end)
  },
  eventClick(info) {
    emit('eventClick', info.event.id)
  },
  eventDrop(info) {
    emit('eventDrop', info.event.id, info.event.start!, info.event.end!, info.revert)
  },
}
</script>

<template>
  <FullCalendar ref="calendarRef" :options="{ ...calendarOptions, events }" />
</template>
