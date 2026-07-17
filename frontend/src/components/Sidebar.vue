<template>
  <aside class="sidebar" :class="{ 'sidebar-open': isSidebarOpen }">
    <div class="logo-container">
      <div class="logo-icon"><Sparkles size="20" color="#fff" /></div>
      <div class="logo-text">
        <h2>ClarifAi</h2>
        <span>Elevate your productivity</span>
      </div>
    </div>

    <nav class="nav-menu">
      <router-link to="/" class="nav-item" active-class="active">
        <LayoutDashboard size="20" />
        <span>Dashboard</span>
      </router-link>
      <router-link to="/recordings" class="nav-item" active-class="active">
        <Mic size="20" />
        <span>Recordings</span>
      </router-link>
      <router-link v-if="authStore.isAdmin" to="/users" class="nav-item" active-class="active">
        <Users size="20" />
        <span>Users</span>
      </router-link>
      <router-link v-if="authStore.isAdmin" to="/settings" class="nav-item" active-class="active">
        <Settings size="20" />
        <span>Settings</span>
      </router-link>
    </nav>

    <div class="sidebar-footer">
      <button class="btn btn-primary w-full new-recording-btn" @click="isRecordingModalOpen = true">
        <Plus size="18" />
        <span>New Recording</span>
      </button>
      
      <button class="nav-item mt-4 w-full justify-start text-left bg-transparent border-none" @click="authStore.logout()">
        <LogOut size="20" />
        <span>Logout</span>
      </button>
    </div>

    <!-- Teleport Modal out of sidebar to avoid z-index issues -->
    <Teleport to="body">
      <NewRecordingModal v-if="isRecordingModalOpen" @close="isRecordingModalOpen = false" />
    </Teleport>
  </aside>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { isSidebarOpen } from '../stores/layout'
import { LayoutDashboard, Mic, Settings, Users, Plus, Sparkles, LogOut } from '@lucide/vue'
import NewRecordingModal from './NewRecordingModal.vue'

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
  background-color: var(--white);
  border-right: 1px solid var(--neutral-200);
  display: flex;
  flex-direction: column;
  padding: 2rem 1.5rem;
  z-index: 40;
  transition: transform 0.3s ease;
}

.logo-container {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 3rem;
}

.logo-icon {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, var(--primary), var(--tertiary));
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--shadow-glow);
}

.logo-text h2 {
  font-size: 1.25rem;
  color: var(--primary);
  margin: 0;
  line-height: 1.2;
}

.logo-text span {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.nav-menu {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  flex: 1;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.875rem 1rem;
  border-radius: var(--radius-lg);
  color: var(--text-secondary);
  font-weight: 500;
  transition: var(--transition);
  cursor: pointer;
}

.nav-item:hover {
  background-color: var(--neutral-50);
  color: var(--primary);
}

.nav-item.active {
  background-color: var(--primary);
  color: var(--white);
  box-shadow: 0 4px 14px 0 rgba(79, 70, 229, 0.39);
}

.sidebar-footer {
  margin-top: auto;
}

.new-recording-btn {
  border-radius: var(--radius-md);
  padding: 1rem;
}

.mt-4 { margin-top: 1rem; }
.text-left { text-align: left; }
.justify-start { justify-content: flex-start; }

@media (max-width: 768px) {
  .sidebar {
    transform: translateX(-100%);
    box-shadow: var(--shadow-lg);
  }
  
  .sidebar.sidebar-open {
    transform: translateX(0);
  }
}
</style>
