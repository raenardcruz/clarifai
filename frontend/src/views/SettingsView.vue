<template>
  <div class="settings-page">
    <div class="page-header mb-8">
      <h1 class="text-2xl font-bold mb-1">System Settings</h1>
      <p class="text-sm text-gray-500">Configure global AI behavior and application rules</p>
    </div>

    <div class="settings-card card">
      <h3 class="font-bold mb-6 flex items-center gap-2"><SettingsIcon size="20" class="text-primary"/> AI Summarization Settings</h3>
      
      <form @submit.prevent="saveSettings">
        <div class="form-group mb-6">
          <label class="block font-semibold mb-2">Summarization Mode</label>
          <div class="radio-group">
            <label class="radio-label">
              <RadioButton v-model="settings.ai_summarization_mode" inputId="mode-auto" name="ai_summarization_mode" value="auto" />
              <div class="radio-content" @click="settings.ai_summarization_mode = 'auto'">
                <strong>Automatic</strong>
                <p class="text-sm text-gray-500">Generate summary immediately after transcription.</p>
              </div>
            </label>
            <label class="radio-label">
              <RadioButton v-model="settings.ai_summarization_mode" inputId="mode-manual" name="ai_summarization_mode" value="manual" />
              <div class="radio-content" @click="settings.ai_summarization_mode = 'manual'">
                <strong>Manual</strong>
                <p class="text-sm text-gray-500">Wait for user to manually trigger summarization (allows editing transcript first).</p>
              </div>
            </label>
          </div>
        </div>

        <div class="form-group mb-6">
          <label class="block font-semibold mb-2">Ollama LLM Model</label>
          <p class="text-sm text-gray-500 mb-2">Choose the local Ollama LLM to use for generating meeting titles and summaries.</p>
          <Select 
            v-model="settings.ollama_model" 
            :options="availableModels" 
            placeholder="Select a Model"
            class="w-full"
          />
        </div>

        <div class="form-group mb-6">
          <label class="block font-semibold mb-2">Executive Summary Prompt</label>
          <p class="text-sm text-gray-500 mb-2">Define the instructions given to the LLM to generate the summary.</p>
          <Textarea 
            v-model="settings.executive_summary_prompt" 
            class="w-full h-32"
            autoResize
            placeholder="E.g. Provide a concise executive summary..."
          ></Textarea>
        </div>

        <div class="form-group mb-6">
          <label class="block font-semibold mb-2">Speechmatics API Key</label>
          <p class="text-sm text-gray-500 mb-2">Provide your Speechmatics API Key to enable audio transcription. The key is masked to protect its secrecy.</p>
          <InputText 
            v-model="settings.speechmatics_api_key" 
            type="password"
            class="w-full"
            placeholder="Enter Speechmatics API Key"
          />
        </div>

        <div class="flex justify-end gap-4 mt-8 pt-4 border-t border-gray-100">
          <Button type="button" label="Cancel" severity="secondary" @click="fetchSettings" />
          <Button type="submit" :loading="loading" label="Save Settings" icon="pi pi-save" />
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useToast } from 'primevue/usetoast'
import { Settings as SettingsIcon } from '@lucide/vue'
import RadioButton from 'primevue/radiobutton'
import Textarea from 'primevue/textarea'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'

const toast = useToast()
const settings = ref({
  ai_summarization_mode: 'auto',
  executive_summary_prompt: 'Analyze the following transcript and write a detailed, cohesive executive summary in a narrative paragraph format. Synthesize the meeting\'s core purpose, main arguments, and final outcomes into smooth, professional prose, explicitly attributing key ideas, decisions, and viewpoints to specific speakers by name (or speaker identifier) directly within the flow of the text. Ensure the summary is comprehensive and captures concrete details, specific project names, and actionable next steps, but deliver it entirely as a sequence of well-structured paragraphs without using any bullet points, lists, or tables. Here is the transcript: [PASTE TRANSCRIPT HERE]',
  speechmatics_api_key: '',
  ollama_model: 'gemma4:12b-mlx'
})
const availableModels = ref([])
const loading = ref(false)

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
    const res = await axios.put('/api/settings', settings.value)
    settings.value = res.data
    toast.add({ severity: 'success', summary: 'Saved', detail: 'Settings saved successfully.', life: 3000 })
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Error', detail: 'Failed to save settings.', life: 5000 })
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchModels()
  fetchSettings()
})
</script>

<style scoped>
.settings-page {
  max-width: 800px;
}

.radio-group {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.radio-label {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 1rem;
  border: 1px solid var(--neutral-200);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: var(--transition);
}

.radio-label:hover {
  border-color: var(--primary);
  background: var(--neutral-50);
}

.radio-label input[type="radio"] {
  margin-top: 0.25rem;
}

.radio-label input[type="radio"]:checked + .radio-content strong {
  color: var(--primary);
}

.radio-label:has(input[type="radio"]:checked) {
  border-color: var(--primary);
  background: var(--neutral-50);
}

.h-32 { height: 8rem; }
.resize-y { resize: vertical; }
.block { display: block; }
.font-semibold { font-weight: 600; }
.border-t { border-top-width: 1px; border-top-style: solid; }
.border-gray-100 { border-color: #F3F4F6; }
.pt-4 { padding-top: 1rem; }
.mt-8 { margin-top: 2rem; }
</style>
