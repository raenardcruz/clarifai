<template>
  <div class="settings-page">
    <div class="page-header">
      <h1 class="page-title">System Settings</h1>
      <p class="page-subtitle">Configure AI models, speech transcription engines, and system rules</p>
    </div>

    <!-- macOS Grouped Settings Card -->
    <div class="settings-card">
      <form @submit.prevent="saveSettings">
        <!-- AI Intelligence Section -->
        <div class="settings-section">
          <div class="section-title-row">
            <div class="section-icon-badge bg-blue">
              <Sparkles size="16" color="#0071E3" />
            </div>
            <div>
              <h3 class="section-heading">AI & Summarization</h3>
              <p class="section-desc">Control how and when ClarifAi generates meeting notes</p>
            </div>
          </div>

          <div class="apple-settings-group">
            <!-- Mode Setting Row -->
            <div class="setting-item">
              <div class="setting-info">
                <span class="setting-name">Summarization Trigger</span>
                <span class="setting-hint">Choose whether meeting summaries are generated automatically or on demand</span>
              </div>
              
              <div class="mode-selection-grid">
                <label 
                  class="apple-choice-card" 
                  :class="{ active: settings.ai_summarization_mode === 'auto' }"
                >
                  <input type="radio" value="auto" v-model="settings.ai_summarization_mode" class="hidden-radio" />
                  <div class="choice-content">
                    <div class="choice-title-row">
                      <strong>Automatic</strong>
                      <span class="choice-check" v-if="settings.ai_summarization_mode === 'auto'">✓</span>
                    </div>
                    <span class="choice-sub">Generate immediately after audio transcription finishes</span>
                  </div>
                </label>

                <label 
                  class="apple-choice-card" 
                  :class="{ active: settings.ai_summarization_mode === 'manual' }"
                >
                  <input type="radio" value="manual" v-model="settings.ai_summarization_mode" class="hidden-radio" />
                  <div class="choice-content">
                    <div class="choice-title-row">
                      <strong>Manual</strong>
                      <span class="choice-check" v-if="settings.ai_summarization_mode === 'manual'">✓</span>
                    </div>
                    <span class="choice-sub">Wait for manual trigger to allow transcript editing first</span>
                  </div>
                </label>
              </div>
            </div>

            <div class="setting-divider"></div>

            <!-- Ollama Model Row -->
            <div class="setting-item">
              <div class="setting-info">
                <span class="setting-name">Ollama LLM Model</span>
                <span class="setting-hint">Local model used for generating executive summaries and titles</span>
              </div>
              <div class="setting-control">
                <Select 
                  v-model="settings.ollama_model" 
                  :options="availableModels" 
                  placeholder="Select a Model"
                  class="apple-select-input"
                />
              </div>
            </div>

            <div class="setting-divider"></div>

            <!-- Executive Prompt Row -->
            <div class="setting-item flex-col-item">
              <div class="setting-info mb-2">
                <span class="setting-name">Executive Summary System Prompt</span>
                <span class="setting-hint">Core instructions passed to the local LLM to shape the summary tone and structure</span>
              </div>
              <div class="setting-control w-full">
                <textarea 
                  v-model="settings.executive_summary_prompt" 
                  class="apple-textarea"
                  rows="4"
                  placeholder="Enter system prompt instructions..."
                ></textarea>
              </div>
            </div>

            <div class="setting-divider"></div>

            <!-- Medical Prompt Row -->
            <div class="setting-item flex-col-item">
              <div class="setting-info mb-2">
                <span class="setting-name">Medical Conversation SOAP System Prompt</span>
                <span class="setting-hint">Specialized clinical instructions used by the local LLM for recordings flagged as medical encounters</span>
              </div>
              <div class="setting-control w-full">
                <textarea 
                  v-model="settings.medical_summary_prompt" 
                  class="apple-textarea"
                  rows="4"
                  placeholder="Enter medical/clinical SOAP prompt instructions..."
                ></textarea>
              </div>
            </div>
          </div>
        </div>


        <!-- Speech Transcription Engine Section -->
        <div class="settings-section mt-8">
          <div class="section-title-row">
            <div class="section-icon-badge bg-purple">
              <Key size="16" color="#AF52DE" />
            </div>
            <div>
              <h3 class="section-heading">Transcription Provider</h3>
              <p class="section-desc">Connect Speechmatics for batch & live speech-to-text</p>
            </div>
          </div>

          <div class="apple-settings-group">
            <div class="setting-item flex-col-item">
              <div class="setting-info mb-2">
                <span class="setting-name">Speechmatics API Key</span>
                <span class="setting-hint">Used to authenticate audio transcription and speaker diarization requests</span>
              </div>

              <div class="api-key-input-wrapper">
                <input 
                  :type="showKey ? 'text' : 'password'"
                  v-model="settings.speechmatics_api_key" 
                  class="apple-input-text font-mono"
                  placeholder="Enter Speechmatics API Key"
                />
                <button type="button" class="eye-toggle-btn" @click="showKey = !showKey" title="Toggle visibility">
                  <EyeOff size="16" v-if="showKey" />
                  <Eye size="16" v-else />
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Mobile & Offline Synchronization Section -->
        <div class="settings-section">
          <div class="section-title-row">
            <div class="section-icon-badge bg-blue">
              <Smartphone size="16" color="#0071E3" />
            </div>
            <div>
              <h3 class="section-heading">Server & Offline Sync</h3>
              <p class="section-desc">Configure backend host for the iOS mobile app and inspect local offline recordings</p>
            </div>
          </div>

          <div class="apple-settings-group">
            <!-- Backend API URL -->
            <div class="setting-item">
              <div class="setting-info">
                <span class="setting-name">Backend Server URL</span>
                <span class="setting-hint">Target URL for API requests on iOS / mobile devices (e.g., http://192.168.1.100:8000)</span>
              </div>
              <div class="setting-control server-input-row">
                <input 
                  type="text" 
                  v-model="serverUrl" 
                  class="apple-input-text font-mono"
                  placeholder="e.g. http://192.168.1.50:8000"
                />
                <button type="button" class="apple-btn-secondary" @click="testServerConnection" :disabled="isTestingServer">
                  <Loader2 class="animate-spin" size="14" v-if="isTestingServer" />
                  <span>{{ isTestingServer ? 'Testing...' : 'Test' }}</span>
                </button>
              </div>
            </div>

            <div class="setting-divider"></div>

            <!-- Offline Storage Diagnostics -->
            <div class="setting-item">
              <div class="setting-info">
                <span class="setting-name">Offline Device Cache</span>
                <span class="setting-hint">{{ offlineStore.offlineList.length }} local recording(s) on device • {{ offlineStore.pendingCount }} waiting for internet upload</span>
              </div>
              <div class="setting-control">
                <button 
                  type="button" 
                  class="apple-primary-btn" 
                  :disabled="!offlineStore.isOnline || offlineStore.isSyncing || offlineStore.pendingCount === 0"
                  @click="offlineStore.syncAll()"
                >
                  <RefreshCw size="14" :class="{ 'animate-spin': offlineStore.isSyncing }" />
                  <span>{{ offlineStore.isSyncing ? 'Syncing...' : 'Sync All Pending' }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Action Bar -->
        <div class="settings-actions">
          <button type="button" class="apple-btn-secondary" @click="fetchSettings">
            Reset Changes
          </button>
          
          <button type="submit" class="apple-primary-btn" :disabled="loading">
            <Loader2 class="animate-spin" size="15" v-if="loading" />
            <Save size="15" v-else />
            <span>Save Settings</span>
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useToast } from 'primevue/usetoast'
import { Sparkles, Key, Save, Loader2, Eye, EyeOff, Smartphone, RefreshCw } from '@lucide/vue'
import Select from 'primevue/select'
import { getApiBaseUrl, setApiBaseUrl, checkBackendHealth } from '../services/apiConfig'
import { useOfflineRecordingsStore } from '../stores/offlineRecordings'

const toast = useToast()
const offlineStore = useOfflineRecordingsStore()
const serverUrl = ref('')
const isTestingServer = ref(false)

const settings = ref({
  ai_summarization_mode: 'auto',
  executive_summary_prompt: 'Analyze the following transcript and write a detailed, cohesive executive summary in a narrative paragraph format. Synthesize the meeting\'s core purpose, main arguments, and final outcomes into smooth, professional prose, explicitly attributing key ideas, decisions, and viewpoints to specific speakers by name (or speaker identifier) directly within the flow of the text. Ensure the summary is comprehensive and captures concrete details, specific project names, and actionable next steps, but deliver it entirely as a sequence of well-structured paragraphs without using any bullet points, lists, or tables. Here is the transcript: [PASTE TRANSCRIPT HERE]',
  medical_summary_prompt: '',
  speechmatics_api_key: '',
  ollama_model: 'gemma4:12b-mlx'
})

const availableModels = ref([])
const loading = ref(false)
const showKey = ref(false)

const testServerConnection = async () => {
  isTestingServer.value = true
  try {
    await setApiBaseUrl(serverUrl.value)
    const res = await checkBackendHealth()
    if (res.ok) {
      toast.add({ severity: 'success', summary: 'Connected', detail: 'Successfully reached backend server!', life: 3000 })
    } else {
      toast.add({ severity: 'warn', summary: 'Connection Warning', detail: res.error || 'Backend not responding', life: 4000 })
    }
  } catch (err) {
    toast.add({ severity: 'error', summary: 'Connection Failed', detail: err.message, life: 4000 })
  } finally {
    isTestingServer.value = false
  }
}

const fetchModels = async () => {
  try {
    const res = await axios.get('/api/settings/ollama-models')
    availableModels.value = res.data
  } catch (e) {
    console.error("Failed to load Ollama models", e)
    availableModels.value = ['gemma4:12b-mlx']
  }
}

const fetchSettings = async () => {
  serverUrl.value = getApiBaseUrl()
  try {
    const res = await axios.get('/api/settings')
    settings.value = res.data
  } catch (e) {
    console.error("Failed to load settings", e)
  }
}

const saveSettings = async () => {
  loading.value = true
  try {
    if (serverUrl.value !== getApiBaseUrl()) {
      await setApiBaseUrl(serverUrl.value)
    }
    const res = await axios.put('/api/settings', settings.value)
    settings.value = res.data
    toast.add({ severity: 'success', summary: 'Saved', detail: 'Settings updated successfully.', life: 3000 })
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to save settings.', life: 5000 })
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  serverUrl.value = getApiBaseUrl()
  fetchModels()
  fetchSettings()
})
</script>

<style scoped>
.settings-page {
  max-width: 860px;
  margin: 0 auto;
  padding: 2.5rem 2rem;
}

.page-header {
  margin-bottom: 2rem;
}

.page-title {
  font-size: 2rem;
  font-weight: 700;
  letter-spacing: -0.03em;
  color: var(--text-primary);
  margin-bottom: 0.35rem;
}

.page-subtitle {
  font-size: 0.95rem;
  color: var(--text-secondary);
}

.settings-card {
  background: var(--bg-secondary);
  border-radius: var(--radius-2xl);
  border: 1px solid rgba(0, 0, 0, 0.06);
  padding: 2.5rem 2.5rem;
  box-shadow: 0 4px 20px -2px rgba(0, 0, 0, 0.03), 0 1px 2px rgba(0, 0, 0, 0.02);
}

.settings-section {
  display: flex;
  flex-direction: column;
}

.section-title-row {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  margin-bottom: 1.25rem;
}

.section-icon-badge {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
}

.bg-blue { background-color: var(--apple-blue-light); }
.bg-purple { background-color: #F6EDFC; }

.section-heading {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.section-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
}

/* Apple Inset Settings Group */
.apple-settings-group {
  background: var(--bg-primary);
  border-radius: var(--radius-lg);
  border: 1px solid rgba(0, 0, 0, 0.06);
  overflow: hidden;
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  gap: 1.5rem;
}

.flex-col-item {
  flex-direction: column;
  align-items: flex-start;
}

.setting-info {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.setting-name {
  font-size: 0.92rem;
  font-weight: 600;
  color: var(--text-primary);
}

.setting-hint {
  font-size: 0.8rem;
  color: var(--text-muted);
  line-height: 1.4;
}

.setting-divider {
  height: 1px;
  background-color: rgba(0, 0, 0, 0.05);
  margin: 0 1.5rem;
}

/* Mode Selection Cards */
.mode-selection-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
  width: 100%;
  max-width: 440px;
}

.apple-choice-card {
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: var(--radius-md);
  padding: 0.85rem 1rem;
  cursor: pointer;
  transition: all 0.2s var(--apple-ease);
  background: var(--white);
}

.apple-choice-card:hover {
  border-color: rgba(0, 113, 227, 0.3);
}

.apple-choice-card.active {
  border-color: var(--apple-blue);
  background: rgba(0, 113, 227, 0.04);
  box-shadow: 0 0 0 1px var(--apple-blue);
}

.hidden-radio {
  display: none;
}

.choice-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.25rem;
}

.choice-title-row strong {
  font-size: 0.88rem;
  color: var(--text-primary);
}

.choice-check {
  font-size: 0.85rem;
  color: var(--apple-blue);
  font-weight: 700;
}

.choice-sub {
  font-size: 0.75rem;
  color: var(--text-muted);
  line-height: 1.35;
  display: block;
}

.apple-select-input {
  min-width: 200px;
}

.apple-textarea {
  width: 100%;
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: var(--radius-md);
  padding: 0.85rem 1rem;
  font-family: inherit;
  font-size: 0.88rem;
  color: var(--text-primary);
  background: var(--white);
  line-height: 1.5;
  outline: none;
  transition: var(--transition-fast);
  resize: vertical;
}

.apple-textarea:focus {
  border-color: var(--apple-blue);
  box-shadow: 0 0 0 3px rgba(0, 113, 227, 0.15);
}

.api-key-input-wrapper {
  position: relative;
  width: 100%;
  display: flex;
  align-items: center;
}

.apple-input-text {
  width: 100%;
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: var(--radius-md);
  padding: 0.75rem 2.5rem 0.75rem 1rem;
  font-size: 0.88rem;
  color: var(--text-primary);
  background: var(--white);
  outline: none;
  transition: var(--transition-fast);
}

.apple-input-text:focus {
  border-color: var(--apple-blue);
  box-shadow: 0 0 0 3px rgba(0, 113, 227, 0.15);
}

.eye-toggle-btn {
  position: absolute;
  right: 0.75rem;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.eye-toggle-btn:hover {
  color: var(--text-primary);
}

.font-mono {
  font-family: var(--font-mono);
}

/* Actions */
.settings-actions {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 0.75rem;
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 1px solid rgba(0, 0, 0, 0.06);
}

.apple-btn-secondary {
  background: var(--white);
  border: 1px solid rgba(0, 0, 0, 0.1);
  padding: 0.65rem 1.25rem;
  border-radius: var(--radius-full);
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-secondary);
  cursor: pointer;
  transition: var(--transition-fast);
}

.apple-btn-secondary:hover {
  background: rgba(0, 0, 0, 0.04);
  color: var(--text-primary);
}

.apple-primary-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--apple-blue);
  color: #FFFFFF;
  border: 1px solid rgba(255, 255, 255, 0.2);
  padding: 0.65rem 1.4rem;
  border-radius: var(--radius-full);
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s var(--apple-ease);
  box-shadow: 0 4px 12px rgba(0, 113, 227, 0.25);
}

.apple-primary-btn:hover {
  background: var(--apple-blue-hover);
  transform: translateY(-1px);
}

.mt-8 { margin-top: 2rem; }
.mb-2 { margin-bottom: 0.5rem; }

@media (max-width: 768px) {
  .settings-card {
    padding: 1.5rem 1.25rem;
  }

  .setting-item {
    flex-direction: column;
    align-items: flex-start;
    padding: 1rem;
  }

  .mode-selection-grid {
    grid-template-columns: 1fr;
    max-width: 100%;
  }
}
</style>
