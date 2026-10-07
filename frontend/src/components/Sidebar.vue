<template>
  <aside class="sidebar" :class="{ 'sidebar-open': isSidebarOpen }">
    <!-- Brand Header -->
    <div class="logo-container">
      <div class="logo-icon-wrapper">
        <div class="logo-icon">
          <Sparkles size="20" color="#ffffff" stroke-width="2.5" />
        </div>
      </div>
      <div class="logo-text">
        <div class="logo-title-row">
          <h2>ClarifAi</h2>
          <span class="pro-badge">PRO</span>
        </div>
        <span class="logo-subtitle">Meeting Intelligence</span>
      </div>
    </div>

    <!-- Quick Action Button -->
    <div class="quick-action-section">
      <button class="apple-primary-btn w-full" @click="isRecordingModalOpen = true">
        <div class="btn-icon-circle">
          <Plus size="16" stroke-width="2.5" />
        </div>
        <span>New Recording</span>
      </button>
    </div>

    <!-- macOS Navigation Menu -->
    <nav class="nav-menu">
      <div class="nav-section-label">WORKSPACE</div>
      
      <router-link to="/" class="nav-item" active-class="active">
        <div class="nav-icon-badge badge-blue">
          <LayoutDashboard size="17" stroke-width="2.2" />
        </div>
        <span class="nav-text">Dashboard</span>
      </router-link>

      <router-link to="/recordings" class="nav-item" active-class="active">
        <div class="nav-icon-badge badge-purple">
          <Mic size="17" stroke-width="2.2" />
        </div>
        <span class="nav-text">All Recordings</span>
      </router-link>

      <template v-if="authStore.isAdmin">
        <div class="nav-section-label mt-6">MANAGEMENT</div>

        <router-link to="/users" class="nav-item" active-class="active">
          <div class="nav-icon-badge badge-green">
            <Users size="17" stroke-width="2.2" />
          </div>
          <span class="nav-text">Users</span>
        </router-link>

        <router-link to="/settings" class="nav-item" active-class="active">
          <div class="nav-icon-badge badge-slate">
            <Settings size="17" stroke-width="2.2" />
          </div>
          <span class="nav-text">Settings</span>
        </router-link>
      </template>
    </nav>

    <!-- Sidebar Footer / User Profile -->
    <div class="sidebar-footer">
      <div class="user-card">
        <div class="user-avatar-initials">
          {{ authStore.user?.email?.[0].toUpperCase() || 'U' }}
        </div>
        <div class="user-card-info">
          <span class="user-card-name">{{ authStore.user?.email || 'User' }}</span>
          <span class="user-card-role">{{ authStore.user?.role === 'admin' ? 'Administrator' : 'Standard Member' }}</span>
        </div>
        <button 
          v-if="isIosApp"
          class="ios-reload-icon-btn" 
          @click="handleReload" 
          title="Reload app"
          aria-label="Reload app"
        >
          <RotateCw size="15" :class="{ 'animate-spin': isReloading }" />
        </button>
        <button class="logout-icon-btn" @click="authStore.logout()" title="Log out">
          <LogOut size="16" />
        </button>
      </div>
    </div>

    <!-- Teleport Modal out of sidebar to avoid z-index issues -->
    <Teleport to="body">
      <NewRecordingModal v-if="isRecordingModalOpen" @close="isRecordingModalOpen = false" />
    </Teleport>
  </aside>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Capacitor } from '@capacitor/core'
import { useAuthStore } from '../stores/auth'
import { isSidebarOpen } from '../stores/layout'
import { LayoutDashboard, Mic, Settings, Users, Plus, Sparkles, LogOut, RotateCw } from '@lucide/vue'
import NewRecordingModal from './NewRecordingModal.vue'

const isIosApp = computed(() => {
  return Capacitor.isNativePlatform() && Capacitor.getPlatform() === 'ios'
})

const isReloading = ref(false)

const handleReload = () => {
  if (isReloading.value) return
  isReloading.value = true
  setTimeout(() => {
    window.location.reload()
  }, 100)
}

const authStore = useAuthStore()
const isRecordingModalOpen = ref(false)
</script>

<style scoped>
.sidebar {
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  width: var(--sidebar-width);
  background-color: rgba(255, 255, 255, 0.82);
  backdrop-filter: blur(28px) saturate(190%);
  -webkit-backdrop-filter: blur(28px) saturate(190%);
  border-right: 1px solid rgba(0, 0, 0, 0.07);
  display: flex;
  flex-direction: column;
  padding: calc(1.5rem + var(--safe-area-top, env(safe-area-inset-top, 0px))) calc(1rem + var(--safe-area-right, env(safe-area-inset-right, 0px))) calc(1.25rem + var(--safe-area-bottom, env(safe-area-inset-bottom, 0px))) calc(1rem + var(--safe-area-left, env(safe-area-inset-left, 0px)));
  z-index: 40;
  transition: transform 0.28s var(--apple-ease);
  user-select: none;
  -webkit-user-select: none;
  box-sizing: border-box;
}

.logo-container {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  padding: 0.25rem 0.5rem 1.5rem 0.5rem;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
}

.logo-icon-wrapper {
  position: relative;
}

.logo-icon {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #0071E3 0%, #AF52DE 100%);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(0, 113, 227, 0.3), 0 0 0 1px rgba(255, 255, 255, 0.3) inset;
}

.logo-title-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.logo-text h2 {
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
  letter-spacing: -0.025em;
  line-height: 1.2;
}

.pro-badge {
  font-size: 0.65rem;
  font-weight: 700;
  background: linear-gradient(135deg, #0071E3, #5856D6);
  color: #ffffff;
  padding: 0.12rem 0.4rem;
  border-radius: var(--radius-full);
  letter-spacing: 0.04em;
}

.logo-subtitle {
  font-size: 0.75rem;
  color: var(--text-muted);
  font-weight: 500;
}

.quick-action-section {
  padding: 1rem 0.25rem;
}

.apple-primary-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  background: linear-gradient(180deg, #0077ED 0%, #0071E3 100%);
  color: #ffffff;
  border: 1px solid rgba(255, 255, 255, 0.25);
  box-shadow: 0 4px 14px rgba(0, 113, 227, 0.28), 0 1px 1px rgba(255, 255, 255, 0.4) inset;
  padding: 0.75rem 1rem;
  border-radius: var(--radius-md);
  font-weight: 600;
  font-size: 0.9rem;
  cursor: pointer;
  transition: var(--transition);
}

.apple-primary-btn:hover {
  background: linear-gradient(180deg, #0A84FF 0%, #0077ED 100%);
  box-shadow: 0 6px 18px rgba(0, 113, 227, 0.36);
  transform: translateY(-1px);
}

.apple-primary-btn:active {
  transform: scale(0.98);
}

.btn-icon-circle {
  width: 22px;
  height: 22px;
  background: rgba(255, 255, 255, 0.22);
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
}

.nav-menu {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  flex: 1;
  padding-top: 0.5rem;
}

.nav-section-label {
  font-size: 0.68rem;
  font-weight: 700;
  color: var(--text-muted);
  letter-spacing: 0.06em;
  padding: 0.5rem 0.75rem 0.25rem 0.75rem;
}

.mt-6 {
  margin-top: 1.25rem;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.6rem 0.75rem;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 0.88rem;
  font-weight: 500;
  transition: var(--transition);
  cursor: pointer;
  text-decoration: none;
}

.nav-item:hover {
  background-color: rgba(0, 0, 0, 0.04);
  color: var(--text-primary);
}

.nav-item.active {
  background-color: rgba(0, 113, 227, 0.1);
  color: var(--apple-blue);
  font-weight: 600;
}

.nav-icon-badge {
  width: 30px;
  height: 30px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: var(--transition);
  flex-shrink: 0;
}

.badge-blue {
  background-color: rgba(0, 113, 227, 0.1);
  color: #0071E3;
}

.badge-purple {
  background-color: rgba(175, 82, 222, 0.1);
  color: #AF52DE;
}

.badge-green {
  background-color: rgba(52, 199, 89, 0.12);
  color: #34C759;
}

.badge-slate {
  background-color: rgba(142, 142, 147, 0.15);
  color: #636366;
}

.nav-item.active .badge-blue {
  background-color: #0071E3;
  color: #ffffff;
  box-shadow: 0 2px 6px rgba(0, 113, 227, 0.3);
}

.nav-item.active .badge-purple {
  background-color: #AF52DE;
  color: #ffffff;
  box-shadow: 0 2px 6px rgba(175, 82, 222, 0.3);
}

.nav-item.active .badge-green {
  background-color: #34C759;
  color: #ffffff;
  box-shadow: 0 2px 6px rgba(52, 199, 89, 0.3);
}

.nav-item.active .badge-slate {
  background-color: #636366;
  color: #ffffff;
}

.nav-text {
  flex: 1;
}

.sidebar-footer {
  margin-top: auto;
  padding-top: 1rem;
  border-top: 1px solid rgba(0, 0, 0, 0.05);
}

.user-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.65rem 0.75rem;
  border-radius: var(--radius-md);
  background: rgba(0, 0, 0, 0.025);
  border: 1px solid rgba(0, 0, 0, 0.04);
}

.user-avatar-initials {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-full);
  background: linear-gradient(135deg, #0071E3, #30B0C7);
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.85rem;
  flex-shrink: 0;
}

.user-card-info {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  flex: 1;
}

.user-card-name {
  font-size: 0.825rem;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-card-role {
  font-size: 0.7rem;
  color: var(--text-muted);
}

.ios-reload-icon-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0.4rem;
  min-width: 34px;
  min-height: 34px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: var(--transition);
  touch-action: manipulation;
  -webkit-tap-highlight-color: transparent;
}

.ios-reload-icon-btn:hover {
  background: rgba(0, 113, 227, 0.1);
  color: var(--apple-blue);
}

.logout-icon-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0.4rem;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: var(--transition);
}

.logout-icon-btn:hover {
  background: rgba(255, 59, 48, 0.1);
  color: var(--apple-red);
}

@media (max-width: 768px) {
  .sidebar {
    transform: translateX(-100%);
    box-shadow: 0 0 30px rgba(0, 0, 0, 0.15);
  }
  
  .sidebar.sidebar-open {
    transform: translateX(0);
  }
}
</style>
