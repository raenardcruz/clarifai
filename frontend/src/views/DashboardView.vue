<template>
  <div class="dashboard-page">
    <!-- Apple visionOS Ambient Hero Section -->
    <div class="apple-hero-section">
      <div class="hero-glow-orb-1"></div>
      <div class="hero-glow-orb-2"></div>
      
      <div class="hero-content">
        <div class="hero-badge">
          <Sparkles size="14" class="text-blue" />
          <span>AI-Powered Speech Intelligence</span>
        </div>

        <h1 class="hero-title">
          Clear your mind.<br />
          <span class="gradient-text">Capture every insight.</span>
        </h1>
        
        <p class="hero-subtitle">
          ClarifAi transforms raw meeting audio into structured executive summaries, 
          attributed speaker dialogues, and concrete action items.
        </p>
        
        <div class="hero-actions">
          <button class="hero-primary-btn" @click="isRecordingModalOpen = true">
            <span class="live-dot"></span>
            <Mic size="16" />
            <span>Start Recording</span>
          </button>
          
          <button class="hero-secondary-btn" @click="isRecordingModalOpen = true">
            <UploadCloud size="16" />
            <span>Upload Audio File</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Apple Metric Widgets Grid -->
    <div class="stats-grid">
      <div class="stat-card" v-if="stats">
        <div class="stat-icon-wrapper bg-blue-subtle text-blue">
          <Mic size="22" stroke-width="2.2" />
        </div>
        <div class="stat-info">
          <span class="stat-label">TOTAL RECORDINGS</span>
          <div class="stat-value-row">
            <h2 class="stat-value">{{ stats.total_recordings }}</h2>
            <span class="stat-trend">Captured</span>
          </div>
        </div>
      </div>
      
      <div class="stat-card" v-if="stats">
        <div class="stat-icon-wrapper bg-purple-subtle text-purple">
          <Sparkles size="22" stroke-width="2.2" />
        </div>
        <div class="stat-info">
          <span class="stat-label">INSIGHTS GENERATED</span>
          <div class="stat-value-row">
            <h2 class="stat-value">{{ stats.insights_generated }}</h2>
            <span class="stat-trend text-purple-trend">AI Summaries</span>
          </div>
        </div>
      </div>
      
      <div class="stat-card" v-if="stats">
        <div class="stat-icon-wrapper bg-teal-subtle text-teal">
          <Clock size="22" stroke-width="2.2" />
        </div>
        <div class="stat-info">
          <span class="stat-label">HOURS SAVED</span>
          <div class="stat-value-row">
            <h2 class="stat-value">{{ stats.total_hours }}</h2>
            <span class="stat-trend">Review time</span>
          </div>
        </div>
      </div>

      <SpeechmaticsUsage type="card" />
    </div>

    <!-- Recent Recordings Section -->
    <div class="section-header">
      <div class="section-title-group">
        <h2>Recent Recordings</h2>
        <span class="section-subtitle">Your latest transcribed meetings and voice notes</span>
      </div>
      
      <div class="section-actions">
        <router-link to="/recordings" class="apple-pill-link">
          <span>View All</span>
          <ArrowRight size="14" />
        </router-link>
      </div>
    </div>

    <!-- Recordings Grid -->
    <div class="grid grid-cols-3">
      <RecordingCard 
        v-for="recording in recordings" 
        :key="recording.id" 
        :recording="recording" 
        @refresh="fetchData"
      />
      
      <div class="apple-create-card group" @click="isRecordingModalOpen = true">
        <div class="create-icon-ring">
          <Plus size="28" stroke-width="2.2" />
        </div>
        <h3>Start New Recording</h3>
        <p>Record microphone live or upload an MP3, WAV, or MP4 file</p>
      </div>
    </div>

    <!-- New Recording Modal -->
    <Teleport to="body">
      <NewRecordingModal v-if="isRecordingModalOpen" @close="isRecordingModalOpen = false" @refresh="fetchData" />
    </Teleport>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { Mic, Sparkles, Clock, Plus, UploadCloud, ArrowRight } from '@lucide/vue'
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
  max-width: 1280px;
  margin: 0 auto;
  padding: 2rem 2.5rem 3rem 2.5rem;
}

/* Apple Hero Section */
.apple-hero-section {
  position: relative;
  background: linear-gradient(145deg, #1C1C1E 0%, #2C2C2E 100%);
  border-radius: var(--radius-2xl);
  padding: 3.5rem 3.5rem;
  color: var(--white);
  margin-bottom: 2rem;
  overflow: hidden;
  box-shadow: 0 20px 40px -15px rgba(0, 0, 0, 0.3), 0 0 0 1px rgba(255, 255, 255, 0.1) inset;
}

.hero-glow-orb-1 {
  position: absolute;
  top: -120px;
  right: -80px;
  width: 450px;
  height: 450px;
  background: radial-gradient(circle, rgba(0, 113, 227, 0.45) 0%, rgba(88, 86, 214, 0.25) 50%, rgba(0, 0, 0, 0) 70%);
  filter: blur(50px);
  pointer-events: none;
}

.hero-glow-orb-2 {
  position: absolute;
  bottom: -150px;
  left: 20%;
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, rgba(175, 82, 222, 0.3) 0%, rgba(0, 0, 0, 0) 70%);
  filter: blur(60px);
  pointer-events: none;
}

.hero-content {
  position: relative;
  z-index: 2;
  max-width: 680px;
}

.hero-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.15);
  padding: 0.3rem 0.85rem;
  border-radius: var(--radius-full);
  font-size: 0.78rem;
  font-weight: 600;
  color: #EBF5FF;
  margin-bottom: 1.25rem;
}

.text-blue {
  color: #64D2FF;
}

.hero-title {
  font-size: 3rem;
  font-weight: 800;
  line-height: 1.15;
  letter-spacing: -0.035em;
  margin-bottom: 1rem;
  color: #FFFFFF;
}

.hero-subtitle {
  font-size: 1.05rem;
  line-height: 1.55;
  color: rgba(255, 255, 255, 0.75);
  margin-bottom: 2.25rem;
}

.hero-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.hero-primary-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.65rem;
  background: #FFFFFF;
  color: #1D1D1F;
  border: none;
  padding: 0.85rem 1.6rem;
  border-radius: var(--radius-full);
  font-weight: 600;
  font-size: 0.95rem;
  cursor: pointer;
  transition: all 0.2s var(--apple-ease);
  box-shadow: 0 4px 16px rgba(255, 255, 255, 0.2);
}

.hero-primary-btn:hover {
  background: #F5F5F7;
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(255, 255, 255, 0.3);
}

.hero-primary-btn:active {
  transform: scale(0.98);
}

.live-dot {
  width: 8px;
  height: 8px;
  background-color: var(--apple-red);
  border-radius: var(--radius-full);
  animation: pulse 1.5s infinite;
}

.hero-secondary-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.65rem;
  background: rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  color: #FFFFFF;
  border: 1px solid rgba(255, 255, 255, 0.2);
  padding: 0.85rem 1.6rem;
  border-radius: var(--radius-full);
  font-weight: 600;
  font-size: 0.95rem;
  cursor: pointer;
  transition: all 0.2s var(--apple-ease);
}

.hero-secondary-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: translateY(-1px);
}

/* Stats Section */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.25rem;
  margin-bottom: 2.75rem;
}

.stat-card {
  background: var(--bg-secondary);
  border-radius: var(--radius-lg);
  padding: 1.4rem;
  display: flex;
  align-items: center;
  gap: 1.2rem;
  border: 1px solid rgba(0, 0, 0, 0.06);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.03), 0 1px 2px rgba(0, 0, 0, 0.02);
  transition: all 0.28s var(--apple-ease);
}

.stat-card:hover {
  box-shadow: 0 12px 30px -4px rgba(0, 0, 0, 0.08);
  border-color: rgba(0, 113, 227, 0.25);
  transform: translateY(-2px);
}

.stat-icon-wrapper {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.bg-blue-subtle { background-color: var(--apple-blue-light); }
.text-blue { color: var(--apple-blue); }

.bg-teal-subtle { background-color: #E6F7F9; }
.text-teal { color: var(--apple-teal); }

.bg-purple-subtle { background-color: #F6EDFC; }
.text-purple { color: var(--apple-purple); }

.stat-info {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.stat-label {
  font-size: 0.68rem;
  font-weight: 700;
  color: var(--text-muted);
  letter-spacing: 0.05em;
  margin-bottom: 0.25rem;
}

.stat-value-row {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
}

.stat-value {
  font-size: 1.85rem;
  font-weight: 700;
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
  line-height: 1;
  letter-spacing: -0.03em;
}

.stat-trend {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-muted);
}

.text-purple-trend {
  color: var(--apple-purple);
}

/* Section Header */
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 1.5rem;
}

.section-title-group h2 {
  font-size: 1.45rem;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.02em;
  margin-bottom: 0.2rem;
}

.section-subtitle {
  font-size: 0.85rem;
  color: var(--text-muted);
}

.apple-pill-link {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.45rem 0.95rem;
  border-radius: var(--radius-full);
  background: rgba(0, 113, 227, 0.08);
  color: var(--apple-blue);
  font-weight: 600;
  font-size: 0.825rem;
  transition: var(--transition);
  text-decoration: none;
}

.apple-pill-link:hover {
  background: rgba(0, 113, 227, 0.15);
  transform: translateX(2px);
}

/* Create Card */
.apple-create-card {
  border: 2px dashed rgba(0, 0, 0, 0.1);
  border-radius: var(--radius-lg);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2.25rem 1.5rem;
  text-align: center;
  cursor: pointer;
  transition: all 0.28s var(--apple-ease);
  background: rgba(0, 0, 0, 0.015);
  min-height: 250px;
}

.apple-create-card:hover {
  border-color: var(--apple-blue);
  background: rgba(0, 113, 227, 0.03);
  transform: translateY(-2px);
}

.create-icon-ring {
  width: 56px;
  height: 56px;
  border-radius: var(--radius-full);
  background: var(--white);
  border: 1px solid rgba(0, 0, 0, 0.06);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
  color: var(--apple-blue);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.04);
  transition: var(--transition);
}

.apple-create-card:hover .create-icon-ring {
  background: var(--apple-blue);
  color: #FFFFFF;
  box-shadow: 0 6px 18px rgba(0, 113, 227, 0.3);
  transform: scale(1.08);
}

.apple-create-card h3 {
  font-size: 1.05rem;
  font-weight: 600;
  margin-bottom: 0.4rem;
  color: var(--text-primary);
}

.apple-create-card p {
  font-size: 0.825rem;
  color: var(--text-muted);
  max-width: 240px;
  line-height: 1.45;
}

@media (max-width: 992px) {
  .apple-hero-section {
    padding: 2.5rem 2rem;
  }
  
  .hero-title {
    font-size: 2.25rem;
  }
}

@media (max-width: 768px) {
  .dashboard-page {
    padding: 1.25rem;
  }

  .apple-hero-section {
    padding: 2rem 1.5rem;
    margin-bottom: 1.5rem;
  }
  
  .hero-title {
    font-size: 1.85rem;
  }
  
  .hero-subtitle {
    font-size: 0.95rem;
    margin-bottom: 1.5rem;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
    gap: 1rem;
    margin-bottom: 2rem;
  }
}
</style>
