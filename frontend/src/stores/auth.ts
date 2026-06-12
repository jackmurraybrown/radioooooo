import { defineStore } from 'pinia'
import { ref } from 'vue'

const BASE = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

// ✮ ⋆ ˚｡𖦹 auth state — single source of truth for tokens
function decodeStationId(token: string): string | null {
  try {
    const payload = JSON.parse(atob(token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')))
    return (payload.station_id as string) ?? null
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string | null>(localStorage.getItem('accessToken'))
  const refreshToken = ref<string | null>(localStorage.getItem('refreshToken'))
  const stationId = ref<string | null>(
    accessToken.value ? decodeStationId(accessToken.value) : null
  )

  function set(access: string, refresh: string) {
    accessToken.value = access
    refreshToken.value = refresh
    stationId.value = decodeStationId(access)
    localStorage.setItem('accessToken', access)
    localStorage.setItem('refreshToken', refresh)
  }

  function clear() {
    accessToken.value = null
    refreshToken.value = null
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
  }

  async function login(email: string, password: string) {
    const res = await fetch(`${BASE}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
    if (!res.ok) throw new Error('invalid credentials')
    const data = await res.json()
    set(data.accessToken, data.refreshToken)
  }

  async function refresh() {
    if (!refreshToken.value) throw new Error('no refresh token')
    const res = await fetch(`${BASE}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken: refreshToken.value }),
    })
    if (!res.ok) throw new Error('refresh failed')
    const data = await res.json()
    set(data.accessToken, data.refreshToken)
  }

  async function logout() {
    if (refreshToken.value) {
      await fetch(`${BASE}/auth/logout`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refreshToken: refreshToken.value }),
      }).catch(() => {})
    }
    clear()
  }

  const isAuthenticated = () => !!accessToken.value

  return { accessToken, refreshToken, stationId, set, clear, login, refresh, logout, isAuthenticated }
})
