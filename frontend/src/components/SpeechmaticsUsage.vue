<template>
  <div v-if="usageData" :class="['speechmatics-usage-container', type]">
    <!-- Card Mode (Dashboard Widget) -->
    <div v-if="type === 'card'" class="apple-usage-widget">
      <div class="widget-header">
        <div class="icon-glyph-container" :class="usageSeverityClass">
          <Activity size="20" stroke-width="2.2" />
        </div>
        <div class="header-info">
          <span class="widget-label">SPEECHMATICS USAGE</span>
          <div class="value-row">
            <span class="value-highlight">{{ usageData.used_hours }}</span>
            <span class="value-divider">/</span>
            <span class="value-limit">{{ usageData.limit_hours }} hrs</span>
          </div>
        </div>
      </div>
      
      <div class="progress-section">
        <div class="apple-track">
          <div 
            class="apple-fill" 
            :class="fillGradientClass"
            :style="{ width: `${Math.min(100, usageData.percentage)}%` }"
          >
            <div class="specular-shine"></div>
          </div>
        </div>
        
        <div class="progress-footer-stats">
          <span class="percentage-pill" :class="usageSeverityClass">
            {{ usageData.percentage }}% used
          </span>
          <span class="remaining-text">
            {{ Math.max(0, (usageData.limit_hours - usageData.used_hours).toFixed(2)) }} hrs left
          </span>
        </div>
      </div>
    </div>

    <!-- Footer Mode (Bottom Dock Toolbar) -->
    <div v-else class="apple-usage-dock">
      <div class="dock-content">
        <div class="dock-left">
          <div class="dock-icon-circle">
            <Activity size="14" stroke-width="2.5" />
          </div>
          <span class="dock-title">Speechmatics Transcription Limit</span>
          <span class="dock-divider">•</span>
          <span class="dock-detail">
            <strong>{{ usageData.used_hours }}</strong> hrs of {{ usageData.limit_hours }} hrs used ({{ usageData.percentage }}%)
          </span>
        </div>
        
        <div class="dock-right">
          <div class="dock-track">
            <div 
              class="dock-fill" 
              :class="fillGradientClass"
              :style="{ width: `${Math.min(100, usageData.percentage)}%` }"
            ></div>
          </div>
          <span class="dock-remaining">
            {{ Math.max(0, (usageData.limit_hours - usageData.used_hours).toFixed(2)) }} hrs remaining
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
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

const usageSeverityClass = computed(() => {
  if (!usageData.value) return 'severity-normal'
  const p = usageData.value.percentage
  if (p >= 90) return 'severity-danger'
  if (p >= 75) return 'severity-warning'
  return 'severity-normal'
})

const fillGradientClass = computed(() => {
  if (!usageData.value) return 'fill-blue'
  const p = usageData.value.percentage
  if (p >= 90) return 'fill-red'
  if (p >= 75) return 'fill-orange'
  return 'fill-blue'
})

onMounted(() => {
  fetchUsage()
})
</script>

<style scoped>
.speechmatics-usage-container {
  width: 100%;
}

/* Apple Card Widget Mode */
.apple-usage-widget {
  background: var(--bg-secondary);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  border: 1px solid rgba(0, 0, 0, 0.06);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.03), 0 1px 2px rgba(0, 0, 0, 0.02);
  transition: all 0.28s var(--apple-ease);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 1.25rem;
  min-height: 145px;
}

.apple-usage-widget:hover {
  box-shadow: 0 12px 30px -4px rgba(0, 0, 0, 0.08);
  border-color: rgba(0, 113, 227, 0.25);
  transform: translateY(-2px);
}

.widget-header {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.icon-glyph-container {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: var(--transition);
}

.severity-normal {
  background-color: var(--apple-blue-light);
  color: var(--apple-blue);
}

.severity-warning {
  background-color: var(--apple-orange-light);
  color: var(--apple-orange);
}

.severity-danger {
  background-color: var(--apple-red-light);
  color: var(--apple-red);
}

.header-info {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.widget-label {
  font-size: 0.7rem;
  font-weight: 700;
  color: var(--text-muted);
  letter-spacing: 0.06em;
}

.value-row {
  display: flex;
  align-items: baseline;
  gap: 0.25rem;
}

.value-highlight {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
  line-height: 1.1;
  letter-spacing: -0.03em;
}

.value-divider {
  color: var(--text-muted);
  font-size: 1rem;
  font-weight: 400;
}

.value-limit {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-secondary);
}

.progress-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

/* Apple Liquid Progress Bar Track */
.apple-track {
  width: 100%;
  height: 7px;
  background-color: rgba(118, 118, 128, 0.12);
  border-radius: var(--radius-full);
  overflow: hidden;
  position: relative;
}

.apple-fill {
  height: 100%;
  border-radius: var(--radius-full);
  transition: width 0.8s cubic-bezier(0.16, 1, 0.3, 1);
  position: relative;
}

.fill-blue {
  background: linear-gradient(90deg, #0071E3 0%, #30B0C7 100%);
}

.fill-orange {
  background: linear-gradient(90deg, #FF9500 0%, #FFCC00 100%);
}

.fill-red {
  background: linear-gradient(90deg, #FF3B30 0%, #FF6482 100%);
}

.specular-shine {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.4) 0%, rgba(255, 255, 255, 0) 100%);
}

.progress-footer-stats {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.75rem;
}

.percentage-pill {
  font-weight: 600;
  padding: 0.1rem 0.5rem;
  border-radius: var(--radius-full);
  font-size: 0.7rem;
}

.remaining-text {
  color: var(--text-muted);
  font-weight: 500;
  font-variant-numeric: tabular-nums;
}

/* Footer Dock Mode */
.apple-usage-dock {
  width: 100%;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  border-top: 1px solid rgba(0, 0, 0, 0.07);
  padding: 0.85rem 2rem;
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.03);
  margin-top: 3rem;
}

.dock-content {
  max-width: 1400px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.dock-left {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  font-size: 0.84rem;
}

.dock-icon-circle {
  width: 22px;
  height: 22px;
  border-radius: var(--radius-full);
  background-color: var(--apple-blue-light);
  color: var(--apple-blue);
  display: flex;
  align-items: center;
  justify-content: center;
}

.dock-title {
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: -0.01em;
}

.dock-divider {
  color: rgba(0, 0, 0, 0.2);
}

.dock-detail {
  color: var(--text-secondary);
}

.dock-detail strong {
  color: var(--text-primary);
}

.dock-right {
  display: flex;
  align-items: center;
  gap: 0.85rem;
}

.dock-track {
  width: 140px;
  height: 6px;
  background-color: rgba(118, 118, 128, 0.14);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.dock-fill {
  height: 100%;
  border-radius: var(--radius-full);
  transition: width 0.6s ease;
}

.dock-remaining {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
}

@media (max-width: 768px) {
  .dock-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.6rem;
  }

  .dock-right {
    width: 100%;
    justify-content: space-between;
  }

  .dock-track {
    flex: 1;
    max-width: 200px;
  }
}
</style>
