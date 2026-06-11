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
