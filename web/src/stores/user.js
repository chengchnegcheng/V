import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token'))
  const user = ref(null)

  const isAuthenticated = computed(() => !!token.value)

  async function login(username, password) {
    try {
      const response = await api.post('/auth/login', { username, password })
      token.value = response.data.token
      user.value = response.data.user
      localStorage.setItem('token', token.value)
      return true
    } catch (error) {
      console.error('Login failed:', error)
      return false
    }
  }

  async function logout() {
    try {
      await api.post('/auth/logout')
      token.value = null
      user.value = null
      localStorage.removeItem('token')
      return true
    } catch (error) {
      console.error('Logout failed:', error)
      return false
    }
  }

  async function fetchUser() {
    try {
      const response = await api.get('/auth/user')
      user.value = response.data
      return true
    } catch (error) {
      console.error('Failed to fetch user:', error)
      return false
    }
  }

  return {
    token,
    user,
    isAuthenticated,
    login,
    logout,
    fetchUser
  }
}) 