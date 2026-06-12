// ✮ ⋆ ˚｡𖦹 ⋆｡°✩ toast state lives at module level — singleton across all components
import { ref } from 'vue'

export type ToastType = 'error' | 'success'

interface Toast {
  id:      number
  message: string
  type:    ToastType
}

const toasts = ref<Toast[]>([])
let seq = 0

export function useToast() {
  function add(message: string, type: ToastType = 'error', duration = 4000) {
    const id = seq++
    toasts.value.push({ id, message, type })
    setTimeout(() => dismiss(id), duration)
  }

  function dismiss(id: number) {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  const error   = (msg: string) => add(msg, 'error')
  const success = (msg: string) => add(msg, 'success')

  return { toasts, error, success, dismiss }
}
