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
    
    <div v-if="isUploading" class="upload-progress">
      <Loader2 class="animate-spin" size="20" />
      <span>Uploading and starting transcription...</span>
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
const isRecording = ref(false)
const recordingTime = ref(0)
const isKeySet = ref(true)
let mediaRecorder = null
let audioChunks = []

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
  const formData = new FormData()
  formData.append('file', file)
  
  try {
    const res = await axios.post('/api/recordings', formData)
    emit('close')
    emit('refresh')
    router.push(`/recordings/${res.data.job_id}`)
  } catch (error) {
    toast.add({ severity: 'error', summary: 'Upload Failed', detail: 'Failed to upload file.', life: 5000 })
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

.upload-progress {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  margin-top: 2rem;
  color: var(--primary);
  font-weight: 500;
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
