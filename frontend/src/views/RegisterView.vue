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

      <form @submit.prevent="handleRegister" v-if="!success">
        <div class="form-group">
          <label>Full Name</label>
          <div class="relative">
            <UserIcon class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 z-10" size="18" />
            <InputText type="text" v-model="name" class="w-full pl-10" placeholder="Alex Rivera" required />
          </div>
        </div>

        <div class="form-group">
          <label>Username</label>
          <div class="relative">
            <Mail class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 z-10" size="18" />
            <InputText type="text" v-model="email" class="w-full pl-10" placeholder="alex123" required />
          </div>
        </div>

        <div class="form-group">
          <label>Password</label>
          <div class="relative">
            <Lock class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 z-10" size="18" />
            <Password v-model="password" class="w-full" inputClass="w-full pl-10" placeholder="••••••••" :feedback="false" toggleMask required />
          </div>
        </div>

        <div v-if="error" class="text-red-400 text-sm mb-4 text-center">
          {{ error }}
        </div>

        <Button type="submit" class="w-full" :loading="loading" label="Create Account">
          <template #icon>
            <ArrowRight size="18" v-if="!loading" class="ml-2" />
          </template>
        </Button>
      </form>

      <div v-else class="text-center py-6">
        <div class="w-16 h-16 bg-green-500/20 text-green-400 rounded-full flex items-center justify-center mx-auto mb-4">
          <Check size="32" />
        </div>
        <h3 class="text-xl font-bold mb-2">Registration Successful</h3>
        <p class="text-sm text-gray-400 mb-6">Your account has been created and is pending admin approval.</p>
        <router-link to="/login" class="btn btn-outline w-full text-white border-gray-600 hover:border-indigo-400 hover:text-indigo-400">Back to Login</router-link>
      </div>

      <p v-if="!success" class="mt-6 text-center text-sm text-gray-400">
        Already have an account? 
        <router-link to="/login" class="text-indigo-400 hover:text-indigo-300 font-medium">Log in</router-link>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import axios from 'axios'
import { Sparkles, Mail, Lock, User as UserIcon, ArrowRight, Check } from '@lucide/vue'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'


const name = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const success = ref(false)

const handleRegister = async () => {
  error.value = ''
  loading.value = true
  
  try {
    await axios.post('/api/auth/register', {
      email: email.value,
      password: password.value
    })
    success.value = true
  } catch (err) {
    if (err.response?.data?.detail) {
      error.value = err.response.data.detail
    } else {
      error.value = 'Registration failed. Please try again.'
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
/* Reusing the scoped classes from LoginView */
.mb-8 { margin-bottom: 2rem; }
.mb-6 { margin-bottom: 1.5rem; }
.mb-4 { margin-bottom: 1rem; }
.mb-2 { margin-bottom: 0.5rem; }
.mt-6 { margin-top: 1.5rem; }
.mx-auto { margin-left: auto; margin-right: auto; }
.py-6 { padding-top: 1.5rem; padding-bottom: 1.5rem; }
.text-center { text-align: center; }
.text-2xl { font-size: 1.5rem; }
.text-xl { font-size: 1.25rem; }
.text-sm { font-size: 0.875rem; }
.font-bold { font-weight: 700; }
.font-medium { font-weight: 500; }
.text-gray-400 { color: #9CA3AF; }
.text-indigo-400 { color: #818CF8; }
.text-red-400 { color: #F87171; }
.text-white { color: #ffffff; }
.text-green-400 { color: #4ADE80; }
.border-gray-600 { border-color: #4B5563; }
.hover\:border-indigo-400:hover { border-color: #818CF8; }
.hover\:text-indigo-400:hover { color: #818CF8; }
.hover\:text-indigo-300:hover { color: #A5B4FC; }
.w-12 { width: 3rem; }
.h-12 { height: 3rem; }
.w-16 { width: 4rem; }
.h-16 { height: 4rem; }
.rounded-xl { border-radius: 0.75rem; }
.rounded-full { border-radius: 9999px; }
.bg-gradient-to-br { background-image: linear-gradient(to bottom right, var(--tw-gradient-stops)); }
.from-indigo-500 { --tw-gradient-from: #6366f1; --tw-gradient-stops: var(--tw-gradient-from), var(--tw-gradient-to, rgba(99, 102, 241, 0)); }
.to-purple-600 { --tw-gradient-to: #9333ea; }
.bg-green-500\/20 { background-color: rgba(34, 197, 94, 0.2); }
.shadow-lg { box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05); }
.shadow-indigo-500\/30 { box-shadow: 0 10px 15px -3px rgba(99, 102, 241, 0.3); }
.relative { position: relative; }
.absolute { position: absolute; }
.left-3 { left: 0.75rem; }
.top-1\/2 { top: 50%; }
.-translate-y-1\/2 { transform: translateY(-50%); }
.pl-10 { padding-left: 2.5rem; }
</style>
