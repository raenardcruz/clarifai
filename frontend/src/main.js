import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import PrimeVue from 'primevue/config'
import Aura from '@primevue/themes/aura'
import ToastService from 'primevue/toastservice'

import './style.css' // Design system tokens and aesthetics
import 'primeicons/primeicons.css'

import { initApiConfig } from './services/apiConfig'
import { useOfflineRecordingsStore } from './stores/offlineRecordings'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(PrimeVue, {
  theme: {
    preset: Aura,
    options: {
      darkModeSelector: '.dark'
    }
  }
})
app.use(ToastService)

async function startApp() {
  try {
    await initApiConfig()
  } catch (err) {
    console.error('Failed to initialize API config:', err)
  }

  const offlineStore = useOfflineRecordingsStore(pinia)
  offlineStore.init()

  app.mount('#app')
}

startApp()
