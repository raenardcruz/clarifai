<template>
  <div class="auth-layout">
    <div class="auth-card">
      <div class="text-center mb-8 flex flex-col items-center">
        <div class="w-12 h-12 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-xl flex items-center justify-center mb-4 shadow-lg shadow-indigo-500/30">
          <Sparkles color="white" />
        </div>
        <h2 class="text-gray-400 font-bold mb-2">ClarifAi</h2>
        <p class="text-sm text-gray-400">Elevate your productivity</p>
      </div>

      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label>Username</label>
          <div class="relative">
            <Mail class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 z-10" size="18" />
            <InputText type="text" v-model="email" class="w-full pl-10" placeholder="admin" required />
          </div>
        </div>

        <div class="form-group">
          <div class="flex justify-between items-center mb-1">
            <label class="mb-0">Password</label>
            <a href="#" class="text-xs text-indigo-400 hover:text-indigo-300">Forgot password?</a>
          </div>
          <div class="relative">
            <Lock class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 z-10" size="18" />
            <Password v-model="password" class="w-full" inputClass="w-full pl-10" placeholder="••••••••" :feedback="false" toggleMask required />
          </div>
        </div>

        <div v-if="error" class="text-red-400 text-sm mb-4 text-center">
          {{ error }}
        </div>

        <Button type="submit" class="w-full" :loading="loading" label="Sign In">
          <template #icon>
            <ArrowRight size="18" v-if="!loading" class="ml-2" />
          </template>
        </Button>
      </form>

      <p class="mt-6 text-center text-sm text-gray-400">
        Don't have an account? 
        <router-link to="/register" class="text-indigo-400 hover:text-indigo-300 font-medium">Sign up</router-link>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import axios from 'axios'
import { Sparkles, Mail, Lock, ArrowRight } from '@lucide/vue'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

const handleLogin = async () => {
  error.value = ''
  loading.value = true
  
  const formData = new URLSearchParams()
  formData.append('username', email.value)
  formData.append('password', password.value)

  try {
    const res = await axios.post('/api/auth/login', formData)
    const token = res.data.access_token
    
    // Fetch user details immediately to set role
    axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
    const userRes = await axios.get('/api/auth/me')
    
    authStore.setAuth(token, userRes.data)
    router.push('/')
  } catch (err) {
    if (err.response?.status === 400) {
      error.value = err.response.data.detail
    } else {
      error.value = 'Invalid email or password'
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
/* Scoped overrides to use the global auth styles and utility classes */
.mb-8 { margin-bottom: 2rem; }
.mb-4 { margin-bottom: 1rem; }
.mb-2 { margin-bottom: 0.5rem; }
.mb-1 { margin-bottom: 0.25rem; }
.mb-0 { margin-bottom: 0; }
.mt-6 { margin-top: 1.5rem; }
.text-center { text-align: center; }
.text-2xl { font-size: 1.5rem; }
.text-sm { font-size: 0.875rem; }
.text-xs { font-size: 0.75rem; }
.font-bold { font-weight: 700; }
.font-medium { font-weight: 500; }
.text-gray-400 { color: #9CA3AF; }
.text-indigo-400 { color: #818CF8; }
.text-red-400 { color: #F87171; }
.hover\:text-indigo-300:hover { color: #A5B4FC; }
.w-12 { width: 3rem; }
.h-12 { height: 3rem; }
.rounded-xl { border-radius: 0.75rem; }
.bg-gradient-to-br { background-image: linear-gradient(to bottom right, var(--tw-gradient-stops)); }
.from-indigo-500 { --tw-gradient-from: #6366f1; --tw-gradient-stops: var(--tw-gradient-from), var(--tw-gradient-to, rgba(99, 102, 241, 0)); }
.to-purple-600 { --tw-gradient-to: #9333ea; }
.shadow-lg { box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05); }
.shadow-indigo-500\/30 { box-shadow: 0 10px 15px -3px rgba(99, 102, 241, 0.3); }
.relative { position: relative; }
.absolute { position: absolute; }
.left-3 { left: 0.75rem; }
.top-1\/2 { top: 50%; }
.-translate-y-1\/2 { transform: translateY(-50%); }
.pl-10 { padding-left: 2.5rem; }
</style>
