<template>
  <div class="recordings-page-wrapper">
    <div class="recordings-page">
      <!-- Page Header -->
      <div class="page-header">
        <div class="header-left">
          <h1 class="page-title">All Recordings</h1>
          <p class="page-subtitle">Manage, search, and review your captured thoughts and transcripts</p>
        </div>
        
        <button class="apple-primary-btn" @click="isRecordingModalOpen = true">
          <Plus size="16" stroke-width="2.5" />
          <span>New Recording</span>
        </button>
      </div>

      <!-- Offline Pending Sync Alert Banner -->
      <div class="offline-queue-banner" v-if="offlineStore.pendingCount > 0">
        <div class="queue-banner-left">
          <HardDrive size="20" class="text-blue" />
          <div class="queue-banner-text">
            <strong>{{ offlineStore.pendingCount }} offline recording{{ offlineStore.pendingCount > 1 ? 's' : '' }} stored locally</strong>
            <span>Recorded on device. {{ offlineStore.isOnline ? 'Ready to sync with backend and Speechmatics.' : 'Will automatically sync once internet connection returns.' }}</span>
          </div>
        </div>
        <div class="queue-banner-actions">
          <button 
            class="banner-sync-btn" 
            :disabled="!offlineStore.isOnline || offlineStore.isSyncing"
            @click="offlineStore.syncAll()"
          >
            <RefreshCw size="13" :class="{ 'animate-spin': offlineStore.isSyncing }" />
            <span>{{ offlineStore.isSyncing ? 'Syncing...' : (offlineStore.isOnline ? 'Sync All Now' : 'Offline') }}</span>
          </button>
          <button class="banner-view-btn" @click="currentFilter = 'offline'">
            <span>View Queue</span>
          </button>
        </div>
      </div>

      <!-- macOS Toolbar: Filters, Sorting, and View Switcher -->
      <div class="macos-toolbar" v-if="recordings.length > 0 || route.query.search || offlineStore.offlineList.length > 0">
        <!-- Status Filter Segmented Control -->
        <div class="segmented-pills">
          <button 
            class="pill-btn" 
            :class="{ active: currentFilter === 'all' }"
            @click="currentFilter = 'all'"
          >
            All <span class="counter-badge">{{ recordings.length }}</span>
          </button>
          <button 
            class="pill-btn" 
            :class="{ active: currentFilter === 'ready' }"
            @click="currentFilter = 'ready'"
          >
            Ready <span class="counter-badge">{{ readyCount }}</span>
          </button>
          <button 
            class="pill-btn" 
            :class="{ active: currentFilter === 'processing' }"
            @click="currentFilter = 'processing'"
          >
            Processing <span class="counter-badge">{{ processingCount }}</span>
          </button>
          <button 
            class="pill-btn" 
            :class="{ active: currentFilter === 'medical' }"
            @click="currentFilter = 'medical'"
          >
            Medical <span class="counter-badge">{{ medicalCount }}</span>
          </button>
          <button 
            class="pill-btn" 
            :class="{ active: currentFilter === 'offline' }"
            @click="currentFilter = 'offline'"
          >
            Offline Queue <span class="counter-badge" :class="{ 'has-pending-badge': offlineStore.pendingCount > 0 }">{{ offlineStore.offlineList.length }}</span>
          </button>
        </div>


        <!-- Right Side: Sort & View Toggle -->
        <div class="toolbar-right">
          <!-- Sort Selector -->
          <div class="sort-selector">
            <span class="sort-label">Sort:</span>
            <select v-model="sortBy" class="apple-select">
              <option value="newest">Newest First</option>
              <option value="oldest">Oldest First</option>
              <option value="duration">Longest Duration</option>
            </select>
          </div>

          <!-- macOS View Toggle (Grid / List) -->
          <div class="view-toggle-segment">
            <button 
              class="view-toggle-btn" 
              :class="{ active: viewMode === 'grid' }"
              @click="viewMode = 'grid'"
              title="Grid View"
            >
              <LayoutGrid size="16" />
            </button>
            <button 
              class="view-toggle-btn" 
              :class="{ active: viewMode === 'list' }"
              @click="viewMode = 'list'"
              title="List View"
            >
              <List size="16" />
            </button>
          </div>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="loading-state">
        <Loader2 class="animate-spin text-blue" size="36" />
        <span class="loading-text">Loading recordings...</span>
      </div>
      
      <!-- Empty State -->
      <div v-else-if="filteredRecordings.length === 0" class="apple-empty-state">
        <div class="concentric-ring-wrapper">
          <div class="ring ring-outer"></div>
          <div class="ring ring-mid"></div>
          <div class="empty-icon-circle">
            <Mic size="28" stroke-width="2.2" />
          </div>
        </div>

        <h3 class="empty-title">
          {{ route.query.search ? 'No matching recordings found' : 'No recordings yet' }}
        </h3>
        <p class="empty-desc">
          {{ route.query.search 
            ? `No recordings found matching "${route.query.search}". Try a different keyword.` 
            : 'Capture audio live with your microphone or upload audio files to start generating AI transcripts and summaries.' 
          }}
        </p>
        
        <button class="apple-primary-btn mt-6" @click="isRecordingModalOpen = true">
          <Plus size="16" />
          <span>{{ route.query.search ? 'Record New Meeting' : 'Create First Recording' }}</span>
        </button>
      </div>

      <!-- Offline Recordings Queue Mode -->
      <div v-else-if="currentFilter === 'offline'" class="offline-queue-view">
        <div v-if="offlineStore.offlineList.length === 0" class="apple-empty-state">
          <div class="concentric-ring-wrapper">
            <div class="ring ring-outer"></div>
            <div class="ring ring-mid"></div>
            <div class="empty-icon-circle">
              <HardDrive size="28" stroke-width="2.2" />
            </div>
          </div>
          <h3 class="empty-title">No offline recordings</h3>
          <p class="empty-desc">
            Recordings you capture on this device while offline or in unstable networks will be stored here and automatically uploaded to Speechmatics once online.
          </p>
          <button class="apple-primary-btn mt-6" @click="isRecordingModalOpen = true">
            <Plus size="16" />
            <span>Record New Audio</span>
          </button>
        </div>

        <div v-else>
          <div class="offline-queue-header-bar">
            <div class="offline-queue-header-info">
              <h3 class="offline-header-title">Recordings Stored on This Device</h3>
              <p class="offline-header-sub">
                {{ offlineStore.offlineList.length }} local file{{ offlineStore.offlineList.length === 1 ? '' : 's' }} cached in device storage
                <span v-if="offlineStore.pendingCount > 0" class="offline-header-pending">
                  • {{ offlineStore.pendingCount }} waiting to upload
                </span>
              </p>
            </div>
            <div class="offline-queue-header-actions">
              <button 
                class="apple-btn-secondary danger-action-btn" 
                @click="showClearAllOfflineDialog = true"
                title="Delete all local recordings from this device"
              >
                <Trash2 size="14" />
                <span>Delete All Local</span>
              </button>
              <button 
                class="apple-primary-btn" 
                :disabled="!offlineStore.isOnline || offlineStore.isSyncing || offlineStore.pendingCount === 0"
                @click="offlineStore.syncAll()"
              >
                <RefreshCw size="14" :class="{ 'animate-spin': offlineStore.isSyncing }" />
                <span>{{ offlineStore.isSyncing ? 'Syncing...' : 'Sync All Pending' }}</span>
              </button>
            </div>
          </div>

          <div class="grid grid-cols-3 gap-4">
            <OfflineRecordingCard 
              v-for="rec in offlineStore.offlineList" 
              :key="rec.id" 
              :recording="rec" 
            />
          </div>
        </div>
      </div>

      <!-- Grid View Mode -->
      <div v-else-if="viewMode === 'grid'" class="grid grid-cols-4">
        <RecordingCard 
          v-for="recording in filteredRecordings" 
          :key="recording.id" 
          :recording="recording" 
          @refresh="fetchRecordings"
        />
      </div>

      <!-- macOS List View Mode -->
      <div v-else class="apple-list-view">
        <div class="list-table-header">
          <span class="col-title">RECORDING TITLE</span>
          <span class="col-status">STATUS</span>
          <span class="col-duration">DURATION</span>
          <span class="col-date">CREATED DATE</span>
          <span class="col-actions">ACTION</span>
        </div>

        <div class="list-rows-group">
          <div 
            v-for="recording in filteredRecordings" 
            :key="recording.id"
            class="list-row"
            @click="goToRecording(recording.id)"
          >
            <!-- Title Column -->
            <div class="col-title row-title-cell">
              <div class="row-icon-circle">
                <Mic size="16" class="text-blue" />
              </div>
              <div class="row-title-info">
                <div class="row-title-line">
                  <span class="row-title-text">{{ recording.title }}</span>
                  <span v-if="recording.is_medical" class="row-medical-pill" title="Medical Conversation">
                    <Stethoscope size="11" />
                    <span>Medical</span>
                  </span>
                </div>
                <span class="row-preview-text" v-if="getSnippet(recording)">{{ getSnippet(recording) }}</span>
              </div>
            </div>


            <!-- Status Column -->
            <div class="col-status">
              <span class="status-pill-compact" :class="getStatusClass(recording.status)">
                <span class="status-dot"></span>
                <span>{{ capitalize(recording.status) }}</span>
              </span>
            </div>

            <!-- Duration Column -->
            <div class="col-duration font-tabular">
              {{ formatDuration(recording.duration) }}
            </div>

            <!-- Date Column -->
            <div class="col-date font-tabular">
              {{ formatDate(recording.created_at) }}
            </div>

            <!-- Action Column -->
            <div class="col-actions" @click.stop>
              <router-link :to="`/recordings/${recording.id}`" class="row-view-btn" title="Open Recording">
                <ChevronRight size="16" />
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Speechmatics Usage Footer Dock -->
    <SpeechmaticsUsage v-if="!loading && recordings.length > 0" type="footer" />

    <!-- Clear All Offline Recordings Dialog -->
    <Dialog 
      v-model:visible="showClearAllOfflineDialog" 
      modal 
      header="Delete All Local Recordings"
      :style="{ width: '90vw', maxWidth: '420px' }"
    >
      <div class="delete-dialog-body">
        <div class="delete-dialog-icon">
          <Trash2 size="24" class="text-red-500" />
        </div>
        <p class="delete-dialog-text">
          Are you sure you want to delete all <strong>{{ offlineStore.offlineList.length }}</strong> local recordings from this iPhone? 
          Any unsynced audio files will be permanently lost and cannot be recovered.
        </p>
      </div>
      <template #footer>
        <div class="delete-dialog-actions">
          <button 
            type="button" 
            class="apple-dialog-cancel-btn" 
            @click="showClearAllOfflineDialog = false"
            :disabled="isClearingOffline"
          >
            Cancel
          </button>
          <button 
            type="button" 
            class="apple-dialog-delete-btn" 
            @click="handleClearAllOffline"
            :disabled="isClearingOffline"
          >
            <Loader2 class="animate-spin" size="14" v-if="isClearingOffline" />
            <span v-else>Delete All</span>
          </button>
        </div>
      </template>
    </Dialog>

    <!-- Teleport Modal out of view -->
    <Teleport to="body">
      <NewRecordingModal v-if="isRecordingModalOpen" @close="isRecordingModalOpen = false" @refresh="fetchRecordings" />
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import Dialog from 'primevue/dialog'
import axios from 'axios'
import moment from 'moment'
import { Loader2, Mic, Plus, LayoutGrid, List, ChevronRight, Stethoscope, HardDrive, RefreshCw, Trash2 } from '@lucide/vue'
import RecordingCard from '../components/RecordingCard.vue'
import OfflineRecordingCard from '../components/OfflineRecordingCard.vue'
import NewRecordingModal from '../components/NewRecordingModal.vue'
import SpeechmaticsUsage from '../components/SpeechmaticsUsage.vue'
import { useOfflineRecordingsStore } from '../stores/offlineRecordings'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const offlineStore = useOfflineRecordingsStore()
const recordings = ref([])
const loading = ref(true)
const isRecordingModalOpen = ref(false)
const currentFilter = ref('all')
const sortBy = ref('newest')
const viewMode = ref('grid')
const showClearAllOfflineDialog = ref(false)
const isClearingOffline = ref(false)

const handleClearAllOffline = async () => {
  isClearingOffline.value = true
  try {
    await offlineStore.clearAllRecordings()
    showClearAllOfflineDialog.value = false
    toast.add({
      severity: 'success',
      summary: 'Deleted',
      detail: 'All local recordings removed from device storage.',
      life: 3000
    })
  } catch (err) {
    console.error('Failed to clear offline recordings:', err)
    toast.add({
      severity: 'error',
      summary: 'Error',
      detail: 'Failed to delete local recordings.',
      life: 4000
    })
  } finally {
    isClearingOffline.value = false
  }
}

const fetchRecordings = async () => {
  loading.value = true
  try {
    const res = await axios.get('/api/recordings', {
      params: { search: route.query.search || undefined }
    })
    recordings.value = res.data
  } catch (e) {
    console.error("Failed to fetch recordings", e)
    // If network failed and we have offline recordings, switch filter to offline
    if (offlineStore.offlineList.length > 0 && recordings.value.length === 0) {
      currentFilter.value = 'offline'
    }
  } finally {
    loading.value = false
  }
}

watch(() => route.query.search, () => {
  fetchRecordings()
})

onMounted(() => {
  fetchRecordings()
})

const readyCount = computed(() => {
  return recordings.value.filter(r => r.status === 'summarized' || r.status === 'completed').length
})

const processingCount = computed(() => {
  return recordings.value.filter(r => ['pending', 'transcribing', 'diarizing', 'summarizing'].includes(r.status)).length
})

const medicalCount = computed(() => {
  return recordings.value.filter(r => r.is_medical).length
})

const filteredRecordings = computed(() => {
  let list = [...recordings.value]

  // Filter
  if (currentFilter.value === 'ready') {
    list = list.filter(r => r.status === 'summarized' || r.status === 'completed')
  } else if (currentFilter.value === 'processing') {
    list = list.filter(r => ['pending', 'transcribing', 'diarizing', 'summarizing'].includes(r.status))
  } else if (currentFilter.value === 'medical') {
    list = list.filter(r => r.is_medical)
  }

  // Sort
  if (sortBy.value === 'newest') {
    list.sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
  } else if (sortBy.value === 'oldest') {
    list.sort((a, b) => new Date(a.created_at) - new Date(b.created_at))
  } else if (sortBy.value === 'duration') {
    list.sort((a, b) => (b.duration || 0) - (a.duration || 0))
  }

  return list
})


const goToRecording = (id) => {
  router.push(`/recordings/${id}`)
}

const formatDuration = (duration) => {
  if (!duration) return '00:00'
  const mins = Math.floor(duration / 60)
  const secs = Math.floor(duration % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

const formatDate = (dateStr) => {
  return moment(dateStr).format('MMM D, YYYY')
}

const capitalize = (str) => {
  if (!str) return ''
  return str.charAt(0).toUpperCase() + str.slice(1)
}

const getStatusClass = (status) => {
  if (['pending', 'transcribing', 'diarizing', 'summarizing'].includes(status)) return 'status-proc'
  if (status === 'error') return 'status-err'
  return 'status-ok'
}

const getSnippet = (recording) => {
  if (recording.summary_md) {
    const lines = recording.summary_md.split('\n')
    const first = lines.find(l => l.trim().length > 10 && !l.startsWith('#'))
    return first ? first.substring(0, 75) + '...' : ''
  }
  return ''
}
</script>

<style scoped>
.recordings-page-wrapper {
  display: flex;
  flex-direction: column;
  min-height: calc(100vh - var(--header-height));
  justify-content: space-between;
}

.recordings-page {
  max-width: 1400px;
  width: 100%;
  margin: 0 auto;
  padding: 2.5rem 2.5rem;
  flex: 1;
}

/* Page Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 2rem;
}

.page-title {
  font-size: 2rem;
  font-weight: 700;
  letter-spacing: -0.03em;
  color: var(--text-primary);
  margin-bottom: 0.35rem;
}

.page-subtitle {
  font-size: 0.95rem;
  color: var(--text-secondary);
}

.apple-primary-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--apple-blue);
  color: #FFFFFF;
  border: 1px solid rgba(255, 255, 255, 0.2);
  padding: 0.65rem 1.25rem;
  border-radius: var(--radius-full);
  font-weight: 600;
  font-size: 0.88rem;
  cursor: pointer;
  transition: all 0.2s var(--apple-ease);
  box-shadow: 0 4px 12px rgba(0, 113, 227, 0.25);
}

.apple-primary-btn:hover {
  background: var(--apple-blue-hover);
  transform: translateY(-1px);
  box-shadow: 0 6px 16px rgba(0, 113, 227, 0.35);
}

.apple-primary-btn:active {
  transform: scale(0.98);
}

/* macOS Toolbar */
.macos-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.75rem;
  padding: 0.6rem 0.85rem;
  background: var(--bg-secondary);
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: var(--radius-md);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.02);
  flex-wrap: wrap;
  gap: 1rem;
}

.segmented-pills {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.pill-btn {
  background: transparent;
  border: none;
  padding: 0.35rem 0.75rem;
  border-radius: var(--radius-sm);
  font-size: 0.825rem;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
  transition: var(--transition-fast);
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.pill-btn:hover {
  background: rgba(0, 0, 0, 0.04);
  color: var(--text-primary);
}

.pill-btn.active {
  background: rgba(0, 113, 227, 0.1);
  color: var(--apple-blue);
  font-weight: 600;
}

.counter-badge {
  font-size: 0.7rem;
  background: rgba(0, 0, 0, 0.06);
  padding: 0.05rem 0.4rem;
  border-radius: var(--radius-full);
}

.pill-btn.active .counter-badge {
  background: rgba(0, 113, 227, 0.2);
  color: var(--apple-blue);
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 1.25rem;
}

.sort-selector {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.825rem;
}

.sort-label {
  color: var(--text-muted);
}

.apple-select {
  background: transparent;
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: var(--radius-sm);
  padding: 0.3rem 0.6rem;
  font-size: 0.825rem;
  font-family: inherit;
  color: var(--text-primary);
  outline: none;
  cursor: pointer;
}

.view-toggle-segment {
  display: inline-flex;
  align-items: center;
  background: rgba(118, 118, 128, 0.12);
  border-radius: var(--radius-sm);
  padding: 2px;
  gap: 2px;
}

.view-toggle-btn {
  background: transparent;
  border: none;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  cursor: pointer;
  transition: var(--transition-fast);
}

.view-toggle-btn.active {
  background: var(--white);
  color: var(--text-primary);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

/* Loading State */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 6rem 2rem;
  gap: 1rem;
}

.loading-text {
  font-size: 0.95rem;
  color: var(--text-muted);
  font-weight: 500;
}

.text-blue {
  color: var(--apple-blue);
}

/* Apple Concentric Ring Empty State */
.apple-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 5rem 2rem;
  background: var(--bg-secondary);
  border-radius: var(--radius-2xl);
  border: 1px solid rgba(0, 0, 0, 0.06);
  text-align: center;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.02);
}

.concentric-ring-wrapper {
  position: relative;
  width: 100px;
  height: 100px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.5rem;
}

.ring {
  position: absolute;
  border-radius: var(--radius-full);
  border: 1px solid rgba(0, 113, 227, 0.15);
}

.ring-outer {
  width: 100px;
  height: 100px;
  opacity: 0.4;
}

.ring-mid {
  width: 76px;
  height: 76px;
  opacity: 0.7;
}

.empty-icon-circle {
  width: 52px;
  height: 52px;
  border-radius: var(--radius-full);
  background: var(--apple-blue-light);
  color: var(--apple-blue);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 14px rgba(0, 113, 227, 0.2);
}

.empty-title {
  font-size: 1.35rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
}

.empty-desc {
  max-width: 440px;
  font-size: 0.9rem;
  line-height: 1.5;
  color: var(--text-secondary);
}

.mt-6 {
  margin-top: 1.5rem;
}

/* Apple List View */
.apple-list-view {
  background: var(--bg-secondary);
  border-radius: var(--radius-lg);
  border: 1px solid rgba(0, 0, 0, 0.06);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.02);
  overflow: hidden;
}

.list-table-header {
  display: grid;
  grid-template-columns: 2.5fr 1fr 1fr 1fr 40px;
  padding: 0.85rem 1.5rem;
  background: rgba(0, 0, 0, 0.02);
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  font-size: 0.7rem;
  font-weight: 700;
  color: var(--text-muted);
  letter-spacing: 0.05em;
}

.list-row {
  display: grid;
  grid-template-columns: 2.5fr 1fr 1fr 1fr 40px;
  padding: 1rem 1.5rem;
  align-items: center;
  border-bottom: 1px solid rgba(0, 0, 0, 0.04);
  transition: var(--transition-fast);
  cursor: pointer;
}

.list-row:last-child {
  border-bottom: none;
}

.list-row:hover {
  background: rgba(0, 113, 227, 0.03);
}

.row-title-cell {
  display: flex;
  align-items: center;
  gap: 0.85rem;
}

.row-icon-circle {
  width: 34px;
  height: 34px;
  border-radius: var(--radius-sm);
  background: var(--apple-blue-light);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.row-title-info {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.row-title-line {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.row-title-text {
  font-weight: 600;
  font-size: 0.92rem;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.row-medical-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.15rem 0.45rem;
  border-radius: var(--radius-full);
  font-size: 0.68rem;
  font-weight: 600;
  background-color: rgba(16, 185, 129, 0.12);
  color: #047857;
  border: 1px solid rgba(16, 185, 129, 0.25);
  flex-shrink: 0;
}

.list-row:hover .row-title-text {
  color: var(--apple-blue);
}


.row-preview-text {
  font-size: 0.75rem;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.status-pill-compact {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.2rem 0.55rem;
  border-radius: var(--radius-full);
  font-size: 0.72rem;
  font-weight: 600;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: var(--radius-full);
}

.status-ok { background: var(--apple-green-light); color: #1E8E3E; }
.status-ok .status-dot { background: #34C759; }

.status-proc { background: var(--apple-blue-light); color: var(--apple-blue); }
.status-proc .status-dot { background: var(--apple-blue); }

.status-err { background: var(--apple-red-light); color: var(--apple-red); }
.status-err .status-dot { background: var(--apple-red); }

.font-tabular {
  font-variant-numeric: tabular-nums;
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.row-view-btn {
  color: var(--text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: var(--transition-fast);
}

.list-row:hover .row-view-btn {
  color: var(--apple-blue);
  transform: translateX(2px);
}

@media (max-width: 900px) {
  .recordings-page {
    padding: 1.5rem;
  }
  
  .list-table-header, .list-row {
    grid-template-columns: 2fr 1fr 1fr 30px;
  }
  
  .col-date {
    display: none;
  }
}

/* Offline Queue Banner */
.offline-queue-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  margin-bottom: 1.5rem;
  border-radius: var(--radius-lg, 12px);
  background: rgba(0, 113, 227, 0.06);
  border: 1px solid rgba(0, 113, 227, 0.2);
  gap: 1rem;
}

.queue-banner-left {
  display: flex;
  align-items: center;
  gap: 0.85rem;
}

.queue-banner-text {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.queue-banner-text strong {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-primary);
}

.queue-banner-text span {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.queue-banner-actions {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  flex-shrink: 0;
}

.banner-sync-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.45rem 0.9rem;
  border-radius: 8px;
  background: var(--apple-blue, #0071E3);
  color: #ffffff;
  border: none;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.banner-sync-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.banner-sync-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.banner-view-btn {
  display: inline-flex;
  align-items: center;
  padding: 0.45rem 0.85rem;
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.05);
  border: 1px solid rgba(0, 0, 0, 0.08);
  font-size: 0.8rem;
  font-weight: 500;
  color: var(--text-primary);
  cursor: pointer;
}

.counter-badge.has-pending-badge {
  background: var(--apple-blue);
  color: white;
}

.offline-queue-view {
  margin-top: 1rem;
}

.offline-queue-header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: 1rem 1.25rem;
  margin-bottom: 1.5rem;
  gap: 1rem;
  flex-wrap: wrap;
}

.offline-queue-header-info {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.offline-header-title {
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.offline-header-sub {
  font-size: 0.82rem;
  color: var(--text-secondary);
  margin: 0;
}

.offline-header-pending {
  color: var(--apple-blue);
  font-weight: 600;
}

.offline-queue-header-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.danger-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.55rem 1rem;
  border-radius: var(--radius-full);
  background: rgba(255, 59, 48, 0.08);
  border: 1px solid rgba(255, 59, 48, 0.2);
  color: var(--apple-red);
  font-size: 0.82rem;
  font-weight: 600;
  cursor: pointer;
  touch-action: manipulation;
  transition: var(--transition);
}

.danger-action-btn:hover {
  background: rgba(255, 59, 48, 0.15);
  color: #b71c1c;
}

/* Apple Dialog Styling */
.delete-dialog-body {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 0.5rem 0 1rem 0;
}

.delete-dialog-icon {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  background: rgba(255, 59, 48, 0.1);
  color: var(--apple-red);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.delete-dialog-text {
  font-size: 0.92rem;
  color: var(--text-primary);
  line-height: 1.5;
  margin: 0;
}

.delete-dialog-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
  width: 100%;
}

.apple-dialog-cancel-btn {
  padding: 0.55rem 1.1rem;
  border-radius: var(--radius-full);
  background: var(--neutral-100);
  border: 1px solid rgba(0, 0, 0, 0.08);
  color: var(--text-primary);
  font-weight: 600;
  font-size: 0.85rem;
  cursor: pointer;
  touch-action: manipulation;
  transition: var(--transition);
}

.apple-dialog-cancel-btn:hover {
  background: var(--neutral-200);
}

.apple-dialog-delete-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.55rem 1.25rem;
  border-radius: var(--radius-full);
  background: var(--apple-red);
  color: #ffffff;
  border: none;
  font-weight: 600;
  font-size: 0.85rem;
  cursor: pointer;
  touch-action: manipulation;
  box-shadow: 0 2px 8px rgba(255, 59, 48, 0.3);
  transition: var(--transition);
}

.apple-dialog-delete-btn:hover:not(:disabled) {
  background: #e02d24;
  transform: translateY(-1px);
}

.apple-dialog-delete-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
