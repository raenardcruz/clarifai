// IndexedDB wrapper for persistent offline recordings
const DB_NAME = 'NoteTakerOfflineDB'
const DB_VERSION = 1
const STORE_RECORDINGS = 'recordings'
const STORE_AUDIO_BLOBS = 'audio_blobs'

function openDB() {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION)

    request.onupgradeneeded = (e) => {
      const db = e.target.result
      if (!db.objectStoreNames.contains(STORE_RECORDINGS)) {
        db.createObjectStore(STORE_RECORDINGS, { keyPath: 'id' })
      }
      if (!db.objectStoreNames.contains(STORE_AUDIO_BLOBS)) {
        db.createObjectStore(STORE_AUDIO_BLOBS, { keyPath: 'id' })
      }
    }

    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

export async function saveOfflineRecording(recordingMeta, audioBlob = null) {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction([STORE_RECORDINGS, STORE_AUDIO_BLOBS], 'readwrite')
    const recStore = tx.objectStore(STORE_RECORDINGS)
    const blobStore = tx.objectStore(STORE_AUDIO_BLOBS)

    recStore.put(recordingMeta)
    if (audioBlob) {
      blobStore.put({ id: recordingMeta.id, blob: audioBlob })
    }

    tx.oncomplete = () => resolve(recordingMeta)
    tx.onerror = () => reject(tx.error)
  })
}

export async function getOfflineRecordings() {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_RECORDINGS, 'readonly')
    const store = tx.objectStore(STORE_RECORDINGS)
    const req = store.getAll()
    req.onsuccess = () => resolve(req.result || [])
    req.onerror = () => reject(req.error)
  })
}

export async function getOfflineRecording(id) {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_RECORDINGS, 'readonly')
    const store = tx.objectStore(STORE_RECORDINGS)
    const req = store.get(id)
    req.onsuccess = () => resolve(req.result)
    req.onerror = () => reject(req.error)
  })
}

export async function updateOfflineRecording(id, updates) {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_RECORDINGS, 'readwrite')
    const store = tx.objectStore(STORE_RECORDINGS)
    const req = store.get(id)
    req.onsuccess = () => {
      const existing = req.result
      if (!existing) {
        reject(new Error(`Recording ${id} not found`))
        return
      }
      const updated = { ...existing, ...updates }
      const putReq = store.put(updated)
      putReq.onsuccess = () => resolve(updated)
      putReq.onerror = () => reject(putReq.error)
    }
    req.onerror = () => reject(req.error)
  })
}

export async function getAudioBlob(id) {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_AUDIO_BLOBS, 'readonly')
    const store = tx.objectStore(STORE_AUDIO_BLOBS)
    const req = store.get(id)
    req.onsuccess = () => resolve(req.result ? req.result.blob : null)
    req.onerror = () => reject(req.error)
  })
}

export async function deleteOfflineRecording(id) {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction([STORE_RECORDINGS, STORE_AUDIO_BLOBS], 'readwrite')
    tx.objectStore(STORE_RECORDINGS).delete(id)
    tx.objectStore(STORE_AUDIO_BLOBS).delete(id)
    tx.oncomplete = () => resolve(true)
    tx.onerror = () => reject(tx.error)
  })
}

export async function clearAllOfflineRecordings() {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction([STORE_RECORDINGS, STORE_AUDIO_BLOBS], 'readwrite')
    tx.objectStore(STORE_RECORDINGS).clear()
    tx.objectStore(STORE_AUDIO_BLOBS).clear()
    tx.oncomplete = () => resolve(true)
    tx.onerror = () => reject(tx.error)
  })
}
