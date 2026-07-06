<script setup lang="ts">
defineProps<{
  error: string | null
  loading: boolean
}>()

const emit = defineEmits<{
  submit: [email: string, password: string]
}>()

import { ref } from 'vue'

const email = ref('')
const password = ref('')
</script>

<style scoped>
form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: 320px;
}

label {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  font-size: 0.8rem;
  color: var(--muted-foreground);
}

input {
  font-size: 0.9rem;
  padding: 0.45rem 0.6rem;
  border: 1px solid var(--border);
  background: var(--input);
  color: var(--foreground);
  outline: none;
  font-family: inherit;
  width: 100%;
}

input:focus { border-color: var(--ring); }

p {
  font-size: 0.85rem;
  color: var(--destructive);
  margin: 0;
}

button {
  padding: 0.5rem 1rem;
  background: var(--primary);
  color: var(--primary-foreground);
  border: none;
  font-size: 0.9rem;
  cursor: pointer;
  font-family: inherit;
  font-weight: 500;
}

button:disabled { opacity: 0.5; cursor: default; }
</style>

<template>
  <form @submit.prevent="emit('submit', email, password)">
    <label>
      email
      <input v-model="email" type="email" autocomplete="email" required />
    </label>
    <label>
      password
      <input v-model="password" type="password" autocomplete="current-password" required />
    </label>
    <p v-if="error">{{ error }}</p>
    <button type="submit" :disabled="loading">
      {{ loading ? 'signing in…' : 'sign in' }}
    </button>
  </form>
</template>
