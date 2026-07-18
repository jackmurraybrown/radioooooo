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
  <form class="flex flex-col gap-4 w-80" @submit.prevent="emit('submit', email, password)">
    <label class="flex flex-col gap-[0.3rem] text-[0.8rem] text-muted-foreground">
      email
      <input
        v-model="email"
        type="email"
        autocomplete="email"
        required
        class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border border-border bg-input text-foreground outline-none font-sans w-full focus:border-ring"
      />
    </label>
    <label class="flex flex-col gap-[0.3rem] text-[0.8rem] text-muted-foreground">
      password
      <input
        v-model="password"
        type="password"
        autocomplete="current-password"
        required
        class="text-[0.9rem] px-[0.6rem] py-[0.45rem] border border-border bg-input text-foreground outline-none font-sans w-full focus:border-ring"
      />
    </label>
    <p v-if="error" class="text-[0.85rem] text-destructive m-0">{{ error }}</p>
    <button
      type="submit"
      :disabled="loading"
      class="px-4 py-2 bg-primary text-primary-foreground border-0 text-[0.9rem] cursor-pointer font-sans font-medium disabled:opacity-50 disabled:cursor-default"
    >
      {{ loading ? 'signing in…' : 'sign in' }}
    </button>
  </form>
</template>
