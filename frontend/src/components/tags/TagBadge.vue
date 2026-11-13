<template>
  <span
    class="tag-badge"
    :class="[`tag-badge-${size}`, { 'is-clickable': clickable }]"
    :style="badgeStyle"
    :title="tag.path || tag.name"
    @click="handleClick"
  >
    <span class="tag-dot" :style="{ backgroundColor: tag.color || '#3b82f6' }"></span>
    <span class="tag-label">{{ displayLabel }}</span>
    <button
      v-if="removable"
      type="button"
      class="tag-remove-btn"
      @click.stop="$emit('remove', tag)"
      aria-label="Remove tag"
    >
      <X :size="size === 'sm' ? 11 : 13" />
    </button>
  </span>
</template>

<script setup>
import { computed } from 'vue'
import { X } from '@lucide/vue'

const props = defineProps({
  tag: { type: Object, required: true },
  showPath: { type: Boolean, default: false },
  removable: { type: Boolean, default: false },
  clickable: { type: Boolean, default: true },
  size: { type: String, default: 'sm' },
})

const emit = defineEmits(['click', 'remove'])

const displayLabel = computed(() => {
  if (props.showPath && props.tag.path) {
    return props.tag.path.replace(/\s*\/\s*/g, ' › ')
  }
  return props.tag.name
})

const badgeStyle = computed(() => {
  const hex = props.tag.color || '#3b82f6'
  return {
    '--tag-color': hex,
    backgroundColor: `${hex}14`,
    borderColor: `${hex}38`,
    color: '#1e293b',
  }
})

const handleClick = (e) => {
  if (!props.clickable) return
  e.stopPropagation()
  emit('click', props.tag, e)
}
</script>

<style scoped>
.tag-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  border: 1px solid;
  border-radius: 9999px;
  font-weight: 500;
  line-height: 1;
  white-space: nowrap;
  max-width: 100%;
  transition: all 0.16s ease;
}

.tag-badge-sm {
  font-size: 0.72rem;
  padding: 0.22rem 0.52rem;
}

.tag-badge-md {
  font-size: 0.8rem;
  padding: 0.32rem 0.68rem;
}

.tag-badge.is-clickable {
  cursor: pointer;
}

.tag-badge.is-clickable:hover {
  filter: brightness(0.95);
  transform: translateY(-0.5px);
}

.tag-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.tag-label {
  overflow: hidden;
  text-overflow: ellipsis;
}

.tag-remove-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  padding: 0;
  margin-left: 0.1rem;
  color: #64748b;
  cursor: pointer;
  border-radius: 50%;
  transition: color 0.15s ease, background-color 0.15s ease;
}

.tag-remove-btn:hover {
  color: #0f172a;
  background-color: rgba(0, 0, 0, 0.08);
}
</style>
