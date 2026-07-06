import { useAuthStore } from '@/stores/auth'

const BASE = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

// ⋆.˚ ⊹₊⟡ attach bearer token to request headers
function withAuth(init: RequestInit = {}): RequestInit {
  const auth = useAuthStore()
  if (!auth.accessToken) return init
  return {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init.headers,
      Authorization: `Bearer ${auth.accessToken}`,
    },
  }
}

// ✶. ݁ ˖ˎˊ˗ fetch wrapper — retries once after a token refresh on 401
export async function apiFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const res = await fetch(`${BASE}${path}`, withAuth(init))

  if (res.status !== 401) return res

  const auth = useAuthStore()
  await auth.refresh()

  return fetch(`${BASE}${path}`, withAuth(init))
}

export function api(path: string) {
  return {
    get: () => apiFetch(path),
    post: (body: unknown) => apiFetch(path, { method: 'POST', body: JSON.stringify(body) }),
    put: (body: unknown) => apiFetch(path, { method: 'PUT', body: JSON.stringify(body) }),
    delete: () => apiFetch(path, { method: 'DELETE' }),
  }
}

// ⊹ ₊ ⟡ multipart upload — no Content-Type so browser sets boundary automatically
export async function apiUpload(path: string, data: FormData): Promise<Response> {
  const auth = useAuthStore()
  const authHeader = auth.accessToken ? { Authorization: `Bearer ${auth.accessToken}` } : {}
  const res = await fetch(`${BASE}${path}`, { method: 'POST', body: data, headers: authHeader })
  if (res.status !== 401) return res
  await auth.refresh()
  const retryHeader = auth.accessToken ? { Authorization: `Bearer ${auth.accessToken}` } : {}
  return fetch(`${BASE}${path}`, { method: 'POST', body: data, headers: retryHeader })
}
