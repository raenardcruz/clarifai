<template>
  <div
    class="app-loading-screen"
    :class="{ 'is-fullscreen': fullscreen, 'is-dark': dark }"
    role="status"
    aria-live="polite"
  >
    <div class="loader-card">
      <div class="brand-emblem-wrap">
        <div class="brand-halo"></div>
        <div class="brand-squircle">
          <Sparkles size="24" color="#FFFFFF" stroke-width="2.4" />
        </div>
      </div>

      <div class="loader-waveform" aria-hidden="true">
        <span class="wave-bar" style="--i: 0"></span>
        <span class="wave-bar" style="--i: 1"></span>
        <span class="wave-bar" style="--i: 2"></span>
        <span class="wave-bar" style="--i: 3"></span>
        <span class="wave-bar" style="--i: 4"></span>
      </div>

      <h3 v-if="title" class="loader-title">{{ title }}</h3>
      <p v-if="subtitle" class="loader-subtitle">{{ subtitle }}</p>

      <div class="loader-track" aria-hidden="true">
        <div class="loader-track-runner"></div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { Sparkles } from '@lucide/vue'

defineProps({
  title: {
    type: String,
    default: 'ClarifAi'
  },
  subtitle: {
    type: String,
    default: 'Preparing your workspace...'
  },
  fullscreen: {
    type: Boolean,
    default: false
  },
  dark: {
    type: Boolean,
    default: false
  }
})
</script>

<style scoped>
.app-loading-screen {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 3rem 1.5rem;
  min-height: 320px;
  width: 100%;
}

.app-loading-screen.is-fullscreen {
  position: fixed;
  inset: 0;
  min-height: 100vh;
  min-height: 100dvh;
  background: var(--bg-primary, #F5F5F7);
  z-index: 9000;
}

.app-loading-screen.is-dark {
  background: #000000;
}

.loader-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  max-width: 300px;
  width: 100%;
}

.brand-emblem-wrap {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.35rem;
}

.brand-halo {
  position: absolute;
  width: 78px;
  height: 78px;
  border-radius: var(--radius-full, 9999px);
  background: radial-gradient(circle, rgba(0, 113, 227, 0.24) 0%, rgba(175, 82, 222, 0.14) 50%, transparent 72%);
  animation: loaderHaloPulse 2.2s ease-in-out infinite;
}

.brand-squircle {
  position: relative;
  width: 52px;
  height: 52px;
  border-radius: 15px;
  background: linear-gradient(135deg, #0071E3 0%, #5856D6 52%, #AF52DE 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 10px 24px -4px rgba(0, 113, 227, 0.35), 0 0 0 1px rgba(255, 255, 255, 0.35) inset;
}

.loader-waveform {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  height: 22px;
  margin-bottom: 0.9rem;
}

.wave-bar {
  width: 3.5px;
  height: 8px;
  border-radius: var(--radius-full, 9999px);
  background: linear-gradient(180deg, #0071E3 0%, #AF52DE 100%);
  animation: loaderWavePulse 1.15s cubic-bezier(0.4, 0, 0.2, 1) infinite;
  animation-delay: calc(var(--i) * 0.12s);
}

.loader-title {
  font-family: var(--font-headline);
  font-size: 1.05rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--text-primary, #1D1D1F);
  margin: 0 0 0.25rem 0;
}

.is-dark .loader-title {
  color: #FFFFFF;
}

.loader-subtitle {
  font-size: 0.84rem;
  color: var(--text-muted, #86868B);
  margin: 0 0 1.15rem 0;
  line-height: 1.4;
}

.is-dark .loader-subtitle {
  color: rgba(255, 255, 255, 0.65);
}

.loader-track {
  width: 140px;
  height: 4px;
  border-radius: var(--radius-full, 9999px);
  background: rgba(0, 0, 0, 0.08);
  overflow: hidden;
  position: relative;
}

.is-dark .loader-track {
  background: rgba(255, 255, 255, 0.14);
}

.loader-track-runner {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 48%;
  border-radius: var(--radius-full, 9999px);
  background: linear-gradient(90deg, #0071E3 0%, #AF52DE 100%);
  animation: loaderTrackSlide 1.35s cubic-bezier(0.4, 0, 0.2, 1) infinite;
}

@keyframes loaderHaloPulse {
  0%, 100% {
    transform: scale(0.92);
    opacity: 0.65;
  }
  50% {
    transform: scale(1.12);
    opacity: 1;
  }
}

@keyframes loaderWavePulse {
  0%, 100% {
    height: 7px;
    opacity: 0.45;
  }
  50% {
    height: 20px;
    opacity: 1;
  }
}

@keyframes loaderTrackSlide {
  0% {
    transform: translateX(-110%);
  }
  100% {
    transform: translateX(230%);
  }
}

@media (prefers-reduced-motion: reduce) {
  .brand-halo,
  .wave-bar,
  .loader-track-runner {
    animation: none;
  }

  .wave-bar {
    height: 12px;
    opacity: 0.7;
  }

  .loader-track-runner {
    width: 100%;
    opacity: 0.55;
  }
}
</style>
