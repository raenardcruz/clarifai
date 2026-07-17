<template>
  <div class="app-container">
    <Toast />
    <Sidebar v-if="authStore.isAuthenticated && showNavigation" />
    <div 
      class="sidebar-backdrop" 
      v-if="authStore.isAuthenticated && showNavigation && isSidebarOpen" 
      @click="closeSidebar"
    ></div>
    <div class="main-content" :class="{'full-width': !authStore.isAuthenticated || !showNavigation}">
      <Topbar v-if="authStore.isAuthenticated && showNavigation" />
      <div class="page-content">
        <router-view />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from './stores/auth'
import { isSidebarOpen, closeSidebar } from './stores/layout'
import Sidebar from './components/Sidebar.vue'
import Topbar from './components/Topbar.vue'
import Toast from 'primevue/toast'

const route = useRoute()
const authStore = useAuthStore()

const showNavigation = computed(() => {
  return route.name !== 'login' && route.name !== 'register' && route.name !== 'share'
})

// Close sidebar on route change on mobile
watch(() => route.path, () => {
  closeSidebar()
})

onMounted(() => {
  authStore.fetchUser()
})
</script>

<style>
.app-container {
  display: flex;
  min-height: 100vh;
  background-color: var(--bg-primary);
  color: var(--text-primary);
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  margin-left: var(--sidebar-width);
  transition: margin-left 0.3s ease;
  min-width: 0; /* Prevents flex items from overflowing */
}

.main-content.full-width {
  margin-left: 0;
}

.page-content {
  flex: 1;
  overflow-y: auto;
}

.sidebar-backdrop {
  display: none;
}

@media (max-width: 768px) {
  .main-content {
    margin-left: 0;
  }
  
  .sidebar-backdrop {
    display: block;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(15, 23, 42, 0.4);
    backdrop-filter: blur(4px);
    z-index: 35;
    transition: opacity 0.3s ease;
  }
}
</style>
