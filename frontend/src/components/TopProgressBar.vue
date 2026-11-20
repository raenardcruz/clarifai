<template>
  <div
    class="top-progress-bar no-print"
    :class="{ 'is-visible': loadingStore.visible }"
    role="progressbar"
    aria-label="Page loading progress"
    :aria-valuenow="Math.round(loadingStore.progress)"
    aria-valuemin="0"
    aria-valuemax="100"
  >
    <div
      class="top-progress-fill"
      :style="{ transform: `scaleX(${Math.min(100, Math.max(0, loadingStore.progress)) / 100})` }"
    >
      <div class="top-progress-glow"></div>
    </div>
  </div>
</template>

<script setup>
import { useLoadingStore } from '../stores/loading'

const loadingStore = useLoadingStore()
</script>

<style scoped>
.top-progress-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 2.5px;
  z-index: 9999;
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.22s ease;
}

.top-progress-bar.is-visible {
  opacity: 1;
}

.top-progress-fill {
  position: relative;
  width: 100%;
  height: 100%;
  transform-origin: left center;
  background: linear-gradient(90deg, #0071E3 0%, #5856D6 55%, #AF52DE 100%);
  transition: transform 0.22s cubic-bezier(0.16, 1, 0.3, 1);
  will-change: transform;
}

.top-progress-glow {
  position: absolute;
  right: 0;
  top: 0;
  width: 60px;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.75));
  box-shadow: 0 0 10px rgba(0, 113, 227, 0.55), 0 0 4px rgba(175, 82, 222, 0.45);
}

@media (prefers-reduced-motion: reduce) {
  .top-progress-fill {
    transition: none;
  }
}
</style>
