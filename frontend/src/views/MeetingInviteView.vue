<template>
  <div class="auth-layout">
    <div class="auth-card">
      <!-- Header -->
      <div class="auth-header">
        <div class="auth-app-icon">
          <Sparkles size="24" color="#FFFFFF" stroke-width="2.2" />
        </div>
        <h2 class="auth-title">Meeting Invite</h2>
        <p class="auth-subtitle" v-if="invite">
          You've been invited to access <strong>{{ invite.title }}</strong>
        </p>
        <p class="auth-subtitle" v-else-if="!loadingInvite && !inviteError">
          Create your username and password to join
        </p>
      </div>

      <!-- Loading Invite -->
      <div v-if="loadingInvite" class="invite-state-box">
        <Loader2 class="animate-spin" size="28" color="#FFFFFF" />
        <p class="state-text">Validating meeting invite...</p>
      </div>

      <!-- Invalid Invite -->
      <div v-else-if="inviteError" class="invite-state-box">
        <div class="error-banner" style="width: 100%; justify-content: center;">
          <span>{{ inviteError }}</span>
        </div>
        <router-link to="/login" class="apple-auth-submit-btn" style="margin-top: 1rem; text-decoration: none;">
          <span>Go to Sign In</span>
        </router-link>
      </div>

      <!-- Already Signed In -->
      <div v-else-if="authStore.isAuthenticated" class="invite-signed-in">
        <div class="signed-in-pill">
          <span>Signed in as <strong>{{ authStore.user?.email }}</strong></span>
        </div>

        <div v-if="error" class="error-banner">
          <span>{{ error }}</span>
        </div>

        <button class="apple-auth-submit-btn" :disabled="submitting" @click="handleJoinExisting">
          <Loader2 class="animate-spin" size="16" v-if="submitting" />
          <span v-else>Join Meeting</span>
          <ArrowRight size="16" v-if="!submitting" />
        </button>

        <button type="button" class="secondary-link-btn" @click="switchToNewAccount">
          Create a new username & password instead
        </button>
      </div>

      <!-- Create Username & Password Form -->
      <form v-else @submit.prevent="handleAcceptInvite" class="auth-form">
        <div class="form-group">
          <label class="input-label">Create Username</label>
          <div class="input-field-wrapper">
            <UserIcon class="input-icon" size="16" />
            <input
              type="text"
              v-model="username"
              class="apple-auth-input"
              placeholder="Choose a username"
              required
              autocomplete="username"
            />
          </div>
        </div>

        <div class="form-group">
          <label class="input-label">Create Password</label>
          <div class="input-field-wrapper">
            <Lock class="input-icon" size="16" />
            <input
              :type="showPassword ? 'text' : 'password'"
              v-model="password"
              class="apple-auth-input"
              placeholder="••••••••"
              required
              minlength="4"
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

        <button type="submit" class="apple-auth-submit-btn" :disabled="submitting">
          <Loader2 class="animate-spin" size="16" v-if="submitting" />
          <span v-else>Create Account & Open Meeting</span>
          <ArrowRight size="16" v-if="!submitting" />
        </button>
      </form>

      <div class="auth-footer" v-if="!loadingInvite && !inviteError && !authStore.isAuthenticated">
        <span>Already have an account?</span>
        <router-link :to="{ path: '/login', query: { redirect: route.fullPath } }" class="auth-link">Sign In</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import { useToast } from 'primevue/usetoast'
import { useAuthStore } from '../stores/auth'
import { Sparkles, User as UserIcon, Lock, ArrowRight, Loader2, Eye, EyeOff } from '@lucide/vue'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const authStore = useAuthStore()

const invite = ref(null)
const loadingInvite = ref(true)
const inviteError = ref('')

const username = ref('')
const password = ref('')
const showPassword = ref(false)
const error = ref('')
const submitting = ref(false)

const fetchInvite = async () => {
  loadingInvite.value = true
  inviteError.value = ''
  try {
    const res = await axios.get(`/api/recordings/invite/${route.params.token}`)
    invite.value = res.data
  } catch (err) {
    inviteError.value = err.response?.data?.detail || 'Invalid or expired meeting invite link.'
  } finally {
    loadingInvite.value = false
  }
}

const handleAcceptInvite = async () => {
  error.value = ''
  submitting.value = true
  try {
    const res = await axios.post(`/api/recordings/invite/${route.params.token}/accept`, {
      username: username.value,
      password: password.value
    })
    authStore.setAuth(res.data.access_token, res.data.user)
    toast.add({ severity: 'success', summary: 'Welcome!', detail: 'Account created and meeting unlocked.', life: 3000 })
    router.push(`/recordings/${res.data.recording_id}`)
  } catch (err) {
    error.value = err.response?.data?.detail || 'Failed to accept invite. Please try again.'
  } finally {
    submitting.value = false
  }
}

const handleJoinExisting = async () => {
  error.value = ''
  submitting.value = true
  try {
    const res = await axios.post(`/api/recordings/invite/${route.params.token}/join`)
    toast.add({ severity: 'success', summary: 'Joined', detail: 'Meeting added to your workspace.', life: 3000 })
    router.push(`/recordings/${res.data.recording_id}`)
  } catch (err) {
    error.value = err.response?.data?.detail || 'Failed to join meeting.'
  } finally {
    submitting.value = false
  }
}

const switchToNewAccount = () => {
  authStore.setAuth(null, null)
}

onMounted(() => {
  fetchInvite()
})
</script>

<style scoped>
.auth-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  margin-bottom: 1.75rem;
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
  color: rgba(255, 255, 255, 0.7);
  line-height: 1.45;
}

.auth-subtitle strong {
  color: #FFFFFF;
}

.invite-state-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1.5rem 0;
  gap: 0.75rem;
}

.state-text {
  font-size: 0.88rem;
  color: rgba(255, 255, 255, 0.7);
}

.invite-signed-in {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.signed-in-pill {
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: var(--radius-md);
  padding: 0.75rem 1rem;
  font-size: 0.86rem;
  color: rgba(255, 255, 255, 0.85);
  text-align: center;
}

.secondary-link-btn {
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.65);
  font-size: 0.82rem;
  cursor: pointer;
  padding: 0.25rem;
  text-align: center;
  transition: color 0.2s;
}

.secondary-link-btn:hover {
  color: #FFFFFF;
  text-decoration: underline;
}

.auth-form {
  display: flex;
  flex-direction: column;
}

.form-group {
  margin-bottom: 1.15rem;
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
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  display: flex;
  align-items: center;
  padding: 0.2rem;
}

.eye-toggle-btn:hover {
  color: rgba(255, 255, 255, 0.85);
}

.error-banner {
  background: rgba(255, 69, 58, 0.15);
  border: 1px solid rgba(255, 69, 58, 0.3);
  color: #FF6961;
  padding: 0.65rem 0.85rem;
  border-radius: var(--radius-sm);
  font-size: 0.82rem;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
}

.apple-auth-submit-btn {
  width: 100%;
  background: var(--apple-blue);
  color: #FFFFFF;
  border: none;
  border-radius: var(--radius-md);
  padding: 0.8rem 1.25rem;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  transition: var(--transition-fast);
}

.apple-auth-submit-btn:hover:not(:disabled) {
  background: #0077ED;
  transform: translateY(-1px);
}

.apple-auth-submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.auth-footer {
  margin-top: 1.5rem;
  text-align: center;
  font-size: 0.82rem;
  color: rgba(255, 255, 255, 0.55);
  display: flex;
  justify-content: center;
  gap: 0.35rem;
}

.auth-link {
  color: #60A5FA;
  font-weight: 600;
  text-decoration: none;
}

.auth-link:hover {
  text-decoration: underline;
}
</style>
