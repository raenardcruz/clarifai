<template>
  <div class="recording-show-page" v-if="recording">
    <!-- Read Only Header for Shared View -->
    <div class="header-section text-center mb-8">
      <div class="inline-flex items-center gap-2 mb-4">
        <div class="w-8 h-8 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-lg flex items-center justify-center">
          <Sparkles size="16" color="white" />
        </div>
        <span class="font-bold text-lg">ClarifAi Shared Insight</span>
      </div>
      <h1 class="page-title">{{ recording.title }}</h1>
      <div class="meta-info justify-center mt-4">
        <span class="meta-item"><Calendar size="14" /> {{ formattedDate }}</span>
        <span class="meta-item"><Clock size="14" /> {{ formattedDuration }}</span>
      </div>
    </div>

    <div class="tabs-container-custom">
      <Tabs v-model:value="activeTab" class="w-full">
        <div class="flex justify-center border-b border-gray-200 mb-4 pb-2">
          <TabList>
            <Tab value="transcript">Transcript</Tab>
            <Tab value="summary">AI Summary</Tab>
          </TabList>
        </div>

        <TabPanels>
          <TabPanel value="transcript">
            <!-- Transcript Tab (Read Only) -->
            <div class="transcript-view">
              <div v-for="segment in recording.segments" :key="segment.id" class="segment-block">
                <div class="segment-header">
                  <div class="speaker-info">
                    <div class="speaker-avatar" :style="{ backgroundColor: getSpeakerColor(segment.speaker) }">
                      {{ getSpeakerInitials(segment.speaker) }}
                    </div>
                    <div class="speaker-name font-bold">{{ segment.speaker }}</div>
                  </div>
                  <div class="timestamp">{{ formatTimestamp(segment.start) }}</div>
                </div>
                <div class="segment-text text-gray-700">
                  {{ segment.text }}
                </div>
              </div>
            </div>
          </TabPanel>

          <TabPanel value="summary">
            <!-- Summary Tab (Read Only) -->
            <div class="summary-view">
              <div class="summary-card glass" v-if="recording.summary_md">
                <div class="flex items-center gap-2 mb-4">
                  <Sparkles size="20" class="text-tertiary" />
                  <h2 class="text-xl font-bold text-tertiary">Executive Summary</h2>
                </div>
                <div class="markdown-content" v-html="parsedSummary"></div>
              </div>
              <div v-else class="text-center py-20 text-gray-500">
                No summary generated yet.
              </div>
            </div>
          </TabPanel>
        </TabPanels>
      </Tabs>
    </div>
  </div>
  <div v-else class="flex justify-center items-center h-screen">
    <Loader2 class="animate-spin text-primary" size="40" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import axios from 'axios'
import moment from 'moment'
import { marked } from 'marked'
import { useToast } from 'primevue/usetoast'
import { Calendar, Clock, Sparkles, Loader2 } from '@lucide/vue'
import { useTitle } from '@vueuse/core'
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'

const route = useRoute()
const toast = useToast()
const recording = ref(null)
const activeTab = ref('summary')

const title = computed(() => recording.value ? `${recording.value.title} - ClarifAi (Shared)` : 'ClarifAi')
useTitle(title)

onMounted(async () => {
  try {
    const res = await axios.get(`/api/recordings/share/${route.params.id}`)
    recording.value = res.data
    if (!recording.value.summary_md) {
      activeTab.value = 'transcript'
    }
  } catch (e) {
    console.error("Failed to load recording", e)
    toast.add({ severity: 'error', summary: 'Access Error', detail: 'Recording not found or is private', life: 5000 })
  }
})

const parsedSummary = computed(() => {
  return recording.value?.summary_md ? marked(recording.value.summary_md) : ''
})

// Helpers
const formattedDate = computed(() => recording.value ? moment(recording.value.created_at).format('MMM DD, YYYY') : '')
const formattedDuration = computed(() => {
  if (!recording.value || !recording.value.duration) return '--:--'
  const mins = Math.floor(recording.value.duration / 60)
  const secs = Math.floor(recording.value.duration % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
})
const formatTimestamp = (sec) => {
  const m = Math.floor(sec / 60)
  const s = Math.floor(sec % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}
const getSpeakerInitials = (name) => name.split(' ').map(n => n[0]).join('').substring(0, 2).toUpperCase()
const getSpeakerColor = (name) => {
  const colors = ['#4F46E5', '#0D9488', '#A855F7', '#EF4444', '#F59E0B']
  let hash = 0
  for(let i=0; i<name.length; i++) hash += name.charCodeAt(i)
  return colors[hash % colors.length]
}
</script>

<style scoped>
/* Re-use styles from RecordingShowView but removed editing capabilities */
.recording-show-page { max-width: 1000px; margin: 0 auto; padding-top: 2rem; }
.page-title { font-size: 2.25rem; margin-bottom: 0.75rem; color: var(--primary); }
.meta-info { display: flex; align-items: center; gap: 1.5rem; }
.meta-item { display: flex; align-items: center; gap: 0.5rem; color: var(--text-secondary); font-size: 0.875rem; }
.tabs-container { display: flex; justify-content: center; align-items: center; border-bottom: 1px solid var(--neutral-200); margin-bottom: 2rem; padding-bottom: 0.5rem; }
.tabs { display: flex; gap: 1.5rem; }
.tab-btn { background: transparent; border: none; font-size: 1rem; font-weight: 600; color: var(--text-muted); cursor: pointer; padding: 0.5rem 1rem; position: relative; transition: var(--transition); }
.tab-btn.active { color: var(--primary); }
.tab-btn.active::after { content: ''; position: absolute; left: 0; right: 0; bottom: -0.6rem; height: 3px; background: var(--primary); border-radius: 3px 3px 0 0; }
.segment-block { background: var(--white); border-radius: var(--radius-lg); padding: 1.5rem; margin-bottom: 1rem; border: 1px solid var(--neutral-200); }
.segment-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
.speaker-info { display: flex; align-items: center; gap: 0.75rem; }
.speaker-avatar { width: 32px; height: 32px; border-radius: 50%; color: white; font-size: 0.75rem; font-weight: 700; display: flex; align-items: center; justify-content: center; }
.speaker-name { font-size: 1rem; color: var(--text-primary); font-family: var(--font-headline); }
.timestamp { font-size: 0.75rem; color: var(--text-muted); font-variant-numeric: tabular-nums; }
.segment-text { font-family: var(--font-body); font-size: 1rem; line-height: 1.6; }
.summary-card { padding: 2.5rem; border-radius: var(--radius-xl); }
.markdown-content :deep(h1) { font-size: 1.5rem; margin-bottom: 1rem; margin-top: 1.5rem; }
.markdown-content :deep(h2) { font-size: 1.25rem; margin-bottom: 0.75rem; margin-top: 1.5rem; }
.markdown-content :deep(h3) { font-size: 1.1rem; margin-bottom: 0.5rem; margin-top: 1.5rem; }
.markdown-content :deep(p) { margin-bottom: 1rem; color: var(--text-secondary); }
.markdown-content :deep(ul), .markdown-content :deep(ol) { margin-bottom: 1rem; padding-left: 1.5rem; }
.markdown-content :deep(li) { margin-bottom: 0.5rem; color: var(--text-secondary); }
.markdown-content :deep(strong) { color: var(--text-primary); font-weight: 600; }
.markdown-content :deep(blockquote) { border-left: 4px solid var(--primary); padding-left: 1rem; margin-left: 0; color: var(--text-muted); font-style: italic; }
.text-tertiary { color: var(--tertiary); }
</style>
