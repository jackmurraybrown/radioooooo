<script setup lang="ts">
// ⊹ ࣪ ˖ toast container — fixed bottom-right, auto-dismiss
import { useToast } from '@/composables/useToast'

const { toasts, dismiss } = useToast()
</script>

<template>
  <Teleport to="body">
    <div class="fixed bottom-6 right-6 flex flex-col gap-2 z-1000 pointer-events-none" aria-live="polite">
      <!-- ⋆˙⟡ semantic border colors only — no bg fill -->
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="flex items-center gap-3 px-4 py-[0.65rem] text-sm pointer-events-auto max-w-90 border bg-muted text-foreground animate-[slide-in_0.15s_ease]"
        :class="toast.type === 'error' ? 'border-destructive text-destructive' : toast.type === 'success' ? 'border-success-border text-success' : 'border-border'"
      >
        <span class="flex-1">{{ toast.message }}</span>
        <button
          class="bg-transparent border-0 cursor-pointer text-inherit opacity-60 text-[0.8rem] p-0 leading-none shrink-0 hover:opacity-100"
          @click="dismiss(toast.id)"
          aria-label="dismiss"
        >✕</button>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
/* ⋆˙⟡ tailwind can't author keyframes — kept minimal, referenced via arbitrary [animation:] above */
@keyframes slide-in {
  from { opacity: 0; transform: translateX(0.5rem); }
  to   { opacity: 1; transform: translateX(0); }
}
</style>
