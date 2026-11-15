<template>
  <div class="settings-section">
    <div class="integrations-section-header">
      <div class="section-title-row mb-0">
        <div class="section-icon-badge bg-teal">
          <Plug size="16" color="#0891B2" />
        </div>
        <div>
          <h3 class="section-heading">Third-Party Note-Taker Integrations</h3>
          <p class="section-desc">
            Import meeting transcripts and summaries from Fathom, Fireflies.ai, and Otter.ai via Direct API or MCP
          </p>
        </div>
      </div>

      <button
        type="button"
        class="apple-btn-secondary sync-all-btn"
        :disabled="!hasAnyEnabled || isSyncingAll || syncingProvider !== null"
        @click="$emit('sync', 'all')"
        title="Save settings and sync all enabled integrations"
      >
        <RefreshCw size="14" :class="{ 'animate-spin': isSyncingAll }" />
        <span>{{ isSyncingAll ? 'Syncing All...' : 'Sync All Enabled' }}</span>
      </button>
    </div>

    <div class="integrations-stack">
      <div
        v-for="provider in providers"
        :key="provider.key"
        class="apple-settings-group integration-card"
        :class="{ 'is-enabled': isEnabled(provider.key) }"
      >
        <!-- Provider Header & Enable Toggle -->
        <div class="setting-item integration-header-row">
          <div class="integration-identity">
            <div class="provider-logo-box" :style="{ backgroundColor: provider.bgColor }">
              <img
                :src="provider.iconUrl"
                :alt="provider.name"
                class="provider-logo-img"
              />
            </div>

            <div class="setting-info">
              <div class="provider-title-line">
                <span class="setting-name">{{ provider.name }}</span>
                <span v-if="isEnabled(provider.key)" class="enabled-pill">Enabled</span>
              </div>
              <span class="setting-hint">{{ provider.description }}</span>
            </div>
          </div>

          <label class="apple-switch" :title="`Enable ${provider.name} integration`">
            <input
              type="checkbox"
              :checked="isEnabled(provider.key)"
              @change="setEnabled(provider.key, $event.target.checked)"
            />
            <span class="apple-switch-slider"></span>
          </label>
        </div>

        <!-- Expanded Configuration & Step-by-Step Guide (shown when enabled) -->
        <div v-if="isEnabled(provider.key)" class="integration-body">
          <div class="setting-divider"></div>

          <!-- Step-by-Step Setup Instructions -->
          <div class="setup-guide-box">
            <div class="setup-guide-header">
              <BookOpen size="15" class="guide-icon" />
              <span>Step-by-Step Setup Guide ({{ getMethod(provider.key) === 'mcp' ? 'MCP Mode' : 'Direct API Mode' }})</span>
            </div>
            <ol class="setup-steps-list">
              <li
                v-for="(step, idx) in getGuideSteps(provider.key)"
                :key="idx"
                v-html="step"
              ></li>
            </ol>
          </div>

          <!-- Connection Method (API vs MCP) -->
          <div class="setting-item">
            <div class="setting-info">
              <span class="setting-name">Connection Method</span>
              <span class="setting-hint">Choose between {{ provider.apiLabel }} or Model Context Protocol (MCP)</span>
            </div>
            <div class="mode-selection-grid">
              <label
                class="apple-choice-card"
                :class="{ active: getMethod(provider.key) === 'api' }"
              >
                <input
                  type="radio"
                  :name="`${provider.key}_method`"
                  value="api"
                  :checked="getMethod(provider.key) === 'api'"
                  @change="setMethod(provider.key, 'api')"
                  class="hidden-radio"
                />
                <div class="choice-content">
                  <div class="choice-title-row">
                    <strong>Direct API</strong>
                    <span class="choice-check" v-if="getMethod(provider.key) === 'api'">✓</span>
                  </div>
                  <span class="choice-sub">{{ provider.apiLabel }}</span>
                </div>
              </label>

              <label
                class="apple-choice-card"
                :class="{ active: getMethod(provider.key) === 'mcp' }"
              >
                <input
                  type="radio"
                  :name="`${provider.key}_method`"
                  value="mcp"
                  :checked="getMethod(provider.key) === 'mcp'"
                  @change="setMethod(provider.key, 'mcp')"
                  class="hidden-radio"
                />
                <div class="choice-content">
                  <div class="choice-title-row">
                    <strong>MCP Server</strong>
                    <span class="choice-check" v-if="getMethod(provider.key) === 'mcp'">✓</span>
                  </div>
                  <span class="choice-sub">Model Context Protocol (JSON-RPC)</span>
                </div>
              </label>
            </div>
          </div>

          <div class="setting-divider"></div>

          <!-- API Key / Bearer Token Input -->
          <div class="setting-item flex-col-item">
            <div class="setting-info mb-2">
              <span class="setting-name">
                {{ getMethod(provider.key) === 'mcp' ? `${provider.name} Bearer / Access Token` : `${provider.name} API Key` }}
              </span>
              <span class="setting-hint">{{ provider.keyHint }}</span>
            </div>
            <div class="api-key-input-wrapper">
              <input
                :type="visibleKeys[provider.key] ? 'text' : 'password'"
                :value="getApiKey(provider.key)"
                @input="setApiKey(provider.key, $event.target.value)"
                class="apple-input-text font-mono"
                :placeholder="provider.keyPlaceholder"
              />
              <button
                type="button"
                class="eye-toggle-btn"
                @click="visibleKeys[provider.key] = !visibleKeys[provider.key]"
                title="Toggle visibility"
              >
                <EyeOff size="16" v-if="visibleKeys[provider.key]" />
                <Eye size="16" v-else />
              </button>
            </div>
          </div>

          <!-- MCP URL (for MCP mode) OR Otter API Base URL -->
          <template v-if="getMethod(provider.key) === 'mcp'">
            <div class="setting-divider"></div>
            <div class="setting-item flex-col-item">
              <div class="setting-info mb-2">
                <span class="setting-name">{{ provider.name }} MCP Server URL</span>
                <span class="setting-hint">
                  Streamable HTTP MCP endpoint ({{ provider.defaultMcpUrl || 'e.g. http://localhost:3100/mcp or hosted MCP URL' }})
                </span>
              </div>
              <input
                type="text"
                :value="getMcpUrl(provider.key)"
                @input="setMcpUrl(provider.key, $event.target.value)"
                class="apple-input-text font-mono"
                :placeholder="provider.defaultMcpUrl || 'https://your-mcp-endpoint/mcp'"
              />
            </div>
          </template>

          <template v-else-if="provider.key === 'otter'">
            <div class="setting-divider"></div>
            <div class="setting-item flex-col-item">
              <div class="setting-info mb-2">
                <span class="setting-name">Otter.ai API Base URL</span>
                <span class="setting-hint">Default official endpoint is https://api.otter.ai</span>
              </div>
              <input
                type="text"
                :value="settings.otter_api_base_url"
                @input="updateField('otter_api_base_url', $event.target.value)"
                class="apple-input-text font-mono"
                placeholder="https://api.otter.ai"
              />
            </div>
          </template>

          <div class="setting-divider"></div>

          <!-- Sync Actions & Status Row -->
          <div class="setting-item integration-footer-row">
            <div class="sync-status-block">
              <div v-if="getLastSyncAt(provider.key)" class="sync-time-row">
                <CheckCircle2 v-if="!isErrorStatus(getLastSyncStatus(provider.key))" size="14" class="text-green" />
                <AlertCircle v-else size="14" class="text-red" />
                <span>Last synced {{ formatSyncTime(getLastSyncAt(provider.key)) }}</span>
              </div>
              <span
                v-if="getLastSyncStatus(provider.key)"
                class="sync-status-msg"
                :class="{ 'is-error': isErrorStatus(getLastSyncStatus(provider.key)) }"
              >
                {{ getLastSyncStatus(provider.key) }}
              </span>
              <span v-else class="sync-status-msg">Not synced yet</span>

              <!-- Inline Test Result -->
              <div
                v-if="testResults[provider.key]"
                class="test-result-pill"
                :class="testResults[provider.key].ok ? 'test-ok' : 'test-fail'"
              >
                <span>{{ testResults[provider.key].message }}</span>
                <span v-if="testResults[provider.key].tools?.length" class="test-tools-list">
                  Tools: {{ testResults[provider.key].tools.slice(0, 6).join(', ') }}{{ testResults[provider.key].tools.length > 6 ? '...' : '' }}
                </span>
              </div>
            </div>

            <div class="integration-actions">
              <button
                type="button"
                class="apple-btn-secondary"
                :disabled="testingProvider === provider.key || syncingProvider === provider.key || isSyncingAll"
                @click="testConnection(provider.key)"
              >
                <Loader2 v-if="testingProvider === provider.key" size="14" class="animate-spin" />
                <Wifi v-else size="14" />
                <span>{{ testingProvider === provider.key ? 'Testing...' : 'Test Connection' }}</span>
              </button>

              <button
                type="button"
                class="apple-primary-btn sync-now-btn"
                :disabled="syncingProvider === provider.key || isSyncingAll"
                @click="$emit('sync', provider.key)"
              >
                <RefreshCw size="14" :class="{ 'animate-spin': syncingProvider === provider.key }" />
                <span>{{ syncingProvider === provider.key ? 'Syncing...' : 'Sync Now' }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, computed, ref } from 'vue'
import axios from 'axios'
import moment from 'moment'
import { Plug, RefreshCw, BookOpen, Eye, EyeOff, Wifi, Loader2, CheckCircle2, AlertCircle } from '@lucide/vue'

const props = defineProps({
  settings: {
    type: Object,
    required: true
  },
  syncingProvider: {
    type: String,
    default: null
  },
  isSyncingAll: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:settings', 'sync'])

const visibleKeys = reactive({
  fathom: false,
  fireflies: false,
  otter: false
})

const testingProvider = ref(null)
const testResults = reactive({
  fathom: null,
  fireflies: null,
  otter: null
})

const providers = [
  {
    key: 'fathom',
    name: 'Fathom',
    description: 'Sync recorded meetings, speaker transcripts, summaries, and action items from Fathom',
    iconUrl: '/logos/fathom-icon.svg',
    color: '#0891B2',
    bgColor: 'rgba(8, 145, 178, 0.12)',
    apiLabel: 'REST API v1 (X-Api-Key)',
    keyHint: 'Copied from Fathom under Settings → API Access (only the API Key is needed; ignore the Webhook Secret)',
    keyPlaceholder: 'Enter your Fathom API Key',
    defaultMcpUrl: 'https://api.fathom.ai/mcp'
  },
  {
    key: 'fireflies',
    name: 'Fireflies.ai',
    description: 'Import Fireflies meeting transcripts, AI overviews, outlines, and action items',
    iconUrl: '/logos/fireflies-icon.svg',
    color: '#9333EA',
    bgColor: 'rgba(147, 51, 234, 0.12)',
    apiLabel: 'GraphQL API (Bearer Token)',
    keyHint: 'Copied from Fireflies.ai under Settings → Developer Settings',
    keyPlaceholder: 'Enter your Fireflies.ai API Key',
    defaultMcpUrl: 'https://api.fireflies.ai/mcp'
  },
  {
    key: 'otter',
    name: 'Otter.ai',
    description: 'Fetch Otter conversations, speaker-attributed transcripts, outlines, and summaries',
    iconUrl: '/logos/otter.svg',
    color: '#2563EB',
    bgColor: 'rgba(37, 99, 235, 0.12)',
    apiLabel: 'Public REST API (Enterprise Bearer)',
    keyHint: 'Created in Otter.ai under Account Settings → Integrations → Developer',
    keyPlaceholder: 'Enter your Otter.ai API Key / Access Token',
    defaultMcpUrl: ''
  }
]

const updateField = (field, value) => {
  emit('update:settings', {
    ...props.settings,
    [field]: value
  })
}

const isEnabled = (key) => !!props.settings[`${key}_enabled`]
const setEnabled = (key, val) => updateField(`${key}_enabled`, val)

const getMethod = (key) => props.settings[`${key}_method`] || 'api'
const setMethod = (key, val) => {
  testResults[key] = null
  updateField(`${key}_method`, val)
}

const getApiKey = (key) => props.settings[`${key}_api_key`] || ''
const setApiKey = (key, val) => updateField(`${key}_api_key`, val)

const getMcpUrl = (key) => props.settings[`${key}_mcp_url`] || ''
const setMcpUrl = (key, val) => updateField(`${key}_mcp_url`, val)

const getLastSyncAt = (key) => props.settings[`${key}_last_sync_at`] || null
const getLastSyncStatus = (key) => props.settings[`${key}_last_sync_status`] || ''

const hasAnyEnabled = computed(() =>
  providers.some((p) => isEnabled(p.key))
)

const isErrorStatus = (status) =>
  typeof status === 'string' && status.toLowerCase().startsWith('error:')

const formatSyncTime = (iso) => {
  if (!iso) return ''
  return moment(iso).fromNow()
}

const getGuideSteps = (key) => {
  const method = getMethod(key)
  if (key === 'fathom') {
    if (method === 'mcp') {
      return [
        'Sign in to <a href="https://fathom.video" target="_blank" rel="noopener noreferrer">fathom.video</a> and confirm MCP access is available for your account.',
        'Keep the MCP Server URL set to <code>https://api.fathom.ai/mcp</code> (or enter your local OAuth MCP bridge URL).',
        'Obtain your Fathom Bearer token or API Key from <strong>Settings → My Settings → API Access</strong> (ignore any Webhook Secret) and paste it below.',
        'Click <strong>Test Connection</strong> to verify discovered MCP tools, then click <strong>Sync Now</strong> to import meetings.'
      ]
    }
    return [
      'Sign in to your Fathom account at <a href="https://fathom.video" target="_blank" rel="noopener noreferrer">fathom.video</a>.',
      'Open <strong>Settings</strong> (top-right profile menu) and scroll to <strong>My Settings → API Access</strong>.',
      'Click <strong>Add</strong> (or <strong>Generate API Key</strong>), enter a name (e.g. <em>ClarifAi</em>), and click <strong>Create API Client</strong>.',
      'Fathom will display both an <strong>API Key</strong> and a <strong>Webhook Secret</strong>. Copy only the <strong>API Key</strong> — the Webhook Secret is not needed because ClarifAi pulls meetings directly via API.',
      'Paste the <strong>API Key</strong> below and click <strong>Test Connection</strong> or <strong>Sync Now</strong>.'
    ]
  }

  if (key === 'fireflies') {
    if (method === 'mcp') {
      return [
        'Log in to <a href="https://app.fireflies.ai" target="_blank" rel="noopener noreferrer">app.fireflies.ai</a>.',
        'Go to <strong>Integrations</strong> or <strong>Settings → Developer Settings</strong> and copy your <strong>API Key</strong>.',
        'Ensure the MCP Server URL is set to <code>https://api.fireflies.ai/mcp</code> and paste your API key as the Bearer token.',
        'Click <strong>Test Connection</strong> to verify MCP tools, then click <strong>Sync Now</strong> to import transcripts and summaries.'
      ]
    }
    return [
      'Log in to your Fireflies.ai workspace at <a href="https://app.fireflies.ai" target="_blank" rel="noopener noreferrer">app.fireflies.ai</a>.',
      'Navigate to <strong>Integrations</strong> in the left sidebar → <strong>Fireflies API</strong> (or <strong>Settings → Developer Settings</strong>).',
      'Copy your personal <strong>API Key</strong>.',
      'Paste the key into the <strong>Fireflies.ai API Key</strong> field below.',
      'Click <strong>Test Connection</strong> or <strong>Sync Now</strong> to import your transcripts, speaker sentences, and AI summaries.'
    ]
  }

  // Otter.ai
  if (method === 'mcp') {
    return [
      'Set up an Otter.ai MCP HTTP endpoint (for example, an <code>otter-agent-cli</code> HTTP MCP bridge, Composio, or Zapier MCP connector).',
      'Enter the <strong>Otter.ai MCP Server URL</strong> below (e.g. <code>http://localhost:3100/mcp</code> or your hosted MCP URL).',
      'Paste your Otter access token or MCP Bearer token in the token field (if required by your MCP endpoint).',
      'Click <strong>Test Connection</strong> to inspect exposed MCP tools, then click <strong>Sync Now</strong> to fetch conversations and summaries.'
    ]
  }
  return [
    'Sign in to your <strong>Otter.ai Enterprise</strong> workspace at <a href="https://otter.ai" target="_blank" rel="noopener noreferrer">otter.ai</a>.',
    'Open <strong>Account Settings → Integrations → Developer</strong>.',
    'Click <strong>Create key</strong> and copy the generated Bearer API token immediately (Otter only displays it once).',
    'Paste the token into the <strong>Otter.ai API Key</strong> field below and keep <code>https://api.otter.ai</code> as the API Base URL.',
    'Click <strong>Test Connection</strong> or <strong>Sync Now</strong> to import conversations, transcripts, and summaries.'
  ]
}

const testConnection = async (providerKey) => {
  testingProvider.value = providerKey
  testResults[providerKey] = null
  try {
    const method = getMethod(providerKey)
    const endpointUrl =
      method === 'mcp'
        ? getMcpUrl(providerKey)
        : providerKey === 'otter'
          ? settingsField('otter_api_base_url')
          : ''

    const res = await axios.post(`/api/integrations/${providerKey}/test`, {
      method,
      api_key: getApiKey(providerKey),
      endpoint_url: endpointUrl
    })
    testResults[providerKey] = res.data
  } catch (err) {
    testResults[providerKey] = {
      ok: false,
      message: err.response?.data?.detail || err.message || 'Connection test failed'
    }
  } finally {
    testingProvider.value = null
  }
}

const settingsField = (name) => props.settings[name] || ''
</script>

<style scoped>
.mt-8 { margin-top: 2rem; }
.mb-2 { margin-bottom: 0.5rem; }

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
  flex-shrink: 0;
}

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
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
}

.apple-btn-secondary:hover:not(:disabled) {
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

.apple-primary-btn:hover:not(:disabled) {
  background: var(--apple-blue-hover);
  transform: translateY(-1px);
}

.integrations-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.25rem;
  flex-wrap: wrap;
}

.mb-0 {
  margin-bottom: 0 !important;
}

.bg-teal {
  background-color: rgba(8, 145, 178, 0.12);
}

.integrations-stack {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.integration-card {
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.integration-card.is-enabled {
  border-color: rgba(0, 113, 227, 0.22);
  box-shadow: 0 2px 10px rgba(0, 113, 227, 0.04);
}

.integration-header-row {
  padding: 1.1rem 1.4rem;
}

.integration-identity {
  display: flex;
  align-items: center;
  gap: 0.9rem;
}

.provider-logo-box {
  width: 38px;
  height: 38px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.provider-logo-img {
  max-width: 26px;
  max-height: 20px;
  width: auto;
  height: auto;
  object-fit: contain;
  display: block;
}

.provider-title-line {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.enabled-pill {
  font-size: 0.68rem;
  font-weight: 700;
  color: #1e8e3e;
  background: rgba(52, 199, 89, 0.14);
  padding: 0.12rem 0.5rem;
  border-radius: 9999px;
}

/* Apple Toggle Switch */
.apple-switch {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 26px;
  flex-shrink: 0;
  cursor: pointer;
}

.apple-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.apple-switch-slider {
  position: absolute;
  inset: 0;
  background-color: rgba(120, 120, 128, 0.28);
  border-radius: 9999px;
  transition: background-color 0.22s ease;
}

.apple-switch-slider::before {
  content: '';
  position: absolute;
  height: 22px;
  width: 22px;
  left: 2px;
  bottom: 2px;
  background-color: #ffffff;
  border-radius: 50%;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  transition: transform 0.22s ease;
}

.apple-switch input:checked + .apple-switch-slider {
  background-color: var(--apple-blue);
}

.apple-switch input:checked + .apple-switch-slider::before {
  transform: translateX(18px);
}

/* Setup Guide Box */
.setup-guide-box {
  margin: 1rem 1.5rem 0.25rem 1.5rem;
  padding: 1rem 1.15rem;
  border-radius: var(--radius-md);
  background: rgba(0, 113, 227, 0.045);
  border: 1px solid rgba(0, 113, 227, 0.15);
}

.setup-guide-header {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.82rem;
  font-weight: 700;
  color: var(--apple-blue);
  margin-bottom: 0.55rem;
}

.setup-steps-list {
  margin: 0;
  padding-left: 1.2rem;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 0.8rem;
  line-height: 1.45;
  color: var(--text-secondary);
}

.setup-steps-list :deep(a) {
  color: var(--apple-blue);
  font-weight: 600;
  text-decoration: underline;
}

.setup-steps-list :deep(code) {
  font-family: var(--font-mono);
  font-size: 0.76rem;
  background: rgba(0, 0, 0, 0.05);
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  color: var(--text-primary);
}

.integration-footer-row {
  flex-wrap: wrap;
  gap: 1rem;
}

.sync-status-block {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
  min-width: 220px;
}

.sync-time-row {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-primary);
}

.sync-status-msg {
  font-size: 0.78rem;
  color: var(--text-muted);
  line-height: 1.35;
}

.sync-status-msg.is-error {
  color: var(--apple-red);
}

.text-green { color: #1e8e3e; }
.text-red { color: var(--apple-red); }

.test-result-pill {
  margin-top: 0.35rem;
  padding: 0.45rem 0.7rem;
  border-radius: var(--radius-sm);
  font-size: 0.76rem;
  font-weight: 500;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.test-ok {
  background: rgba(52, 199, 89, 0.1);
  color: #166534;
  border: 1px solid rgba(52, 199, 89, 0.28);
}

.test-fail {
  background: rgba(255, 59, 48, 0.08);
  color: #b91c1c;
  border: 1px solid rgba(255, 59, 48, 0.24);
}

.test-tools-list {
  font-family: var(--font-mono);
  font-size: 0.72rem;
  opacity: 0.9;
}

.integration-actions {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  flex-wrap: wrap;
}

.apple-btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
}

.apple-btn-secondary:disabled,
.apple-primary-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

@media (max-width: 768px) {
  .setting-item {
    flex-direction: column;
    align-items: flex-start;
    padding: 1rem;
  }

  .mode-selection-grid {
    grid-template-columns: 1fr;
    max-width: 100%;
  }

  .setup-guide-box {
    margin: 0.75rem 1rem 0.25rem 1rem;
  }
}
</style>
