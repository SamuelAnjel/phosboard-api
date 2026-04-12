<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

interface LoginForm {
  email: string
  password: string
}

const router = useRouter()
const authStore = useAuthStore()

const form = ref<LoginForm>({
  email: '',
  password: '',
})

const error = ref('')
const loading = ref(false)

async function handleSubmit() {
  error.value = ''
  loading.value = true

  const success = await authStore.login(form.value.email, form.value.password)

  loading.value = false

  if (success) {
    router.push('/')
  } else {
    error.value = 'Invalid email or password'
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-100">
    <div class="bg-white p-8 rounded-lg shadow-md w-full max-w-md">
      <h1 class="text-2xl font-bold text-gray-800 mb-6 text-center">Phosboard</h1>

      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div v-if="error" class="bg-red-50 text-red-600 p-3 rounded text-sm">
          {{ error }}
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
          <input
            v-model="form.email"
            type="email"
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="admin@phosboard.cl"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Password</label>
          <input
            v-model="form.password"
            type="password"
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="password123"
          />
        </div>

        <button
          type="submit"
          :disabled="loading"
          class="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ loading ? 'Signing in...' : 'Sign In' }}
        </button>
      </form>
    </div>
  </div>
</template>
