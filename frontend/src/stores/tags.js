import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'

export const TAG_COLORS = [
  '#3b82f6', // Blue
  '#8b5cf6', // Purple
  '#ec4899', // Pink
  '#ef4444', // Red
  '#f97316', // Orange
  '#eab308', // Amber
  '#10b981', // Emerald
  '#14b8a6', // Teal
  '#64748b', // Slate
]

export const useTagsStore = defineStore('tags', () => {
  const tags = ref([])
  const tree = ref([])
  const untaggedCount = ref(0)
  const loading = ref(false)
  const loaded = ref(false)

  const tagMap = computed(() => {
    const map = new Map()
    for (const t of tags.value) {
      map.set(t.id, t)
    }
    return map
  })

  async function fetchTags(force = false) {
    if (loading.value) return
    if (loaded.value && !force) return
    loading.value = true
    try {
      const res = await axios.get('/api/tags')
      tags.value = res.data.tags || []
      tree.value = res.data.tree || []
      untaggedCount.value = res.data.untagged_count || 0
      loaded.value = true
    } catch (err) {
      console.error('Failed to fetch tags', err)
    } finally {
      loading.value = false
    }
  }

  async function createTag(payload) {
    const res = await axios.post('/api/tags', payload)
    await fetchTags(true)
    return res.data
  }

  async function updateTag(id, payload) {
    const res = await axios.patch(`/api/tags/${id}`, payload)
    await fetchTags(true)
    return res.data
  }

  async function deleteTag(id, mode = 'reparent') {
    await axios.delete(`/api/tags/${id}?mode=${mode}`)
    await fetchTags(true)
  }

  async function shareTag(id) {
    const res = await axios.post(`/api/tags/${id}/share`)
    await fetchTags(true)
    return res.data.share_token
  }

  async function unshareTag(id) {
    await axios.delete(`/api/tags/${id}/share`)
    await fetchTags(true)
  }

  async function setRecordingTags(recordingId, tagIds) {
    const res = await axios.put(`/api/recordings/${recordingId}/tags`, {
      tag_ids: tagIds,
    })
    await fetchTags(true)
    return res.data.tags || []
  }

  async function bulkUpdateTags(recordingIds, tagIds, action) {
    const res = await axios.post('/api/recordings/bulk-tags', {
      recording_ids: recordingIds,
      tag_ids: tagIds,
      action,
    })
    await fetchTags(true)
    return res.data
  }

  function getDescendantIds(tagId) {
    const result = [tagId]
    const queue = [tagId]
    const visited = new Set([tagId])
    while (queue.length > 0) {
      const curr = queue.shift()
      for (const t of tags.value) {
        if (t.parent_id === curr && !visited.has(t.id)) {
          visited.add(t.id)
          result.push(t.id)
          queue.push(t.id)
        }
      }
    }
    return result
  }

  function reset() {
    tags.value = []
    tree.value = []
    untaggedCount.value = 0
    loaded.value = false
  }

  return {
    tags,
    tree,
    untaggedCount,
    loading,
    loaded,
    tagMap,
    fetchTags,
    createTag,
    updateTag,
    deleteTag,
    shareTag,
    unshareTag,
    setRecordingTags,
    bulkUpdateTags,
    getDescendantIds,
    reset,
  }
})
