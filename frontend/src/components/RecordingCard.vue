<template>
  <div class="card recording-card relative group">
    <div class="card-header">
      <div class="status-badge" :class="statusClass">
        <Loader2 v-if="isProcessing" class="animate-spin" size="14" />
        <CheckCircle2 v-else-if="recording.status === 'summarized' || recording.status === 'completed'" size="14" />
        <AlertCircle v-else-if="recording.status === 'error'" size="14" />
        <Mic v-else size="14" />
        <span>{{ formattedStatus }}</span>
      </div>
      
      <div class="dropdown" v-if="!hideMenu">
        <button class="menu-btn" @click.stop="toggleMenu">
          <MoreVertical size="20" />
        </button>
        <div class="dropdown-menu" v-if="showMenu">
          <router-link :to="`/recordings/${recording.id}`" class="dropdown-item">View Details</router-link>
          <a :href="`/api/recordings/${recording.id}/download/transcript`" class="dropdown-item" v-if="isReady">Download Transcript</a>
          <a :href="`/api/recordings/${recording.id}/download/summary`" class="dropdown-item" v-if="isReady">Download Summary</a>
          <div class="dropdown-divider"></div>
          <button @click.stop="deleteRecording" class="dropdown-item delete-item">Delete</button>
        </div>
      </div>
    </div>
    
    <div class="card-body cursor-pointer" @click="goToDetails">
      <h3 class="recording-title">{{ recording.title }}</h3>
      
      <!-- Preview Snippet for summarized ones -->
      <div v-if="isReady && preview" class="preview-snippet">
        <div class="insight-label">KEY INSIGHT</div>
        <p class="insight-text">"{{ preview }}"</p>
      </div>
    </div>
    
    <div class="card-footer">
      <div class="meta-item">
        <Calendar size="14" />
        <span>{{ formattedDate }}</span>
      </div>
      <div class="meta-item">
        <Clock size="14" />
        <span>{{ formattedDuration }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Loader2, CheckCircle2, AlertCircle, Mic, MoreVertical, Calendar, Clock } from '@lucide/vue'
import { useToast } from 'primevue/usetoast'
import moment from 'moment'
import axios from 'axios'

const props = defineProps({
  recording: { type: Object, required: true },
  hideMenu: { type: Boolean, default: false }
})

const emit = defineEmits(['refresh'])
const router = useRouter()
const toast = useToast()
const showMenu = ref(false)

const toggleMenu = () => { showMenu.value = !showMenu.value }
const closeMenu = () => { showMenu.value = false }

const deleteRecording = async () => {
  if (confirm(`Are you sure you want to delete "${props.recording.title}"?`)) {
    try {
      await axios.delete(`/api/recordings/${props.recording.id}`)
      toast.add({ severity: 'success', summary: 'Deleted', detail: 'Recording deleted successfully.', life: 3000 })
      emit('refresh')
    } catch (e) {
      console.error("Failed to delete recording", e)
      toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to delete recording.', life: 5000 })
    }
  }
}

onMounted(() => { document.addEventListener('click', closeMenu) })
onUnmounted(() => { document.removeEventListener('click', closeMenu) })

const goToDetails = () => {
  router.push(`/recordings/${props.recording.id}`)
}

const isProcessing = computed(() => ['pending', 'transcribing', 'diarizing', 'summarizing'].includes(props.recording.status))
const isReady = computed(() => props.recording.status === 'summarized' || props.recording.status === 'completed')

const formattedStatus = computed(() => {
  return props.recording.status.charAt(0).toUpperCase() + props.recording.status.slice(1)
})

const statusClass = computed(() => {
  const status = props.recording.status
  if (['pending', 'transcribing', 'diarizing', 'summarizing'].includes(status)) return 'status-processing animate-pulse-slow'
  if (status === 'error') return 'status-error'
  return 'status-success'
})

const formattedDate = computed(() => {
  return moment(props.recording.created_at).format('MMM DD, YYYY')
})

const formattedDuration = computed(() => {
  if (!props.recording.duration) return '--:--'
  const mins = Math.floor(props.recording.duration / 60)
  const secs = Math.floor(props.recording.duration % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
})

const preview = computed(() => {
  // If there's a summary, we can extract the first sentence as a preview
  if (props.recording.summary_md) {
    const lines = props.recording.summary_md.split('\n')
    const firstTextLine = lines.find(l => l.trim().length > 10 && !l.startsWith('#'))
    return firstTextLine ? firstTextLine.substring(0, 80) + '...' : null
  }
  return null
})
</script>

<style scoped>
.recording-card {
  display: flex;
  flex-direction: column;
  min-height: 250px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.75rem;
  border-radius: var(--radius-full);
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.025em;
}

.status-processing {
  background-color: #E0E7FF;
  color: #4F46E5;
}

.status-success {
  background-color: #CCFBF1;
  color: #0D9488;
}

.status-error {
  background-color: #FEE2E2;
  color: #EF4444;
}

.menu-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 50%;
  transition: var(--transition);
}

.menu-btn:hover {
  background: var(--neutral-100);
  color: var(--text-primary);
}

.dropdown {
  position: relative;
}

.dropdown-menu {
  position: absolute;
  right: 0;
  top: 100%;
  background: var(--white);
  border: 1px solid var(--neutral-200);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-md);
  min-width: 160px;
  z-index: 10;
  overflow: hidden;
}

.dropdown-item {
  display: block;
  padding: 0.75rem 1rem;
  font-size: 0.875rem;
  color: var(--text-primary);
  cursor: pointer;
  transition: var(--transition);
}

.dropdown-item:hover {
  background: var(--neutral-50);
  color: var(--primary);
}

button.dropdown-item {
  width: 100%;
  text-align: left;
  border: none;
  background: none;
  font-family: inherit;
}

.dropdown-divider {
  height: 1px;
  background-color: var(--neutral-200);
  margin: 0.25rem 0;
}

.delete-item {
  color: #EF4444;
}

.delete-item:hover {
  background-color: #FEE2E2 !important;
  color: #EF4444 !important;
}

.card-body {
  flex: 1;
}

.recording-title {
  font-size: 1.125rem;
  margin-bottom: 0.5rem;
  line-height: 1.3;
}

.preview-snippet {
  margin-top: 1rem;
  background: var(--neutral-50);
  padding: 1rem;
  border-radius: var(--radius-md);
  position: relative;
}

.preview-snippet::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: var(--secondary);
  border-top-left-radius: var(--radius-md);
  border-bottom-left-radius: var(--radius-md);
}

.insight-label {
  font-size: 0.65rem;
  font-weight: 700;
  color: var(--secondary);
  letter-spacing: 0.05em;
  margin-bottom: 0.25rem;
}

.insight-text {
  font-size: 0.875rem;
  color: var(--text-secondary);
  font-style: italic;
  line-height: 1.4;
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--neutral-100);
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  color: var(--text-muted);
  font-size: 0.75rem;
  font-weight: 500;
}
</style>
