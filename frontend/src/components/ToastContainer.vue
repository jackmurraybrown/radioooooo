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
  border-radius: 8px;
  font-size: 0.875rem;
  box-shadow: 0 4px 12px rgba(0,0,0,0.12);
  pointer-events: all;
  max-width: 360px;
  animation: slide-in 0.15s ease;
}

.toast.error {
  background: #fef2f2;
  color: #991b1b;
  border: 1px solid #fecaca;
}

.toast.success {
  background: #f0fdf4;
  color: #166534;
  border: 1px solid #bbf7d0;
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
