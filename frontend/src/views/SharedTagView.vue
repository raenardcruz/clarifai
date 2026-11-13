<template>
  <div class="shared-tag-container" v-if="sharedTag">
    <!-- Read-Only Header for Shared Tag -->
    <header class="shared-header no-print">
      <div class="brand-row">
        <div class="shared-logo-icon">
          <Sparkles size="16" color="#FFFFFF" stroke-width="2.5" />
        </div>
        <span class="brand-title">ClarifAi Shared Collection</span>
      </div>

      <div class="tag-title-row">
        <span class="tag-title-dot" :style="{ backgroundColor: sharedTag.color || '#3b82f6' }"></span>
        <h1 class="page-title">{{ sharedTag.name }}</h1>
      </div>

      <div class="meta-pills-row">
        <span class="meta-pill">
          <FileText size="13" />
          <span>{{ recordings.length }} {{ recordings.length === 1 ? 'Recording' : 'Recordings' }}</span>
        </span>
        <span class="meta-pill pill-shared">
          <span>Read-Only Preview</span>
        </span>
      </div>

      <!-- Client-side Search & Sub-tag Filter -->
      <div class="shared-filter-bar" v-if="recordings.length > 0">
        <div class="search-input-wrap">
          <Search size="15" class="search-icon" />
          <input
            v-model="searchQuery"
            type="text"
            class="shared-search-input"
            placeholder="Search recordings in this tag..."
          />
          <button v-if="searchQuery" type="button" class="clear-search-btn" @click="searchQuery = ''">×</button>
        </div>

        <div class="subtag-pills" v-if="subtags.length > 0">
          <button
            type="button"
            class="subtag-pill"
            :class="{ active: selectedSubtagId === null }"
            @click="selectedSubtagId = null"
          >
            All
          </button>
          <button
            v-for="st in subtags"
            :key="st.id"
            type="button"
            class="subtag-pill"
            :class="{ active: selectedSubtagId === st.id }"
            @click="selectedSubtagId = selectedSubtagId === st.id ? null : st.id"
          >
            <span class="subtag-dot" :style="{ backgroundColor: st.color || '#3b82f6' }"></span>
            <span>{{ st.name }}</span>
          </button>
        </div>
      </div>
    </header>

    <!-- Empty Tag State -->
    <div v-if="recordings.length === 0" class="apple-doc-card empty-card">
      <p>No recordings have been added to this tag yet.</p>
    </div>

    <!-- No Search Match State -->
    <div v-else-if="filteredRecordings.length === 0" class="apple-doc-card empty-card">
      <p>No recordings match your current filter.</p>
    </div>

    <!-- Master-Detail Read-Only Layout -->
    <main v-else class="shared-layout">
      <!-- Recording List -->
      <aside class="recording-list-pane no-print">
        <button
          v-for="rec in filteredRecordings"
          :key="rec.id"
          type="button"
          class="rec-list-item"
          :class="{ active: selectedRecording && selectedRecording.id === rec.id }"
          @click="selectedRecordingId = rec.id"
        >
          <div class="rec-item-top">
            <span class="rec-item-title">{{ rec.title }}</span>
            <span v-if="rec.is_medical" class="mini-medical-badge" title="Clinical Consultation">
              <Stethoscope size="12" />
            </span>
          </div>
          <div class="rec-item-meta">
            <span>{{ formatDate(rec.created_at) }}</span>
            <span v-if="rec.duration" class="font-tabular">· {{ formatDuration(rec.duration) }}</span>
          </div>
          <div class="rec-item-tags" v-if="rec.tags && rec.tags.length > 0">
            <span
              v-for="t in rec.tags"
              :key="t.id"
              class="mini-tag-chip"
            >
              <span class="mini-tag-dot" :style="{ backgroundColor: t.color || '#3b82f6' }"></span>
              <span>{{ t.name }}</span>
            </span>
          </div>
        </button>
      </aside>

      <!-- Selected Recording Summary Reader -->
      <section class="recording-detail-pane" v-if="selectedRecording">
        <div class="detail-top-bar no-print">
          <div class="detail-meta-block">
            <h2 class="detail-title">{{ selectedRecording.title }}</h2>
            <div class="detail-meta-row">
              <span class="meta-pill">
                <Calendar size="13" />
                <span>{{ formatLongDate(selectedRecording.created_at) }}</span>
              </span>
              <span class="meta-pill" v-if="selectedRecording.duration">
                <Clock size="13" />
                <span class="font-tabular">{{ formatDuration(selectedRecording.duration) }}</span>
              </span>
              <span class="meta-pill medical-pill" v-if="selectedRecording.is_medical">
                <Stethoscope size="13" />
                <span>Clinical Consultation</span>
              </span>
            </div>
          </div>

          <div class="detail-actions" v-if="selectedRecording.summary_md">
            <button type="button" class="apple-btn-secondary" @click="copySummary">
              <Check size="14" v-if="copied" class="text-green" />
              <Copy size="14" v-else />
              <span>{{ copied ? 'Copied!' : 'Copy Summary' }}</span>
            </button>
            <button type="button" class="apple-btn-secondary" @click="printSummary" v-if="!isIosApp">
              <Printer size="14" />
              <span>Print / Save PDF</span>
            </button>
          </div>
        </div>

        <div class="apple-doc-card" v-if="selectedRecording.summary_md">
          <div class="doc-header">
            <div class="doc-header-left">
              <div class="doc-icon-badge" :class="{ 'medical-icon-badge': selectedRecording.is_medical }">
                <Stethoscope v-if="selectedRecording.is_medical" size="18" class="text-emerald-600" />
                <Sparkles v-else size="18" color="#0071E3" />
              </div>
              <div>
                <h3 class="doc-heading">{{ selectedRecording.is_medical ? 'Clinical SOAP Note' : 'Executive Summary' }}</h3>
                <span class="doc-subheading">{{ selectedRecording.is_medical ? 'Structured Clinical Consultation Summary' : 'Read-Only Shared Summary' }}</span>
              </div>
            </div>
          </div>
          <div class="markdown-content" v-html="parsedSummary"></div>
        </div>

        <div v-else class="apple-doc-card empty-card">
          <p>No summary generated yet for this recording.</p>
        </div>
      </section>
    </main>
  </div>

  <!-- Error State -->
  <div v-else-if="loadError" class="full-page-loader">
    <div class="apple-doc-card error-card">
      <h3 class="error-title">Access Error</h3>
      <p class="error-desc">{{ loadError }}</p>
      <button type="button" class="apple-primary-btn" @click="fetchSharedTag">
        Try Again
      </button>
    </div>
  </div>

  <!-- Loading State -->
  <div v-else class="full-page-loader">
    <Loader2 class="animate-spin text-blue" size="44" />
    <span class="loader-label">Loading shared collection...</span>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Capacitor } from '@capacitor/core'
import axios from 'axios'
import moment from 'moment'
import { marked } from 'marked'
import { useToast } from 'primevue/usetoast'
import { useTitle } from '@vueuse/core'
import {
  Sparkles,
  Calendar,
  Clock,
  Loader2,
  Printer,
  Copy,
  Check,
  Stethoscope,
  Search,
  FileText,
} from '@lucide/vue'

const isIosApp = computed(() => {
  return Capacitor.isNativePlatform() && Capacitor.getPlatform() === 'ios'
})

const route = useRoute()
const toast = useToast()

const sharedTag = ref(null)
const subtags = ref([])
const recordings = ref([])
const loadError = ref(null)
const searchQuery = ref('')
const selectedSubtagId = ref(null)
const selectedRecordingId = ref(null)
const copied = ref(false)

const pageTitle = computed(() =>
  sharedTag.value ? `${sharedTag.value.name} — ClarifAi (Shared)` : 'ClarifAi Shared Collection'
)
useTitle(pageTitle)

const fetchSharedTag = async () => {
  loadError.value = null
  try {
    const res = await axios.get(`/api/tags/share/${route.params.token}`)
    sharedTag.value = res.data.tag
    subtags.value = res.data.subtags || []
    recordings.value = res.data.recordings || []
    if (recordings.value.length > 0) {
      selectedRecordingId.value = recordings.value[0].id
    }
  } catch (e) {
    console.error('Failed to load shared tag', e)
    loadError.value = e.response?.data?.detail || 'Shared tag not found or link has been revoked.'
    toast.add({ severity: 'error', summary: 'Access Error', detail: loadError.value, life: 5000 })
  }
}

onMounted(() => {
  fetchSharedTag()
})

const filteredRecordings = computed(() => {
  let list = recordings.value
  if (selectedSubtagId.value !== null) {
    list = list.filter((rec) => (rec.tags || []).some((t) => t.id === selectedSubtagId.value))
  }
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    list = list.filter(
      (rec) =>
        (rec.title || '').toLowerCase().includes(q) ||
        (rec.summary_md || '').toLowerCase().includes(q)
    )
  }
  return list
})

watch(filteredRecordings, (next) => {
  if (next.length === 0) {
    selectedRecordingId.value = null
  } else if (!next.some((r) => r.id === selectedRecordingId.value)) {
    selectedRecordingId.value = next[0].id
  }
})

const selectedRecording = computed(() => {
  return filteredRecordings.value.find((r) => r.id === selectedRecordingId.value) || null
})

const parsedSummary = computed(() => {
  return selectedRecording.value?.summary_md ? marked(selectedRecording.value.summary_md) : ''
})

const copySummary = () => {
  if (!selectedRecording.value?.summary_md) return
  navigator.clipboard.writeText(selectedRecording.value.summary_md)
  copied.value = true
  toast.add({ severity: 'success', summary: 'Copied', detail: 'Summary copied to clipboard.', life: 2500 })
  setTimeout(() => { copied.value = false }, 2500)
}

const printSummary = () => {
  window.print()
}

const formatDate = (iso) => moment(iso).format('MMM D, YYYY')
const formatLongDate = (iso) => moment(iso).format('MMMM D, YYYY')
const formatDuration = (secs) => {
  if (!secs) return '00:00'
  const mins = Math.floor(secs / 60)
  const rem = Math.floor(secs % 60)
  return `${mins}:${rem.toString().padStart(2, '0')}`
}
</script>

<style scoped>
.shared-tag-container {
  max-width: 1180px;
  margin: 0 auto;
  padding: 2.5rem 1.75rem 5rem 1.75rem;
}

.shared-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  margin-bottom: 2rem;
}

.brand-row {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--bg-secondary);
  border: 1px solid rgba(0, 0, 0, 0.08);
  padding: 0.35rem 0.85rem;
  border-radius: var(--radius-full);
  margin-bottom: 1.1rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.03);
}

.shared-logo-icon {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  background: linear-gradient(135deg, #0071E3, #AF52DE);
  display: flex;
  align-items: center;
  justify-content: center;
}

.brand-title {
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.01em;
}

.tag-title-row {
  display: inline-flex;
  align-items: center;
  gap: 0.65rem;
  margin-bottom: 0.75rem;
}

.tag-title-dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  flex-shrink: 0;
}

.page-title {
  font-size: 2.15rem;
  font-weight: 800;
  color: var(--text-primary);
  letter-spacing: -0.03em;
  margin: 0;
  line-height: 1.15;
}

.meta-pills-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  flex-wrap: wrap;
  margin-bottom: 1.25rem;
}

.meta-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  background: var(--bg-secondary);
  border: 1px solid rgba(0, 0, 0, 0.07);
  padding: 0.3rem 0.75rem;
  border-radius: var(--radius-full);
  font-size: 0.78rem;
  font-weight: 500;
  color: var(--text-secondary);
}

.pill-shared {
  background: rgba(0, 113, 227, 0.08);
  color: #0071E3;
  border-color: rgba(0, 113, 227, 0.16);
  font-weight: 600;
}

.medical-pill {
  background: rgba(16, 185, 129, 0.1);
  color: #059669;
  border-color: rgba(16, 185, 129, 0.2);
}

.shared-filter-bar {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
  max-width: 560px;
}

.search-input-wrap {
  position: relative;
  width: 100%;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 0.9rem;
  color: var(--text-muted);
  pointer-events: none;
}

.shared-search-input {
  width: 100%;
  padding: 0.6rem 2.2rem 0.6rem 2.4rem;
  border-radius: 12px;
  border: 1px solid rgba(0, 0, 0, 0.1);
  background: var(--bg-secondary);
  font-size: 0.9rem;
  color: var(--text-primary);
  outline: none;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.shared-search-input:focus {
  border-color: #0071E3;
  box-shadow: 0 0 0 3px rgba(0, 113, 227, 0.12);
}

.clear-search-btn {
  position: absolute;
  right: 0.7rem;
  border: none;
  background: transparent;
  font-size: 1.1rem;
  color: var(--text-muted);
  cursor: pointer;
}

.subtag-pills {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.45rem;
}

.subtag-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.3rem 0.7rem;
  border-radius: 999px;
  border: 1px solid rgba(0, 0, 0, 0.08);
  background: var(--bg-secondary);
  font-size: 0.78rem;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s ease;
}

.subtag-pill.active {
  background: #0071E3;
  border-color: #0071E3;
  color: #ffffff;
}

.subtag-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.subtag-pill.active .subtag-dot {
  background: #ffffff !important;
}

.shared-layout {
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  gap: 1.5rem;
  align-items: start;
}

.recording-list-pane {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  max-height: 74vh;
  overflow-y: auto;
  padding-right: 0.25rem;
}

.rec-list-item {
  text-align: left;
  width: 100%;
  padding: 0.9rem 1rem;
  border-radius: 14px;
  border: 1px solid rgba(0, 0, 0, 0.07);
  background: var(--bg-secondary);
  cursor: pointer;
  transition: all 0.15s ease;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.rec-list-item:hover {
  border-color: rgba(0, 113, 227, 0.3);
  transform: translateY(-1px);
}

.rec-list-item.active {
  border-color: #0071E3;
  background: rgba(0, 113, 227, 0.05);
  box-shadow: 0 2px 10px rgba(0, 113, 227, 0.08);
}

.rec-item-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.4rem;
}

.rec-item-title {
  font-size: 0.92rem;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.3;
}

.mini-medical-badge {
  color: #059669;
  flex-shrink: 0;
  margin-top: 2px;
}

.rec-item-meta {
  font-size: 0.76rem;
  color: var(--text-muted);
  display: flex;
  gap: 0.3rem;
}

.rec-item-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
  margin-top: 0.15rem;
}

.mini-tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.28rem;
  font-size: 0.7rem;
  padding: 0.12rem 0.45rem;
  border-radius: 999px;
  background: rgba(0, 0, 0, 0.045);
  color: var(--text-secondary);
}

.mini-tag-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.recording-detail-pane {
  min-width: 0;
}

.detail-top-bar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
  margin-bottom: 1rem;
}

.detail-title {
  font-size: 1.45rem;
  font-weight: 750;
  color: var(--text-primary);
  margin: 0 0 0.5rem 0;
  letter-spacing: -0.02em;
}

.detail-meta-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.detail-actions {
  display: flex;
  gap: 0.55rem;
  flex-wrap: wrap;
}

.apple-btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 0.9rem;
  border-radius: 10px;
  border: 1px solid rgba(0, 0, 0, 0.1);
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 0.84rem;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.apple-btn-secondary:hover {
  background: rgba(0, 0, 0, 0.04);
}

.apple-doc-card {
  background: var(--bg-secondary);
  border: 1px solid rgba(0, 0, 0, 0.07);
  border-radius: 18px;
  padding: 1.75rem 2rem;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03);
}

.empty-card {
  text-align: center;
  color: var(--text-muted);
  padding: 3rem 1.5rem;
}

.doc-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 1.15rem;
  margin-bottom: 1.25rem;
  border-bottom: 1px solid rgba(0, 0, 0, 0.07);
}

.doc-header-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.doc-icon-badge {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: rgba(0, 113, 227, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
}

.medical-icon-badge {
  background: rgba(16, 185, 129, 0.12);
}

.doc-heading {
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.doc-subheading {
  font-size: 0.78rem;
  color: var(--text-muted);
}

.markdown-content {
  color: var(--text-primary);
  font-size: 0.96rem;
  line-height: 1.7;
}

.markdown-content :deep(h1),
.markdown-content :deep(h2),
.markdown-content :deep(h3) {
  margin-top: 1.4rem;
  margin-bottom: 0.6rem;
  font-weight: 700;
  color: var(--text-primary);
}

.markdown-content :deep(p) {
  margin-bottom: 0.9rem;
}

.markdown-content :deep(ul),
.markdown-content :deep(ol) {
  padding-left: 1.35rem;
  margin-bottom: 0.9rem;
}

.full-page-loader {
  min-height: 80vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
}

.loader-label {
  font-size: 0.92rem;
  color: var(--text-secondary);
}

.error-card {
  max-width: 440px;
  margin: 0 1rem;
  text-align: center;
}

.error-title {
  font-size: 1.2rem;
  font-weight: 700;
  color: #E03131;
  margin-bottom: 0.5rem;
}

.error-desc {
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin-bottom: 1.5rem;
}

.text-green {
  color: #10b981;
}

.text-blue {
  color: #0071E3;
}

@media (max-width: 860px) {
  .shared-layout {
    grid-template-columns: 1fr;
  }

  .recording-list-pane {
    max-height: 260px;
  }

  .apple-doc-card {
    padding: 1.25rem;
  }
}

@media print {
  .no-print {
    display: none !important;
  }
  .shared-layout {
    display: block;
  }
  .apple-doc-card {
    border: none;
    box-shadow: none;
    padding: 0;
  }
}
</style>
