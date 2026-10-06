import axios from 'axios'
import { Capacitor } from '@capacitor/core'
import { Preferences } from '@capacitor/preferences'

const SERVER_URL_KEY = 'note_taker_server_url'
const DEFAULT_URL = import.meta.env.VITE_API_BASE_URL || (Capacitor.isNativePlatform() ? 'http://localhost:8000' : '')

export async function initApiConfig() {
  let savedUrl = null
  try {
    const { value } = await Preferences.get({ key: SERVER_URL_KEY })
    savedUrl = value
  } catch {
    savedUrl = localStorage.getItem(SERVER_URL_KEY)
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
  localStorage.setItem(SERVER_URL_KEY, cleanUrl)
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
