<template>
  <div class="list-skeleton-card" aria-hidden="true">
    <div class="list-skeleton-header">
      <span>RECORDING TITLE</span>
      <span>STATUS</span>
      <span class="hide-sm">DURATION</span>
      <span class="hide-md">CREATED DATE</span>
      <span class="text-right">ACTION</span>
    </div>

    <div class="list-skeleton-rows">
      <div v-for="n in rows" :key="n" class="list-skeleton-row">
        <div class="skel-title-cell">
          <span class="skeleton-shimmer skeleton-circle skel-row-icon"></span>
          <div class="skel-row-text">
            <span class="skeleton-shimmer skel-row-title" :style="{ width: titleWidths[(n - 1) % titleWidths.length] }"></span>
            <span class="skeleton-shimmer skel-row-sub" :style="{ width: subWidths[(n - 1) % subWidths.length] }"></span>
          </div>
        </div>

        <div>
          <span class="skeleton-shimmer skeleton-pill skel-row-status"></span>
        </div>

        <div class="hide-sm">
          <span class="skeleton-shimmer skel-row-dur"></span>
        </div>

        <div class="hide-md">
          <span class="skeleton-shimmer skel-row-date"></span>
        </div>

        <div class="skel-action-cell">
          <span class="skeleton-shimmer skeleton-circle skel-row-chevron"></span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  rows: {
    type: Number,
    default: 6
  }
})

const titleWidths = ['68%', '54%', '76%', '62%', '72%', '58%']
const subWidths = ['86%', '72%', '80%', '64%', '88%', '70%']
</script>

<style scoped>
.list-skeleton-card {
  background: var(--bg-secondary);
  border-radius: var(--radius-lg);
  border: 1px solid rgba(0, 0, 0, 0.06);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.03), 0 1px 2px rgba(0, 0, 0, 0.02);
  overflow: hidden;
}

.list-skeleton-header {
  display: grid;
  grid-template-columns: minmax(220px, 2.5fr) 130px 110px 130px 64px;
  align-items: center;
  padding: 0.85rem 1.25rem;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  background: rgba(0, 0, 0, 0.015);
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  color: var(--text-muted);
}

.text-right {
  text-align: right;
}

.list-skeleton-rows {
  display: flex;
  flex-direction: column;
}

.list-skeleton-row {
  display: grid;
  grid-template-columns: minmax(220px, 2.5fr) 130px 110px 130px 64px;
  align-items: center;
  padding: 0.95rem 1.25rem;
  border-bottom: 1px solid rgba(0, 0, 0, 0.045);
  gap: 0.75rem;
}

.list-skeleton-row:last-child {
  border-bottom: none;
}

.skel-title-cell {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  min-width: 0;
}

.skel-row-icon {
  width: 34px;
  height: 34px;
  flex-shrink: 0;
}

.skel-row-text {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  flex: 1;
  min-width: 0;
}

.skel-row-title {
  height: 13px;
  border-radius: 5px;
}

.skel-row-sub {
  height: 10px;
  border-radius: 4px;
}

.skel-row-status {
  width: 82px;
  height: 22px;
}

.skel-row-dur {
  width: 46px;
  height: 12px;
}

.skel-row-date {
  width: 82px;
  height: 12px;
}

.skel-action-cell {
  display: flex;
  justify-content: flex-end;
}

.skel-row-chevron {
  width: 26px;
  height: 26px;
}

@media (max-width: 900px) {
  .list-skeleton-header,
  .list-skeleton-row {
    grid-template-columns: minmax(180px, 2fr) 110px 90px 48px;
  }

  .hide-md {
    display: none;
  }
}

@media (max-width: 640px) {
  .list-skeleton-header,
  .list-skeleton-row {
    grid-template-columns: 1fr 96px 40px;
  }

  .hide-sm {
    display: none;
  }
}
</style>
