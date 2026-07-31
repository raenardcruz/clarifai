<template>
  <Dialog :visible="true" modal header="Create New Recording" :style="{ width: '90vw', maxWidth: '800px' }" @update:visible="$emit('close')">
    <p class="subtitle">Start capturing your thoughts or upload an existing file.</p>

    <div v-if="!isKeySet" class="warning-banner">
      <AlertTriangle size="20" class="text-warning-icon" />
      <span>Speechmatics API key is not configured. Ask an administrator to set the API Key in System Settings.</span>
    </div>

    <div class="upload-options">
      <div class="option-card live-record" :class="{ 'is-recording': isRecording }">
        <div class="icon-wrapper">
          <Mic size="24" :color="isRecording ? '#EF4444' : 'var(--primary)'" />
        </div>
        <h3>Record Live</h3>
        <p class="card-desc">Real-time transcription</p>
        
        <div class="record-controls">
          <Button v-if="!isRecording" label="Start Recording" class="w-full start-record" severity="primary" @click="startRecording" :disabled="!isKeySet" />
          <Button v-else label="Stop" icon="pi pi-stop" class="w-full stop-record" severity="danger" @click="stopRecording" />
          
          <p v-if="isRecording" class="recording-time animate-pulse">Recording... {{ formattedTime }}</p>
        </div>
      </div>

      <div class="option-card upload-file">
        <div class="icon-wrapper">
          <UploadCloud size="24" color="var(--primary)" />
        </div>
        <h3>Upload Content</h3>
        <p class="card-desc">MP3, WAV, or MP4 files</p>
        <input type="file" ref="fileInput" class="hidden" accept="audio/*,video/*" @change="handleFileUpload" :disabled="!isKeySet" />
        <Button label="Browse Files" icon="pi pi-file" class="w-full" severity="secondary" @click="isKeySet && $refs.fileInput.click()" :disabled="!isKeySet" />
      </div>
    </div>
    
    <div v-if="isUploading" class="upload-progress-container">
      <div class="upload-progress-header">
        <Loader2 class="animate-spin" size="20" />
        <span>{{ uploadStatusText }}</span>
      </div>
      <div class="progress-bar-bg">
        <div class="progress-bar-fill" :style="{ width: uploadProgress + '%' }"></div>
      </div>
      <span class="progress-percentage">{{ uploadProgress }}%</span>
    </div>
  </Dialog>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import { useToast } from 'primevue/usetoast'
import { Mic, UploadCloud, Loader2, AlertTriangle } from '@lucide/vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { useIntervalFn } from '@vueuse/core'

const emit = defineEmits(['close', 'refresh'])
const router = useRouter()
const toast = useToast()

const fileInput = ref(null)
const isUploading = ref(false)
const uploadProgress = ref(0)
const uploadStatusText = ref('')
const isRecording = ref(false)
const recordingTime = ref(0)
const isKeySet = ref(true)
let mediaRecorder = null
let audioChunks = []

const CHUNK_SIZE = 5 * 1024 * 1024 // 5 MB chunks
const MAX_CONCURRENT_UPLOADS = 3 // Parallel chunk upload concurrency limit

const checkApiKey = async () => {
  try {
    const res = await axios.get('/api/settings')
    isKeySet.value = !!res.data.speechmatics_api_key
  } catch (e) {
    console.error("Failed to fetch settings keys", e)
    isKeySet.value = false
  }
}

onMounted(() => {
  checkApiKey()
})

const { pause, resume } = useIntervalFn(() => {
  recordingTime.value++
}, 1000, { immediate: false })

const formattedTime = computed(() => {
  const mins = Math.floor(recordingTime.value / 60)
  const secs = recordingTime.value % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
})

const startRecording = async () => {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    mediaRecorder = new MediaRecorder(stream)
    
    mediaRecorder.ondataavailable = (event) => {
      audioChunks.push(event.data)
    }
    
    mediaRecorder.onstop = async () => {
      const audioBlob = new Blob(audioChunks, { type: 'audio/webm' })
      const file = new File([audioBlob], `Live_Recording_${new Date().getTime()}.webm`, { type: 'audio/webm' })
      await uploadFile(file)
      audioChunks = []
    }
    
    mediaRecorder.start()
    isRecording.value = true
    recordingTime.value = 0
    resume()
  } catch (err) {
    toast.add({ severity: 'error', summary: 'Microphone Error', detail: 'Microphone access denied or not available.', life: 5000 })
  }
}

const stopRecording = () => {
  if (mediaRecorder && isRecording.value) {
    mediaRecorder.stop()
    isRecording.value = false
    pause()
    mediaRecorder.stream.getTracks().forEach(t => t.stop())
  }
}

const handleFileUpload = (event) => {
  const file = event.target.files[0]
  if (file) {
    uploadFile(file)
  }
}

const uploadFile = async (file) => {
  isUploading.value = true
  uploadProgress.value = 0
  uploadStatusText.value = 'Initializing upload...'

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
.subtitle {
  color: var(--text-secondary);
  margin-bottom: 2rem;
}

.warning-banner {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background-color: #FFFBEB;
  border: 1px solid #FDE68A;
  color: #92400E;
  padding: 1rem;
  border-radius: var(--radius-md);
  margin-bottom: 1.5rem;
  font-size: 0.875rem;
  font-weight: 500;
}

.text-warning-icon {
  color: #D97706;
}

.upload-options {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
}

.option-card {
  border: 2px dashed var(--neutral-200);
  border-radius: var(--radius-lg);
  padding: 2rem;
  text-align: center;
  transition: var(--transition);
  background: var(--neutral-50);
  display: flex;
  flex-direction: column;
  align-items: center;
}

.option-card:hover {
  border-color: var(--primary);
  background: var(--white);
}

.icon-wrapper {
  width: 64px;
  height: 64px;
  background: var(--white);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1.5rem auto;
  box-shadow: var(--shadow-sm);
}

.option-card.is-recording {
  border-color: #EF4444;
  background: rgba(239, 68, 68, 0.05);
}

.record-controls {
  margin-top: 1.5rem;
  width: 100%;
}

.card-desc {
  margin-bottom: 1.5rem;
  color: var(--text-muted);
}

.recording-time {
  margin-top: 1rem;
  color: #EF4444;
  font-weight: 600;
}

.hidden { display: none; }

.upload-progress-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-top: 1.5rem;
  background: var(--neutral-50);
  padding: 1rem 1.25rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--neutral-200);
}

.upload-progress-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: var(--primary);
  font-weight: 500;
  font-size: 0.9rem;
}

.progress-bar-bg {
  width: 100%;
  height: 8px;
  background-color: var(--neutral-200);
  border-radius: 4px;
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  background-color: var(--primary);
  transition: width 0.3s ease;
}

.progress-percentage {
  font-size: 0.8rem;
  color: var(--text-muted);
  align-self: flex-end;
}

@media (max-width: 640px) {
  .subtitle {
    margin-bottom: 1.25rem;
    font-size: 0.875rem;
  }
  
  .upload-options {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
  
  .option-card {
    padding: 1.25rem;
  }
  
  .icon-wrapper {
    width: 48px;
    height: 48px;
    margin-bottom: 0.75rem;
  }
  
  .card-desc {
    margin-bottom: 1rem;
  }
  
  .warning-banner {
    padding: 0.75rem;
    font-size: 0.8rem;
  }
}
</style>
