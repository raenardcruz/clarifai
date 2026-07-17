<template>
  <div v-if="usageData" :class="['speechmatics-usage-container', type]">
    <div v-if="type === 'card'" class="usage-card glass">
      <div class="card-header">
        <div class="icon-wrapper">
          <Activity size="20" class="text-primary animate-pulse-slow" />
        </div>
        <div class="header-info">
          <span class="label">SPEECHMATICS USAGE</span>
          <h3 class="value">{{ usageData.used_hours }} / {{ usageData.limit_hours }} hrs</h3>
        </div>
      </div>
      
      <div class="progress-section">
        <div class="progress-bar-track">
          <div 
            class="progress-bar-fill" 
            :style="{ width: `${usageData.percentage}%` }"
          >
            <div class="progress-glow"></div>
          </div>
        </div>
        <div class="progress-stats">
          <span class="percentage-label">{{ usageData.percentage }}% used</span>
          <span class="remaining-label">{{ Math.max(0, (usageData.limit_hours - usageData.used_hours).toFixed(2)) }} hrs left</span>
        </div>
      </div>
    </div>

    <div v-else class="usage-footer-wrapper">
      <div class="usage-footer-content">
        <div class="footer-left">
          <Activity size="16" class="text-primary animate-pulse-slow" />
          <span class="footer-title">Speechmatics Transcription Limit</span>
          <span class="footer-divider">|</span>
          <span class="footer-detail font-medium">{{ usageData.used_hours }} hrs of {{ usageData.limit_hours }} hrs used ({{ usageData.percentage }}%)</span>
        </div>
        <div class="footer-right">
          <div class="progress-bar-track footer-track">
            <div 
              class="progress-bar-fill" 
              :style="{ width: `${usageData.percentage}%` }"
            >
              <div class="progress-glow"></div>
            </div>
          </div>
          <span class="remaining-text">{{ Math.max(0, (usageData.limit_hours - usageData.used_hours).toFixed(2)) }} hrs remaining</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { Activity } from '@lucide/vue'

const props = defineProps({
  type: {
    type: String,
    default: 'card',
    validator: (value) => ['card', 'footer'].includes(value)
  }
})

const usageData = ref(null)

const fetchUsage = async () => {
  try {
    const res = await axios.get('/api/recordings/speechmatics-usage')
    usageData.value = res.data
  } catch (e) {
    console.error("Failed to fetch Speechmatics usage details:", e)
  }
}

onMounted(() => {
  fetchUsage()
})
</script>

<style scoped>
.speechmatics-usage-container {
  width: 100%;
}

/* Card Mode */
.usage-card {
  background: var(--bg-secondary);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  border: 1px solid var(--neutral-200);
  box-shadow: var(--shadow-sm);
  transition: var(--transition);
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  min-height: 145px;
}

.usage-card:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.icon-wrapper {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #EEF2FF;
}

.header-info {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.label {
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--text-muted);
  letter-spacing: 0.05em;
}

.value {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.2;
}

.progress-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.progress-bar-track {
  width: 100%;
  height: 8px;
  background-color: var(--neutral-200);
  border-radius: var(--radius-full);
  overflow: hidden;
  position: relative;
}

.progress-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--primary), var(--tertiary));
  border-radius: var(--radius-full);
  transition: width 0.8s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
}

.progress-glow {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: 30px;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.4));
  filter: blur(2px);
}

.progress-stats {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-secondary);
}

.percentage-label {
  color: var(--primary);
}

.remaining-label {
  color: var(--text-muted);
}

/* Footer Mode */
.usage-footer-wrapper {
  width: 100%;
  background: var(--bg-secondary);
  border-top: 1px solid var(--neutral-200);
  padding: 1rem 2rem;
  box-shadow: 0 -4px 10px -4px rgba(0, 0, 0, 0.05);
  margin-top: 3rem;
}

.usage-footer-content {
  max-width: 1400px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.footer-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.875rem;
}

.footer-title {
  font-weight: 700;
  color: var(--text-primary);
  font-family: var(--font-headline);
}

.footer-divider {
  color: var(--neutral-200);
}

.footer-detail {
  color: var(--text-secondary);
}

.footer-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.footer-track {
  width: 150px;
  height: 6px;
}

.remaining-text {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-muted);
}

@media (max-width: 768px) {
  .usage-footer-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }
  
  .footer-right {
    width: 100%;
    justify-content: space-between;
  }
  
  .footer-track {
    flex: 1;
    max-width: 200px;
  }
}

.animate-pulse-slow {
  animation: pulse 2.5s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: .5;
  }
}
</style>
