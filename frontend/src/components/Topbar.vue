<template>
  <header class="topbar">
    <div class="left-section">
      <button class="hamburger-btn" @click="toggleSidebar" aria-label="Toggle navigation menu">
        <Menu size="24" />
      </button>
      
      <div class="search-container">
        <Search class="search-icon" size="18" />
        <input v-model="searchQuery" type="text" class="search-input" placeholder="Search recordings or insights..." />
      </div>
    </div>

    <div class="user-profile">
      <button class="notification-btn">
        <Bell size="20" />
        <span class="badge"></span>
      </button>
      
      <div class="user-info">
        <div class="user-details text-right hidden sm:block">
          <p class="name">{{ authStore.user?.email || 'User' }}</p>
          <p class="plan">{{ authStore.user?.role === 'admin' ? 'Admin' : 'Pro Plan' }}</p>
        </div>
        <div class="avatar" :title="authStore.user?.email || 'User'">
          {{ authStore.user?.email?.[0].toUpperCase() || 'U' }}
        </div>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Search, Bell, Menu } from '@lucide/vue'
import { useAuthStore } from '../stores/auth'
import { toggleSidebar } from '../stores/layout'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()
const searchQuery = ref(route.query.search || '')

// Watch for search query input change
watch(searchQuery, (newVal) => {
  router.push({
    path: '/recordings',
    query: { ...route.query, search: newVal || undefined }
  })
})

// Sync search query with route if it changes externally
watch(() => route.query.search, (newVal) => {
  searchQuery.value = newVal || ''
})
</script>

<style scoped>
.topbar {
  height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 2rem;
  background-color: var(--bg-primary);
  border-bottom: 1px solid var(--neutral-200);
}

.left-section {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex: 1;
  max-width: 400px;
}

.hamburger-btn {
  display: none;
  background: transparent;
  border: none;
  color: var(--text-primary);
  cursor: pointer;
  padding: 0.5rem;
  border-radius: var(--radius-md);
  transition: var(--transition);
}

.hamburger-btn:hover {
  background-color: var(--neutral-100);
}

.search-container {
  position: relative;
  width: 100%;
}

.search-icon {
  position: absolute;
  left: 1rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
}

.search-input {
  width: 100%;
  padding: 0.75rem 1rem 0.75rem 2.75rem;
  border-radius: var(--radius-full);
  border: 1px solid transparent;
  background-color: var(--white);
  font-family: var(--font-body);
  transition: var(--transition);
}

.search-input:focus {
  outline: none;
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  margin-left: 1rem;
}

.notification-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  position: relative;
  transition: var(--transition);
}

.notification-btn:hover {
  color: var(--primary);
}

.badge {
  position: absolute;
  top: 0;
  right: 0;
  width: 8px;
  height: 8px;
  background-color: #EF4444;
  border-radius: 50%;
  border: 2px solid var(--bg-primary);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.user-details .name {
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--text-primary);
  max-width: 180px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-details .plan {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--secondary), var(--primary));
  color: var(--white);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-family: var(--font-headline);
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .topbar {
    padding: 0 1.25rem;
  }
  
  .hamburger-btn {
    display: flex;
    align-items: center;
    justify-content: center;
  }
}

@media (max-width: 576px) {
  .user-details {
    display: none;
  }
  
  .search-container {
    display: none; /* Hide search bar on tiny screens or keep it small, let's keep topbar clean */
  }
}
</style>
