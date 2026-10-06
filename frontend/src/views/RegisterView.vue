<template>
  <div class="auth-layout">
    <div class="auth-card">
      <!-- Apple App Icon & Header -->
      <div class="auth-header">
        <div class="auth-app-icon">
          <Sparkles size="24" color="#FFFFFF" stroke-width="2.2" />
        </div>
        <h2 class="auth-title">ClarifAi</h2>
        <p class="auth-subtitle">Create your personal AI productivity workspace</p>
      </div>

      <!-- Registration Form -->
      <form @submit.prevent="handleRegister" class="auth-form" v-if="!success">
        <div class="form-group">
          <label class="input-label">Full Name</label>
          <div class="input-field-wrapper">
            <UserIcon class="input-icon" size="16" />
            <input 
              type="text" 
              v-model="name" 
              class="apple-auth-input" 
              placeholder="e.g. Alex Rivera" 
              required 
            />
          </div>
        </div>

        <div class="form-group">
          <label class="input-label">Username or Email</label>
          <div class="input-field-wrapper">
            <Mail class="input-icon" size="16" />
            <input 
              type="text" 
              v-model="email" 
              class="apple-auth-input" 
              placeholder="alex123" 
              required 
              autocomplete="username"
            />
          </div>
        </div>

        <div class="form-group">
          <label class="input-label">Password</label>
          <div class="input-field-wrapper">
            <Lock class="input-icon" size="16" />
            <input 
              :type="showPassword ? 'text' : 'password'" 
              v-model="password" 
              class="apple-auth-input" 
              placeholder="••••••••" 
              required 
              autocomplete="new-password"
            />
            <button type="button" class="eye-toggle-btn" @click="showPassword = !showPassword">
              <EyeOff size="15" v-if="showPassword" />
              <Eye size="15" v-else />
            </button>
          </div>
        </div>

        <div v-if="error" class="error-banner">
          <span>{{ error }}</span>
        </div>

        <button type="submit" class="apple-auth-submit-btn" :disabled="loading">
          <Loader2 class="animate-spin" size="16" v-if="loading" />
          <span v-else>Create Account</span>
          <ArrowRight size="16" v-if="!loading" />
        </button>
      </form>

      <!-- Success State -->
      <div v-else class="auth-success-state">
        <div class="success-icon-circle">
          <Check size="28" stroke-width="2.5" />
        </div>
        <h3 class="success-title">Registration Submitted</h3>
        <p class="success-desc">Your account has been registered and is pending administrator authorization.</p>
        
        <router-link to="/login" class="apple-auth-submit-btn w-full text-center">
          Back to Sign In
        </router-link>
      </div>

      <div class="auth-footer" v-if="!success">
        <span>Already have an account?</span>
        <router-link to="/login" class="auth-link">Sign In</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import axios from 'axios'
import { Sparkles, Mail, Lock, User as UserIcon, ArrowRight, Check, Loader2, Eye, EyeOff } from '@lucide/vue'

const name = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const success = ref(false)
const showPassword = ref(false)

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
.auth-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  margin-bottom: 2rem;
}

.auth-app-icon {
  width: 54px;
  height: 54px;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, #0071E3 0%, #AF52DE 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.25rem;
  box-shadow: 0 8px 24px rgba(0, 113, 227, 0.4), 0 0 0 1px rgba(255, 255, 255, 0.25) inset;
}

.auth-title {
  font-size: 1.75rem;
  font-weight: 700;
  color: #FFFFFF;
  letter-spacing: -0.03em;
  margin-bottom: 0.35rem;
}

.auth-subtitle {
  font-size: 0.88rem;
  color: rgba(255, 255, 255, 0.65);
}

.auth-form {
  display: flex;
  flex-direction: column;
}

.input-label {
  font-size: 0.8rem;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.8);
  margin-bottom: 0.4rem;
  display: block;
}

.input-field-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: 0.85rem;
  color: rgba(255, 255, 255, 0.4);
  pointer-events: none;
}

.apple-auth-input {
  width: 100%;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: var(--radius-md);
  padding: 0.75rem 2.5rem 0.75rem 2.4rem;
  font-size: 0.9rem;
  font-family: inherit;
  color: #FFFFFF;
  outline: none;
  transition: var(--transition-fast);
}

.apple-auth-input:focus {
  background: rgba(255, 255, 255, 0.12);
  border-color: var(--apple-blue);
  box-shadow: 0 0 0 3px rgba(0, 113, 227, 0.35);
}

.apple-auth-input::placeholder {
  color: rgba(255, 255, 255, 0.35);
}

.eye-toggle-btn {
  position: absolute;
  right: 0.85rem;
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.45);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.eye-toggle-btn:hover {
  color: #FFFFFF;
}

.error-banner {
  background: rgba(255, 59, 48, 0.15);
  border: 1px solid rgba(255, 59, 48, 0.3);
  color: #FF6961;
  padding: 0.65rem 0.85rem;
  border-radius: var(--radius-sm);
  font-size: 0.8rem;
  margin-bottom: 1.25rem;
  text-align: center;
}

.apple-auth-submit-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  background: linear-gradient(180deg, #0077ED 0%, #0071E3 100%);
  color: #FFFFFF;
  border: 1px solid rgba(255, 255, 255, 0.25);
  border-radius: var(--radius-full);
  padding: 0.8rem 1.5rem;
  font-size: 0.92rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s var(--apple-ease);
  box-shadow: 0 4px 16px rgba(0, 113, 227, 0.4);
  margin-top: 0.5rem;
  text-decoration: none;
}

.apple-auth-submit-btn:hover {
  background: linear-gradient(180deg, #0A84FF 0%, #0077ED 100%);
  transform: translateY(-1px);
}

.apple-auth-submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.auth-success-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 1rem 0;
}

.success-icon-circle {
  width: 60px;
  height: 60px;
  border-radius: var(--radius-full);
  background: rgba(52, 199, 89, 0.18);
  color: #30D158;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.25rem;
}

.success-title {
  font-size: 1.35rem;
  font-weight: 700;
  color: #FFFFFF;
  margin-bottom: 0.5rem;
}

.success-desc {
  font-size: 0.85rem;
  color: rgba(255, 255, 255, 0.65);
  margin-bottom: 1.75rem;
  line-height: 1.5;
}

.auth-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  margin-top: 1.75rem;
  font-size: 0.825rem;
  color: rgba(255, 255, 255, 0.55);
}

.auth-link {
  color: #64D2FF;
  font-weight: 600;
  text-decoration: none;
}

.auth-link:hover {
  text-decoration: underline;
}
</style>
