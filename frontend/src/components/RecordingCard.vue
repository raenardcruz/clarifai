<template>
  <div class="card apple-recording-card group" @click="goToDetails">
    <!-- Header with Status Badge and macOS Menu -->
    <div class="card-header" @click.stop>
      <div class="card-header-badges">
        <div class="status-pill" :class="statusClass">
          <span class="status-dot-beacon" :class="beaconClass"></span>
          <Loader2 v-if="isProcessing" class="animate-spin" size="13" />
          <CheckCircle2 v-else-if="isReady" size="13" />
          <AlertCircle v-else-if="recording.status === 'error'" size="13" />
          <Mic v-else size="13" />
          <span class="status-text">{{ formattedStatus }}</span>
        </div>
        
        <div v-if="recording.is_medical" class="medical-badge" title="Medical Conversation">
          <Stethoscope size="11" />
          <span>Medical</span>
        </div>
      </div>
      
      <div class="dropdown-container" v-if="!hideMenu">
        <button class="menu-pill-btn" @click.stop="toggleMenu" aria-label="More options">
          <MoreHorizontal size="18" />
        </button>
        <div class="apple-dropdown-menu" v-if="showMenu" @click.stop>
          <router-link :to="`/recordings/${recording.id}`" class="apple-dropdown-item">
            <Eye size="15" />
            <span>View Details</span>
          </router-link>
          <a :href="`/api/recordings/${recording.id}/download/transcript`" class="apple-dropdown-item" v-if="isReady">
            <FileText size="15" />
            <span>Download Transcript</span>
          </a>
          <a :href="`/api/recordings/${recording.id}/download/summary`" class="apple-dropdown-item" v-if="isReady">
            <Sparkles size="15" />
            <span>Download Summary</span>
          </a>
          <div class="apple-dropdown-divider"></div>
          <button @click.stop="deleteRecording" class="apple-dropdown-item delete-item">
            <Trash2 size="15" />
            <span>Delete Recording</span>
          </button>
        </div>
      </div>
    </div>

    
    <!-- Body: Title & Soundwave & Preview -->
    <div class="card-body">
      <h3 class="recording-title" :title="recording.title">{{ recording.title }}</h3>
      
      <!-- Mini Apple Soundwave Graphic Teaser -->
      <div class="soundwave-container" :class="{ 'is-active': isProcessing }">
        <span class="bar bar-1"></span>
        <span class="bar bar-2"></span>
        <span class="bar bar-3"></span>
        <span class="bar bar-4"></span>
        <span class="bar bar-5"></span>
        <span class="bar bar-6"></span>
        <span class="bar bar-7"></span>
        <span class="bar bar-8"></span>
        <span class="bar bar-9"></span>
        <span class="bar bar-10"></span>
        <span class="bar bar-11"></span>
        <span class="bar bar-12"></span>
        <span class="bar bar-13"></span>
        <span class="bar bar-14"></span>
        <span class="bar bar-15"></span>
        <span class="bar bar-16"></span>
      </div>

      <!-- Preview Snippet for summarized ones -->
      <div v-if="isReady && preview" class="insight-callout">
        <div class="insight-label-row">
          <Sparkles size="12" class="sparkle-icon" />
          <span class="insight-label">KEY TAKEAWAY</span>
        </div>
        <p class="insight-text">"{{ preview }}"</p>
      </div>

      <!-- Processing State Teaser -->
      <div v-else-if="isProcessing" class="processing-callout">
        <Loader2 class="animate-spin text-blue" size="14" />
        <span>ClarifAi is synthesizing transcript and notes...</span>
      </div>
    </div>
    
    <!-- Footer with Metadata and subtle view arrow -->
    <div class="card-footer">
      <div class="meta-group">
        <div class="meta-item">
          <Calendar size="13" stroke-width="2" />
          <span>{{ formattedDate }}</span>
        </div>
        <div class="meta-separator">•</div>
        <div class="meta-item">
          <Clock size="13" stroke-width="2" />
          <span class="duration-num">{{ formattedDuration }}</span>
        </div>
      </div>

      <div class="view-arrow-circle">
        <ChevronRight size="15" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { 
  Loader2, CheckCircle2, AlertCircle, Mic, MoreHorizontal, 
  Calendar, Clock, Eye, FileText, Sparkles, Trash2, ChevronRight, Stethoscope 
} from '@lucide/vue'
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
      toast.add({ severity: 'success', summary: 'Deleted', detail: 'Recording removed successfully.', life: 3000 })
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
  if (isProcessing.value) return 'status-processing'
  if (status === 'error') return 'status-error'
  return 'status-ready'
})

const beaconClass = computed(() => {
  if (isProcessing.value) return 'beacon-blue animate-pulse'
  if (props.recording.status === 'error') return 'beacon-red'
  return 'beacon-green'
})

const formattedDate = computed(() => {
  return moment(props.recording.created_at).format('MMM D, YYYY')
})

const formattedDuration = computed(() => {
  if (!props.recording.duration) return '00:00'
  const mins = Math.floor(props.recording.duration / 60)
  const secs = Math.floor(props.recording.duration % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
})

const preview = computed(() => {
  if (props.recording.summary_md) {
    const lines = props.recording.summary_md.split('\n')
    const firstTextLine = lines.find(l => l.trim().length > 10 && !l.startsWith('#'))
    return firstTextLine ? firstTextLine.substring(0, 95) + '...' : null
  }
  return null
})
</script>

<style scoped>
.apple-recording-card {
  display: flex;
  flex-direction: column;
  background: var(--bg-secondary);
  border-radius: var(--radius-lg);
  padding: 1.4rem;
  border: 1px solid rgba(0, 0, 0, 0.06);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.03), 0 1px 2px rgba(0, 0, 0, 0.02);
  transition: all 0.28s var(--apple-ease);
  cursor: pointer;
  position: relative;
  overflow: visible;
}

.apple-recording-card:hover {
  box-shadow: 0 12px 30px -4px rgba(0, 0, 0, 0.08), 0 4px 10px rgba(0, 0, 0, 0.03);
  border-color: rgba(0, 113, 227, 0.25);
  transform: translateY(-2px);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.85rem;
}

.card-header-badges {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.medical-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.28rem;
  padding: 0.2rem 0.55rem;
  border-radius: var(--radius-full);
  font-size: 0.7rem;
  font-weight: 600;
  background-color: rgba(16, 185, 129, 0.12);
  color: #047857;
  border: 1px solid rgba(16, 185, 129, 0.28);
  letter-spacing: -0.01em;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.25rem 0.65rem;
  border-radius: var(--radius-full);
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: -0.01em;
}


.status-dot-beacon {
  width: 6px;
  height: 6px;
  border-radius: var(--radius-full);
}

.beacon-blue { background-color: var(--apple-blue); }
.beacon-green { background-color: var(--apple-green); }
.beacon-red { background-color: var(--apple-red); }

.status-processing {
  background-color: var(--apple-blue-light);
  color: var(--apple-blue);
}

.status-ready {
  background-color: var(--apple-green-light);
  color: #1E8E3E;
}

.status-error {
  background-color: var(--apple-red-light);
  color: var(--apple-red);
}

.dropdown-container {
  position: relative;
}

.menu-pill-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: var(--transition-fast);
}

.menu-pill-btn:hover {
  background: rgba(0, 0, 0, 0.06);
  color: var(--text-primary);
}

/* Apple Dropdown Menu */
.apple-dropdown-menu {
  position: absolute;
  right: 0;
  top: 100%;
  margin-top: 4px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: var(--radius-md);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.12), 0 1px 3px rgba(0, 0, 0, 0.06);
  min-width: 190px;
  z-index: 50;
  padding: 0.35rem;
}

.apple-dropdown-item {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.55rem 0.75rem;
  font-size: 0.84rem;
  font-weight: 500;
  color: var(--text-primary);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: var(--transition-fast);
  text-decoration: none;
  border: none;
  background: none;
  width: 100%;
  text-align: left;
}

.apple-dropdown-item:hover {
  background: rgba(0, 113, 227, 0.08);
  color: var(--apple-blue);
}

.apple-dropdown-divider {
  height: 1px;
  background-color: rgba(0, 0, 0, 0.06);
  margin: 0.3rem 0;
}

.delete-item {
  color: var(--apple-red);
}

.delete-item:hover {
  background: rgba(255, 59, 48, 0.08) !important;
  color: var(--apple-red) !important;
}

.card-body {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.recording-title {
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 0.65rem;
  line-height: 1.35;
  letter-spacing: -0.02em;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  transition: var(--transition);
}

.apple-recording-card:hover .recording-title {
  color: var(--apple-blue);
}

/* Soundwave Graphic */
.soundwave-container {
  display: flex;
  align-items: center;
  gap: 2.5px;
  height: 20px;
  margin-bottom: 0.85rem;
  opacity: 0.45;
}

.apple-recording-card:hover .soundwave-container {
  opacity: 0.8;
}

.soundwave-container .bar {
  width: 2.5px;
  background-color: var(--apple-blue);
  border-radius: var(--radius-full);
  display: inline-block;
}

.bar-1 { height: 4px; }
.bar-2 { height: 10px; }
.bar-3 { height: 16px; }
.bar-4 { height: 8px; }
.bar-5 { height: 14px; }
.bar-6 { height: 18px; }
.bar-7 { height: 11px; }
.bar-8 { height: 7px; }
.bar-9 { height: 15px; }
.bar-10 { height: 12px; }
.bar-11 { height: 9px; }
.bar-12 { height: 17px; }
.bar-13 { height: 14px; }
.bar-14 { height: 8px; }
.bar-15 { height: 11px; }
.bar-16 { height: 5px; }

.soundwave-container.is-active .bar {
  animation: soundwaveBounce 1.2s ease-in-out infinite alternate;
}

@keyframes soundwaveBounce {
  0% { transform: scaleY(0.4); }
  100% { transform: scaleY(1.3); }
}

/* Insight Callout */
.insight-callout {
  background: rgba(0, 113, 227, 0.04);
  border: 1px solid rgba(0, 113, 227, 0.09);
  padding: 0.75rem 0.85rem;
  border-radius: var(--radius-md);
  margin-top: auto;
  position: relative;
  border-left: 3px solid var(--apple-blue);
}

.insight-label-row {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  margin-bottom: 0.25rem;
}

.sparkle-icon {
  color: var(--apple-blue);
}

.insight-label {
  font-size: 0.65rem;
  font-weight: 700;
  color: var(--apple-blue);
  letter-spacing: 0.05em;
}

.insight-text {
  font-size: 0.8rem;
  color: var(--text-secondary);
  font-style: italic;
  line-height: 1.4;
  margin: 0;
}

.processing-callout {
  background: rgba(0, 113, 227, 0.04);
  border-radius: var(--radius-md);
  padding: 0.75rem 0.85rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.78rem;
  color: var(--apple-blue);
  margin-top: auto;
  font-weight: 500;
}

.text-blue {
  color: var(--apple-blue);
}

/* Footer */
.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 1.15rem;
  padding-top: 0.85rem;
  border-top: 1px solid rgba(0, 0, 0, 0.05);
}

.meta-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  color: var(--text-muted);
  font-size: 0.75rem;
  font-weight: 500;
}

.duration-num {
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  color: var(--text-secondary);
}

.meta-separator {
  color: rgba(0, 0, 0, 0.15);
  font-size: 0.75rem;
}

.view-arrow-circle {
  width: 24px;
  height: 24px;
  border-radius: var(--radius-full);
  background: rgba(0, 113, 227, 0.06);
  color: var(--apple-blue);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transform: translateX(-4px);
  transition: all 0.2s var(--apple-ease);
}

.apple-recording-card:hover .view-arrow-circle {
  opacity: 1;
  transform: translateX(0);
}
</style>
