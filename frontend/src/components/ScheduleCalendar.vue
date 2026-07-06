<script setup lang="ts">
// ✮⋆‧°—°‧⋆✮ fullcalendar wrapper — timegrid view for radio scheduling
import FullCalendar from '@fullcalendar/vue3'
import dayGridPlugin from '@fullcalendar/daygrid'
import timeGridPlugin from '@fullcalendar/timegrid'
import interactionPlugin from '@fullcalendar/interaction'
import type { CalendarOptions, EventInput } from '@fullcalendar/core'

defineProps<{
  events: EventInput[]
}>()

const emit = defineEmits<{
  eventClick: [id: string]
  dateSelect: [start: Date, end: Date]
  eventDrop: [id: string, start: Date, end: Date, revert: () => void]
}>()

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
  <FullCalendar :options="{ ...calendarOptions, events }" />
</template>
