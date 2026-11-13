<template>
  <div class="tag-picker" ref="pickerRef">
    <div class="selected-tags-row">
      <TagBadge
        v-for="tag in selectedTags"
        :key="tag.id"
        :tag="tag"
        :show-path="showPath"
        :removable="!disabled"
        :clickable="clickableBadges"
        @click="(t, e) => $emit('badge-click', t, e)"
        @remove="removeTag(tag.id)"
      />

      <button
        v-if="!disabled"
        type="button"
        class="add-tag-trigger"
        :class="{ 'is-open': isOpen }"
        @click.stop="toggleOpen"
      >
        <TagIcon :size="12" />
        <span>{{ selectedTags.length === 0 ? placeholder : 'Add tag' }}</span>
        <Plus :size="12" />
      </button>
    </div>

    <div v-if="isOpen" class="tag-popover" @click.stop>
      <div class="popover-search">
        <Search :size="14" class="search-icon" />
        <input
          ref="searchInputRef"
          v-model="searchQuery"
          type="text"
          placeholder="Search or create tag..."
          class="popover-search-input"
          @keydown.escape="isOpen = false"
        />
        <button v-if="searchQuery" type="button" class="clear-search-btn" @click="searchQuery = ''">
          <X :size="12" />
        </button>
      </div>

      <div class="popover-list" v-if="flattenedTree.length > 0">
        <label
          v-for="item in flattenedTree"
          :key="item.id"
          class="popover-tag-item"
          :style="{ paddingLeft: `${item.depth * 14 + 10}px` }"
        >
          <input
            type="checkbox"
            class="tag-checkbox"
            :checked="modelValue.includes(item.id)"
            @change="toggleTag(item.id)"
          />
          <span class="item-dot" :style="{ backgroundColor: item.color || '#3b82f6' }"></span>
          <span class="item-name">{{ item.name }}</span>
          <span v-if="item.depth > 0 && searchQuery" class="item-path-hint">{{ item.path }}</span>
        </label>
      </div>
      <div v-else class="popover-empty">
        No tags matching "{{ searchQuery }}"
      </div>

      <!-- Quick Create Section -->
      <div class="popover-create">
        <button
          v-if="!showQuickCreate"
          type="button"
          class="quick-create-toggle"
          @click="openQuickCreate"
        >
          <Plus :size="14" />
          <span>{{ searchQuery.trim() ? `Create "${searchQuery.trim()}"` : 'Create new tag' }}</span>
        </button>

        <div v-else class="quick-create-form">
          <input
            v-model="newTagName"
            type="text"
            placeholder="Tag name"
            class="quick-input"
            @keydown.enter.prevent="handleQuickCreate"
          />
          <select v-model="newTagParentId" class="quick-select">
            <option :value="null">No parent (Root level)</option>
            <option v-for="t in tagsStore.tags" :key="t.id" :value="t.id">
              Under: {{ t.path }}
            </option>
          </select>
          <div class="quick-colors-row">
            <button
              v-for="c in TAG_COLORS"
              :key="c"
              type="button"
              class="color-swatch"
              :class="{ active: newTagColor === c }"
              :style="{ backgroundColor: c }"
              @click="newTagColor = c"
            ></button>
          </div>
          <div class="quick-actions-row">
            <button type="button" class="quick-cancel-btn" @click="showQuickCreate = false">
              Cancel
            </button>
            <button
              type="button"
              class="quick-save-btn"
              :disabled="!newTagName.trim() || creating"
              @click="handleQuickCreate"
            >
              {{ creating ? 'Creating...' : 'Create & Select' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { Tag as TagIcon, Plus, Search, X } from '@lucide/vue'
import { useTagsStore, TAG_COLORS } from '../../stores/tags'
import TagBadge from './TagBadge.vue'

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  placeholder: { type: String, default: 'Add tags' },
  showPath: { type: Boolean, default: true },
  disabled: { type: Boolean, default: false },
  clickableBadges: { type: Boolean, default: false },
})

const emit = defineEmits(['update:modelValue', 'badge-click'])

const tagsStore = useTagsStore()
const pickerRef = ref(null)
const searchInputRef = ref(null)
const isOpen = ref(false)
const searchQuery = ref('')

const showQuickCreate = ref(false)
const newTagName = ref('')
const newTagParentId = ref(null)
const newTagColor = ref(TAG_COLORS[0])
const creating = ref(false)

const selectedTags = computed(() => {
  return props.modelValue
    .map((id) => tagsStore.tagMap.get(id))
    .filter(Boolean)
})

const flattenedTree = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  const result = []

  const walk = (nodes, depth) => {
    for (const node of nodes) {
      const matches = !q || node.name.toLowerCase().includes(q) || (node.path && node.path.toLowerCase().includes(q))
      if (matches) {
        result.push({ ...node, depth: q ? 0 : depth })
      }
      if (node.children && node.children.length > 0) {
        walk(node.children, depth + 1)
      }
    }
  }

  walk(tagsStore.tree, 0)
  return result
})

const toggleOpen = async () => {
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    await tagsStore.fetchTags()
    await nextTick()
    searchInputRef.value?.focus()
  } else {
    showQuickCreate.value = false
  }
}

const toggleTag = (tagId) => {
  const next = new Set(props.modelValue)
  if (next.has(tagId)) {
    next.delete(tagId)
  } else {
    next.add(tagId)
  }
  emit('update:modelValue', Array.from(next))
}

const removeTag = (tagId) => {
  emit('update:modelValue', props.modelValue.filter((id) => id !== tagId))
}

const openQuickCreate = () => {
  newTagName.value = searchQuery.value.trim()
  newTagParentId.value = null
  showQuickCreate.value = true
}

const handleQuickCreate = async () => {
  if (!newTagName.value.trim() || creating.value) return
  creating.value = true
  try {
    const created = await tagsStore.createTag({
      name: newTagName.value.trim(),
      parent_id: newTagParentId.value,
      color: newTagColor.value,
    })
    const next = Array.from(new Set([...props.modelValue, created.id]))
    emit('update:modelValue', next)
    newTagName.value = ''
    searchQuery.value = ''
    showQuickCreate.value = false
  } catch (err) {
    console.error('Quick create tag failed:', err)
  } finally {
    creating.value = false
  }
}

const handleOutsideClick = (e) => {
  if (pickerRef.value && !pickerRef.value.contains(e.target)) {
    isOpen.value = false
    showQuickCreate.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleOutsideClick)
})

onUnmounted(() => {
  document.removeEventListener('click', handleOutsideClick)
})
</script>

<style scoped>
.tag-picker {
  position: relative;
  display: inline-block;
  width: 100%;
}

.selected-tags-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.4rem;
}

.add-tag-trigger {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.28rem 0.65rem;
  border-radius: 9999px;
  border: 1px dashed #cbd5e1;
  background: rgba(255, 255, 255, 0.8);
  color: #475569;
  font-size: 0.78rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.add-tag-trigger:hover,
.add-tag-trigger.is-open {
  border-color: #3b82f6;
  color: #2563eb;
  background: rgba(59, 130, 246, 0.06);
}

.tag-popover {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  z-index: 60;
  width: 290px;
  max-width: 90vw;
  background: rgba(255, 255, 255, 0.96);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 12px;
  box-shadow: 0 12px 32px rgba(15, 23, 42, 0.14);
  overflow: hidden;
}

.popover-search {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.55rem 0.75rem;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}

.search-icon {
  color: #94a3b8;
  flex-shrink: 0;
}

.popover-search-input {
  border: none;
  outline: none;
  background: transparent;
  width: 100%;
  font-size: 0.84rem;
  color: #0f172a;
}

.clear-search-btn {
  border: none;
  background: transparent;
  color: #94a3b8;
  cursor: pointer;
  padding: 2px;
  display: flex;
}

.popover-list {
  max-height: 220px;
  overflow-y: auto;
  padding: 0.35rem 0;
}

.popover-tag-item {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.38rem 0.75rem;
  cursor: pointer;
  font-size: 0.84rem;
  color: #1e293b;
  transition: background-color 0.12s ease;
}

.popover-tag-item:hover {
  background-color: rgba(59, 130, 246, 0.08);
}

.tag-checkbox {
  accent-color: #3b82f6;
  cursor: pointer;
}

.item-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.item-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-path-hint {
  font-size: 0.7rem;
  color: #94a3b8;
  max-width: 110px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.popover-empty {
  padding: 0.9rem;
  text-align: center;
  font-size: 0.8rem;
  color: #64748b;
}

.popover-create {
  border-top: 1px solid rgba(0, 0, 0, 0.06);
  padding: 0.5rem 0.65rem;
  background: rgba(248, 250, 252, 0.8);
}

.quick-create-toggle {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.35rem 0.4rem;
  border: none;
  background: transparent;
  color: #2563eb;
  font-size: 0.82rem;
  font-weight: 600;
  cursor: pointer;
  border-radius: 6px;
}

.quick-create-toggle:hover {
  background: rgba(59, 130, 246, 0.1);
}

.quick-create-form {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.quick-input,
.quick-select {
  width: 100%;
  padding: 0.38rem 0.55rem;
  border: 1px solid #cbd5e1;
  border-radius: 7px;
  font-size: 0.82rem;
  background: #ffffff;
  color: #0f172a;
  box-sizing: border-box;
}

.quick-colors-row {
  display: flex;
  gap: 0.35rem;
  flex-wrap: wrap;
}

.color-swatch {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  padding: 0;
}

.color-swatch.active {
  border-color: #0f172a;
  transform: scale(1.1);
}

.quick-actions-row {
  display: flex;
  justify-content: flex-end;
  gap: 0.4rem;
}

.quick-cancel-btn {
  padding: 0.3rem 0.65rem;
  border: 1px solid #cbd5e1;
  background: #ffffff;
  color: #475569;
  border-radius: 6px;
  font-size: 0.78rem;
  cursor: pointer;
}

.quick-save-btn {
  padding: 0.3rem 0.7rem;
  border: none;
  background: #2563eb;
  color: #ffffff;
  border-radius: 6px;
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
}

.quick-save-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
</style>
