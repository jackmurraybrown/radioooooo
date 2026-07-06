import { useLocalStorage } from '@vueuse/core'

// ✮ ⋆ ˚｡𖦹 calendar display prefs — persisted to localStorage
export interface CalendarPrefs {
  firstDay: 0 | 1     // 0 = sunday, 1 = monday
  scrollTime: string  // HH:MM:SS
  slotDuration: string // HH:MM:SS
}

const defaults: CalendarPrefs = {
  firstDay: 1,
  scrollTime: '08:00:00',
  slotDuration: '00:30:00',
}

export function useCalendarPrefs() {
  const prefs = useLocalStorage<CalendarPrefs>('calendarPrefs', defaults)
  return { prefs }
}
