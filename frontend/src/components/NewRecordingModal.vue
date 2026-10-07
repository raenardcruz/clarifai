<template>
  <Dialog 
    :visible="true" 
    modal 
    header="New Recording" 
    :style="{ width: '90vw', maxWidth: '720px' }" 
    @update:visible="$emit('close')"
  >
    <div class="apple-modal-body">
      <p class="modal-subtitle">
        Record audio live with your microphone or import an existing audio/video file.
      </p>

      <!-- Warning Banner if API key not set -->
      <div v-if="!isKeySet && isOnline" class="apple-warning-banner">
        <AlertTriangle size="18" class="warning-icon" />
        <div class="warning-text">
          <strong>Speechmatics API Key required:</strong>
          <span> Please set your API key in System Settings to enable AI transcription.</span>
        </div>
      </div>

      <!-- Segmented Mode Switcher -->
      <div class="mode-switcher-container">
        <div class="segmented-control">
          <button 
            type="button" 
            class="segmented-btn" 
            :class="{ active: activeMode === 'live' }"
            @click="activeMode = 'live'"
          >
            <Mic size="15" />
            <span>Live Microphone</span>
          </button>
          
          <button 
            type="button" 
            class="segmented-btn" 
            :class="{ active: activeMode === 'upload' }"
            @click="activeMode = 'upload'"
          >
            <UploadCloud size="15" />
            <span>Upload Media</span>
          </button>
        </div>
      </div>

      <!-- Medical Conversation Flag Toggle Option (set at beginning before recording or upload) -->
      <div 
        class="medical-option-card" 
        :class="{ 
          'is-medical-active': isMedical,
          'is-disabled': isRecording || isUploading 
        }" 
        @click="(!isRecording && !isUploading) && (isMedical = !isMedical)"
      >
        <div class="medical-option-left">
          <div class="medical-icon-badge">
            <Stethoscope size="18" />
          </div>
          <div class="medical-text-group">
            <div class="medical-title-row">
              <span class="medical-title">Medical / Clinical Conversation</span>
              <span class="medical-pill-tag">Clinical SOAP</span>
            </div>
            <p class="medical-desc">
              Optimizes speech recognition for medical vocabulary and formats notes into a structured clinical SOAP summary. Must be set prior to starting recording or upload.
            </p>
          </div>
        </div>
        <div 
          class="apple-toggle-switch" 
          :class="{ checked: isMedical, disabled: isRecording || isUploading }" 
          @click.stop="(!isRecording && !isUploading) && (isMedical = !isMedical)"
        >
          <span class="switch-handle"></span>
        </div>
      </div>

      <!-- Live Recording Mode -->
      <div v-if="activeMode === 'live'" class="live-recording-studio">
        <div class="studio-card" :class="{ 'recording-active': isRecording }">
          <!-- Background Audio pill indicator -->
          <div class="bg-audio-pill" v-if="isRecording">
            <span class="pulse-beacon"></span>
            <span>{{ isNative ? 'Background audio recording active • Continues when screen locks or app switches' : 'Offline recording in progress' }}</span>
          </div>

          <!-- Live Timer Display -->
          <div class="timer-display-group">
            <span class="recording-beacon" v-if="isRecording"></span>
            <span class="digital-timer font-tabular">{{ formattedTime }}</span>
          </div>

          <!-- Dynamic Audio Soundwave Animation (Driven by real-time audio meter) -->
          <div class="studio-waveform-visualizer" :class="{ 'is-dancing': isRecording }">
            <span 
              v-for="n in 28" 
              :key="n" 
              class="waveform-stick" 
              :style="{ 
                transform: isRecording ? `scaleY(${Math.max(0.2, (audioMeterLevel * 1.6) * (0.4 + (n % 6) * 0.14))})` : 'scaleY(0.2)',
                animationDelay: `${(n % 7) * 0.12}s` 
              }"
            ></span>
          </div>

          <!-- Apple Record Action Trigger -->
          <div class="record-action-center">
            <div 
              class="apple-record-outer-ring" 
              :class="{ 'recording-ring': isRecording }"
              @click="toggleLiveRecord"
            >
              <div class="apple-record-inner-circle" :class="{ 'recording-inner': isRecording }">
                <Square size="16" fill="white" v-if="isRecording" />
                <Mic size="24" color="white" v-else stroke-width="2.5" />
              </div>
            </div>

            <span class="record-state-hint">
              {{ isRecording ? 'Tap button to stop and save recording' : 'Tap to start recording' }}
            </span>
          </div>
        </div>
      </div>

      <!-- Upload Content Mode -->
      <div v-else class="file-upload-studio">
        <div 
          class="apple-dropzone" 
          :class="{ 'dropzone-active': isDragging, 'disabled': !isKeySet }"
          @dragover.prevent="isDragging = true"
          @dragleave.prevent="isDragging = false"
          @drop.prevent="onFileDrop"
          @click="isKeySet && triggerFileSelect()"
        >
          <input 
            type="file" 
            ref="fileInputRef" 
            class="hidden-file-input" 
            accept="audio/*,video/*" 
            @change="handleFileUpload" 
            :disabled="!isKeySet" 
          />

          <div class="dropzone-icon-circle">
            <UploadCloud size="32" class="text-blue" />
          </div>

          <h3 class="dropzone-title">Drag & drop your file here</h3>
          <p class="dropzone-subtitle">or click to browse from your Mac</p>

          <div class="supported-formats-pills">
            <span class="format-pill">MP3</span>
            <span class="format-pill">WAV</span>
            <span class="format-pill">M4A</span>
            <span class="format-pill">MP4</span>
            <span class="format-pill">WEBM</span>
          </div>
        </div>
      </div>
      
      <!-- Upload / Stitching Progress Modal Overlay -->
      <div v-if="isUploading" class="upload-progress-panel">
        <div class="progress-info-row">
          <div class="flex items-center gap-2">
            <Loader2 class="animate-spin text-blue" size="18" />
            <span class="progress-status-label">{{ uploadStatusText }}</span>
          </div>
          <span class="progress-percentage-label font-tabular">{{ uploadProgress }}%</span>
        </div>

        <div class="apple-progress-track">
          <div class="apple-progress-bar" :style="{ width: uploadProgress + '%' }"></div>
        </div>
      </div>
    </div>
  </Dialog>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import Dialog from 'primevue/dialog'
import { useToast } from 'primevue/usetoast'
import { 
  Mic, 
  UploadCloud, 
  Loader2, 
  AlertTriangle, 
  Square, 
  Stethoscope, 
  Wifi, 
  WifiOff, 
  CheckCircle2, 
  HardDrive 
} from '@lucide/vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { useIntervalFn } from '@vueuse/core'
import { audioRecorder } from '../services/audioRecorder'
import { useOfflineRecordingsStore } from '../stores/offlineRecordings'
import { storeToRefs } from 'pinia'

const emit = defineEmits(['close', 'refresh'])
const router = useRouter()
const toast = useToast()
const offlineStore = useOfflineRecordingsStore()
const { isOnline } = storeToRefs(offlineStore)

const isNative = audioRecorder.isNative
const activeMode = ref('live')
const fileInputRef = ref(null)
const isMedical = ref(false)
const isDragging = ref(false)

const isUploading = ref(false)
const uploadProgress = ref(0)
const uploadStatusText = ref('')
const isRecording = ref(false)
const recordingTime = ref(0)
const audioMeterLevel = ref(0.2)
const isKeySet = ref(true)

const CHUNK_SIZE = 5 * 1024 * 1024 // 5 MB chunks
const MAX_CONCURRENT_UPLOADS = 3 // Concurrency limit

const checkApiKey = async () => {
  if (!isOnline.value) {
    // Offline mode: proceed without blocking
    isKeySet.value = true
    return
  }
  try {
    const res = await axios.get('/api/settings')
    isKeySet.value = !!res.data.speechmatics_api_key
  } catch (e) {
    console.error("Failed to fetch settings keys", e)
    isKeySet.value = true // Don't block offline recording
  }
}

onMounted(() => {
  checkApiKey()
})

let meterInterval = null

const { pause, resume } = useIntervalFn(() => {
  recordingTime.value++
}, 1000, { immediate: false })

const startMetering = () => {
  meterInterval = setInterval(async () => {
    if (!isRecording.value) return
    const status = await audioRecorder.getStatus()
    audioMeterLevel.value = status.meterLevel || 0.2
    if (status.duration > 0 && Math.abs(status.duration - recordingTime.value) > 2) {
      recordingTime.value = Math.floor(status.duration)
    }
  }, 120)
}

const stopMetering = () => {
  if (meterInterval) {
    clearInterval(meterInterval)
    meterInterval = null
  }
  audioMeterLevel.value = 0.2
}

onUnmounted(() => {
  stopMetering()
})

const formattedTime = computed(() => {
  const hrs = Math.floor(recordingTime.value / 3600)
  const mins = Math.floor((recordingTime.value % 3600) / 60)
  const secs = recordingTime.value % 60
  
  if (hrs > 0) {
    return `${hrs.toString().padStart(2, '0')}:${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
})

const toggleLiveRecord = () => {
  if (isRecording.value) {
    stopRecording()
  } else {
    startRecording()
  }
}

const startRecording = async () => {
  try {
    const hasPermission = await audioRecorder.requestPermission()
    if (!hasPermission) {
      toast.add({
        severity: 'error',
        summary: 'Microphone Permission',
        detail: 'Please enable microphone access in your settings to record audio.',
        life: 5000
      })
      return
    }

    await audioRecorder.start()
    isRecording.value = true
    recordingTime.value = 0
    resume()
    startMetering()

    toast.add({
      severity: 'info',
      summary: isNative ? 'iOS Recording Active' : 'Recording Started',
      detail: isNative 
        ? 'Background audio is active. Recording will continue even when phone sleeps or changes apps.'
        : 'Audio is recorded locally on your device.',
      life: 4000
    })
  } catch (err) {
    console.error('Record error:', err)
    toast.add({
      severity: 'error',
      summary: 'Recording Error',
      detail: err.message || 'Could not start audio recorder.',
      life: 5000
    })
  }
}

const stopRecording = async () => {
  if (!isRecording.value) return
  isRecording.value = false
  pause()
  stopMetering()

  try {
    isUploading.value = true
    uploadProgress.value = 0
    uploadStatusText.value = 'Saving audio locally to device...'

    const result = await audioRecorder.stop()

    // 1. Save immediately to persistent offline storage
    const titlePrefix = isMedical.value ? 'Clinical Consultation' : 'Live Note'
    const savedItem = await offlineStore.addRecording({
      ...result,
      title: `${titlePrefix} ${new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`,
      isMedical: isMedical.value
    }, result.blob)

    // 2. If offline, alert user and finish
    if (!isOnline.value) {
      toast.add({
        severity: 'success',
        summary: 'Saved Offline',
        detail: 'Audio saved to phone storage. It will automatically upload and transcribe once internet returns.',
        life: 6000
      })
      emit('refresh')
      emit('close')
      return
    }

    // 3. If online, stream chunks to backend
    uploadStatusText.value = 'Uploading audio to server...'
    uploadProgress.value = 20

    try {
      await offlineStore.syncRecording(savedItem.id)
      uploadProgress.value = 100
      uploadStatusText.value = 'Processing complete! Redirecting...'

      toast.add({
        severity: 'success',
        summary: 'Upload Complete',
        detail: 'Transcription and AI diarization started.',
        life: 4000
      })

      emit('refresh')
      emit('close')

      // Navigate to the newly created recording
      const updated = offlineStore.offlineList.find(r => r.id === savedItem.id)
      if (updated && updated.jobId) {
        router.push(`/recordings/${updated.jobId}`)
      }
    } catch (syncErr) {
      toast.add({
        severity: 'warn',
        summary: 'Saved to Device',
        detail: 'Upload interrupted. Recording is safely preserved on your phone and will retry automatically.',
        life: 6000
      })
      emit('refresh')
      emit('close')
    }
  } catch (err) {
    console.error('Stop recording error:', err)
    toast.add({
      severity: 'error',
      summary: 'Save Failed',
      detail: err.message || 'Failed to save recording.',
      life: 5000
    })
  } finally {
    isUploading.value = false
  }
}

const triggerFileSelect = () => {
  if (fileInputRef.value) {
    fileInputRef.value.click()
  }
}

const handleFileUpload = (event) => {
  const file = event.target.files[0]
  if (file) {
    uploadFile(file)
  }
}

const onFileDrop = (event) => {
  isDragging.value = false
  const file = event.dataTransfer.files[0]
  if (file) {
    uploadFile(file)
  }
}

const uploadFile = async (file) => {
  isUploading.value = true
  uploadProgress.value = 0
  uploadStatusText.value = 'Preparing audio stream...'

  try {
    const totalChunks = Math.ceil(file.size / CHUNK_SIZE)

    // 1. Initialize upload session on backend
    const initRes = await axios.post('/api/recordings/upload/init', {
      filename: file.name,
      total_chunks: totalChunks,
      file_size: file.size,
    })

    const uploadId = initRes.data.upload_id

    // 2. Prepare chunk upload slice list
    let completedChunks = 0
    uploadStatusText.value = `Uploading (0/${totalChunks} chunks)...`

    const chunks = []
    for (let i = 0; i < totalChunks; i++) {
      const start = i * CHUNK_SIZE
      const end = Math.min(file.size, start + CHUNK_SIZE)
      const chunkBlob = file.slice(start, end)
      chunks.push({ index: i, blob: chunkBlob })
    }

    // Function to upload a single chunk with retries
    const uploadSingleChunk = async (chunk) => {
      let retries = 3
      while (retries > 0) {
        try {
          const formData = new FormData()
          formData.append('upload_id', uploadId)
          formData.append('chunk_index', chunk.index.toString())
          formData.append('chunk', chunk.blob, `chunk_${chunk.index}`)

          await axios.post('/api/recordings/upload/chunk', formData)
          completedChunks++
          uploadProgress.value = Math.round((completedChunks / totalChunks) * 100)
          uploadStatusText.value = `Uploading (${completedChunks}/${totalChunks} chunks)...`
          return
        } catch (err) {
          retries--
          if (retries === 0) throw err
          await new Promise((resolve) => setTimeout(resolve, 1000))
        }
      }
    }

    // Concurrent worker pool
    const queue = [...chunks]
    const workerCount = Math.min(MAX_CONCURRENT_UPLOADS, queue.length)
    const workers = Array(workerCount)
      .fill(null)
      .map(async () => {
        while (queue.length > 0) {
          const chunk = queue.shift()
          if (chunk) {
            await uploadSingleChunk(chunk)
          }
        }
      })

    await Promise.all(workers)

    // 3. Complete chunked upload & trigger stitching/transcription
    uploadStatusText.value = 'Stitching file & starting transcription...'
    const completeRes = await axios.post('/api/recordings/upload/complete', {
      upload_id: uploadId,
      filename: file.name,
      total_chunks: totalChunks,
      is_medical: isMedical.value,
    })

    emit('close')
    emit('refresh')
    router.push(`/recordings/${completeRes.data.job_id}`)
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: 'Upload Failed',
      detail: error.response?.data?.detail || 'Failed to upload file.',
      life: 5000,
    })
    console.error(error)
  } finally {
    isUploading.value = false
  }
}
</script>

<style scoped>
.apple-modal-body {
  padding-top: 0.5rem;
}

/* Medical Option Toggle Card */
.medical-option-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.85rem 1.15rem;
  background: var(--bg-surface-secondary, rgba(0, 0, 0, 0.02));
  border: 1.5px solid var(--border-color, rgba(0, 0, 0, 0.08));
  border-radius: var(--radius-lg, 12px);
  margin-bottom: 1.5rem;
  cursor: pointer;
  transition: all 0.2s ease;
  user-select: none;
}

.medical-option-card:hover {
  background: rgba(16, 185, 129, 0.04);
  border-color: rgba(16, 185, 129, 0.35);
}

.medical-option-card.is-medical-active {
  background: rgba(16, 185, 129, 0.07);
  border-color: rgba(16, 185, 129, 0.5);
}

.medical-option-card.is-disabled {
  opacity: 0.5;
  cursor: not-allowed;
  pointer-events: none;
}

.medical-option-left {
  display: flex;
  align-items: center;
  gap: 0.85rem;
}

.medical-icon-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  border-radius: 10px;
  background: rgba(16, 185, 129, 0.12);
  color: #059669;
  flex-shrink: 0;
  transition: all 0.2s ease;
}

.medical-option-card.is-medical-active .medical-icon-badge {
  background: #059669;
  color: #ffffff;
}

.medical-text-group {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.medical-title-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.medical-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-primary);
}

.medical-pill-tag {
  font-size: 0.68rem;
  font-weight: 600;
  padding: 0.15rem 0.45rem;
  border-radius: 9999px;
  background: rgba(16, 185, 129, 0.15);
  color: #047857;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.medical-desc {
  font-size: 0.78rem;
  color: var(--text-secondary);
  line-height: 1.35;
  margin: 0;
}

/* Apple Toggle Switch */
.apple-toggle-switch {
  width: 44px;
  height: 26px;
  background: #E5E7EB;
  border-radius: 13px;
  padding: 2px;
  cursor: pointer;
  transition: background-color 0.25s ease;
  position: relative;
  flex-shrink: 0;
}

.apple-toggle-switch.checked {
  background: #10B981;
}

.switch-handle {
  display: block;
  width: 22px;
  height: 22px;
  background: #FFFFFF;
  border-radius: 50%;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  transition: transform 0.25s ease;
}

.apple-toggle-switch.checked .switch-handle {
  transform: translateX(18px);
}


.modal-subtitle {
  color: var(--text-secondary);
  font-size: 0.9rem;
  margin-bottom: 1.5rem;
  line-height: 1.45;
}

/* Apple Amber Warning Banner */
.apple-warning-banner {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background-color: var(--apple-orange-light);
  border: 1px solid rgba(255, 149, 0, 0.25);
  color: #92400E;
  padding: 0.85rem 1rem;
  border-radius: var(--radius-md);
  margin-bottom: 1.5rem;
  font-size: 0.84rem;
}

.warning-icon {
  color: var(--apple-orange);
  flex-shrink: 0;
}

.warning-text strong {
  color: #78350F;
}

/* Mode Switcher */
.mode-switcher-container {
  display: flex;
  justify-content: center;
  margin-bottom: 1.75rem;
}

.segmented-control {
  display: inline-flex;
  background: rgba(118, 118, 128, 0.12);
  padding: 3px;
  border-radius: var(--radius-full);
  gap: 3px;
}

.segmented-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.5rem 1.25rem;
  border-radius: var(--radius-full);
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text-secondary);
  border: none;
  background: transparent;
  cursor: pointer;
  transition: var(--transition-fast);
}

.segmented-btn.active {
  background: var(--white);
  color: var(--text-primary);
  font-weight: 600;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
}

/* Live Recording Studio */
.live-recording-studio {
  display: flex;
  flex-direction: column;
}

.studio-card {
  background: var(--bg-primary);
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: var(--radius-xl);
  padding: 2.5rem 2rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  transition: all 0.3s var(--apple-ease);
}

.studio-card.recording-active {
  background: rgba(255, 59, 48, 0.03);
  border-color: rgba(255, 59, 48, 0.2);
  box-shadow: 0 0 30px rgba(255, 59, 48, 0.08);
}

.timer-display-group {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  margin-bottom: 1.5rem;
}

.recording-beacon {
  width: 10px;
  height: 10px;
  background-color: var(--apple-red);
  border-radius: var(--radius-full);
  animation: applePulseBeacon 1.5s infinite;
}

.digital-timer {
  font-size: 2.5rem;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

/* Waveform Visualizer */
.studio-waveform-visualizer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 3.5px;
  height: 48px;
  margin-bottom: 2rem;
  width: 100%;
  max-width: 380px;
}

.waveform-stick {
  width: 3.5px;
  height: 6px;
  background: rgba(0, 0, 0, 0.15);
  border-radius: var(--radius-full);
  transition: height 0.2s ease;
}

.studio-waveform-visualizer.is-dancing .waveform-stick {
  background: var(--apple-red);
  animation: waveformDance 0.8s ease-in-out infinite alternate;
}

@keyframes waveformDance {
  0% { height: 6px; }
  100% { height: 38px; }
}

/* Apple Record Button */
.record-action-center {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.apple-record-outer-ring {
  width: 72px;
  height: 72px;
  border-radius: var(--radius-full);
  border: 3px solid var(--apple-blue);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s var(--apple-ease);
  box-shadow: 0 4px 16px rgba(0, 113, 227, 0.2);
}

.apple-record-outer-ring:hover {
  transform: scale(1.06);
}

.apple-record-outer-ring:active {
  transform: scale(0.96);
}

.apple-record-inner-circle {
  width: 54px;
  height: 54px;
  border-radius: var(--radius-full);
  background: var(--apple-blue);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.25s var(--apple-ease);
}

.apple-record-outer-ring.recording-ring {
  border-color: var(--apple-red);
  box-shadow: 0 0 24px rgba(255, 59, 48, 0.35);
  animation: applePulseBeacon 2s infinite;
}

.apple-record-inner-circle.recording-inner {
  background: var(--apple-red);
  border-radius: 12px;
  width: 44px;
  height: 44px;
}

.apple-record-outer-ring.disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.record-state-hint {
  font-size: 0.825rem;
  color: var(--text-muted);
  font-weight: 500;
}

/* Apple AirDrop Style Dropzone */
.file-upload-studio {
  display: flex;
  flex-direction: column;
}

.apple-dropzone {
  border: 2px dashed rgba(0, 0, 0, 0.12);
  border-radius: var(--radius-xl);
  padding: 3.5rem 2rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: var(--bg-primary);
  cursor: pointer;
  transition: all 0.25s var(--apple-ease);
}

.apple-dropzone:hover, .apple-dropzone.dropzone-active {
  border-color: var(--apple-blue);
  background: rgba(0, 113, 227, 0.04);
  transform: translateY(-2px);
}

.apple-dropzone.disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.dropzone-icon-circle {
  width: 68px;
  height: 68px;
  border-radius: var(--radius-full);
  background: var(--white);
  border: 1px solid rgba(0, 0, 0, 0.06);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.25rem;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.04);
  transition: var(--transition);
}

.apple-dropzone:hover .dropzone-icon-circle {
  transform: scale(1.08);
  box-shadow: 0 6px 20px rgba(0, 113, 227, 0.2);
}

.dropzone-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 0.35rem;
}

.dropzone-subtitle {
  font-size: 0.875rem;
  color: var(--text-muted);
  margin-bottom: 1.5rem;
}

.supported-formats-pills {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-wrap: wrap;
  justify-content: center;
}

.format-pill {
  font-size: 0.7rem;
  font-weight: 700;
  background: rgba(0, 0, 0, 0.05);
  color: var(--text-secondary);
  padding: 0.2rem 0.55rem;
  border-radius: var(--radius-full);
  letter-spacing: 0.04em;
}

.hidden-file-input {
  display: none;
}

/* Upload Progress Overlay */
.upload-progress-panel {
  margin-top: 1.5rem;
  background: var(--bg-primary);
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: var(--radius-lg);
  padding: 1rem 1.25rem;
}

.progress-info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.65rem;
}

.progress-status-label {
  font-size: 0.84rem;
  font-weight: 600;
  color: var(--text-primary);
}

.progress-percentage-label {
  font-size: 0.825rem;
  font-weight: 700;
  color: var(--apple-blue);
}

.apple-progress-track {
  width: 100%;
  height: 6px;
  background: rgba(0, 0, 0, 0.08);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.apple-progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #0071E3, #30B0C7);
  border-radius: var(--radius-full);
  transition: width 0.3s ease;
}

.font-tabular {
  font-variant-numeric: tabular-nums;
}

.text-blue {
  color: var(--apple-blue);
}

.text-emerald {
  color: #10b981;
}

.text-amber {
  color: #f59e0b;
}

/* Offline-First Banner */
.apple-offline-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-radius: var(--radius-md, 10px);
  background: rgba(16, 185, 129, 0.08);
  border: 1px solid rgba(16, 185, 129, 0.25);
  margin-bottom: 1rem;
  gap: 0.75rem;
}

.apple-offline-banner.offline-mode {
  background: rgba(245, 158, 11, 0.09);
  border-color: rgba(245, 158, 11, 0.3);
}

.offline-banner-left {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-size: 0.825rem;
  color: var(--text-primary);
}

.offline-banner-text strong {
  font-weight: 600;
}

.offline-banner-pill {
  font-size: 0.72rem;
  font-weight: 600;
  padding: 0.2rem 0.55rem;
  border-radius: var(--radius-full, 9999px);
  background: rgba(0, 0, 0, 0.06);
  color: var(--text-secondary);
  white-space: nowrap;
}

.offline-banner-pill.native-pill {
  background: rgba(0, 113, 227, 0.12);
  color: var(--apple-blue, #0071E3);
}

/* Background Audio Pill */
.bg-audio-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.35rem 0.85rem;
  border-radius: var(--radius-full, 9999px);
  background: rgba(16, 185, 129, 0.12);
  border: 1px solid rgba(16, 185, 129, 0.25);
  font-size: 0.76rem;
  font-weight: 600;
  color: #059669;
  margin-bottom: 0.5rem;
}

.pulse-beacon {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #10b981;
  animation: pulseBeacon 1.5s infinite;
}

@keyframes pulseBeacon {
  0% { transform: scale(0.95); opacity: 0.7; }
  50% { transform: scale(1.3); opacity: 1; }
  100% { transform: scale(0.95); opacity: 0.7; }
}
</style>
