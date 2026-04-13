import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../plugins/axios'

interface User {
  user_id: string
  tenant_id: string
  role_id: string
}

interface LoginResponse {
  token: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<User | null>(null)

  const isAuthenticated = computed(() => !!token.value)

  async function login(email: string, password: string): Promise<boolean> {
    try {
      const response = await api.post<LoginResponse>('/auth/login', { email, password })
      token.value = response.data.token
      localStorage.setItem('token', response.data.token)
      return true
    } catch {
      return false
    }
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
  }

  function initialize() {
    const storedToken = localStorage.getItem('token')
    if (storedToken) {
      token.value = storedToken
    }
  }

  return {
    token,
    user,
    isAuthenticated,
    login,
    logout,
    initialize,
  }
})
