<template>
  <div class="recording-show-page-wrapper" v-if="recording">
    <div class="recording-show-page">
      <div class="header-section">
        <div class="breadcrumb">
          <router-link to="/recordings">Recordings</router-link>
          <span class="separator">/</span>
          <span class="current">{{ recording.title }}</span>
        </div>

        <div class="title-row">
          <div>
            <h1 class="page-title">{{ recording.title }}</h1>
            <div class="meta-info">
              <span class="meta-item"><Calendar size="14" /> {{ formattedDate }}</span>
              <span class="meta-item"><Clock size="14" /> {{ formattedDuration }}</span>
              <span class="status-badge" :class="statusClass">{{ formattedStatus }}</span>
            </div>
          </div>
          
          <div class="action-buttons gap-2 flex">
            <Button label="Share" icon="pi pi-share-alt" severity="secondary" @click="shareLink" />
            <Button label="Delete" icon="pi pi-trash" severity="danger" @click="deleteRecording" />
          </div>
        </div>
      </div>

      <div class="tabs-container-custom">
        <Tabs v-model:value="activeTab" class="w-full">
          <div class="tabs-header border-b border-gray-200 mb-4 pb-2">
            <TabList>
              <Tab value="transcript">Transcript</Tab>
              <Tab value="summary">AI Summary</Tab>
            </TabList>
            
            <div class="tab-actions flex gap-2">
              <Button v-if="activeTab === 'transcript'" label="Identify Speakers" severity="secondary" size="small" @click="identifySpeakers" :loading="isIdentifyingSpeakers" :disabled="!recording.segments || recording.segments.length === 0">
                <template #icon>
                  <UserCheck size="14" v-if="!isIdentifyingSpeakers" class="mr-1" />
                </template>
              </Button>
              <Button v-if="activeTab === 'summary'" label="Regenerate" severity="secondary" size="small" @click="openSummaryModal" :loading="isGenerating">
                <template #icon>
                  <Wand2 size="14" v-if="!isGenerating" class="mr-1" />
                </template>
              </Button>
              <Button v-if="activeTab === 'summary' && recording?.summary_md" label="Print" severity="secondary" size="small" @click="printSummary" class="no-print">
                <template #icon>
                  <Printer size="14" class="mr-1" />
                </template>
              </Button>
              <a :href="`/api/recordings/${recording.id}/download/${activeTab}`" class="no-underline">
                <Button label="Export" icon="pi pi-download" severity="secondary" size="small" />
              </a>
            </div>
          </div>

          <TabPanels>
            <TabPanel value="transcript">
              <div class="transcript-view">
                <div v-for="segment in recording.segments" :key="segment.id" class="segment-block">
                  <div class="segment-header">
                    <div class="speaker-info">
                      <div class="speaker-avatar" :style="{ backgroundColor: getSpeakerColor(segment.speaker) }">
                        {{ getSpeakerInitials(segment.speaker) }}
                      </div>
                      <div class="speaker-name-wrapper">
                        <InputText type="text" class="speaker-name-input" v-model="segment.speaker" @focus="startEditingSpeaker(segment.speaker)" @blur="updateSegment(segment)" />
                        <Pencil size="12" class="edit-icon" />
                      </div>
                    </div>
                    <div class="timestamp">{{ formatTimestamp(segment.start) }}</div>
                  </div>
                  <div class="segment-text">
                    <Textarea class="text-edit-input w-full" v-model="segment.text" @blur="updateSegment(segment)" rows="2" autoResize />
                  </div>
                </div>
                <div v-if="!recording.segments || recording.segments.length === 0" class="text-center py-10 text-gray-500">
                  No transcript available yet.
                </div>
              </div>
            </TabPanel>

            <TabPanel value="summary">
              <div class="summary-view">
                <div v-if="isProcessing || isGenerating" class="summary-card glass flex flex-col items-center justify-center py-20 text-center">
                  <div class="sparkle-loader mb-6">
                    <Loader2 size="100" class="text-tertiary animate-spin absolute-center" />
                  </div>
                  <h3 class="text-lg font-bold mb-2">ClarifAi is generating meeting insights...</h3>
                  <p class="text-sm text-gray-500 max-w-sm">We're analyzing the conversation, capturing key topics, and drafting your summary.</p>
                </div>
                <div class="summary-card glass" v-else-if="recording.summary_md">
                  <div class="flex items-center justify-between mb-4 screen-summary-header">
                    <div class="flex items-center gap-2">
                      <Sparkles size="20" class="text-tertiary" />
                      <h2 class="text-xl font-bold text-tertiary">Executive Summary</h2>
                    </div>
                  </div>
                  <div class="markdown-content" v-html="parsedSummary"></div>
                </div>
                <div v-else class="text-center py-20 text-gray-500 flex flex-col items-center">
                  <Wand2 size="48" class="text-gray-300 mb-4" />
                  <p>No summary generated yet.</p>
                  <Button label="Generate Summary" severity="primary" class="mt-4" @click="openSummaryModal" :disabled="isGenerating" />
                </div>
              </div>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </div>
    </div>
    <SpeechmaticsUsage type="footer" />

    <Dialog v-model:visible="showSummaryModal" modal header="Generate AI Summary" :style="{ width: '90vw', maxWidth: '600px' }">
      <div class="flex flex-col gap-4 py-2">
        <p class="text-sm text-gray-600">
          Add optional special instructions or focus topics to append to the summary prompt.
        </p>
        <div class="flex flex-col gap-2">
          <label for="special-instruction" class="font-semibold text-sm">Special Instructions (Optional)</label>
          <Textarea
            id="special-instruction"
            v-model="specialInstruction"
            rows="4"
            placeholder="e.g. Focus on action items for the marketing team, key decision points, or specific meeting topics..."
            class="w-full"
            autoResize
          />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-2 pt-2">
          <Button label="Cancel" severity="secondary" @click="showSummaryModal = false" />
          <Button label="Generate Summary" icon="pi pi-sparkles" severity="primary" @click="submitRegenerateSummary" :loading="isGenerating" />
        </div>
      </template>
    </Dialog>
  </div>
  <div v-else class="flex justify-center items-center h-screen">
    <Loader2 class="animate-spin text-primary" size="40" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import moment from 'moment'
import { marked } from 'marked'
import { useToast } from 'primevue/usetoast'
import { Calendar, Clock, Share2, Download, Wand2, Sparkles, Loader2, Pencil, Trash2, UserCheck, Printer } from '@lucide/vue'
import { useIntervalFn, useClipboard, useTitle } from '@vueuse/core'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import SpeechmaticsUsage from '../components/SpeechmaticsUsage.vue'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const recording = ref(null)
const activeTab = ref('summary')
const isGenerating = ref(false)
const isIdentifyingSpeakers = ref(false)
const editingSpeakerOrigName = ref('')
const showSummaryModal = ref(false)
const specialInstruction = ref('')

const { copy } = useClipboard()
const title = computed(() => recording.value ? `${recording.value.title} - ClarifAi` : 'ClarifAi')
useTitle(title)

const deleteRecording = async () => {
  if (confirm(`Are you sure you want to delete "${recording.value.title}"?`)) {
    try {
      await axios.delete(`/api/recordings/${recording.value.id}`)
      toast.add({ severity: 'success', summary: 'Deleted', detail: 'Recording deleted successfully.', life: 3000 })
      router.push('/recordings')
    } catch (e) {
      console.error("Failed to delete recording", e)
      toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to delete recording', life: 5000 })
    }
  }
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
    
    // Stop polling when done or error
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

const startPolling = () => {
  resume()
}

const stopPolling = () => {
  pause()
}

const fetchData = async () => {
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
    toast.add({ severity: 'error', summary: 'Error', detail: 'Recording not found', life: 5000 })
  }
}

onMounted(() => {
  fetchData()
})

onUnmounted(() => {
  stopPolling()
})

watch(isProcessing, (newValue) => {
  if (newValue) {
    startPolling()
  } else {
    stopPolling()
  }
})

const parsedSummary = computed(() => {
  return recording.value?.summary_md ? marked(recording.value.summary_md) : ''
})

const updateSegment = async (segment) => {
  const newSpeakerName = segment.speaker.trim()
  const oldSpeakerName = editingSpeakerOrigName.value.trim()
  
  if (oldSpeakerName && oldSpeakerName !== newSpeakerName) {
    const renameGlobally = confirm(`Do you want to rename "${oldSpeakerName}" to "${newSpeakerName}" for all segments in this recording? Click OK to rename globally, or Cancel to update only this segment.`)
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
      instruction: specialInstruction.value
    })
    toast.add({ severity: 'info', summary: 'Summarization Started', detail: 'Checking progress automatically...', life: 5000 })
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
      toast.add({ severity: 'success', summary: 'Success', detail: `Successfully identified and renamed ${count} speaker(s).`, life: 5000 })
    } else {
      toast.add({ severity: 'info', summary: 'Info', detail: 'No new speaker names could be identified.', life: 5000 })
    }
    await fetchData()
  } catch (e) {
    console.error("Failed to identify speakers", e)
    toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to identify speakers.', life: 5000 })
  } finally {
    isIdentifyingSpeakers.value = false
  }
}


const shareLink = () => {
  const url = `${window.location.origin}/share/${recording.value.id}`
  navigator.clipboard.writeText(url)
  toast.add({ severity: 'success', summary: 'Copied', detail: 'Public link copied to clipboard!', life: 3000 })
}

const printSummary = () => {
  window.print()
}

// Helpers
const formattedDate = computed(() => recording.value ? moment(recording.value.created_at).format('MMM DD, YYYY') : '')
const formattedDuration = computed(() => {
  if (!recording.value || !recording.value.duration) return '--:--'
  const mins = Math.floor(recording.value.duration / 60)
  const secs = Math.floor(recording.value.duration % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
})
const formattedStatus = computed(() => recording.value ? recording.value.status.toUpperCase() : '')
const statusClass = computed(() => {
  if (!recording.value) return ''
  const s = recording.value.status
  if (['pending', 'transcribing', 'diarizing', 'summarizing'].includes(s)) {
    return 'status-processing animate-pulse-slow'
  }
  if (s === 'error') {
    return 'status-error'
  }
  return 'status-success'
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
  const colors = ['#4F46E5', '#0D9488', '#A855F7', '#EF4444', '#F59E0B']
  let hash = 0
  for(let i=0; i<name.length; i++) hash += name.charCodeAt(i)
  return colors[hash % colors.length]
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
  max-width: 1000px;
  width: 100%;
  margin: 0 auto;
  padding: 2rem;
  flex: 1;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  margin-bottom: 1.5rem;
}

.breadcrumb a { color: var(--text-muted); }
.breadcrumb .separator { color: var(--neutral-200); }
.breadcrumb .current { font-weight: 600; color: var(--text-primary); }

.title-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
}

.action-buttons {
  display: flex;
  gap: 0.75rem;
}

.btn-delete {
  color: #EF4444;
  border-color: var(--neutral-200);
}

.btn-delete:hover {
  background-color: #FEE2E2;
  border-color: #EF4444;
  color: #EF4444;
}

.page-title {
  font-size: 2.25rem;
  margin-bottom: 0.75rem;
  color: var(--primary);
}

.meta-info {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.status-badge {
  padding: 0.25rem 0.75rem;
  border-radius: var(--radius-full);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
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



.tabs-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--neutral-200);
  margin-bottom: 2rem;
  padding-bottom: 0.5rem;
}

.tabs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.tabs {
  display: flex;
  gap: 1.5rem;
}

.tab-btn {
  background: transparent;
  border: none;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0.5rem 1rem;
  position: relative;
  transition: var(--transition);
}

.tab-btn.active { color: var(--primary); }
.tab-btn.active::after {
  content: ''; position: absolute; left: 0; right: 0; bottom: -0.6rem;
  height: 3px; background: var(--primary); border-radius: 3px 3px 0 0;
}

.tab-actions { display: flex; gap: 1rem; }
.btn-sm { padding: 0.5rem 1rem; font-size: 0.875rem; }

/* Transcript View */
.segment-block {
  background: var(--white);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  margin-bottom: 1rem;
  border: 1px solid transparent;
  transition: var(--transition);
}

.segment-block:hover { border-color: var(--neutral-200); box-shadow: var(--shadow-sm); }

.segment-header {
  display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem;
}

.speaker-info { display: flex; align-items: center; gap: 0.75rem; }

.speaker-avatar {
  width: 32px; height: 32px; border-radius: 50%;
  color: white; font-size: 0.75rem; font-weight: 700;
  display: flex; align-items: center; justify-content: center;
}

.speaker-name-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.speaker-name-input {
  border: 1px solid transparent; 
  background: transparent; 
  font-weight: 600;
  font-size: 1rem; 
  color: var(--text-primary); 
  font-family: var(--font-headline);
  border-radius: var(--radius-md);
  padding: 0.2rem 0.5rem;
  transition: var(--transition);
  cursor: pointer;
}

.speaker-name-input:hover {
  background: var(--neutral-50);
  border-color: var(--neutral-200);
}

.speaker-name-input:focus { 
  outline: none; 
  border-color: var(--primary);
  background: var(--white);
  cursor: text;
}

.edit-icon {
  color: var(--text-muted);
  opacity: 0;
  transition: var(--transition);
  pointer-events: none;
}

.speaker-name-wrapper:hover .edit-icon,
.speaker-name-input:focus ~ .edit-icon {
  opacity: 1;
}

.timestamp { font-size: 0.75rem; color: var(--text-muted); font-variant-numeric: tabular-nums; }

.text-edit-input {
  width: 100%; border: none; background: transparent; resize: none;
  font-family: var(--font-body); font-size: 1rem; color: var(--text-secondary); line-height: 1.6;
}
.text-edit-input:focus { outline: none; }

/* Summary View */
.summary-card {
  padding: 2.5rem;
  border-radius: var(--radius-xl);
}

.markdown-content :deep(h1) { font-size: 1.5rem; margin-bottom: 1rem; margin-top: 1.5rem; }
.markdown-content :deep(h2) { font-size: 1.25rem; margin-bottom: 0.75rem; margin-top: 1.5rem; }
.markdown-content :deep(h3) { font-size: 1.1rem; margin-bottom: 0.5rem; margin-top: 1.5rem; }
.markdown-content :deep(p) { margin-bottom: 1rem; color: var(--text-secondary); }
.markdown-content :deep(ul), .markdown-content :deep(ol) { margin-bottom: 1rem; padding-left: 1.5rem; }
.markdown-content :deep(li) { margin-bottom: 0.5rem; color: var(--text-secondary); }
.markdown-content :deep(strong) { color: var(--text-primary); font-weight: 600; }
.markdown-content :deep(blockquote) { 
  border-left: 4px solid var(--primary); padding-left: 1rem; margin-left: 0; color: var(--text-muted); font-style: italic;
}

/* Utils */
.text-tertiary { color: var(--tertiary); }

.sparkle-loader {
  position: relative;
  width: 80px;
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.absolute-center {
  position: absolute;
}

@media (max-width: 768px) {
  .page-title {
    font-size: 1.75rem;
  }
  
  .title-row {
    flex-direction: column;
    align-items: stretch;
    gap: 1.25rem;
  }
  
  .meta-info {
    flex-wrap: wrap;
    gap: 0.75rem 1rem;
  }
  
  .tabs-header {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 1rem;
  }
  
  .tab-actions {
    justify-content: flex-end;
  }
  
  .segment-block {
    padding: 1rem;
  }
  
  .segment-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }
  
  .timestamp {
    align-self: flex-end;
  }
  
  .summary-card {
    padding: 1.5rem;
  }
}
</style>
