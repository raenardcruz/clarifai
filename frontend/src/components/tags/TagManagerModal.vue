<template>
  <Dialog
    :visible="true"
    modal
    :header="editingTag ? 'Edit Tag' : (presetParent ? `New Sub-tag under ${presetParent.name}` : 'Manage Tag Taxonomy')"
    :style="{ width: '92vw', maxWidth: '540px' }"
    @update:visible="$emit('close')"
  >
    <div class="tag-manager-body">
      <!-- Create / Edit Form -->
      <div class="tag-form-card">
        <div class="form-row">
          <label class="form-label">Tag Name</label>
          <input
            v-model="formName"
            type="text"
            class="apple-input"
            placeholder="e.g. Cardiology, Q4 Planning, Acme Corp"
            @keydown.enter.prevent="saveForm"
          />
        </div>

        <div class="form-row">
          <label class="form-label">Parent Tag (Optional — for multi-level hierarchy)</label>
          <select v-model="formParentId" class="apple-input">
            <option :value="null">None (Root Level)</option>
            <option
              v-for="t in availableParents"
              :key="t.id"
              :value="t.id"
            >
              {{ t.path }}
            </option>
          </select>
        </div>

        <div class="form-row">
          <label class="form-label">Color</label>
          <div class="color-swatches">
            <button
              v-for="c in TAG_COLORS"
              :key="c"
              type="button"
              class="color-swatch-lg"
              :class="{ active: formColor === c }"
              :style="{ backgroundColor: c }"
              @click="formColor = c"
            ></button>
          </div>
        </div>

        <div class="form-actions">
          <button
            v-if="editingTag"
            type="button"
            class="btn-secondary"
            @click="resetForm"
          >
            Cancel Edit
          </button>
          <button
            type="button"
            class="btn-primary"
            :disabled="!formName.trim() || saving"
            @click="saveForm"
          >
            {{ saving ? 'Saving...' : (editingTag ? 'Save Changes' : 'Create Tag') }}
          </button>
        </div>
      </div>

      <!-- Existing Hierarchy List -->
      <div class="taxonomy-section" v-if="tagsStore.tags.length > 0">
        <div class="taxonomy-header">
          <span>Existing Tag Hierarchy</span>
          <span class="taxonomy-count">{{ tagsStore.tags.length }} tags</span>
        </div>

        <div class="taxonomy-list">
          <div
            v-for="item in flattenedTags"
            :key="item.id"
            class="taxonomy-item"
            :style="{ paddingLeft: `${item.depth * 18 + 10}px` }"
          >
            <span class="item-dot" :style="{ backgroundColor: item.color }"></span>
            <span class="item-name">{{ item.name }}</span>
            <span class="item-count">{{ item.total_count }}</span>
            <div class="item-actions">
              <button
                type="button"
                class="icon-btn"
                :class="{ 'is-shared': !!item.share_token }"
                :title="item.share_token ? 'Copy shared tag link' : 'Share tag link'"
                @click="shareTagLink(item)"
              >
                <Share2 :size="14" />
              </button>
              <button
                v-if="item.share_token"
                type="button"
                class="icon-btn danger"
                title="Revoke share link"
                @click="revokeShareLink(item)"
              >
                <Link2Off :size="14" />
              </button>
              <button type="button" class="icon-btn" title="Add sub-tag" @click="startAddChild(item)">
                <Plus :size="14" />
              </button>
              <button type="button" class="icon-btn" title="Edit" @click="startEdit(item)">
                <Pencil :size="14" />
              </button>
              <button type="button" class="icon-btn danger" title="Delete" @click="confirmDelete(item)">
                <Trash2 :size="14" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Dialog>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import Dialog from 'primevue/dialog'
import { useToast } from 'primevue/usetoast'
import { Plus, Pencil, Trash2, Share2, Link2Off } from '@lucide/vue'
import { useTagsStore, TAG_COLORS } from '../../stores/tags'
import { getPublicBaseUrl } from '../../services/apiConfig'

const props = defineProps({
  initialEditTag: { type: Object, default: null },
  initialParentTag: { type: Object, default: null },
})

const emit = defineEmits(['close'])

const tagsStore = useTagsStore()
const toast = useToast()

const editingTag = ref(props.initialEditTag)
const presetParent = ref(props.initialParentTag)
const formName = ref(props.initialEditTag?.name || '')
const formParentId = ref(props.initialEditTag ? props.initialEditTag.parent_id : (props.initialParentTag?.id || null))
const formColor = ref(props.initialEditTag?.color || TAG_COLORS[0])
const saving = ref(false)

onMounted(() => {
  tagsStore.fetchTags()
})

const availableParents = computed(() => {
  if (!editingTag.value) return tagsStore.tags
  const excluded = new Set(tagsStore.getDescendantIds(editingTag.value.id))
  return tagsStore.tags.filter((t) => !excluded.has(t.id))
})

const flattenedTags = computed(() => {
  const result = []
  const walk = (nodes, depth) => {
    for (const node of nodes) {
      result.push({ ...node, depth })
      if (node.children?.length) {
        walk(node.children, depth + 1)
      }
    }
  }
  walk(tagsStore.tree, 0)
  return result
})

const resetForm = () => {
  editingTag.value = null
  presetParent.value = null
  formName.value = ''
  formParentId.value = null
  formColor.value = TAG_COLORS[0]
}

const startAddChild = (tag) => {
  editingTag.value = null
  presetParent.value = tag
  formName.value = ''
  formParentId.value = tag.id
  formColor.value = tag.color || TAG_COLORS[0]
}

const shareTagLink = async (tag) => {
  try {
    const token = await tagsStore.shareTag(tag.id)
    const url = `${getPublicBaseUrl()}/share/tag/${token}`
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(url)
    }
    toast.add({
      severity: 'success',
      summary: 'Tag Share Link Copied',
      detail: `Read-only link for "${tag.name}" copied to clipboard.`,
      life: 3500,
    })
  } catch (err) {
    toast.add({
      severity: 'error',
      summary: 'Share Failed',
      detail: err.response?.data?.detail || 'Could not generate share link.',
      life: 3500,
    })
  }
}

const revokeShareLink = async (tag) => {
  try {
    await tagsStore.unshareTag(tag.id)
    toast.add({
      severity: 'info',
      summary: 'Share Link Revoked',
      detail: `Public link for "${tag.name}" has been disabled.`,
      life: 3000,
    })
  } catch (err) {
    toast.add({
      severity: 'error',
      summary: 'Error',
      detail: err.response?.data?.detail || 'Could not revoke share link.',
      life: 3500,
    })
  }
}

const startEdit = (tag) => {
  presetParent.value = null
  editingTag.value = tag
  formName.value = tag.name
  formParentId.value = tag.parent_id ?? null
  formColor.value = tag.color || TAG_COLORS[0]
}

const saveForm = async () => {
  if (!formName.value.trim() || saving.value) return
  saving.value = true
  try {
    if (editingTag.value) {
      await tagsStore.updateTag(editingTag.value.id, {
        name: formName.value.trim(),
        parent_id: formParentId.value,
        color: formColor.value,
      })
      toast.add({ severity: 'success', summary: 'Tag updated', life: 2500 })
    } else {
      await tagsStore.createTag({
        name: formName.value.trim(),
        parent_id: formParentId.value,
        color: formColor.value,
      })
      toast.add({ severity: 'success', summary: 'Tag created', life: 2500 })
    }
    resetForm()
  } catch (err) {
    toast.add({
      severity: 'error',
      summary: 'Error',
      detail: err.response?.data?.detail || 'Failed to save tag',
      life: 3500,
    })
  } finally {
    saving.value = false
  }
}

const confirmDelete = async (tag) => {
  const hasChildren = tag.children && tag.children.length > 0
  const msg = hasChildren
    ? `Delete "${tag.name}"? Its sub-tags will be moved up one level.`
    : `Delete tag "${tag.name}"? Recordings will not be deleted.`
  if (!confirm(msg)) return

  try {
    await tagsStore.deleteTag(tag.id, 'reparent')
    if (editingTag.value?.id === tag.id) resetForm()
    toast.add({ severity: 'success', summary: 'Tag deleted', life: 2500 })
  } catch (err) {
    toast.add({
      severity: 'error',
      summary: 'Error',
      detail: err.response?.data?.detail || 'Failed to delete tag',
      life: 3500,
    })
  }
}
</script>

<style scoped>
.tag-manager-body {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.tag-form-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.form-label {
  font-size: 0.78rem;
  font-weight: 600;
  color: #475569;
}

.apple-input {
  width: 100%;
  padding: 0.5rem 0.7rem;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  font-size: 0.88rem;
  background: #ffffff;
  color: #0f172a;
  box-sizing: border-box;
}

.color-swatches {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.color-swatch-lg {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
}

.color-swatch-lg.active {
  border-color: #0f172a;
  transform: scale(1.12);
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.btn-primary {
  padding: 0.48rem 0.95rem;
  border: none;
  border-radius: 8px;
  background: #2563eb;
  color: #ffffff;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  padding: 0.48rem 0.85rem;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  background: #ffffff;
  color: #475569;
  font-size: 0.85rem;
  cursor: pointer;
}

.taxonomy-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.taxonomy-header {
  display: flex;
  justify-content: space-between;
  font-size: 0.78rem;
  font-weight: 700;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.taxonomy-list {
  max-height: 260px;
  overflow-y: auto;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 0.35rem 0;
}

.taxonomy-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.42rem 0.75rem;
  font-size: 0.86rem;
  color: #1e293b;
}

.taxonomy-item:hover {
  background: #f1f5f9;
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

.item-count {
  font-size: 0.74rem;
  color: #64748b;
  background: #e2e8f0;
  padding: 0.08rem 0.42rem;
  border-radius: 9999px;
}

.item-actions {
  display: flex;
  gap: 0.25rem;
}

.icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border: none;
  background: transparent;
  color: #64748b;
  border-radius: 6px;
  cursor: pointer;
}

.icon-btn:hover {
  background: rgba(0, 0, 0, 0.07);
  color: #0f172a;
}

.icon-btn.danger:hover {
  background: rgba(239, 68, 68, 0.12);
  color: #dc2626;
}

.icon-btn.is-shared {
  color: #0071e3;
}
</style>
