<template>
  <div class="recording-card-skeleton" aria-hidden="true">
    <!-- Header: status pill + menu placeholder -->
    <div class="skel-card-header">
      <div class="skel-badges">
        <span class="skeleton-shimmer skeleton-pill skel-status-pill" :style="{ width: statusWidth }"></span>
        <span v-if="showSecondaryBadge" class="skeleton-shimmer skeleton-pill skel-source-pill"></span>
      </div>
      <span class="skeleton-shimmer skeleton-circle skel-menu-dot"></span>
    </div>

    <!-- Body: title, mini soundwave, takeaway callout -->
    <div class="skel-card-body">
      <div class="skel-title-group">
        <span class="skeleton-shimmer skel-title-line" :style="{ width: titleWidth1 }"></span>
        <span class="skeleton-shimmer skel-title-line skel-title-line-short" :style="{ width: titleWidth2 }"></span>
      </div>

      <div class="skel-soundwave">
        <span
          v-for="(h, idx) in barHeights"
          :key="idx"
          class="skel-wave-bar skeleton-shimmer-subtle"
          :style="{ height: `${h}px`, animationDelay: `${idx * 0.04}s` }"
        ></span>
      </div>

      <div class="skel-insight-box">
        <div class="skel-insight-header">
          <span class="skeleton-shimmer-subtle skeleton-circle skel-sparkle-dot"></span>
          <span class="skeleton-shimmer-subtle skel-insight-label"></span>
        </div>
        <span class="skeleton-shimmer skel-insight-line" style="width: 96%"></span>
        <span class="skeleton-shimmer skel-insight-line" :style="{ width: insightWidth2 }"></span>
      </div>
    </div>

    <!-- Footer: metadata & arrow -->
    <div class="skel-card-footer">
      <div class="skel-meta-group">
        <span class="skeleton-shimmer skel-meta-date"></span>
        <span class="skel-meta-dot">•</span>
        <span class="skeleton-shimmer skel-meta-dur"></span>
      </div>
      <span class="skeleton-shimmer skeleton-circle skel-arrow"></span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  variant: {
    type: Number,
    default: 0
  }
})

const barHeights = [4, 10, 16, 8, 14, 18, 11, 7, 15, 12, 9, 17, 14, 8, 11, 5]

const titleWidths1 = ['86%', '74%', '92%', '80%', '88%', '70%']
const titleWidths2 = ['54%', '46%', '62%', '40%', '58%', '50%']
const statusWidths = ['92px', '78px', '86px', '96px', '82px', '88px']
const insightWidths2 = ['72%', '64%', '82%', '68%', '78%', '60%']

const idx = computed(() => Math.abs(props.variant) % titleWidths1.length)
const titleWidth1 = computed(() => titleWidths1[idx.value])
const titleWidth2 = computed(() => titleWidths2[idx.value])
const statusWidth = computed(() => statusWidths[idx.value])
const insightWidth2 = computed(() => insightWidths2[idx.value])
const showSecondaryBadge = computed(() => idx.value % 2 === 0)
</script>

<style scoped>
.recording-card-skeleton {
  display: flex;
  flex-direction: column;
  background: var(--bg-secondary);
  border-radius: var(--radius-lg);
  padding: 1.4rem;
  border: 1px solid rgba(0, 0, 0, 0.06);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.03), 0 1px 2px rgba(0, 0, 0, 0.02);
  min-height: 256px;
}

.skel-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.9rem;
}

.skel-badges {
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.skel-status-pill {
  height: 22px;
}

.skel-source-pill {
  width: 58px;
  height: 20px;
}

.skel-menu-dot {
  width: 24px;
  height: 24px;
}

.skel-card-body {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.skel-title-group {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  margin-bottom: 0.75rem;
}

.skel-title-line {
  height: 15px;
  border-radius: 5px;
}

.skel-title-line-short {
  height: 14px;
}

.skel-soundwave {
  display: flex;
  align-items: center;
  gap: 2.5px;
  height: 20px;
  margin-bottom: 0.9rem;
}

.skel-wave-bar {
  width: 2.5px;
  border-radius: var(--radius-full);
}

.skel-insight-box {
  background: rgba(0, 113, 227, 0.035);
  border: 1px solid rgba(0, 113, 227, 0.07);
  border-left: 3px solid rgba(0, 113, 227, 0.22);
  padding: 0.75rem 0.85rem;
  border-radius: var(--radius-md);
  margin-top: auto;
  display: flex;
  flex-direction: column;
  gap: 0.42rem;
}

.skel-insight-header {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  margin-bottom: 0.1rem;
}

.skel-sparkle-dot {
  width: 12px;
  height: 12px;
}

.skel-insight-label {
  width: 78px;
  height: 9px;
  border-radius: 4px;
}

.skel-insight-line {
  height: 10px;
  border-radius: 4px;
}

.skel-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 1.15rem;
  padding-top: 0.85rem;
  border-top: 1px solid rgba(0, 0, 0, 0.05);
}

.skel-meta-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.skel-meta-date {
  width: 74px;
  height: 11px;
  border-radius: 4px;
}

.skel-meta-dot {
  color: rgba(0, 0, 0, 0.14);
  font-size: 0.75rem;
}

.skel-meta-dur {
  width: 40px;
  height: 11px;
  border-radius: 4px;
}

.skel-arrow {
  width: 22px;
  height: 22px;
}
</style>
