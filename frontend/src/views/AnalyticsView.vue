<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { VisXYContainer, VisLine, VisAxis, VisCrosshair } from '@unovis/vue'
import { api } from '@/api/client'
import type { Channel } from '@/api/types'

interface HourlyCount { hour: string; listeners: number; peak: number }
interface CountryCount { countryCode: string; listeners: number }
interface ChannelStats { currentListeners: number; peakListeners: number; countries: CountryCount[] }

type Range = '24h' | '7d' | '30d'

const channels = ref<Channel[]>([])
const channelId = ref('')
const range = ref<Range>('24h')
const liveCount = ref<number | null>(null)
const stats = ref<ChannelStats | null>(null)
const points = ref<HourlyCount[]>([])
const loading = ref(false)

const RANGES: { label: string; value: Range }[] = [
  { label: '24h', value: '24h' },
  { label: '7d', value: '7d' },
  { label: '30d', value: '30d' },
]

function rangeWindow() {
  const to = new Date()
  const from = new Date(to)
  if (range.value === '24h') from.setHours(from.getHours() - 24)
  else if (range.value === '7d') from.setDate(from.getDate() - 7)
  else from.setDate(from.getDate() - 30)
  return { from: from.toISOString(), to: to.toISOString() }
}

async function fetchChannels() {
  try {
    const res = await api('/channels').get()
    if (!res.ok) return
    const data = await res.json()
    channels.value = data.channels ?? []
    if (channels.value.length && !channelId.value) channelId.value = channels.value[0].id
  } catch { }
}

async function fetchLive() {
  if (!channelId.value) return
  try {
    const res = await api(`/channels/${channelId.value}/listeners/live`).get()
    if (res.ok) liveCount.value = (await res.json()).listeners
  } catch { }
}

async function fetchStats() {
  if (!channelId.value) return
  loading.value = true
  const { from, to } = rangeWindow()
  const q = `?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`
  try {
    const [sRes, tRes] = await Promise.all([
      api(`/channels/${channelId.value}/stats${q}`).get(),
      api(`/channels/${channelId.value}/stats/timeseries${q}`).get(),
    ])
    if (sRes.ok) stats.value = await sRes.json()
    if (tRes.ok) points.value = (await tRes.json()).points ?? []
  } catch { }
  finally { loading.value = false }
}

const x = (d: HourlyCount) => new Date(d.hour).getTime()
const y = (d: HourlyCount) => d.listeners
const yDomain: [number, undefined] = [0, undefined]
const LINE_COLOR = 'oklch(0.72 0.18 250)'

function tickFormat(tick: number | Date) {
  const d = new Date(typeof tick === 'number' ? tick : (tick as Date).getTime())
  return range.value === '24h'
    ? d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    : d.toLocaleDateString([], { month: 'short', day: 'numeric' })
}

const crosshairTemplate = (d: HourlyCount) =>
  `<div class="px-[9px] py-[5px] text-[0.72rem] font-sans whitespace-nowrap">${d.listeners} listening · peak ${d.peak}</div>`

const countryTotal = computed(() =>
  stats.value?.countries.reduce((s, c) => s + c.listeners, 0) ?? 0
)

function pct(n: number) {
  return countryTotal.value ? `${Math.round((n / countryTotal.value) * 100)}%` : '0%'
}

let liveTimer: ReturnType<typeof setInterval> | undefined

watch(channelId, () => { fetchStats(); fetchLive() })
watch(range, fetchStats)

onMounted(async () => {
  await fetchChannels()
  await fetchStats()
  fetchLive()
  liveTimer = setInterval(fetchLive, 30_000)
})

onUnmounted(() => clearInterval(liveTimer))
</script>

<template>
  <div class="flex flex-col gap-8 p-8 max-w-[820px]">
    <header class="flex items-center justify-between gap-4 flex-wrap">
      <h1>analytics</h1>
      <div class="flex items-center gap-3">
        <select v-if="channels.length > 1" v-model="channelId"
          class="text-[0.82rem] px-2 py-[0.3rem] border border-border bg-input text-foreground font-sans">
          <option v-for="ch in channels" :key="ch.id" :value="ch.id">{{ ch.name }}</option>
        </select>
        <div class="flex border border-border divide-x divide-border">
          <button v-for="r in RANGES" :key="r.value" type="button"
            class="px-[0.65rem] py-[0.3rem] text-[0.78rem] font-sans bg-transparent border-0 cursor-pointer text-muted-foreground hover:text-foreground"
            :class="{ 'bg-muted text-foreground': range === r.value }" @click="range = r.value">{{ r.label }}</button>
        </div>
      </div>
    </header>

    <div class="flex gap-4">
      <div class="flex-1 px-6 py-5 border border-border flex flex-col gap-[0.3rem]">
        <div class="text-[2.2rem] font-semibold tracking-[-0.02em] leading-none text-foreground">{{ liveCount ?? '—' }}
        </div>
        <div class="text-xs text-muted-foreground">live now</div>
      </div>
      <div class="flex-1 px-6 py-5 border border-border flex flex-col gap-[0.3rem]">
        <div class="text-[2.2rem] font-semibold tracking-[-0.02em] leading-none text-foreground">{{ stats?.peakListeners
          ?? '—' }}</div>
        <div class="text-xs text-muted-foreground">peak ({{ range }})</div>
      </div>
    </div>

    <section class="flex flex-col gap-3">
      <h2>listeners over time</h2>
      <div v-if="loading"
        class="h-[200px] flex items-center justify-center text-[0.85rem] text-muted-foreground border border-dashed border-border">
        loading…</div>
      <div v-else-if="points.length === 0"
        class="h-[200px] flex items-center justify-center text-[0.85rem] text-muted-foreground border border-dashed border-border">
        no listener data for this period</div>
      <!-- ⋆˙⟡ unovis needs a block container to measure width -->
      <div v-else class="relative w-full">
        <VisXYContainer :data="points" :height="200" :y-domain="yDomain">
          <VisLine :x="x" :y="y" :color="LINE_COLOR" />
          <VisAxis type="x" :tick-format="tickFormat" :num-ticks="6" />
          <VisAxis type="y" :num-ticks="4" />
          <VisCrosshair :x="x" :y="y" :color="LINE_COLOR" :template="crosshairTemplate" />
        </VisXYContainer>
      </div>
    </section>

    <section v-if="stats && stats.countries.length" class="flex flex-col gap-3">
      <h2>top locations</h2>
      <table class="w-full border-collapse">
        <thead>
          <tr>
            <th
              class="w-8 text-center text-left text-[0.68rem] uppercase tracking-[0.06em] text-muted-foreground px-2 py-[0.3rem] border-b border-border font-medium">
              #</th>
            <th
              class="text-left text-[0.68rem] uppercase tracking-[0.06em] text-muted-foreground px-2 py-[0.3rem] border-b border-border font-medium">
              country</th>
            <th
              class="w-[40%] text-left text-[0.68rem] uppercase tracking-[0.06em] text-muted-foreground px-2 py-[0.3rem] border-b border-border font-medium">
            </th>
            <th
              class="w-24 text-right text-[0.68rem] uppercase tracking-[0.06em] text-muted-foreground px-2 py-[0.3rem] border-b border-border font-medium">
              listeners</th>
            <th
              class="w-16 text-right text-[0.68rem] uppercase tracking-[0.06em] text-muted-foreground px-2 py-[0.3rem] border-b border-border font-medium">
              share</th>
          </tr>
        </thead>
        <tbody class="[&>tr:last-child>td]:border-b-0">
          <tr v-for="(c, i) in stats.countries" :key="c.countryCode" class="group">
            <td
              class="w-8 text-center px-2 py-[0.35rem] text-[0.82rem] border-b border-border/40 align-middle text-muted-foreground group-hover:bg-muted">
              {{ i + 1 }}</td>
            <td
              class="font-medium uppercase tracking-[0.04em] px-2 py-[0.35rem] text-[0.82rem] border-b border-border/40 align-middle text-muted-foreground group-hover:bg-muted">
              {{ c.countryCode }}</td>
            <td
              class="w-[40%] px-2 py-[0.35rem] text-[0.82rem] border-b border-border/40 align-middle group-hover:bg-muted">
              <div class="h-1 bg-muted overflow-hidden rounded-[2px]">
                <div
                  class="h-full bg-[oklch(0.72_0.18_250)] rounded-[2px] transition-[width] duration-[400ms] ease-[ease]"
                  :style="{ width: pct(c.listeners) }"></div>
              </div>
            </td>
            <td
              class="w-24 text-right tabular-nums px-2 py-[0.35rem] text-[0.82rem] border-b border-border/40 align-middle text-muted-foreground group-hover:bg-muted">
              {{ c.listeners.toLocaleString() }}</td>
            <td
              class="w-16 text-right tabular-nums px-2 py-[0.35rem] text-[0.82rem] border-b border-border/40 align-middle text-muted-foreground group-hover:bg-muted">
              {{ pct(c.listeners) }}</td>
          </tr>
        </tbody>
      </table>
    </section>
  </div>
</template>
