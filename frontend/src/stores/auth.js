import { defineStore } from 'pinia'
import axios from 'axios'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useStorage } from '@vueuse/core'

export const useAuthStore = defineStore('auth', () => {
  const token = useStorage('token', null)
  const user = useStorage('user', null)
  const router = useRouter()

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  const setAuth = (newToken, newUser) => {
    token.value = newToken
    user.value = newUser
    axios.defaults.headers.common['Authorization'] = `Bearer ${newToken}`
  }

  const logout = () => {
    token.value = null
    user.value = null
    delete axios.defaults.headers.common['Authorization']
    router.push('/login')
  }

  const fetchUser = async () => {
    if (!token.value) return
    try {
      axios.defaults.headers.common['Authorization'] = `Bearer ${token.value}`
      const res = await axios.get('/api/auth/me')
      user.value = res.data
    } catch (e) {
      logout()
    }
  }

  return { token, user, isAuthenticated, isAdmin, setAuth, logout, fetchUser }
})
