<template>
  <div class="tag-tree-node">
    <div
      class="node-row"
      :class="{ 'is-selected': isSelected }"
      :style="{ paddingLeft: `${depth * 14 + 8}px` }"
      @click="onSelect($event)"
    >
      <button
        v-if="hasChildren"
        type="button"
        class="expand-btn"
        @click.stop="$emit('toggle-expand', node.id)"
        :aria-label="isExpanded ? 'Collapse' : 'Expand'"
      >
        <ChevronDown v-if="isExpanded" :size="13" />
        <ChevronRight v-else :size="13" />
      </button>
      <span v-else class="expand-spacer"></span>

      <span class="node-dot" :style="{ backgroundColor: node.color || '#3b82f6' }"></span>
      <span class="node-name" :title="node.path || node.name">{{ node.name }}</span>

      <div class="node-right">
        <span v-if="showCounts && node.total_count > 0" class="node-count">
          {{ node.total_count }}
        </span>
        <div class="node-actions" @click.stop>
          <button
            type="button"
            class="node-action-btn"
            :class="{ 'is-shared': !!node.share_token }"
            :title="node.share_token ? 'Copy shared tag link' : 'Share tag link'"
            @click="$emit('share', node)"
          >
            <Share2 :size="11" />
          </button>
          <button
            type="button"
            class="node-action-btn"
            title="Add sub-tag"
            @click="$emit('add-child', node)"
          >
            <Plus :size="12" />
          </button>
          <button
            type="button"
            class="node-action-btn"
            title="Edit tag"
            @click="$emit('edit', node)"
          >
            <Pencil :size="11" />
          </button>
        </div>
      </div>
    </div>

    <div v-if="hasChildren && isExpanded" class="node-children">
      <TagTreeNode
        v-for="child in node.children"
        :key="child.id"
        :node="child"
        :depth="depth + 1"
        :selected-ids="selectedIds"
        :expanded-ids="expandedIds"
        :show-counts="showCounts"
        @select="(payload) => $emit('select', payload)"
        @toggle-expand="(id) => $emit('toggle-expand', id)"
        @add-child="(n) => $emit('add-child', n)"
        @edit="(n) => $emit('edit', n)"
        @share="(n) => $emit('share', n)"
      />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { ChevronRight, ChevronDown, Plus, Pencil, Share2 } from '@lucide/vue'

defineOptions({ name: 'TagTreeNode' })

const props = defineProps({
  node: { type: Object, required: true },
  depth: { type: Number, default: 0 },
  selectedIds: { type: Array, default: () => [] },
  expandedIds: { type: Object, required: true }, // Set
  showCounts: { type: Boolean, default: true },
})

const emit = defineEmits(['select', 'toggle-expand', 'add-child', 'edit', 'share'])

const hasChildren = computed(() => props.node.children && props.node.children.length > 0)
const isExpanded = computed(() => props.expandedIds.has(props.node.id))
const isSelected = computed(() => props.selectedIds.includes(props.node.id))

const onSelect = (event) => {
  emit('select', { tag: props.node, event })
}
</script>

<style scoped>
.node-row {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding-top: 0.36rem;
  padding-bottom: 0.36rem;
  padding-right: 0.5rem;
  border-radius: 8px;
  cursor: pointer;
  color: #334155;
  font-size: 0.84rem;
  transition: background-color 0.14s ease, color 0.14s ease;
  user-select: none;
}

.node-row:hover {
  background-color: rgba(0, 0, 0, 0.045);
  color: #0f172a;
}

.node-row.is-selected {
  background-color: rgba(59, 130, 246, 0.12);
  color: #1d4ed8;
  font-weight: 600;
}

.expand-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border: none;
  background: transparent;
  color: #64748b;
  cursor: pointer;
  padding: 0;
  border-radius: 4px;
  flex-shrink: 0;
}

.expand-btn:hover {
  background-color: rgba(0, 0, 0, 0.07);
  color: #0f172a;
}

.expand-spacer {
  width: 16px;
  flex-shrink: 0;
}

.node-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.node-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-right {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  flex-shrink: 0;
}

.node-count {
  font-size: 0.72rem;
  font-weight: 600;
  color: #64748b;
  background: rgba(0, 0, 0, 0.05);
  padding: 0.08rem 0.4rem;
  border-radius: 9999px;
}

.node-row.is-selected .node-count {
  background: rgba(59, 130, 246, 0.18);
  color: #1d4ed8;
}

.node-actions {
  display: none;
  align-items: center;
  gap: 0.15rem;
}

.node-row:hover .node-actions {
  display: flex;
}

.node-row:hover .node-count {
  display: none;
}

.node-action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border: none;
  background: transparent;
  color: #64748b;
  border-radius: 5px;
  cursor: pointer;
  padding: 0;
}

.node-action-btn:hover {
  background: rgba(0, 0, 0, 0.08);
  color: #0f172a;
}

.node-action-btn.is-shared {
  color: #0071e3;
}
</style>
