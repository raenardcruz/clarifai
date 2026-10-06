import { ref } from 'vue'
import { Network } from '@capacitor/network'
import axios from 'axios'
import {
  getOfflineRecordings,
  updateOfflineRecording,
  getAudioBlob,
  deleteOfflineRecording
} from './db'
import { audioRecorder } from './audioRecorder'

const CHUNK_SIZE = 5 * 1024 * 1024 // 5MB chunk size

export const isOnline = ref(navigator.onLine)
export const isSyncing = ref(false)
export const currentSyncId = ref(null)

// Initialize network listener
export async function initNetworkMonitor(onOnlineCallback = null) {
  try {
    const status = await Network.getStatus()
    isOnline.value = status.connected

    Network.addListener('networkStatusChange', async (newStatus) => {
      const wasOffline = !isOnline.value
      isOnline.value = newStatus.connected
      if (wasOffline && newStatus.connected) {
        if (onOnlineCallback) {
          onOnlineCallback()
        } else {
          syncPendingRecordings()
        }
      }
    })
  } catch (e) {
    console.warn('Network plugin listener unavailable, using window events', e)
  }

  window.addEventListener('online', () => {
    isOnline.value = true
    if (onOnlineCallback) {
      onOnlineCallback()
    } else {
      syncPendingRecordings()
    }
  })

  window.addEventListener('offline', () => {
    isOnline.value = false
  })
}

// Upload a single recording to the backend with chunking and retries
export async function syncSingleRecording(recording, onProgress = null) {
  if (!isOnline.value) {
    throw new Error('Device is currently offline')
  }

  const token = localStorage.getItem('token') || (JSON.parse(localStorage.getItem('token') || 'null'))
  if (token) {
    axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
  }

  await updateOfflineRecording(recording.id, {
    status: 'uploading',
    errorMessage: null
  })

  try {
    const totalChunks = Math.max(1, Math.ceil(recording.fileSize / CHUNK_SIZE))
    let uploadId = recording.uploadId

    // 1. Initialize upload session if not already initialized
    if (!uploadId) {
      const initRes = await axios.post('/api/recordings/upload/init', {
        filename: recording.filename || `Recording_${recording.id}.m4a`,
        total_chunks: totalChunks,
        file_size: recording.fileSize
      })
      uploadId = initRes.data.upload_id
      await updateOfflineRecording(recording.id, { uploadId })
    }

    const uploadedChunks = new Set(recording.completedChunks || [])

    // 2. Upload chunks sequentially
    for (let i = 0; i < totalChunks; i++) {
      if (uploadedChunks.has(i)) {
        continue
      }

      // Check online status before starting each chunk
      if (!isOnline.value) {
        await updateOfflineRecording(recording.id, {
          status: 'saved_locally',
          completedChunks: Array.from(uploadedChunks)
        })
        throw new Error('Connection lost while uploading. Will resume when online.')
      }

      const offset = i * CHUNK_SIZE
      const length = Math.min(CHUNK_SIZE, recording.fileSize - offset)

      // Read chunk slice directly from storage (native or web blob)
      const chunkBlob = await audioRecorder.readChunk(recording, offset, length)

      // Upload chunk with retries
      let retries = 3
      let uploaded = false
      while (retries > 0 && !uploaded) {
        try {
          const formData = new FormData()
          formData.append('upload_id', uploadId)
          formData.append('chunk_index', i.toString())
          formData.append('chunk', chunkBlob, `chunk_${i}`)

          await axios.post('/api/recordings/upload/chunk', formData)
          uploadedChunks.add(i)
          uploaded = true

          const progress = Math.round((uploadedChunks.size / totalChunks) * 100)
          await updateOfflineRecording(recording.id, {
            uploadProgress: progress,
            completedChunks: Array.from(uploadedChunks)
          })

          if (onProgress) {
            onProgress(progress, uploadedChunks.size, totalChunks)
          }
        } catch (err) {
          retries--
          if (retries === 0) throw err
          await new Promise((r) => setTimeout(r, 1200))
        }
      }
    }

    // 3. Complete chunked upload on backend
    const completeRes = await axios.post('/api/recordings/upload/complete', {
      upload_id: uploadId,
      filename: recording.filename || `Recording_${recording.id}.m4a`,
      total_chunks: totalChunks,
      is_medical: !!recording.isMedical
    })

    const jobId = completeRes.data.job_id

    // Mark recording as synced
    await updateOfflineRecording(recording.id, {
      status: 'synced',
      uploadProgress: 100,
      jobId: jobId,
      syncedAt: new Date().toISOString()
    })

    return { success: true, jobId }
  } catch (err) {
    console.error(`Failed to sync recording ${recording.id}:`, err)
    const errorMsg = err.response?.data?.detail || err.message || 'Upload failed'
    await updateOfflineRecording(recording.id, {
      status: isOnline.value ? 'error' : 'saved_locally',
      errorMessage: errorMsg
    })
    throw err
  }
}

// Sync all pending recordings in the queue
export async function syncPendingRecordings(onItemSynced = null) {
  if (isSyncing.value || !isOnline.value) return

  try {
    isSyncing.value = true
    const allRecordings = await getOfflineRecordings()
    const pendingList = allRecordings.filter(r => r.status === 'saved_locally' || r.status === 'error' || r.status === 'uploading')

    for (const rec of pendingList) {
      if (!isOnline.value) break // Stop if network disconnected
      currentSyncId.value = rec.id
      try {
        const result = await syncSingleRecording(rec)
        if (onItemSynced) {
          onItemSynced(rec.id, result.jobId)
        }
      } catch (e) {
        console.warn(`Sync skipped for ${rec.id}: ${e.message}`)
      }
    }
  } finally {
    isSyncing.value = false
    currentSyncId.value = null
  }
}
