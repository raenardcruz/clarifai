<template>
  <div class="recording-show-page" v-if="recording">
    <!-- Read Only Header for Shared View -->
    <div class="header-section text-center mb-8 no-print">
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

    <!-- Summary Only View -->
    <div class="summary-view">
      <div class="summary-card glass" v-if="recording.summary_md">
        <div class="flex items-center justify-between mb-4 screen-summary-header">
          <div class="flex items-center gap-2">
            <Sparkles size="20" class="text-tertiary" />
            <h2 class="text-xl font-bold text-tertiary">Executive Summary</h2>
          </div>
          <Button label="Print Summary" severity="secondary" size="small" @click="printSummary" class="no-print">
            <template #icon>
              <Printer size="14" class="mr-1" />
            </template>
          </Button>
        </div>
        <div class="markdown-content" v-html="parsedSummary"></div>
      </div>
      <div v-else class="text-center py-20 text-gray-500">
        No summary generated yet.
      </div>
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
import { Calendar, Clock, Sparkles, Loader2, Printer } from '@lucide/vue'
import { useTitle } from '@vueuse/core'
import Button from 'primevue/button'

const route = useRoute()
const toast = useToast()
const recording = ref(null)

const title = computed(() => recording.value ? `${recording.value.title} - ClarifAi (Shared)` : 'ClarifAi')
useTitle(title)

onMounted(async () => {
  try {
    const res = await axios.get(`/api/recordings/share/${route.params.id}`)
    recording.value = res.data
  } catch (e) {
    console.error("Failed to load recording", e)
    toast.add({ severity: 'error', summary: 'Access Error', detail: 'Recording not found or is private', life: 5000 })
  }
})

const parsedSummary = computed(() => {
  return recording.value?.summary_md ? marked(recording.value.summary_md) : ''
})

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
</script>

<style scoped>
.recording-show-page { max-width: 1000px; margin: 0 auto; padding-top: 2rem; }
.page-title { font-size: 2.25rem; margin-bottom: 0.75rem; color: var(--primary); }
.meta-info { display: flex; align-items: center; gap: 1.5rem; }
.meta-item { display: flex; align-items: center; gap: 0.5rem; color: var(--text-secondary); font-size: 0.875rem; }
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

