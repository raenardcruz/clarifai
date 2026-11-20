<template>
  <div class="detail-skeleton-page" :class="{ 'is-shared': shared }" aria-hidden="true">
    <!-- Top Header / Breadcrumb -->
    <div class="detail-skel-header">
      <div v-if="!shared" class="skel-breadcrumb-row">
        <span class="skeleton-shimmer skel-crumb-back"></span>
        <span class="skel-crumb-sep">/</span>
        <span class="skeleton-shimmer skel-crumb-curr"></span>
      </div>
      <div v-else class="skel-shared-brand">
        <span class="skeleton-shimmer-subtle skel-brand-icon"></span>
        <span class="skeleton-shimmer skel-brand-name"></span>
      </div>

      <div class="skel-title-row">
        <div class="skel-title-col">
          <span class="skeleton-shimmer skel-page-title"></span>
          <div class="skel-meta-pills">
            <span class="skeleton-shimmer skeleton-pill" style="width: 112px; height: 26px;"></span>
            <span class="skeleton-shimmer skeleton-pill" style="width: 78px; height: 26px;"></span>
            <span class="skeleton-shimmer skeleton-pill" style="width: 98px; height: 26px;"></span>
            <span class="skeleton-shimmer-subtle skeleton-pill" style="width: 92px; height: 26px;"></span>
          </div>
          <div v-if="!shared" class="skel-tags-bar">
            <span class="skeleton-shimmer skeleton-pill" style="width: 140px; height: 28px;"></span>
          </div>
        </div>

        <div class="skel-actions-col">
          <span class="skeleton-shimmer skeleton-pill" style="width: 82px; height: 34px;"></span>
          <span class="skeleton-shimmer skeleton-pill" style="width: 82px; height: 34px;"></span>
          <span v-if="!shared" class="skeleton-shimmer skeleton-pill" style="width: 86px; height: 34px;"></span>
        </div>
      </div>
    </div>

    <!-- Segmented Tabs Bar (authenticated view) -->
    <div v-if="!shared" class="skel-tabs-toolbar">
      <div class="skel-segmented">
        <span class="skeleton-shimmer skel-seg-active"></span>
        <span class="skeleton-shimmer skel-seg-idle"></span>
      </div>
      <div class="skel-tab-actions">
        <span class="skeleton-shimmer skeleton-pill" style="width: 112px; height: 32px;"></span>
        <span class="skeleton-shimmer skeleton-pill" style="width: 100px; height: 32px;"></span>
      </div>
    </div>

    <!-- Document Card -->
    <div class="skel-doc-card">
      <div class="skel-doc-header">
        <span class="skeleton-shimmer-subtle skel-doc-badge"></span>
        <div class="skel-doc-headings">
          <span class="skeleton-shimmer" style="width: 180px; height: 18px;"></span>
          <span class="skeleton-shimmer" style="width: 240px; height: 12px;"></span>
        </div>
      </div>

      <div class="skel-doc-divider"></div>

      <!-- Summary Section 1 -->
      <div class="skel-doc-section">
        <span class="skeleton-shimmer skel-section-heading" style="width: 34%;"></span>
        <span class="skeleton-shimmer skel-body-line" style="width: 98%;"></span>
        <span class="skeleton-shimmer skel-body-line" style="width: 94%;"></span>
        <span class="skeleton-shimmer skel-body-line" style="width: 78%;"></span>
      </div>

      <!-- Summary Bullet Section -->
      <div class="skel-doc-section">
        <span class="skeleton-shimmer skel-section-heading" style="width: 28%;"></span>
        <div class="skel-bullet-row" v-for="w in ['88%', '82%', '91%', '68%']" :key="w">
          <span class="skeleton-shimmer-subtle skeleton-circle skel-bullet-dot"></span>
          <span class="skeleton-shimmer skel-body-line" :style="{ width: w }"></span>
        </div>
      </div>

      <!-- Summary Section 3 -->
      <div class="skel-doc-section">
        <span class="skeleton-shimmer skel-section-heading" style="width: 30%;"></span>
        <span class="skeleton-shimmer skel-body-line" style="width: 95%;"></span>
        <span class="skeleton-shimmer skel-body-line" style="width: 84%;"></span>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  shared: {
    type: Boolean,
    default: false
  }
})
</script>

<style scoped>
.detail-skeleton-page {
  max-width: 1120px;
  width: 100%;
  margin: 0 auto;
  padding: 2rem 2.5rem 3rem 2.5rem;
}

.detail-skeleton-page.is-shared {
  max-width: 920px;
  padding-top: 2.75rem;
}

.detail-skel-header {
  margin-bottom: 1.75rem;
}

.skel-breadcrumb-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.skel-crumb-back {
  width: 84px;
  height: 14px;
}

.skel-crumb-sep {
  color: var(--neutral-300);
  font-size: 0.85rem;
}

.skel-crumb-curr {
  width: 160px;
  height: 14px;
}

.skel-shared-brand {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  margin-bottom: 1.25rem;
}

.skel-brand-icon {
  width: 30px;
  height: 30px;
  border-radius: 9px;
}

.skel-brand-name {
  width: 170px;
  height: 14px;
}

.skel-title-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1.5rem;
  flex-wrap: wrap;
}

.skel-title-col {
  flex: 1;
  min-width: 260px;
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.skel-page-title {
  width: min(460px, 85%);
  height: 32px;
  border-radius: 9px;
}

.skel-meta-pills {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.skel-tags-bar {
  margin-top: 0.1rem;
}

.skel-actions-col {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  flex-wrap: wrap;
}

.skel-tabs-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
  margin-bottom: 1.5rem;
}

.skel-segmented {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 4px;
  border-radius: var(--radius-md);
  background: rgba(118, 118, 128, 0.1);
}

.skel-seg-active {
  width: 152px;
  height: 32px;
  border-radius: 9px;
}

.skel-seg-idle {
  width: 136px;
  height: 32px;
  border-radius: 9px;
  opacity: 0.6;
}

.skel-tab-actions {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.skel-doc-card {
  background: var(--bg-secondary);
  border-radius: var(--radius-2xl);
  padding: 2.25rem 2.5rem;
  border: 1px solid rgba(0, 0, 0, 0.06);
  box-shadow: 0 4px 20px -2px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.02);
  display: flex;
  flex-direction: column;
  gap: 1.6rem;
}

.skel-doc-header {
  display: flex;
  align-items: center;
  gap: 0.9rem;
}

.skel-doc-badge {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  flex-shrink: 0;
}

.skel-doc-headings {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.skel-doc-divider {
  height: 1px;
  background: rgba(0, 0, 0, 0.06);
}

.skel-doc-section {
  display: flex;
  flex-direction: column;
  gap: 0.7rem;
}

.skel-section-heading {
  height: 18px;
  border-radius: 6px;
  margin-bottom: 0.25rem;
}

.skel-body-line {
  height: 13px;
  border-radius: 5px;
}

.skel-bullet-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding-left: 0.25rem;
}

.skel-bullet-dot {
  width: 7px;
  height: 7px;
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .detail-skeleton-page {
    padding: 1.25rem 1rem 2.5rem 1rem;
  }

  .skel-doc-card {
    padding: 1.5rem 1.25rem;
    border-radius: var(--radius-xl);
  }
}
</style>
