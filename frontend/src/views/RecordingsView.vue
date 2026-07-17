<template>
  <div class="recordings-page-wrapper">
    <div class="recordings-page">
      <div class="page-header flex justify-between items-center mb-8">
        <div>
          <h1 class="text-2xl font-bold mb-1">All Recordings</h1>
          <p class="text-sm text-gray-500">Manage and review your captured thoughts</p>
        </div>
        
        <Button label="New Recording" icon="pi pi-plus" @click="isRecordingModalOpen = true" />
      </div>

      <div v-if="loading" class="flex justify-center items-center py-20">
        <Loader2 class="animate-spin text-primary" size="32" />
      </div>
      
      <div v-else-if="recordings.length === 0" class="empty-state">
        <div class="empty-icon"><Mic size="32" /></div>
        <h3>No recordings yet</h3>
        <p>Start capturing your thoughts to see them here.</p>
        <Button label="Create First Recording" severity="secondary" @click="isRecordingModalOpen = true" class="mt-4" />
      </div>

      <div v-else class="grid grid-cols-4">
        <RecordingCard 
          v-for="recording in recordings" 
          :key="recording.id" 
          :recording="recording" 
          @refresh="fetchRecordings"
        />
      </div>
    </div>

    <SpeechmaticsUsage v-if="!loading && recordings.length > 0" type="footer" />

    <Teleport to="body">
      <NewRecordingModal v-if="isRecordingModalOpen" @close="isRecordingModalOpen = false" @refresh="fetchRecordings" />
    </Teleport>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import axios from 'axios'
import { Loader2, Mic, Sparkles } from '@lucide/vue'
import RecordingCard from '../components/RecordingCard.vue'
import NewRecordingModal from '../components/NewRecordingModal.vue'
import Button from 'primevue/button'
import SpeechmaticsUsage from '../components/SpeechmaticsUsage.vue'

const route = useRoute()
const recordings = ref([])
const loading = ref(true)
const isRecordingModalOpen = ref(false)

const fetchRecordings = async () => {
  loading.value = true
  try {
    const res = await axios.get('/api/recordings', {
      params: { search: route.query.search || undefined }
    })
    recordings.value = res.data
  } catch (e) {
    console.error("Failed to fetch recordings", e)
  } finally {
    loading.value = false
  }
}

watch(() => route.query.search, () => {
  fetchRecordings()
})

onMounted(() => {
  fetchRecordings()
})
</script>

<style scoped>
.recordings-page-wrapper {
  display: flex;
  flex-direction: column;
  min-height: calc(100vh - var(--header-height));
  justify-content: space-between;
}

.recordings-page {
  max-width: 1400px;
  width: 100%;
  margin: 0 auto;
  padding: 2rem;
  flex: 1;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 5rem 2rem;
  background: var(--bg-secondary);
  border-radius: var(--radius-xl);
  border: 1px dashed var(--neutral-200);
  text-align: center;
}

.empty-icon {
  width: 64px;
  height: 64px;
  background: var(--neutral-100);
  color: var(--text-muted);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
}

.empty-state h3 {
  font-size: 1.25rem;
  margin-bottom: 0.5rem;
}

.empty-state p {
  color: var(--text-secondary);
}

.footer-stats {
  margin-top: 3rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--neutral-200);
}

.text-gray-500 { color: #6B7280; }
.text-2xl { font-size: 1.5rem; }
.font-bold { font-weight: 700; }
.mb-8 { margin-bottom: 2rem; }
.mb-1 { margin-bottom: 0.25rem; }
.py-20 { padding-top: 5rem; padding-bottom: 5rem; }
.text-primary { color: var(--primary); }
</style>
