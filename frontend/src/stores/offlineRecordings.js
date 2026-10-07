import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  getOfflineRecordings,
  saveOfflineRecording,
  updateOfflineRecording,
  deleteOfflineRecording,
  clearAllOfflineRecordings,
  getAudioBlob
} from '../services/db'
import {
  isOnline,
  isSyncing,
  currentSyncId,
  syncPendingRecordings,
  syncSingleRecording,
  initNetworkMonitor
} from '../services/syncService'
import { audioRecorder } from '../services/audioRecorder'

export const useOfflineRecordingsStore = defineStore('offlineRecordings', () => {
  const offlineList = ref([])
  const initialized = ref(false)

  const pendingCount = computed(() => {
    return offlineList.value.filter(r => r.status === 'saved_locally' || r.status === 'error' || r.status === 'uploading').length
  })

  const refresh = async () => {
    try {
      offlineList.value = await getOfflineRecordings()
      // Sort newest first
      offlineList.value.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt))
    } catch (e) {
      console.error('Failed to load offline recordings:', e)
    }
  }

  const init = async () => {
    if (initialized.value) return
    initialized.value = true
    await refresh()
    initNetworkMonitor(() => {
      // When network restored, sync pending and refresh
      syncPendingRecordings(() => {
        refresh()
      })
    })
  }

  const addRecording = async (meta, audioBlob = null) => {
    const item = {
      id: meta.id || `rec_${Date.now()}`,
      title: meta.title || `Recording ${new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`,
      filename: meta.filename,
      filePath: meta.filePath,
      duration: meta.duration || 0,
      fileSize: meta.fileSize || 0,
      mimeType: meta.mimeType || 'audio/m4a',
      isMedical: !!meta.isMedical,
      isNative: !!meta.isNative,
      createdAt: meta.createdAt || new Date().toISOString(),
      status: 'saved_locally',
      uploadProgress: 0,
      uploadId: null,
      completedChunks: [],
      jobId: null,
      errorMessage: null
    }

    await saveOfflineRecording(item, audioBlob)
    await refresh()

    // If online, immediately initiate background sync
    if (isOnline.value) {
      syncRecording(item.id)
    }

    return item
  }

  const syncRecording = async (id) => {
    const rec = offlineList.value.find(r => r.id === id)
    if (!rec) return
    try {
      await syncSingleRecording(rec, (progress) => {
        const target = offlineList.value.find(r => r.id === id)
        if (target) target.uploadProgress = progress
      })
      await refresh()
    } catch (err) {
      await refresh()
      throw err
    }
  }

  const syncAll = async () => {
    await syncPendingRecordings(() => {
      refresh()
    })
    await refresh()
  }

  const removeRecording = async (id) => {
    const target = offlineList.value.find(r => r.id === id)
    if (target) {
      await audioRecorder.deleteLocalFile(target)
    }
    await deleteOfflineRecording(id)
    await refresh()
  }

  const clearAllRecordings = async () => {
    for (const rec of offlineList.value) {
      try {
        await audioRecorder.deleteLocalFile(rec)
      } catch (e) {
        console.warn('Error deleting local file:', e)
      }
    }
    await clearAllOfflineRecordings()
    await refresh()
  }

  const getAudioUrl = async (id) => {
    const target = offlineList.value.find(r => r.id === id)
    if (!target) return null

    if (target.isNative && target.filePath) {
      // In Capacitor native WebView, convert native file path to webview URL
      const { Capacitor } = await import('@capacitor/core')
      return Capacitor.convertFileSrc(target.filePath)
    }

    const blob = await getAudioBlob(id)
    if (blob) {
      return URL.createObjectURL(blob)
    }
    return null
  }

  return {
    offlineList,
    pendingCount,
    isOnline,
    isSyncing,
    currentSyncId,
    init,
    refresh,
    addRecording,
    syncRecording,
    syncAll,
    removeRecording,
    clearAllRecordings,
    getAudioUrl
  }
})
