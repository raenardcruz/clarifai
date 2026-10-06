import { defineStore } from 'pinia'
import axios from 'axios'
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useStorage } from '@vueuse/core'

function parseToken(raw) {
  if (!raw) return null
  let t = raw
  if (typeof t === 'string' && t.startsWith('"') && t.endsWith('"')) {
    try { t = JSON.parse(t) } catch (_) {}
  }
  if (t === 'null' || t === 'undefined' || !t) return null
  return t
}

export const useAuthStore = defineStore('auth', () => {
  const token = useStorage('token', null)
  const user = useStorage('user', null)
  const router = useRouter()

  const getCleanToken = () => {
    return parseToken(token.value) || (typeof localStorage !== 'undefined' ? parseToken(localStorage.getItem('token')) : null)
  }

  // Set default header immediately if token exists on load
  const initialToken = getCleanToken()
  if (initialToken) {
    axios.defaults.headers.common['Authorization'] = `Bearer ${initialToken}`
  }

  const isAuthenticated = computed(() => !!getCleanToken())
  const isAdmin = computed(() => user.value?.role === 'admin')

  const setAuth = (newToken, newUser) => {
    token.value = newToken
    user.value = newUser
    const clean = parseToken(newToken)
    if (clean) {
      axios.defaults.headers.common['Authorization'] = `Bearer ${clean}`
    } else {
      delete axios.defaults.headers.common['Authorization']
    }
  }

  const logout = () => {
    token.value = null
    user.value = null
    delete axios.defaults.headers.common['Authorization']
    if (router && router.currentRoute?.value?.name !== 'login') {
      router.push('/login')
    }
  }

  const fetchUser = async () => {
    const rawToken = getCleanToken()
    if (!rawToken) return
    try {
      axios.defaults.headers.common['Authorization'] = `Bearer ${rawToken}`
      const res = await axios.get('/api/auth/me')
      user.value = res.data
    } catch (e) {
      // ONLY log out if the server explicitly responded with 401 Unauthorized or 403 Forbidden.
      // Do NOT log out on network connectivity issues, server restarts, 5xx errors, or timeouts!
      if (e.response && (e.response.status === 401 || e.response.status === 403)) {
        console.warn('Session expired or unauthorized. Logging out.', e)
        logout()
      } else {
        console.warn('Network or server error while validating user session; keeping local session:', e.message)
      }
    }
  }

  return { token, user, isAuthenticated, isAdmin, setAuth, logout, fetchUser }
})

// Global Axios Request Interceptor to ensure Authorization header is always present on requests
axios.interceptors.request.use((config) => {
  if (!config.headers) config.headers = {}
  if (!config.headers.Authorization) {
    let t = typeof localStorage !== 'undefined' ? parseToken(localStorage.getItem('token')) : null
    if (t) {
      config.headers.Authorization = `Bearer ${t}`
    }
  }
  return config
})
