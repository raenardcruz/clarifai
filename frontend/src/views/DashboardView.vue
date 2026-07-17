<template>
  <div class="dashboard-page">
    <div class="header-section">
      <h1 class="page-title">Clear your mind.</h1>
      <p class="page-subtitle">Capture every thought. ClarifAi transforms your raw audio into structured summaries and actionable tasks instantly.</p>
      
      <div class="hero-actions">
        <button class="btn btn-primary" @click="isRecordingModalOpen = true">Start Recording</button>
      </div>
    </div>

    <!-- Stats Section -->
    <div class="stats-grid">
      <div class="stat-card" v-if="stats">
        <div class="stat-icon bg-indigo-100 text-indigo-600"><Mic size="24" /></div>
        <div class="stat-info">
          <p class="stat-label">TOTAL RECORDINGS</p>
          <h2 class="stat-value">{{ stats.total_recordings }}</h2>
        </div>
      </div>
      
      <div class="stat-card" v-if="stats">
        <div class="stat-icon bg-teal-100 text-teal-600"><Sparkles size="24" /></div>
        <div class="stat-info">
          <p class="stat-label">INSIGHTS GENERATED</p>
          <h2 class="stat-value">{{ stats.insights_generated }}</h2>
        </div>
      </div>
      
      <div class="stat-card" v-if="stats">
        <div class="stat-icon bg-purple-100 text-purple-600"><Clock size="24" /></div>
        <div class="stat-info">
          <p class="stat-label">HOURS SAVED</p>
          <h2 class="stat-value">{{ stats.total_hours }}</h2>
        </div>
      </div>

      <SpeechmaticsUsage type="card" />
    </div>

    <div class="section-header">
      <h2>Recent Recordings</h2>
      <div class="section-actions">
        <router-link to="/recordings" class="btn btn-outline">View All</router-link>
      </div>
    </div>

    <div class="grid grid-cols-3">
      <RecordingCard 
        v-for="recording in recordings" 
        :key="recording.id" 
        :recording="recording" 
        @refresh="fetchData"
      />
      
      <div class="create-card" @click="isRecordingModalOpen = true">
        <div class="create-icon">
          <Plus size="32" />
        </div>
        <h3>Start New Recording</h3>
        <p>Capture audio or upload a file</p>
      </div>
    </div>

    <Teleport to="body">
      <NewRecordingModal v-if="isRecordingModalOpen" @close="isRecordingModalOpen = false" @refresh="fetchData" />
    </Teleport>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { Mic, Sparkles, Clock, Plus } from '@lucide/vue'
import RecordingCard from '../components/RecordingCard.vue'
import NewRecordingModal from '../components/NewRecordingModal.vue'
import SpeechmaticsUsage from '../components/SpeechmaticsUsage.vue'

const recordings = ref([])
const stats = ref(null)
const isRecordingModalOpen = ref(false)

const fetchData = async () => {
  try {
    const res = await axios.get('/api/recordings/recent')
    recordings.value = res.data.recordings
    stats.value = res.data.statistics
  } catch (e) {
    console.error("Failed to fetch dashboard data", e)
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.dashboard-page {
  max-width: 1200px;
  margin: 0 auto;
}

.header-section {
  background: linear-gradient(135deg, var(--primary), var(--tertiary));
  border-radius: var(--radius-xl);
  padding: 3rem 4rem;
  color: var(--white);
  margin-bottom: 2rem;
  position: relative;
  overflow: hidden;
  box-shadow: 0 20px 25px -5px rgba(79, 70, 229, 0.4), 0 10px 10px -5px rgba(79, 70, 229, 0.2);
}

.page-title {
  font-size: 3rem;
  margin-bottom: 1rem;
  color: var(--white);
}

.page-subtitle {
  font-size: 1.1rem;
  opacity: 0.9;
  max-width: 600px;
  margin-bottom: 2rem;
}

.hero-actions {
  display: flex;
  gap: 1rem;
}

.hero-actions .btn-primary {
  background: var(--white);
  color: var(--primary);
}

.hero-actions .btn-primary:hover {
  background: var(--neutral-50);
  transform: translateY(-2px);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.5rem;
  margin-bottom: 3rem;
}

.stat-card {
  background: var(--bg-secondary);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  display: flex;
  align-items: center;
  gap: 1.5rem;
  border: 1px solid var(--neutral-200);
  box-shadow: var(--shadow-sm);
  transition: var(--transition);
}

.stat-card:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
}

.bg-indigo-100 { background-color: #E0E7FF; }
.text-indigo-600 { color: #4F46E5; }
.bg-teal-100 { background-color: #CCFBF1; }
.text-teal-600 { color: #0D9488; }
.bg-purple-100 { background-color: #F3E8FF; }
.text-purple-600 { color: #9333EA; }

.stat-label {
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--text-muted);
  letter-spacing: 0.05em;
  margin-bottom: 0.25rem;
}

.stat-value {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.create-card {
  border: 2px dashed var(--neutral-200);
  border-radius: var(--radius-lg);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  text-align: center;
  cursor: pointer;
  transition: var(--transition);
  background: var(--neutral-50);
  min-height: 250px;
}

.create-card:hover {
  border-color: var(--primary);
  background: var(--white);
}

.create-icon {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: var(--white);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
  color: var(--primary);
  box-shadow: var(--shadow-sm);
}

.create-card h3 {
  font-size: 1.1rem;
  margin-bottom: 0.5rem;
}

.create-card p {
  font-size: 0.875rem;
  color: var(--text-muted);
}

@media (max-width: 992px) {
  .header-section {
    padding: 2.5rem 3rem;
  }
  
  .page-title {
    font-size: 2.5rem;
  }
}

@media (max-width: 768px) {
  .header-section {
    padding: 2rem 1.5rem;
    margin-bottom: 1.5rem;
  }
  
  .page-title {
    font-size: 2rem;
  }
  
  .page-subtitle {
    font-size: 1rem;
    margin-bottom: 1.5rem;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
    gap: 1rem;
    margin-bottom: 2rem;
  }
  
  .stat-card {
    padding: 1.25rem;
    gap: 1rem;
  }
  
  .stat-icon {
    width: 48px;
    height: 48px;
  }
}
</style>
