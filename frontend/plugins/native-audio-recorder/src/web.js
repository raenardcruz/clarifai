import { WebPlugin } from '@capacitor/core'

export class NativeAudioRecorderWeb extends WebPlugin {
  constructor() {
    super()
    this.mediaRecorder = null
    this.audioChunks = []
    this.recordingId = null
    this.startTime = null
    this.accumulatedDuration = 0
    this.isPaused = false
    this.audioBlob = null
    this.analyser = null
    this.audioContext = null
  }

  async checkPermissions() {
    return { record: 'granted' }
  }

  async requestPermissions() {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      stream.getTracks().forEach(t => t.stop())
      return { record: 'granted' }
    } catch {
      return { record: 'denied' }
    }
  }

  async startRecording(options = {}) {
    this.recordingId = options.id || `web_${Date.now()}`
    this.audioChunks = []
    this.accumulatedDuration = 0
    this.isPaused = false

    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    
    // Set up analyser for metering
    try {
      this.audioContext = new (window.AudioContext || window.webkitAudioContext)()
      const source = this.audioContext.createMediaStreamSource(stream)
      this.analyser = this.audioContext.createAnalyser()
      this.analyser.fftSize = 64
      source.connect(this.analyser)
    } catch (e) {
      console.warn('AudioContext metering setup warning:', e)
    }

    this.mediaRecorder = new MediaRecorder(stream)
    this.mediaRecorder.ondataavailable = (e) => {
      if (e.data && e.data.size > 0) {
        this.audioChunks.push(e.data)
      }
    }

    this.mediaRecorder.start(1000)
    this.startTime = Date.now()

    return {
      recordingId: this.recordingId,
      filePath: `web://${this.recordingId}`,
      status: 'recording'
    }
  }

  async pauseRecording() {
    if (this.mediaRecorder && this.mediaRecorder.state === 'recording') {
      this.mediaRecorder.pause()
      this.accumulatedDuration += (Date.now() - this.startTime) / 1000
      this.startTime = null
      this.isPaused = true
    }
    return { status: 'paused', duration: this.accumulatedDuration }
  }

  async resumeRecording() {
    if (this.mediaRecorder && this.mediaRecorder.state === 'paused') {
      this.mediaRecorder.resume()
      this.startTime = Date.now()
      this.isPaused = false
    }
    return { status: 'recording', duration: this.accumulatedDuration }
  }

  async stopRecording() {
    return new Promise((resolve) => {
      if (!this.mediaRecorder) {
        resolve({ recordingId: this.recordingId, duration: 0, fileSize: 0 })
        return
      }

      if (this.startTime) {
        this.accumulatedDuration += (Date.now() - this.startTime) / 1000
      }

      this.mediaRecorder.onstop = () => {
        this.audioBlob = new Blob(this.audioChunks, { type: 'audio/webm' })
        const duration = this.accumulatedDuration
        const size = this.audioBlob.size
        
        // Stop all tracks
        this.mediaRecorder.stream.getTracks().forEach(t => t.stop())
        if (this.audioContext) {
          this.audioContext.close()
          this.audioContext = null
        }

        resolve({
          recordingId: this.recordingId,
          filePath: `web://${this.recordingId}`,
          filename: `Live_Recording_${Date.now()}.webm`,
          duration: duration,
          fileSize: size,
          mimeType: 'audio/webm',
          blob: this.audioBlob
        })
      }

      this.mediaRecorder.stop()
    })
  }

  async getStatus() {
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

  async readChunk(options) {
    if (!this.audioBlob) throw new Error('No blob available')
    const { offset, length } = options
    const slice = this.audioBlob.slice(offset, offset + length)
    const arrayBuffer = await slice.arrayBuffer()
    const base64 = btoa(String.fromCharCode(...new Uint8Array(arrayBuffer)))
    return { base64, bytesRead: slice.size }
  }

  async readEntireFileBase64() {
    if (!this.audioBlob) throw new Error('No blob available')
    const arrayBuffer = await this.audioBlob.arrayBuffer()
    const base64 = btoa(String.fromCharCode(...new Uint8Array(arrayBuffer)))
    return { base64, fileSize: this.audioBlob.size }
  }

  async deleteAudioFile() {
    this.audioBlob = null
    this.audioChunks = []
    return { deleted: true }
  }

  async getRecordedFiles() {
    return { files: [] }
  }
}
