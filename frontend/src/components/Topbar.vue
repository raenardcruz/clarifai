<template>
  <header class="topbar">
    <div class="left-section">
      <button class="hamburger-btn" @click="toggleSidebar" aria-label="Toggle navigation menu">
        <Menu size="20" />
      </button>
      
      <!-- Apple Spotlight-Style Search Field -->
      <div class="spotlight-search-container" :class="{ 'is-focused': isSearchFocused }">
        <Search class="search-icon" size="15" stroke-width="2.2" />
        <input 
          ref="searchInputRef"
          v-model="searchQuery" 
          type="text" 
          class="search-input" 
          placeholder="Search recordings, summaries or speakers..." 
          @focus="isSearchFocused = true"
          @blur="isSearchFocused = false"
        />
        <button v-if="searchQuery" class="clear-search-btn" @click="clearSearch" title="Clear">
          <X size="13" />
        </button>
        <div v-else class="keyboard-shortcut-hint">
          <span>⌘K</span>
        </div>
      </div>
    </div>

    <!-- Right Utility Controls -->
    <div class="right-section">
      <!-- iOS App Reload Button (Only visible on installed iOS app) -->
      <button 
        v-if="isIosApp"
        class="ios-refresh-btn"
        :class="{ 'is-refreshing': isRefreshing }"
        @click="handlePageRefresh"
        title="Reload Page"
        aria-label="Reload Page"
      >
        <RotateCw size="15" :class="{ 'animate-spin': isRefreshing }" />
      </button>

      <!-- Network & Offline Sync Pill -->
      <button 
        class="network-sync-pill"
        :class="{ 'is-offline': !offlineStore.isOnline, 'is-syncing': offlineStore.isSyncing, 'has-pending': offlineStore.pendingCount > 0 }"
        @click="handleSyncClick"
        :title="offlineStore.isOnline ? (offlineStore.pendingCount > 0 ? 'Click to sync pending recordings' : 'Online and synchronized') : 'Device offline - recordings are safely cached locally'"
      >
        <RefreshCw size="13" class="animate-spin" v-if="offlineStore.isSyncing" />
        <WifiOff size="13" class="text-amber" v-else-if="!offlineStore.isOnline" />
        <CloudUpload size="13" class="text-blue" v-else-if="offlineStore.pendingCount > 0" />
        <Wifi size="13" class="text-emerald" v-else />

        <span class="pill-text">
          {{ !offlineStore.isOnline ? 'Offline' : (offlineStore.isSyncing ? 'Syncing...' : (offlineStore.pendingCount > 0 ? `Sync (${offlineStore.pendingCount})` : 'Online')) }}
        </span>
      </button>

      <div class="user-profile">
        <div class="user-details text-right hidden sm:block">
          <p class="name">{{ authStore.user?.email || 'User' }}</p>
          <span class="role-chip">{{ authStore.user?.role === 'admin' ? 'Admin' : 'Pro Member' }}</span>
        </div>
        
        <div class="avatar-ring" :title="authStore.user?.email || 'User'">
          <div class="avatar">
            {{ authStore.user?.email?.[0].toUpperCase() || 'U' }}
          </div>
          <span class="status-indicator"></span>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Capacitor } from '@capacitor/core'
import { Search, Menu, X, Wifi, WifiOff, CloudUpload, RefreshCw, RotateCw } from '@lucide/vue'
import { useAuthStore } from '../stores/auth'
import { useOfflineRecordingsStore } from '../stores/offlineRecordings'
import { toggleSidebar } from '../stores/layout'

const isIosApp = computed(() => {
  return Capacitor.isNativePlatform() && Capacitor.getPlatform() === 'ios'
})

const isRefreshing = ref(false)

const handlePageRefresh = () => {
  if (isRefreshing.value) return
  isRefreshing.value = true
  setTimeout(() => {
    window.location.reload()
  }, 100)
}

const authStore = useAuthStore()
const offlineStore = useOfflineRecordingsStore()
const router = useRouter()
const route = useRoute()
const searchQuery = ref(route.query.search || '')
const isSearchFocused = ref(false)
const searchInputRef = ref(null)

const handleSyncClick = () => {
  if (offlineStore.isOnline && offlineStore.pendingCount > 0) {
    offlineStore.syncAll()
  }
}

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

const clearSearch = () => {
  searchQuery.value = ''
  if (searchInputRef.value) {
    searchInputRef.value.focus()
  }
}

// Global ⌘K shortcut listener
const handleKeyDown = (e) => {
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    if (searchInputRef.value) {
      searchInputRef.value.focus()
      searchInputRef.value.select()
    }
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
})
</script>

<style scoped>
.topbar {
  height: var(--header-height);
  min-height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: var(--safe-area-top, env(safe-area-inset-top, 0px));
  padding-bottom: 0;
  padding-left: calc(2rem + var(--safe-area-left, env(safe-area-inset-left, 0px)));
  padding-right: calc(2rem + var(--safe-area-right, env(safe-area-inset-right, 0px)));
  background-color: rgba(245, 245, 247, 0.85);
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  position: sticky;
  top: 0;
  z-index: 30;
  box-sizing: border-box;
}

.left-section {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex: 1;
  max-width: 460px;
}

.hamburger-btn {
  display: none;
  background: transparent;
  border: none;
  color: var(--text-primary);
  cursor: pointer;
  padding: 0.5rem;
  border-radius: var(--radius-sm);
  transition: var(--transition);
  -webkit-tap-highlight-color: transparent;
  touch-action: manipulation;
}

.hamburger-btn:hover {
  background-color: rgba(0, 0, 0, 0.05);
}

/* Apple Spotlight Search Field */
.spotlight-search-container {
  position: relative;
  width: 100%;
  display: flex;
  align-items: center;
  background: rgba(118, 118, 128, 0.08);
  border: 1px solid rgba(0, 0, 0, 0.05);
  border-radius: var(--radius-full);
  transition: var(--transition);
  padding: 0 0.85rem 0 0.75rem;
  height: 38px;
}

.spotlight-search-container:hover {
  background: rgba(118, 118, 128, 0.12);
}

.spotlight-search-container.is-focused {
  background: var(--white);
  border-color: var(--apple-blue);
  box-shadow: 0 0 0 3px rgba(0, 113, 227, 0.18), 0 2px 8px rgba(0, 0, 0, 0.04);
}

.search-icon {
  color: var(--text-muted);
  flex-shrink: 0;
  margin-right: 0.5rem;
}

.search-input {
  width: 100%;
  border: none;
  background: transparent;
  font-family: var(--font-body);
  font-size: 0.875rem;
  color: var(--text-primary);
  outline: none;
}

.search-input::placeholder {
  color: var(--text-muted);
  font-size: 0.85rem;
}

.clear-search-btn {
  background: rgba(0, 0, 0, 0.1);
  border: none;
  border-radius: var(--radius-full);
  width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  cursor: pointer;
  transition: var(--transition-fast);
}

.clear-search-btn:hover {
  background: rgba(0, 0, 0, 0.2);
  color: var(--text-primary);
}

.keyboard-shortcut-hint {
  font-size: 0.7rem;
  font-family: var(--font-headline);
  font-weight: 600;
  color: var(--text-muted);
  background: rgba(0, 0, 0, 0.05);
  border: 1px solid rgba(0, 0, 0, 0.06);
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
}

.right-section {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  flex-shrink: 0;
}

.ios-refresh-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  min-width: 36px;
  min-height: 36px;
  border-radius: var(--radius-full);
  background: rgba(0, 0, 0, 0.04);
  border: 1px solid rgba(0, 0, 0, 0.08);
  color: var(--text-primary);
  cursor: pointer;
  touch-action: manipulation;
  -webkit-tap-highlight-color: transparent;
  transition: all 0.2s var(--apple-ease);
}

.ios-refresh-btn:hover {
  background: rgba(0, 0, 0, 0.08);
  color: var(--apple-blue);
  transform: translateY(-1px);
}

.ios-refresh-btn:active,
.ios-refresh-btn.is-refreshing {
  background: rgba(0, 113, 227, 0.12);
  color: var(--apple-blue);
  transform: scale(0.95);
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 0.85rem;
}

.user-details .name {
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-primary);
  max-width: 180px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.2;
}

.role-chip {
  display: inline-block;
  font-size: 0.68rem;
  font-weight: 600;
  color: var(--apple-blue);
  background: rgba(0, 113, 227, 0.08);
  padding: 0.1rem 0.45rem;
  border-radius: var(--radius-full);
  margin-top: 0.15rem;
}

.avatar-ring {
  position: relative;
}

.avatar {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-full);
  background: linear-gradient(135deg, #0071E3 0%, #AF52DE 100%);
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-family: var(--font-headline);
  font-size: 0.85rem;
  box-shadow: 0 2px 6px rgba(0, 113, 227, 0.25);
  border: 2px solid #ffffff;
}

.status-indicator {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 9px;
  height: 9px;
  background-color: var(--apple-green);
  border-radius: var(--radius-full);
  border: 2px solid #ffffff;
}

@media (max-width: 768px) {
  .topbar {
    padding-left: calc(1rem + var(--safe-area-left, env(safe-area-inset-left, 0px)));
    padding-right: calc(1rem + var(--safe-area-right, env(safe-area-inset-right, 0px)));
  }
  
  .hamburger-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 44px;
    min-height: 44px;
    padding: 0.5rem;
  }
  
  .spotlight-search-container {
    height: 38px;
  }

  .search-input {
    font-size: 16px;
  }
}

@media (max-width: 576px) {
  .user-details {
    display: none;
  }
}

/* Network Sync Pill */
.network-sync-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.35rem 0.75rem;
  border-radius: var(--radius-full, 9999px);
  background: rgba(0, 0, 0, 0.04);
  border: 1px solid rgba(0, 0, 0, 0.08);
  font-size: 0.76rem;
  font-weight: 600;
  color: var(--text-secondary);
  cursor: default;
  transition: all 0.2s ease;
  user-select: none;
}

.network-sync-pill.has-pending {
  background: rgba(0, 113, 227, 0.08);
  border-color: rgba(0, 113, 227, 0.25);
  color: var(--apple-blue);
  cursor: pointer;
}

.network-sync-pill.has-pending:hover {
  background: rgba(0, 113, 227, 0.14);
}

.network-sync-pill.is-offline {
  background: rgba(245, 158, 11, 0.1);
  border-color: rgba(245, 158, 11, 0.3);
  color: #b45309;
}

.network-sync-pill.is-syncing {
  background: rgba(0, 113, 227, 0.12);
  color: var(--apple-blue);
}

.text-emerald {
  color: #10b981;
}

.text-amber {
  color: #f59e0b;
}

.text-blue {
  color: var(--apple-blue);
}
</style>
