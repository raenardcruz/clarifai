<template>
  <div class="offline-card" :class="`status-${recording.status}`">
    <div class="card-header">
      <div class="header-left">
        <div class="icon-circle" :class="{ 'icon-synced': recording.status === 'synced' }">
          <HardDrive size="18" v-if="recording.status !== 'synced'" />
          <CheckCircle2 size="18" v-else />
        </div>
        <div class="title-meta">
          <div class="title-row">
            <h3 class="card-title">{{ recording.title }}</h3>
            <span v-if="recording.isMedical" class="medical-badge">Medical</span>
          </div>
          <span class="card-time font-tabular">{{ formatTime(recording.createdAt) }} • {{ formatSize(recording.fileSize) }}</span>
        </div>
      </div>

      <div class="status-badge" :class="recording.status">
        <RefreshCw size="12" class="animate-spin" v-if="recording.status === 'uploading'" />
        <WifiOff size="12" v-else-if="recording.status === 'saved_locally'" />
        <CheckCircle2 size="12" v-else-if="recording.status === 'synced'" />
        <AlertCircle size="12" v-else />
        <span>{{ statusLabel }}</span>
      </div>
    </div>

    <!-- Upload Progress Bar (when syncing) -->
    <div v-if="recording.status === 'uploading'" class="upload-progress-box">
      <div class="progress-labels">
        <span>Uploading chunks to backend...</span>
        <span class="font-tabular">{{ recording.uploadProgress || 0 }}%</span>
      </div>
      <div class="progress-track">
        <div class="progress-fill" :style="{ width: `${recording.uploadProgress || 0}%` }"></div>
      </div>
    </div>

    <!-- Error message if upload failed -->
    <div v-if="recording.status === 'error' && recording.errorMessage" class="error-banner">
      <AlertCircle size="14" />
      <span>{{ recording.errorMessage }}</span>
    </div>

    <!-- Audio Player Preview -->
    <div class="audio-player-row">
      <audio 
        v-if="audioUrl" 
        :src="audioUrl" 
        controls 
        preload="metadata" 
        class="apple-audio-player"
      ></audio>
      <button v-else class="load-audio-btn" @click="loadAudio" :disabled="isLoadingAudio">
        <Play size="14" />
        <span>{{ isLoadingAudio ? 'Loading Audio...' : 'Listen Offline' }}</span>
      </button>
    </div>

    <!-- Action Bar -->
    <div class="card-actions">
      <div class="duration-badge font-tabular">
        <Clock size="13" />
        <span>{{ formatDuration(recording.duration) }}</span>
      </div>

      <div class="action-buttons">
        <button 
          v-if="recording.status === 'synced' && recording.jobId" 
          class="btn-action btn-view" 
          @click="openCloudRecording"
        >
          <span>View Transcript</span>
          <ChevronRight size="14" />
        </button>

        <button 
          v-else-if="recording.status !== 'uploading'" 
          class="btn-action btn-sync" 
          :disabled="!isOnline" 
          @click="handleSync"
          :title="!isOnline ? 'Connect to internet to sync' : 'Sync now'"
        >
          <CloudUpload size="14" />
          <span>Sync Now</span>
        </button>

        <button class="btn-action btn-delete" @click="showDeleteDialog = true" title="Delete local recording" aria-label="Delete local recording">
          <Trash2 size="14" />
          <span class="btn-delete-text">Delete</span>
        </button>
      </div>
    </div>

    <!-- Apple-Style Delete Confirmation Dialog (Reliable on iOS WKWebView) -->
    <Dialog 
      v-model:visible="showDeleteDialog" 
      modal 
      header="Delete Local Recording"
      :style="{ width: '90vw', maxWidth: '420px' }"
    >
      <div class="delete-dialog-body">
        <div class="delete-dialog-icon">
          <Trash2 size="24" class="text-red-500" />
        </div>
        <p class="delete-dialog-text">
          Are you sure you want to delete <strong>"{{ recording.title }}"</strong>? 
          The local audio file will be permanently removed from this iPhone's storage.
        </p>
      </div>
      <template #footer>
        <div class="delete-dialog-actions">
          <button 
            type="button" 
            class="apple-dialog-cancel-btn" 
            @click="showDeleteDialog = false"
            :disabled="isDeleting"
          >
            Cancel
          </button>
          <button 
            type="button" 
            class="apple-dialog-delete-btn" 
            @click="executeDelete"
            :disabled="isDeleting"
          >
            <Loader2 class="animate-spin" size="14" v-if="isDeleting" />
            <span v-else>Delete Recording</span>
          </button>
        </div>
      </template>
    </Dialog>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import Dialog from 'primevue/dialog'
import { 
  HardDrive, 
  CheckCircle2, 
  WifiOff, 
  RefreshCw, 
  AlertCircle, 
  Play, 
  Clock, 
  CloudUpload, 
  ChevronRight, 
  Trash2,
  Loader2
} from '@lucide/vue'
import { useOfflineRecordingsStore } from '../stores/offlineRecordings'
import moment from 'moment'

const props = defineProps({
  recording: {
    type: Object,
    required: true
  }
})

const router = useRouter()
const toast = useToast()
const offlineStore = useOfflineRecordingsStore()
const isOnline = computed(() => offlineStore.isOnline)

const audioUrl = ref(null)
const isLoadingAudio = ref(false)
const showDeleteDialog = ref(false)
const isDeleting = ref(false)

const statusLabel = computed(() => {
  switch (props.recording.status) {
    case 'synced':
      return 'Synced to Cloud'
    case 'uploading':
      return `Uploading ${props.recording.uploadProgress || 0}%`
    case 'error':
      return 'Upload Failed'
    case 'saved_locally':
    default:
      return isOnline.value ? 'Queued for Sync' : 'Offline'
  }
})

const loadAudio = async () => {
  isLoadingAudio.value = true
  try {
    audioUrl.value = await offlineStore.getAudioUrl(props.recording.id)
  } catch (err) {
    console.error('Failed to load local audio:', err)
  } finally {
    isLoadingAudio.value = false
  }
}

const handleSync = async () => {
  try {
    await offlineStore.syncRecording(props.recording.id)
  } catch (e) {
    console.error('Manual sync failed:', e)
  }
}

const executeDelete = async () => {
  isDeleting.value = true
  try {
    await offlineStore.removeRecording(props.recording.id)
    showDeleteDialog.value = false
    toast.add({ 
      severity: 'success', 
      summary: 'Deleted', 
      detail: `"${props.recording.title}" was removed from device storage.`, 
      life: 3000 
    })
  } catch (e) {
    console.error('Failed to delete local recording:', e)
    toast.add({ 
      severity: 'error', 
      summary: 'Error', 
      detail: 'Failed to delete local recording.', 
      life: 4000 
    })
  } finally {
    isDeleting.value = false
  }
}

const openCloudRecording = () => {
  if (props.recording.jobId) {
    router.push(`/recordings/${props.recording.jobId}`)
  }
}

const formatTime = (iso) => {
  if (!iso) return ''
  return moment(iso).format('MMM D, h:mm A')
}

const formatDuration = (secs) => {
  if (!secs) return '00:00'
  const m = Math.floor(secs / 60)
  const s = Math.floor(secs % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}

const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const mb = bytes / (1024 * 1024)
  if (mb >= 1) return `${mb.toFixed(1)} MB`
  const kb = bytes / 1024
  return `${kb.toFixed(0)} KB`
}
</script>

<style scoped>
.offline-card {
  background: var(--bg-surface, #ffffff);
  border: 1px solid var(--border-color, rgba(0, 0, 0, 0.08));
  border-radius: var(--radius-lg, 14px);
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  transition: all 0.2s ease;
}

.offline-card:hover {
  border-color: rgba(0, 113, 227, 0.3);
  box-shadow: 0 4px 16px rgba(0, 113, 227, 0.08);
}

.card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  min-width: 0;
}

.icon-circle {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: rgba(0, 113, 227, 0.1);
  color: var(--apple-blue, #0071E3);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.icon-circle.icon-synced {
  background: rgba(16, 185, 129, 0.12);
  color: #10b981;
}

.title-meta {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  min-width: 0;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.card-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.medical-badge {
  font-size: 0.68rem;
  font-weight: 700;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  background: rgba(16, 185, 129, 0.12);
  color: #059669;
}

.card-time {
  font-size: 0.78rem;
  color: var(--text-muted);
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.3rem 0.65rem;
  border-radius: var(--radius-full, 9999px);
  font-size: 0.72rem;
  font-weight: 600;
  flex-shrink: 0;
}

.status-badge.saved_locally {
  background: rgba(245, 158, 11, 0.12);
  color: #b45309;
}

.status-badge.uploading {
  background: rgba(0, 113, 227, 0.12);
  color: var(--apple-blue);
}

.status-badge.synced {
  background: rgba(16, 185, 129, 0.12);
  color: #059669;
}

.status-badge.error {
  background: rgba(239, 68, 68, 0.12);
  color: #dc2626;
}

/* Upload Progress */
.upload-progress-box {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.progress-labels {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--apple-blue);
}

.progress-track {
  width: 100%;
  height: 5px;
  background: rgba(0, 0, 0, 0.06);
  border-radius: 999px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #0071E3, #30B0C7);
  transition: width 0.3s ease;
}

.error-banner {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border-radius: 8px;
  background: rgba(239, 68, 68, 0.08);
  color: #b91c1c;
  font-size: 0.76rem;
}

/* Audio Player */
.audio-player-row {
  width: 100%;
}

.apple-audio-player {
  width: 100%;
  height: 36px;
  border-radius: 8px;
}

.load-audio-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.4rem 0.85rem;
  background: rgba(0, 0, 0, 0.04);
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 8px;
  font-size: 0.78rem;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s ease;
}

.load-audio-btn:hover {
  background: rgba(0, 0, 0, 0.08);
  color: var(--text-primary);
}

/* Card Actions */
.card-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 0.5rem;
  border-top: 1px solid rgba(0, 0, 0, 0.05);
}

.duration-badge {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.78rem;
  color: var(--text-muted);
  font-weight: 500;
}

.action-buttons {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-action {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4rem 0.75rem;
  border-radius: 8px;
  font-size: 0.78rem;
  font-weight: 600;
  border: none;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-sync {
  background: var(--apple-blue, #0071E3);
  color: #ffffff;
}

.btn-sync:hover:not(:disabled) {
  opacity: 0.9;
}

.btn-sync:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-view {
  background: rgba(16, 185, 129, 0.12);
  color: #059669;
}

.btn-delete {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.4rem 0.65rem;
  min-height: 34px;
  background: rgba(255, 59, 48, 0.08);
  color: #dc2626;
  border: 1px solid rgba(255, 59, 48, 0.2);
  border-radius: 8px;
  font-size: 0.76rem;
  font-weight: 600;
  cursor: pointer;
  touch-action: manipulation;
  -webkit-tap-highlight-color: transparent;
  transition: all 0.15s ease;
}

.btn-delete:hover,
.btn-delete:active {
  background: rgba(239, 68, 68, 0.16);
  color: #b91c1c;
  border-color: rgba(239, 68, 68, 0.35);
  transform: scale(0.98);
}

.btn-delete-text {
  font-size: 0.74rem;
}

/* Dialog Styling */
.delete-dialog-body {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 0.5rem 0 1rem 0;
}

.delete-dialog-icon {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md, 12px);
  background: rgba(255, 59, 48, 0.1);
  color: var(--apple-red, #ff3b30);
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
  border-radius: var(--radius-full, 9999px);
  background: var(--neutral-100, #f2f2f7);
  border: 1px solid rgba(0, 0, 0, 0.08);
  color: var(--text-primary);
  font-weight: 600;
  font-size: 0.85rem;
  cursor: pointer;
  touch-action: manipulation;
  transition: var(--transition, all 0.2s ease);
}

.apple-dialog-cancel-btn:hover {
  background: var(--neutral-200, #e5e5ea);
}

.apple-dialog-delete-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.55rem 1.25rem;
  border-radius: var(--radius-full, 9999px);
  background: var(--apple-red, #ff3b30);
  color: #ffffff;
  border: none;
  font-weight: 600;
  font-size: 0.85rem;
  cursor: pointer;
  touch-action: manipulation;
  box-shadow: 0 2px 8px rgba(255, 59, 48, 0.3);
  transition: var(--transition, all 0.2s ease);
}

.apple-dialog-delete-btn:hover:not(:disabled) {
  background: #e02d24;
  transform: translateY(-1px);
}

.apple-dialog-delete-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.font-tabular {
  font-variant-numeric: tabular-nums;
}
</style>
