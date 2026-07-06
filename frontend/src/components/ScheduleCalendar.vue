<script setup lang="ts">
// ✮⋆‧°—°‧⋆✮ fullcalendar wrapper — timegrid view for radio scheduling
import FullCalendar from '@fullcalendar/vue3'
import dayGridPlugin from '@fullcalendar/daygrid'
import timeGridPlugin from '@fullcalendar/timegrid'
import interactionPlugin from '@fullcalendar/interaction'
import type { CalendarOptions, EventInput } from '@fullcalendar/core'
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useCalendarPrefs } from '@/composables/useCalendarPrefs'

defineProps<{
  events: EventInput[]
}>()

const emit = defineEmits<{
  eventClick: [id: string]
  dateSelect: [start: Date, end: Date]
  eventDrop: [id: string, start: Date, end: Date, revert: () => void]
}>()

const calendarRef = ref<InstanceType<typeof FullCalendar>>()
const { prefs } = useCalendarPrefs()

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

const calendarOptions = computed<CalendarOptions>(() => ({
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
  firstDay:     prefs.value.firstDay,
  scrollTime:   prefs.value.scrollTime,
  slotDuration: prefs.value.slotDuration,
  height: '100%',
  stickyHeaderDates: true,
  // ⊹ ₊ ⟡ show type badge top-right for non-live episodes
  eventContent(info) {
    const type = info.event.extendedProps.type as string
    const wrap = document.createElement('div')
    wrap.className = 'ep-event'

    const title = document.createElement('div')
    title.className = 'ep-title'
    title.textContent = info.event.title
    wrap.appendChild(title)

    if (type && type !== 'live') {
      const badge = document.createElement('span')
      badge.className = 'ep-type-badge'
      badge.textContent = type
      wrap.appendChild(badge)
    }

    return { domNodes: [wrap] }
  },
  select(info) {
    emit('dateSelect', info.start, info.end)
  },
  eventClick(info) {
    emit('eventClick', info.event.id)
  },
  eventDrop(info) {
    emit('eventDrop', info.event.id, info.event.start!, info.event.end!, info.revert)
  },
}))
</script>

<template>
  <div class="cal-wrap">
    <div class="cal-header-extra">
      <slot name="header-right" />
    </div>
    <FullCalendar ref="calendarRef" :options="{ ...calendarOptions, events }" />
  </div>
</template>

<style scoped>
.cal-wrap {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  position: relative;
  padding-top: 1rem;
}

/* ⊹ ₊ ⟡ overlays right of fc header — sits left of the view buttons */
.cal-header-extra {
  position: absolute;
  top: 1rem;
  right: 0;
  height: 41px;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding-right: 0.75rem;
  z-index: 10;
  pointer-events: none;
}

.cal-header-extra > * {
  pointer-events: auto;
}

/* ⋆˙⟡ push view buttons left so channel select has room on the right */
.cal-wrap :deep(.fc-toolbar .fc-toolbar-chunk:last-child) {
  padding-right: 160px;
}

/* ⊹ ₊ ⟡ breathing room on the toolbar */
.cal-wrap :deep(.fc-toolbar) {
  padding: 0 1rem;
}

/* ✶. ݁ ˖ smaller date heading — matches toolbar text size */
.cal-wrap :deep(.fc-col-header-cell-cushion),
.cal-wrap :deep(.fc-col-header-cell a),
.cal-wrap :deep(.fc-col-header-cell) {
  font-size: 0.75rem;
  font-weight: 400;
  padding: 0.2rem 0.4rem;
}

/* ⊹ ₊ toolbar title also smaller ˎˊ˗ */
.cal-wrap :deep(.fc-toolbar-title) {
  font-size: 0.9rem;
  font-weight: 500;
}

/* ⊹ ₊ ⟡ fc hardcodes color:#fff on active — override so dark text shows on light bg */
.cal-wrap :deep(.fc-button-active),
.cal-wrap :deep(.fc-button-active:hover) {
  color: var(--primary-foreground) !important;
}

/* ✮ ⋆ ˚｡𖦹 custom event content layout */
.cal-wrap :deep(.ep-event) {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 1px 3px;
  position: relative;
  overflow: hidden;
  min-height: 0;
}

.cal-wrap :deep(.ep-title) {
  font-size: 0.78rem;
  font-weight: 500;
  line-height: 1.3;
  flex: 1;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
}

/* ⋆˙⟡ type badge — sits bottom-right, small + semi-transparent */
.cal-wrap :deep(.ep-type-badge) {
  position: absolute;
  bottom: 2px;
  right: 3px;
  font-size: 0.6rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  opacity: 0.7;
  line-height: 1;
}
</style>
