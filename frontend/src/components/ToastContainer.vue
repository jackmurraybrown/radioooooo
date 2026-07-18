<script setup lang="ts">
// ⊹ ࣪ ˖ toast container — fixed bottom-right, auto-dismiss
import { useToast } from '@/composables/useToast'

const { toasts, dismiss } = useToast()
</script>

<template>
  <Teleport to="body">
    <div class="toast-container" aria-live="polite">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="toast"
        :class="toast.type"
      >
        <span>{{ toast.message }}</span>
        <button @click="dismiss(toast.id)" aria-label="dismiss">✕</button>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-container {
  position: fixed;
  bottom: 1.5rem;
  right: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  z-index: 1000;
  pointer-events: none;
}

.toast {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.65rem 1rem;
  font-size: 0.875rem;
  pointer-events: all;
  max-width: 360px;
  animation: slide-in 0.15s ease;
  border: 1px solid var(--border);
  background: var(--muted);
  color: var(--foreground);
}

/* ⋆˙⟡ semantic border colors only — no bg fill */
.toast.error {
  border-color: var(--destructive);
  color: var(--destructive);
}

.toast.success {
  border-color: oklch(0.45 0.15 145);
  color: oklch(0.7 0.18 145);
}

.toast span { flex: 1; }

.toast button {
  background: none;
  border: none;
  cursor: pointer;
  color: inherit;
  opacity: 0.6;
  font-size: 0.8rem;
  padding: 0;
  line-height: 1;
  flex-shrink: 0;
}

.toast button:hover { opacity: 1; }

@keyframes slide-in {
  from { opacity: 0; transform: translateX(0.5rem); }
  to   { opacity: 1; transform: translateX(0); }
}
</style>
