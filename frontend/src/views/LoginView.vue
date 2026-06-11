<script setup lang="ts">
// ⋆˙⟡ thin view — delegates to LoginForm, owns async state
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import LoginForm from '@/components/auth/LoginForm.vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const error = ref<string | null>(null)
const loading = ref(false)

async function onLogin(email: string, password: string) {
  error.value = null
  loading.value = true
  try {
    await auth.login(email, password)
    const redirect = (route.query.redirect as string) ?? '/schedule'
    router.push(redirect)
  } catch {
    error.value = 'invalid email or password'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main>
    <LoginForm :error="error" :loading="loading" @submit="onLogin" />
  </main>
</template>
