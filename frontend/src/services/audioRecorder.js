import { Capacitor } from '@capacitor/core'
import { NativeAudioRecorder } from 'native-audio-recorder'

class UnifiedAudioRecorder {
  constructor() {
    this.isNative = Capacitor.isNativePlatform()
    this.activeRecordingId = null
    this.currentRecording = null
    this.mediaRecorder = null
    this.audioChunks = []
    this.audioContext = null
    this.analyser = null
    this.startTime = null
    this.accumulatedDuration = 0
    this.isPaused = false
    this.lastRecordedBlob = null
  }

  async checkPermission() {
    if (this.isNative) {
      try {
        const res = await NativeAudioRecorder.checkPermissions()
        return res.record === 'granted'
      } catch (e) {
        console.warn('Native permission check error:', e)
        return false
      }
    }
    return true
  }

  async requestPermission() {
    if (this.isNative) {
      try {
        const res = await NativeAudioRecorder.requestPermissions()
        return res.record === 'granted'
      } catch (e) {
        console.warn('Native permission request error:', e)
        return false
      }
    }
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      stream.getTracks().forEach(t => t.stop())
      return true
    } catch {
      return false
    }
  }

  async start(recordingId = null) {
    const id = recordingId || `rec_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`
    this.activeRecordingId = id

    if (this.isNative) {
      const res = await NativeAudioRecorder.startRecording({ id })
      this.currentRecording = {
        id,
        filePath: res.filePath,
        status: 'recording',
        isNative: true
      }
      return this.currentRecording
    }

    // Web Fallback
    this.audioChunks = []
    this.accumulatedDuration = 0
    this.isPaused = false
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })

    try {
      const AudioCtx = window.AudioContext || window.webkitAudioContext
      if (AudioCtx) {
        this.audioContext = new AudioCtx()
        const source = this.audioContext.createMediaStreamSource(stream)
        this.analyser = this.audioContext.createAnalyser()
        this.analyser.fftSize = 64
        source.connect(this.analyser)
      }
    } catch (e) {
      console.warn('Web AudioContext warning:', e)
    }

    this.mediaRecorder = new MediaRecorder(stream)
    this.mediaRecorder.ondataavailable = (e) => {
      if (e.data && e.data.size > 0) {
        this.audioChunks.push(e.data)
      }
    }

    this.mediaRecorder.start(1000)
    this.startTime = Date.now()

    this.currentRecording = {
      id,
      filePath: `web://${id}`,
      status: 'recording',
      isNative: false
    }
    return this.currentRecording
  }

  async pause() {
    if (this.isNative) {
      return await NativeAudioRecorder.pauseRecording()
    }

    if (this.mediaRecorder && this.mediaRecorder.state === 'recording') {
      this.mediaRecorder.pause()
      this.accumulatedDuration += (Date.now() - this.startTime) / 1000
      this.startTime = null
      this.isPaused = true
    }
    return { status: 'paused', duration: this.accumulatedDuration }
  }

  async resume() {
    if (this.isNative) {
      return await NativeAudioRecorder.resumeRecording()
    }

    if (this.mediaRecorder && this.mediaRecorder.state === 'paused') {
      this.mediaRecorder.resume()
      this.startTime = Date.now()
      this.isPaused = false
    }
    return { status: 'recording', duration: this.accumulatedDuration }
  }

  async stop() {
    if (this.isNative) {
      const res = await NativeAudioRecorder.stopRecording()
      return {
        id: this.activeRecordingId,
        filePath: res.filePath,
        filename: res.filename || `Recording_${Date.now()}.m4a`,
        duration: Math.round(res.duration || 0),
        fileSize: res.fileSize || 0,
        mimeType: 'audio/m4a',
        isNative: true,
        blob: null
      }
    }

    return new Promise((resolve) => {
      if (!this.mediaRecorder) {
        resolve({
          id: this.activeRecordingId,
          duration: 0,
          fileSize: 0,
          mimeType: 'audio/webm',
          isNative: false,
          blob: null
        })
        return
      }

      if (this.startTime) {
        this.accumulatedDuration += (Date.now() - this.startTime) / 1000
      }

      this.mediaRecorder.onstop = () => {
        const audioBlob = new Blob(this.audioChunks, { type: 'audio/webm' })
        this.lastRecordedBlob = audioBlob
        const duration = Math.round(this.accumulatedDuration)
        const size = audioBlob.size

        if (this.mediaRecorder.stream) {
          this.mediaRecorder.stream.getTracks().forEach(t => t.stop())
        }
        if (this.audioContext) {
          this.audioContext.close()
          this.audioContext = null
        }

        resolve({
          id: this.activeRecordingId,
          filePath: `web://${this.activeRecordingId}`,
          filename: `Recording_${Date.now()}.webm`,
          duration,
          fileSize: size,
          mimeType: 'audio/webm',
          isNative: false,
          blob: audioBlob
        })
      }

      this.mediaRecorder.stop()
    })
  }

  async getStatus() {
    if (this.isNative) {
      try {
        const res = await NativeAudioRecorder.getStatus()
        return {
          isRecording: !!res.isRecording,
          isPaused: !!res.isPaused,
          duration: res.duration || 0,
          meterLevel: res.meterLevel || 0
        }
      } catch {
        return { isRecording: false, isPaused: false, duration: 0, meterLevel: 0 }
      }
    }

    let currentDuration = this.accumulatedDuration
    if (this.startTime && !this.isPaused) {
      currentDuration += (Date.now() - this.startTime) / 1000
    }

    let meterLevel = 0
    if (this.analyser) {
      const dataArray = new Uint8Array(this.analyser.frequencyBinCount)
      this.analyser.getByteFrequencyData(dataArray)
      const avg = dataArray.reduce((acc, val) => acc + val, 0) / dataArray.length
      meterLevel = Math.min(1.0, avg / 128)
    }

    return {
      isRecording: this.mediaRecorder ? this.mediaRecorder.state === 'recording' : false,
      isPaused: this.isPaused,
      duration: currentDuration,
      meterLevel: meterLevel
    }
  }

  async readChunk(recording, offset, length) {
    if (recording.isNative) {
      const res = await NativeAudioRecorder.readChunk({
        filePath: recording.filePath,
        offset,
        length
      })
      // Convert base64 string to Blob
      const byteCharacters = atob(res.base64)
      const byteNumbers = new Array(byteCharacters.length)
      for (let i = 0; i < byteCharacters.length; i++) {
        byteNumbers[i] = byteCharacters.charCodeAt(i)
      }
      const byteArray = new Uint8Array(byteNumbers)
      return new Blob([byteArray], { type: recording.mimeType || 'audio/m4a' })
    }

    // Web mode: get blob from IndexedDB if not in memory
    const { getAudioBlob } = await import('./db')
    const blob = recording.blob || await getAudioBlob(recording.id)
    if (!blob) throw new Error('Audio blob not found on device')
    return blob.slice(offset, offset + length)
  }

  async deleteLocalFile(recording) {
    if (recording.isNative && recording.filePath) {
      try {
        await NativeAudioRecorder.deleteAudioFile({ filePath: recording.filePath })
      } catch (e) {
        console.warn('Native delete error:', e)
      }
    }
  }
}

export const audioRecorder = new UnifiedAudioRecorder()
