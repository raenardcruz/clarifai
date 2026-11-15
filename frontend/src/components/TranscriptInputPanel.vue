<template>
  <div class="transcript-input-panel">
    <div class="transcript-toolbar-row">
      <label :for="textareaId" class="transcript-field-label">
        {{ label }}
      </label>

      <div class="transcript-toolbar-actions">
        <input
          ref="fileInputRef"
          type="file"
          class="hidden-file-input"
          accept=".txt,.md,.vtt,.srt,.json,text/plain,text/markdown,text/vtt,application/json"
          @change="handleFileSelect"
          :disabled="disabled"
        />
        <button
          type="button"
          class="transcript-mini-btn"
          @click="triggerFilePicker"
          :disabled="disabled"
          title="Load transcript from a .txt, .md, .vtt, .srt, or .json file"
        >
          <Upload size="13" />
          <span>Import File</span>
        </button>
        <button
          v-if="modelValue"
          type="button"
          class="transcript-mini-btn danger"
          @click="$emit('update:modelValue', '')"
          :disabled="disabled"
          title="Clear transcript text"
        >
          <X size="13" />
          <span>Clear</span>
        </button>
      </div>
    </div>

    <div
      class="transcript-textarea-wrap"
      :class="{ 'is-dragging': isDragging, 'is-disabled': disabled }"
      @dragover.prevent="!disabled && (isDragging = true)"
      @dragleave.prevent="isDragging = false"
      @drop.prevent="handleDrop"
    >
      <textarea
        :id="textareaId"
        :value="modelValue"
        @input="$emit('update:modelValue', $event.target.value)"
        :rows="rows"
        :disabled="disabled"
        class="transcript-textarea"
        :placeholder="placeholder"
      ></textarea>
      <div v-if="isDragging" class="drop-overlay">
        <FileText size="22" />
        <span>Drop transcript file to load</span>
      </div>
    </div>

    <div class="transcript-meta-row">
      <span class="format-hint">
        Supports plain text, <code>Speaker: text</code>, timestamps (<code>[00:12]</code>), <code>.vtt</code>, <code>.srt</code>, or <code>.json</code>.
      </span>
      <span class="transcript-stats font-tabular" v-if="wordCount > 0">
        {{ wordCount.toLocaleString() }} {{ wordCount === 1 ? 'word' : 'words' }} · {{ lineCount }} {{ lineCount === 1 ? 'line' : 'lines' }}
      </span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Upload, X, FileText } from '@lucide/vue'
import { useToast } from 'primevue/usetoast'

const props = defineProps({
  modelValue: { type: String, default: '' },
  label: { type: String, default: 'Full Transcript' },
  rows: { type: Number, default: 9 },
  disabled: { type: Boolean, default: false },
  textareaId: { type: String, default: 'transcript-input-textarea' },
  placeholder: {
    type: String,
    default: `Paste your entire transcript here, or drag & drop a .txt / .vtt / .srt file...\n\nExamples:\nAlice: Welcome everyone, let's review the Q4 milestones.\nBob: Sounds good. Engineering is on track for Friday.\n\nOr paste plain paragraphs without speaker labels.`
  }
})

const emit = defineEmits(['update:modelValue', 'file-loaded'])

const toast = useToast()
const fileInputRef = ref(null)
const isDragging = ref(false)

const wordCount = computed(() => {
  const trimmed = (props.modelValue || '').trim()
  if (!trimmed) return 0
  return trimmed.split(/\s+/).length
})

const lineCount = computed(() => {
  const trimmed = (props.modelValue || '').trim()
  if (!trimmed) return 0
  return trimmed.split(/\r?\n/).filter((l) => l.trim() !== '').length
})

const triggerFilePicker = () => {
  if (fileInputRef.value) {
    fileInputRef.value.click()
  }
}

const loadTextFile = async (file) => {
  if (!file) return
  try {
    const text = await file.text()
    if (!text.trim()) {
      toast.add({ severity: 'warn', summary: 'Empty File', detail: 'Selected file does not contain any text.', life: 3500 })
      return
    }
    emit('update:modelValue', text)
    const cleanTitle = file.name.replace(/\.[^/.]+$/, '').replace(/[_-]+/g, ' ').trim()
    emit('file-loaded', { filename: file.name, suggestedTitle: cleanTitle })
  } catch (err) {
    console.error('Failed to read transcript file', err)
    toast.add({ severity: 'error', summary: 'Read Failed', detail: 'Could not read the selected transcript file.', life: 4000 })
  }
}

const handleFileSelect = async (event) => {
  const file = event.target.files?.[0]
  if (file) {
    await loadTextFile(file)
  }
  if (event.target) {
    event.target.value = ''
  }
}

const handleDrop = async (event) => {
  isDragging.value = false
  if (props.disabled) return
  const file = event.dataTransfer?.files?.[0]
  if (file) {
    await loadTextFile(file)
  }
}
</script>

<style scoped>
.transcript-input-panel {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.transcript-toolbar-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.transcript-field-label {
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--text-primary);
}

.transcript-toolbar-actions {
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.hidden-file-input {
  display: none;
}

.transcript-mini-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.32rem 0.65rem;
  border-radius: 9999px;
  font-size: 0.76rem;
  font-weight: 600;
  border: 1px solid rgba(0, 0, 0, 0.1);
  background: var(--white, #ffffff);
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.15s ease;
}

.transcript-mini-btn:hover:not(:disabled) {
  border-color: var(--apple-blue, #0071e3);
  color: var(--apple-blue, #0071e3);
  background: rgba(0, 113, 227, 0.05);
}

.transcript-mini-btn.danger:hover:not(:disabled) {
  border-color: #ef4444;
  color: #dc2626;
  background: rgba(239, 68, 68, 0.06);
}

.transcript-mini-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.transcript-textarea-wrap {
  position: relative;
  border-radius: var(--radius-lg, 12px);
}

.transcript-textarea-wrap.is-dragging .transcript-textarea {
  border-color: var(--apple-blue, #0071e3);
  background: rgba(0, 113, 227, 0.03);
}

.transcript-textarea {
  width: 100%;
  min-height: 190px;
  padding: 0.85rem 1rem;
  border-radius: var(--radius-lg, 12px);
  border: 1.5px solid rgba(0, 0, 0, 0.12);
  background: var(--bg-primary, #f8fafc);
  color: var(--text-primary);
  font-family: inherit;
  font-size: 0.875rem;
  line-height: 1.55;
  resize: vertical;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
  box-sizing: border-box;
}

.transcript-textarea:focus {
  outline: none;
  border-color: var(--apple-blue, #0071e3);
  box-shadow: 0 0 0 3px rgba(0, 113, 227, 0.14);
  background: var(--white, #ffffff);
}

.transcript-textarea:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.drop-overlay {
  position: absolute;
  inset: 0;
  border-radius: var(--radius-lg, 12px);
  background: rgba(0, 113, 227, 0.08);
  border: 2px dashed var(--apple-blue, #0071e3);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  color: var(--apple-blue, #0071e3);
  font-size: 0.88rem;
  font-weight: 600;
  pointer-events: none;
}

.transcript-meta-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.format-hint {
  font-size: 0.74rem;
  color: var(--text-muted, #64748b);
  line-height: 1.4;
}

.format-hint code {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.71rem;
  background: rgba(0, 0, 0, 0.05);
  padding: 0.08rem 0.3rem;
  border-radius: 4px;
}

.transcript-stats {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--apple-blue, #0071e3);
  white-space: nowrap;
}

.font-tabular {
  font-variant-numeric: tabular-nums;
}
</style>
