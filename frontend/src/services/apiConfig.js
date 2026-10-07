import axios from 'axios'
import { Capacitor } from '@capacitor/core'
import { Preferences } from '@capacitor/preferences'

const SERVER_URL_KEY = 'note_taker_server_url'
const DEFAULT_URL = import.meta.env.VITE_API_BASE_URL || (Capacitor.isNativePlatform() ? 'http://localhost:8000' : '')

// Immediately initialize axios.defaults.baseURL synchronously from localStorage or DEFAULT_URL
// to prevent initial requests from firing with an undefined/empty baseURL.
const savedSync = typeof localStorage !== 'undefined' ? localStorage.getItem(SERVER_URL_KEY) : null
const initialUrl = (savedSync || DEFAULT_URL || '').trim().replace(/\/+$/, '')
if (initialUrl) {
  axios.defaults.baseURL = initialUrl
}

export async function initApiConfig() {
  let savedUrl = null
  try {
    const { value } = await Preferences.get({ key: SERVER_URL_KEY })
    savedUrl = value
  } catch {
    savedUrl = typeof localStorage !== 'undefined' ? localStorage.getItem(SERVER_URL_KEY) : null
  }

  const effectiveUrl = savedUrl || DEFAULT_URL
  setApiBaseUrl(effectiveUrl)
  return effectiveUrl
}

export function getApiBaseUrl() {
  return axios.defaults.baseURL || ''
}

export async function setApiBaseUrl(url) {
  const cleanUrl = (url || '').trim().replace(/\/+$/, '')
  axios.defaults.baseURL = cleanUrl
  try {
    await Preferences.set({ key: SERVER_URL_KEY, value: cleanUrl })
  } catch {
    // fallback
  }
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem(SERVER_URL_KEY, cleanUrl)
  }
}

export async function checkBackendHealth() {
  try {
    const url = axios.defaults.baseURL ? `${axios.defaults.baseURL}/api/settings` : '/api/settings'
    const res = await axios.get(url, { timeout: 4000 })
    return { ok: true, status: res.status }
  } catch (err) {
    return { ok: false, error: err.message }
  }
}
