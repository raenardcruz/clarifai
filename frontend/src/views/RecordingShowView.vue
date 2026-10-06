<template>
  <div class="recording-show-page-wrapper" v-if="recording">
    <div class="recording-show-page">
      <!-- Apple Header & Breadcrumb -->
      <div class="header-section">
        <div class="breadcrumb-bar">
          <router-link to="/recordings" class="back-link">
            <ChevronLeft size="16" />
            <span>Recordings</span>
          </router-link>
          <span class="separator">/</span>
          <span class="current-title">{{ recording.title }}</span>
        </div>

        <div class="title-row">
          <div class="title-meta-block">
            <h1 class="page-title">{{ recording.title }}</h1>
            
            <div class="meta-pill-group">
              <span class="meta-pill">
                <Calendar size="13" />
                <span>{{ formattedDate }}</span>
              </span>
              <span class="meta-pill">
                <Clock size="13" />
                <span class="font-tabular">{{ formattedDuration }}</span>
              </span>
              <span class="meta-pill" v-if="speakerCount > 0">
                <Users size="13" />
                <span>{{ speakerCount }} {{ speakerCount === 1 ? 'Speaker' : 'Speakers' }}</span>
              </span>
              <!-- Medical Conversation Badge (Read-only) -->
              <span 
                v-if="recording.is_medical" 
                class="meta-pill medical-pill" 
                title="Medical Conversation"
              >
                <Stethoscope size="13" />
                <span>Medical Conversation</span>
              </span>
              <span class="status-pill" :class="statusClass">
                <span class="beacon-dot" :class="beaconClass"></span>
                <span>{{ formattedStatus }}</span>
              </span>
            </div>
          </div>

          
          <!-- Top Action Bar -->
          <div class="action-buttons-group">
            <button class="apple-btn-secondary" @click="shareLink" title="Share public link">
              <Share2 size="15" />
              <span>Share</span>
            </button>
            
            <a :href="`/api/recordings/${recording.id}/download/${activeTab}`" class="no-underline">
              <button class="apple-btn-secondary" title="Export file">
                <Download size="15" />
                <span>Export</span>
              </button>
            </a>

            <button class="apple-btn-danger" @click="deleteRecording" title="Delete recording">
              <Trash2 size="15" />
              <span>Delete</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Apple macOS Segmented Navigation Bar -->
      <div class="tabs-toolbar">
        <div class="segmented-control-macos">
          <button 
            class="segment-item" 
            :class="{ active: activeTab === 'summary' }"
            @click="activeTab = 'summary'"
          >
            <Sparkles size="15" class="tab-icon" />
            <span>Executive Summary</span>
          </button>
          <button 
            class="segment-item" 
            :class="{ active: activeTab === 'transcript' }"
            @click="activeTab = 'transcript'"
          >
            <FileText size="15" class="tab-icon" />
            <span>Full Transcript</span>
            <span class="counter-badge" v-if="recording.segments?.length">{{ recording.segments.length }}</span>
          </button>
        </div>

        <div class="tab-context-actions">
          <!-- Summary Actions -->
          <template v-if="activeTab === 'summary'">
            <button 
              class="apple-tool-btn" 
              @click="copySummary" 
              v-if="recording.summary_md" 
              title="Copy summary text"
            >
              <Check size="14" v-if="copiedSummary" class="text-green" />
              <Copy size="14" v-else />
              <span>{{ copiedSummary ? 'Copied!' : 'Copy Summary' }}</span>
            </button>

            <button 
              class="apple-tool-btn" 
              @click="openSummaryModal" 
              :disabled="isGenerating"
            >
              <Loader2 size="14" class="animate-spin text-blue" v-if="isGenerating" />
              <Wand2 size="14" v-else />
              <span>{{ recording.summary_md ? 'Regenerate' : 'Generate Summary' }}</span>
            </button>

            <button 
              class="apple-tool-btn no-print" 
              @click="printSummary" 
              v-if="recording.summary_md"
            >
              <Printer size="14" />
              <span>Print / PDF</span>
            </button>
          </template>

          <!-- Transcript Actions -->
          <template v-if="activeTab === 'transcript'">
            <div class="transcript-search-box">
              <Search size="14" class="search-ico" />
              <input 
                v-model="transcriptSearch" 
                type="text" 
                placeholder="Search transcript..." 
                class="transcript-search-input"
              />
              <button v-if="transcriptSearch" @click="transcriptSearch = ''" class="clear-mini-btn">✕</button>
            </div>

            <button 
              class="apple-tool-btn" 
              @click="toggleTranscriptSort" 
              :title="transcriptSortOrder === 'asc' ? 'Sorted by time (Earliest first). Click to sort latest first.' : 'Sorted by time (Latest first). Click to sort earliest first.'"
            >
              <ArrowUpDown size="14" />
              <span>{{ transcriptSortOrder === 'asc' ? 'Time: Earliest' : 'Time: Latest' }}</span>
            </button>

            <button 
              class="apple-tool-btn" 
              @click="identifySpeakers" 
              :disabled="isIdentifyingSpeakers || !recording.segments || recording.segments.length === 0"
              title="Detect speaker names with AI"
            >
              <Loader2 size="14" class="animate-spin" v-if="isIdentifyingSpeakers" />
              <UserCheck size="14" v-else />
              <span>Identify Speakers</span>
            </button>
          </template>
        </div>
      </div>

      <!-- Tab Content Area -->
      <div class="tab-content-container">
        <!-- AI Summary Panel (Apple Notes / Pages Document) -->
        <div v-if="activeTab === 'summary'" class="summary-view">
          <!-- Processing State -->
          <div v-if="isProcessing || isGenerating" class="apple-doc-card generating-card">
            <div class="ambient-sparkle-halo">
              <Loader2 size="50" class="animate-spin text-blue" />
            </div>
            <h3 class="generating-title">ClarifAi is analyzing meeting insights...</h3>
            <p class="generating-subtitle">
              Synthesizing conversations, attributing speaker viewpoints, and extracting key decisions.
            </p>
          </div>

          <!-- Document Rendered State -->
          <div class="apple-doc-card" v-else-if="recording.summary_md">
            <div class="doc-header screen-summary-header">
              <div class="doc-title-group">
                <div class="doc-icon-badge" :class="{ 'medical-icon-badge': recording.is_medical }">
                  <Stethoscope v-if="recording.is_medical" size="18" class="text-emerald-600" />
                  <Sparkles v-else size="18" color="#0071E3" />
                </div>
                <div>
                  <h2 class="doc-heading">{{ recording.is_medical ? 'Clinical SOAP Note' : 'Executive Summary' }}</h2>
                  <span class="doc-subheading">{{ recording.is_medical ? 'Structured Clinical Consultation Summary' : 'Generated by ClarifAi Local LLM' }}</span>
                </div>
              </div>


              <div class="doc-meta-stats">
                <span class="meta-stat-item">{{ wordCount }} words</span>
                <span class="meta-stat-divider">•</span>
                <span class="meta-stat-item">{{ readingTime }} min read</span>
              </div>
            </div>

            <div class="markdown-content" v-html="parsedSummary"></div>
          </div>

          <!-- Empty Summary State -->
          <div v-else class="apple-doc-card empty-summary-card">
            <div class="empty-sparkle-circle">
              <Wand2 size="32" stroke-width="1.8" />
            </div>
            <h3>No Executive Summary Yet</h3>
            <p>Generate structured meeting notes, decisions, and action items with one click.</p>
            <button class="apple-primary-btn mt-4" @click="openSummaryModal">
              <Sparkles size="16" />
              <span>Generate Summary</span>
            </button>
          </div>
        </div>

        <!-- Full Transcript Panel (Apple Messages / Dialogue Timeline) -->
        <div v-if="activeTab === 'transcript'" class="transcript-view">
          <div v-if="filteredSegments.length > 0" class="segments-timeline">
            <div 
              v-for="segment in filteredSegments" 
              :key="segment.id" 
              class="segment-card"
            >
              <div class="segment-header">
                <div class="speaker-profile">
                  <div class="speaker-avatar" :style="{ background: getSpeakerColor(segment.speaker) }">
                    {{ getSpeakerInitials(segment.speaker) }}
                  </div>
                  
                  <div class="speaker-name-container">
                    <input 
                      type="text" 
                      class="speaker-name-input" 
                      v-model="segment.speaker" 
                      @focus="startEditingSpeaker(segment.speaker)" 
                      @blur="updateSegment(segment)" 
                      title="Click to rename speaker"
                    />
                    <Pencil size="11" class="pencil-icon" />
                  </div>
                </div>

                <span class="segment-timestamp font-tabular">
                  {{ formatTimestamp(segment.start) }}
                </span>
              </div>

              <div class="segment-body">
                <textarea 
                  class="segment-textarea" 
                  v-model="segment.text" 
                  @blur="updateSegment(segment)" 
                  rows="2"
                  @input="autoGrow"
                ></textarea>
              </div>
            </div>
          </div>

          <!-- Empty Search / No Segments -->
          <div v-else class="apple-doc-card empty-transcript-card">
            <p v-if="transcriptSearch">No segments found matching "{{ transcriptSearch }}".</p>
            <p v-else>No transcript segments available for this recording.</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Speechmatics Usage Footer -->
    <SpeechmaticsUsage type="footer" />

    <!-- Apple AI Prompt Dialog -->
    <Dialog 
      v-model:visible="showSummaryModal" 
      modal 
      header="Generate AI Summary" 
      :style="{ width: '90vw', maxWidth: '580px' }"
    >
      <div class="modal-dialog-content">
        <p class="dialog-desc">
          Customize instructions or select quick focus topics for the AI summarization engine.
        </p>

        <!-- Medical Encounter Status (Set before recording/upload) -->
        <div 
          v-if="recording?.is_medical" 
          class="modal-medical-info-banner"
        >
          <div class="modal-medical-icon">
            <Stethoscope size="16" />
          </div>
          <div class="modal-medical-text">
            <span class="modal-medical-title">Clinical SOAP / Medical Encounter</span>
            <span class="modal-medical-sub">Summary will follow structured Subjective, Objective, Assessment, and Plan format</span>
          </div>
        </div>

        <!-- Quick Focus Topic Pills -->

        <div class="quick-prompt-chips">
          <span class="chips-label">Quick Presets:</span>
          <button 
            type="button" 
            class="chip-btn" 
            v-for="preset in promptPresets" 
            :key="preset.label"
            @click="applyPreset(preset.text)"
          >
            {{ preset.label }}
          </button>
        </div>

        <div class="form-group mt-3">
          <label for="special-instruction" class="field-label">Special Focus Instructions</label>
          <textarea
            id="special-instruction"
            v-model="specialInstruction"
            rows="4"
            placeholder="e.g. Emphasize marketing deliverables, key deadlines, and assigned action owners..."
            class="apple-textarea"
          ></textarea>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer-actions">
          <button class="apple-btn-secondary" @click="showSummaryModal = false">
            Cancel
          </button>
          <button class="apple-primary-btn" @click="submitRegenerateSummary" :disabled="isGenerating">
            <Sparkles size="15" />
            <span>Generate Summary</span>
          </button>
        </div>
      </template>
    </Dialog>
  </div>

  <!-- Error State -->
  <div v-else-if="loadError" class="full-page-loader">
    <div class="apple-doc-card text-center max-w-md p-8">
      <h3 style="font-size: 1.15rem; font-weight: 700; color: #E03131; margin-bottom: 0.5rem;">Unable to Load Recording</h3>
      <p style="font-size: 0.9rem; color: var(--text-secondary); margin-bottom: 1.25rem;">{{ loadError }}</p>
      <div style="display: flex; justify-content: center; gap: 0.75rem;">
        <router-link to="/recordings" class="apple-btn-secondary" style="text-decoration: none;">
          Back to Recordings
        </router-link>
        <button class="apple-primary-btn" @click="fetchData">
          Retry
        </button>
      </div>
    </div>
  </div>

  <!-- Loading Full Page -->
  <div v-else class="full-page-loader">
    <Loader2 class="animate-spin text-blue" size="44" />
    <span class="loader-label">Loading recording...</span>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import moment from 'moment'
import { marked } from 'marked'
import { useToast } from 'primevue/usetoast'
import { 
  Calendar, Clock, Share2, Download, Wand2, Sparkles, Loader2, 
  Pencil, Trash2, UserCheck, Printer, ChevronLeft, Users, 
  FileText, Copy, Check, Search, Stethoscope, ArrowUpDown 
} from '@lucide/vue'
import { useIntervalFn, useTitle } from '@vueuse/core'
import Dialog from 'primevue/dialog'
import { Capacitor } from '@capacitor/core'
import SpeechmaticsUsage from '../components/SpeechmaticsUsage.vue'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const recording = ref(null)
const loadError = ref(null)
const activeTab = ref('summary')
const isGenerating = ref(false)
const isIdentifyingSpeakers = ref(false)
const editingSpeakerOrigName = ref('')
const showSummaryModal = ref(false)
const transcriptSortOrder = ref('asc')
const specialInstruction = ref('')
const copiedSummary = ref(false)
const transcriptSearch = ref('')

const title = computed(() => recording.value ? `${recording.value.title} — ClarifAi` : 'ClarifAi')
useTitle(title)

const promptPresets = [
  { label: 'Clinical SOAP Note', text: 'Structure as a comprehensive clinical SOAP Note with Subjective history, Objective clinical observations, Assessment, and Plan.' },
  { label: 'Action Items & Next Steps', text: 'Highlight key action items, tasks, assignees, and deadlines in explicit detail.' },
  { label: 'Key Decisions', text: 'Focus exclusively on critical decisions made during this meeting and their rationale.' },
  { label: 'Executive Brief', text: 'Provide a concise, high-level summary suitable for executive leadership review.' },
  { label: 'Technical Details', text: 'Capture architectural, technical, and engineering details thoroughly.' }
]


const applyPreset = (text) => {
  if (specialInstruction.value) {
    specialInstruction.value += ' ' + text
  } else {
    specialInstruction.value = text
  }
}

const autoGrow = (e) => {
  e.target.style.height = 'auto'
  e.target.style.height = (e.target.scrollHeight) + 'px'
}

const speakerCount = computed(() => {
  if (!recording.value?.segments) return 0
  const unique = new Set(recording.value.segments.map(s => s.speaker))
  return unique.size
})

const wordCount = computed(() => {
  if (!recording.value?.summary_md) return 0
  return recording.value.summary_md.trim().split(/\s+/).length
})

const readingTime = computed(() => {
  return Math.max(1, Math.ceil(wordCount.value / 200))
})

const toggleTranscriptSort = () => {
  transcriptSortOrder.value = transcriptSortOrder.value === 'asc' ? 'desc' : 'asc'
}

const sortedSegments = computed(() => {
  if (!recording.value?.segments) return []
  return [...recording.value.segments].sort((a, b) => {
    const timeA = typeof a.start === 'number' ? a.start : (parseFloat(a.start) || 0)
    const timeB = typeof b.start === 'number' ? b.start : (parseFloat(b.start) || 0)
    return transcriptSortOrder.value === 'asc' ? timeA - timeB : timeB - timeA
  })
})

const filteredSegments = computed(() => {
  const list = sortedSegments.value
  if (!transcriptSearch.value) return list
  const q = transcriptSearch.value.toLowerCase()
  return list.filter(s => 
    s.speaker.toLowerCase().includes(q) || s.text.toLowerCase().includes(q)
  )
})

const deleteRecording = async () => {
  if (confirm(`Are you sure you want to delete "${recording.value.title}"?`)) {
    try {
      await axios.delete(`/api/recordings/${recording.value.id}`)
      toast.add({ severity: 'success', summary: 'Deleted', detail: 'Recording removed successfully.', life: 3000 })
      router.push('/recordings')
    } catch (e) {
      console.error("Failed to delete recording", e)
      toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to delete recording.', life: 5000 })
    }
  }
}

const copySummary = () => {
  if (!recording.value?.summary_md) return
  navigator.clipboard.writeText(recording.value.summary_md)
  copiedSummary.value = true
  toast.add({ severity: 'success', summary: 'Copied', detail: 'Summary copied to clipboard.', life: 2500 })
  setTimeout(() => { copiedSummary.value = false }, 2500)
}

const startEditingSpeaker = (name) => {
  editingSpeakerOrigName.value = name
}

const isProcessing = computed(() => {
  if (!recording.value) return false
  return ['pending', 'transcribing', 'diarizing', 'summarizing'].includes(recording.value.status)
})

const { pause, resume } = useIntervalFn(async () => {
  try {
    const res = await axios.get(`/api/recordings/${route.params.id}`)
    recording.value = res.data
    
    if (!isProcessing.value) {
      pause()
      if (recording.value.summary_md) {
        activeTab.value = 'summary'
      }
    }
  } catch (e) {
    console.error("Polling failed", e)
    pause()
  }
}, 3000, { immediate: false })

const startPolling = () => { resume() }
const stopPolling = () => { pause() }

const fetchData = async () => {
  loadError.value = null
  try {
    const res = await axios.get(`/api/recordings/${route.params.id}`)
    recording.value = res.data
    if (!recording.value.summary_md) {
      activeTab.value = 'transcript'
    }
    
    if (isProcessing.value) {
      startPolling()
    }
  } catch (e) {
    console.error("Failed to load recording", e)
    loadError.value = e.response?.data?.detail || e.message || 'Recording not found'
    toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to load recording. Please check connection or try again.', life: 5000 })
  }
}

onMounted(() => { fetchData() })
onUnmounted(() => { stopPolling() })

watch(isProcessing, (newValue) => {
  if (newValue) { startPolling() } else { stopPolling() }
})

const parsedSummary = computed(() => {
  return recording.value?.summary_md ? marked(recording.value.summary_md) : ''
})

const updateSegment = async (segment) => {
  const newSpeakerName = segment.speaker.trim()
  const oldSpeakerName = editingSpeakerOrigName.value.trim()
  
  if (oldSpeakerName && oldSpeakerName !== newSpeakerName) {
    const renameGlobally = confirm(`Rename all occurrences of "${oldSpeakerName}" to "${newSpeakerName}" across this recording?`)
    if (renameGlobally) {
      try {
        const promises = []
        for (let seg of recording.value.segments) {
          if (seg.speaker === oldSpeakerName) {
            seg.speaker = newSpeakerName
            promises.push(
              axios.put(`/api/recordings/${recording.value.id}/segments/${seg.id}`, {
                speaker: newSpeakerName,
                text: seg.text
              })
            )
          }
        }
        await Promise.all(promises)
        toast.add({ severity: 'success', summary: 'Updated', detail: `Renamed speaker globally to "${newSpeakerName}".`, life: 3000 })
      } catch (e) {
        console.error("Failed to rename speaker globally", e)
      }
      return
    }
  }

  try {
    await axios.put(`/api/recordings/${recording.value.id}/segments/${segment.id}`, {
      speaker: newSpeakerName,
      text: segment.text
    })
  } catch (e) {
    console.error("Failed to update segment", e)
  }
}

const openSummaryModal = () => {
  showSummaryModal.value = true
}

const submitRegenerateSummary = async () => {
  isGenerating.value = true
  showSummaryModal.value = false
  try {
    await axios.post(`/api/recordings/${recording.value.id}/summarize`, {
      instruction: specialInstruction.value,
      is_medical: !!recording.value?.is_medical
    })
    toast.add({ severity: 'info', summary: 'AI Synthesis Started', detail: 'Generating summary...', life: 5000 })
    fetchData()
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to start summarization.', life: 5000 })
  } finally {
    isGenerating.value = false
  }
}


const identifySpeakers = async () => {
  isIdentifyingSpeakers.value = true
  try {
    const res = await axios.post(`/api/recordings/${recording.value.id}/detect-speakers`)
    const mapping = res.data.mapping
    const count = Object.keys(mapping || {}).length
    if (count > 0) {
      toast.add({ severity: 'success', summary: 'Speakers Identified', detail: `Successfully identified ${count} speaker(s).`, life: 5000 })
    } else {
      toast.add({ severity: 'info', summary: 'Info', detail: 'No new speaker names identified.', life: 5000 })
    }
    await fetchData()
  } catch (e) {
    console.error("Failed to identify speakers", e)
    toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to detect speakers.', life: 5000 })
  } finally {
    isIdentifyingSpeakers.value = false
  }
}

const shareLink = () => {
  const origin = window.location.origin || ''
  const isLocalOrNative = Capacitor.isNativePlatform() || 
    origin.includes('localhost') || 
    origin.startsWith('notetaker') || 
    origin.startsWith('capacitor')
    
  const publicBase = import.meta.env.VITE_APP_URL || (isLocalOrNative ? 'https://clarifai.raenardcruz.com' : origin)
  const url = `${publicBase.replace(/\/+$/, '')}/share/${recording.value.id}`

  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(url)
  }
  toast.add({ 
    severity: 'success', 
    summary: 'Copied Public Link', 
    detail: 'Public sharing link copied to clipboard!', 
    life: 3500 
  })
}

const printSummary = () => {
  window.print()
}

// Formatting Helpers
const formattedDate = computed(() => recording.value ? moment(recording.value.created_at).format('MMM D, YYYY') : '')
const formattedDuration = computed(() => {
  if (!recording.value || !recording.value.duration) return '00:00'
  const mins = Math.floor(recording.value.duration / 60)
  const secs = Math.floor(recording.value.duration % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
})

const formattedStatus = computed(() => recording.value ? recording.value.status.toUpperCase() : '')

const statusClass = computed(() => {
  if (!recording.value) return ''
  const s = recording.value.status
  if (['pending', 'transcribing', 'diarizing', 'summarizing'].includes(s)) return 'status-proc'
  if (s === 'error') return 'status-err'
  return 'status-ok'
})

const beaconClass = computed(() => {
  if (isProcessing.value) return 'beacon-blue animate-pulse'
  if (recording.value?.status === 'error') return 'beacon-red'
  return 'beacon-green'
})

const formatTimestamp = (sec) => {
  const m = Math.floor(sec / 60)
  const s = Math.floor(sec % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}

const getSpeakerInitials = (name) => {
  return name.split(' ').map(n => n[0]).join('').substring(0, 2).toUpperCase()
}

const getSpeakerColor = (name) => {
  const gradients = [
    'linear-gradient(135deg, #0071E3, #30B0C7)',
    'linear-gradient(135deg, #AF52DE, #5856D6)',
    'linear-gradient(135deg, #34C759, #30B0C7)',
    'linear-gradient(135deg, #FF9500, #FF3B30)',
    'linear-gradient(135deg, #5856D6, #0071E3)'
  ]
  let hash = 0
  for(let i=0; i<name.length; i++) hash += name.charCodeAt(i)
  return gradients[hash % gradients.length]
}
</script>

<style scoped>
.recording-show-page-wrapper {
  display: flex;
  flex-direction: column;
  min-height: calc(100vh - var(--header-height));
  justify-content: space-between;
}

.recording-show-page {
  max-width: 1080px;
  width: 100%;
  margin: 0 auto;
  padding: 2.25rem 2rem;
  flex: 1;
}

/* Breadcrumb */
.breadcrumb-bar {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
  margin-bottom: 1.25rem;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  color: var(--apple-blue);
  font-weight: 500;
  text-decoration: none;
  transition: var(--transition-fast);
}

.back-link:hover {
  color: var(--apple-blue-hover);
  transform: translateX(-2px);
}

.separator {
  color: rgba(0, 0, 0, 0.2);
}

.current-title {
  color: var(--text-secondary);
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 400px;
}

/* Title Row */
.title-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
  gap: 1.5rem;
  flex-wrap: wrap;
}

.page-title {
  font-size: 2.25rem;
  font-weight: 700;
  letter-spacing: -0.03em;
  color: var(--text-primary);
  margin-bottom: 0.75rem;
  line-height: 1.2;
}

.meta-pill-group {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  flex-wrap: wrap;
}

.meta-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  background: var(--bg-secondary);
  border: 1px solid rgba(0, 0, 0, 0.06);
  padding: 0.25rem 0.65rem;
  border-radius: var(--radius-full);
  font-size: 0.75rem;
  color: var(--text-secondary);
  font-weight: 500;
}

.medical-pill {
  background: rgba(16, 185, 129, 0.12);
  border: 1px solid rgba(16, 185, 129, 0.3);
  color: #047857;
  font-weight: 600;
}

.medical-icon-badge {
  background: rgba(16, 185, 129, 0.12) !important;
}

.modal-medical-info-banner {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background: rgba(16, 185, 129, 0.08);
  border: 1.5px solid rgba(16, 185, 129, 0.25);
  border-radius: var(--radius-md);
  margin-bottom: 1rem;
}

.modal-medical-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: rgba(16, 185, 129, 0.15);
  color: #059669;
}

.modal-medical-text {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.modal-medical-title {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-primary);
}

.modal-medical-sub {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.apple-toggle-switch-sm {
  width: 38px;
  height: 22px;
  background: #E5E7EB;
  border-radius: 11px;
  padding: 2px;
  cursor: pointer;
  transition: background-color 0.25s ease;
  position: relative;
  flex-shrink: 0;
}

.apple-toggle-switch-sm.checked {
  background: #10B981;
}

.switch-handle-sm {
  display: block;
  width: 18px;
  height: 18px;
  background: #FFFFFF;
  border-radius: 50%;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  transition: transform 0.25s ease;
}

.apple-toggle-switch-sm.checked .switch-handle-sm {
  transform: translateX(16px);
}


.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.25rem 0.65rem;
  border-radius: var(--radius-full);
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.beacon-dot {
  width: 6px;
  height: 6px;
  border-radius: var(--radius-full);
}

.beacon-blue { background-color: var(--apple-blue); }
.beacon-green { background-color: var(--apple-green); }
.beacon-red { background-color: var(--apple-red); }

.status-ok { background: var(--apple-green-light); color: #1E8E3E; }
.status-proc { background: var(--apple-blue-light); color: var(--apple-blue); }
.status-err { background: var(--apple-red-light); color: var(--apple-red); }

.font-tabular {
  font-variant-numeric: tabular-nums;
}

/* Action Buttons Group */
.action-buttons-group {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.apple-btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  background: var(--bg-secondary);
  border: 1px solid rgba(0, 0, 0, 0.08);
  padding: 0.55rem 0.95rem;
  border-radius: var(--radius-full);
  font-size: 0.825rem;
  font-weight: 600;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s var(--apple-ease);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.02);
}

.apple-btn-secondary:hover {
  background: rgba(0, 0, 0, 0.04);
  transform: translateY(-1px);
}

.apple-btn-danger {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  background: var(--bg-secondary);
  border: 1px solid rgba(255, 59, 48, 0.15);
  padding: 0.55rem 0.95rem;
  border-radius: var(--radius-full);
  font-size: 0.825rem;
  font-weight: 600;
  color: var(--apple-red);
  cursor: pointer;
  transition: all 0.2s var(--apple-ease);
}

.apple-btn-danger:hover {
  background: var(--apple-red-light);
  border-color: var(--apple-red);
}

.apple-primary-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--apple-blue);
  color: #FFFFFF;
  border: 1px solid rgba(255, 255, 255, 0.2);
  padding: 0.6rem 1.2rem;
  border-radius: var(--radius-full);
  font-weight: 600;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.2s var(--apple-ease);
}

.apple-primary-btn:hover {
  background: var(--apple-blue-hover);
  transform: translateY(-1px);
}

/* Tabs Toolbar */
.tabs-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid rgba(0, 0, 0, 0.07);
  flex-wrap: wrap;
  gap: 1rem;
}

.segmented-control-macos {
  display: inline-flex;
  align-items: center;
  background: rgba(118, 118, 128, 0.12);
  border-radius: var(--radius-md);
  padding: 3px;
  gap: 3px;
}

.segment-item {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.45rem 1rem;
  border-radius: 9px;
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text-secondary);
  border: none;
  background: transparent;
  cursor: pointer;
  transition: var(--transition-fast);
}

.segment-item:hover {
  color: var(--text-primary);
}

.segment-item.active {
  background: var(--white);
  color: var(--text-primary);
  font-weight: 600;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08), 0 1px 1px rgba(0, 0, 0, 0.04);
}

.segment-item.active .tab-icon {
  color: var(--apple-blue);
}

.counter-badge {
  font-size: 0.7rem;
  background: rgba(0, 0, 0, 0.06);
  padding: 0.1rem 0.45rem;
  border-radius: var(--radius-full);
  margin-left: 0.2rem;
}

.tab-context-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.apple-tool-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  background: var(--bg-secondary);
  border: 1px solid rgba(0, 0, 0, 0.07);
  padding: 0.45rem 0.85rem;
  border-radius: var(--radius-full);
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-secondary);
  cursor: pointer;
  transition: var(--transition-fast);
}

.apple-tool-btn:hover {
  background: rgba(0, 0, 0, 0.04);
  color: var(--text-primary);
  border-color: rgba(0, 0, 0, 0.12);
}

.apple-tool-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.transcript-search-box {
  position: relative;
  display: flex;
  align-items: center;
  background: rgba(118, 118, 128, 0.08);
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: var(--radius-full);
  padding: 0.35rem 0.75rem;
  height: 32px;
}

.search-ico {
  color: var(--text-muted);
  margin-right: 0.35rem;
}

.transcript-search-input {
  border: none;
  background: transparent;
  font-size: 0.8rem;
  font-family: inherit;
  color: var(--text-primary);
  outline: none;
  width: 150px;
}

.clear-mini-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 0.75rem;
}

/* Apple Document Card (Summary) */
.apple-doc-card {
  background: var(--white);
  border-radius: var(--radius-xl);
  padding: 3rem 3.5rem;
  border: 1px solid rgba(0, 0, 0, 0.06);
  box-shadow: 0 4px 20px -2px rgba(0, 0, 0, 0.03), 0 1px 2px rgba(0, 0, 0, 0.02);
}

.doc-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 1.5rem;
  margin-bottom: 2rem;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}

.doc-title-group {
  display: flex;
  align-items: center;
  gap: 0.85rem;
}

.doc-icon-badge {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  background: var(--apple-blue-light);
  display: flex;
  align-items: center;
  justify-content: center;
}

.doc-heading {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.doc-subheading {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.doc-meta-stats {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8rem;
  color: var(--text-muted);
}

.meta-stat-divider {
  color: rgba(0, 0, 0, 0.15);
}

/* Generating Animation Card */
.generating-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 5rem 2rem;
  text-align: center;
}

.ambient-sparkle-halo {
  width: 90px;
  height: 90px;
  border-radius: var(--radius-full);
  background: var(--apple-blue-light);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.5rem;
}

.generating-title {
  font-size: 1.35rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
}

.generating-subtitle {
  font-size: 0.9rem;
  color: var(--text-secondary);
  max-width: 460px;
}

.empty-summary-card, .empty-transcript-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 5rem 2rem;
  text-align: center;
}

.empty-sparkle-circle {
  width: 64px;
  height: 64px;
  border-radius: var(--radius-full);
  background: rgba(0, 0, 0, 0.04);
  color: var(--text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
}

/* Markdown Styling */
.markdown-content {
  font-size: 1.025rem;
  line-height: 1.7;
  color: var(--text-primary);
}

.markdown-content :deep(h1) {
  font-size: 1.6rem;
  font-weight: 700;
  margin-top: 2rem;
  margin-bottom: 0.85rem;
  color: var(--text-primary);
  letter-spacing: -0.025em;
}

.markdown-content :deep(h2) {
  font-size: 1.3rem;
  font-weight: 700;
  margin-top: 1.75rem;
  margin-bottom: 0.65rem;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.markdown-content :deep(h3) {
  font-size: 1.125rem;
  font-weight: 600;
  margin-top: 1.25rem;
  margin-bottom: 0.5rem;
  color: var(--text-primary);
}

.markdown-content :deep(p) {
  margin-bottom: 1.2rem;
  color: #333336;
}

.markdown-content :deep(ul), .markdown-content :deep(ol) {
  margin-bottom: 1.25rem;
  padding-left: 1.5rem;
}

.markdown-content :deep(li) {
  margin-bottom: 0.5rem;
  color: #333336;
}

.markdown-content :deep(strong) {
  color: #000000;
  font-weight: 600;
}

.markdown-content :deep(blockquote) {
  border-left: 3.5px solid var(--apple-blue);
  background: var(--neutral-50);
  padding: 0.85rem 1.25rem;
  margin: 1.5rem 0;
  border-radius: 0 var(--radius-md) var(--radius-md) 0;
  color: var(--text-secondary);
  font-style: italic;
}

/* Transcript View & Dialogue Timeline */
.segments-timeline {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.segment-card {
  background: var(--white);
  border-radius: var(--radius-lg);
  padding: 1.25rem 1.5rem;
  border: 1px solid rgba(0, 0, 0, 0.05);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.02);
  transition: all 0.2s var(--apple-ease);
}

.segment-card:hover {
  border-color: rgba(0, 113, 227, 0.25);
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.04);
}

.segment-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.85rem;
}

.speaker-profile {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.speaker-avatar {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-full);
  color: #ffffff;
  font-size: 0.75rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
  flex-shrink: 0;
}

.speaker-name-container {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.speaker-name-input {
  border: 1px solid transparent;
  background: transparent;
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text-primary);
  font-family: inherit;
  border-radius: var(--radius-sm);
  padding: 0.15rem 0.45rem;
  transition: var(--transition-fast);
  cursor: pointer;
}

.speaker-name-input:hover {
  background: rgba(0, 0, 0, 0.04);
  border-color: rgba(0, 0, 0, 0.08);
}

.speaker-name-input:focus {
  outline: none;
  background: var(--white);
  border-color: var(--apple-blue);
  box-shadow: 0 0 0 2px rgba(0, 113, 227, 0.2);
  cursor: text;
}

.pencil-icon {
  color: var(--text-muted);
  opacity: 0;
  transition: opacity 0.2s ease;
}

.speaker-name-container:hover .pencil-icon {
  opacity: 0.7;
}

.segment-timestamp {
  font-size: 0.75rem;
  color: var(--text-muted);
  font-weight: 500;
}

.segment-textarea {
  width: 100%;
  border: none;
  background: transparent;
  resize: vertical;
  font-family: inherit;
  font-size: 0.95rem;
  color: #2c2c2e;
  line-height: 1.6;
  outline: none;
  padding: 0.25rem;
  border-radius: var(--radius-xs);
  transition: var(--transition-fast);
}

.segment-textarea:focus {
  background: rgba(0, 113, 227, 0.02);
  box-shadow: 0 0 0 1px rgba(0, 113, 227, 0.3);
}

/* Modal Dialog */
.modal-dialog-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.dialog-desc {
  font-size: 0.88rem;
  color: var(--text-secondary);
}

.quick-prompt-chips {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.45rem;
  margin-top: 0.5rem;
}

.chips-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-muted);
  margin-right: 0.2rem;
}

.chip-btn {
  background: rgba(0, 113, 227, 0.07);
  color: var(--apple-blue);
  border: 1px solid rgba(0, 113, 227, 0.15);
  padding: 0.3rem 0.65rem;
  border-radius: var(--radius-full);
  font-size: 0.75rem;
  font-weight: 500;
  cursor: pointer;
  transition: var(--transition-fast);
}

.chip-btn:hover {
  background: rgba(0, 113, 227, 0.15);
  transform: translateY(-1px);
}

.field-label {
  font-size: 0.825rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 0.4rem;
  display: block;
}

.apple-textarea {
  width: 100%;
  border: 1px solid rgba(0, 0, 0, 0.15);
  border-radius: var(--radius-md);
  padding: 0.75rem 1rem;
  font-family: inherit;
  font-size: 0.9rem;
  outline: none;
  transition: var(--transition-fast);
  resize: vertical;
}

.apple-textarea:focus {
  border-color: var(--apple-blue);
  box-shadow: 0 0 0 3px rgba(0, 113, 227, 0.15);
}

.dialog-footer-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.65rem;
}

/* Full Page Loader */
.full-page-loader {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 80vh;
  gap: 1rem;
}

.loader-label {
  font-size: 0.95rem;
  color: var(--text-muted);
  font-weight: 500;
}

.text-green { color: var(--apple-green); }

@media (max-width: 768px) {
  .recording-show-page {
    padding: 1.5rem 1rem;
  }
  
  .page-title {
    font-size: 1.75rem;
  }

  .apple-doc-card {
    padding: 1.75rem 1.25rem;
  }

  .doc-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }

  .tabs-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .tab-context-actions {
    flex-wrap: wrap;
    justify-content: flex-end;
  }

  .transcript-search-box {
    width: 100%;
  }

  .transcript-search-input {
    width: 100%;
  }
}
</style>
