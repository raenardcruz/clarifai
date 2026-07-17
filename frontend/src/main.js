import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import axios from 'axios'
import PrimeVue from 'primevue/config'
import Aura from '@primevue/themes/aura'
import ToastService from 'primevue/toastservice'

import './style.css' // Design system tokens and aesthetics
import 'primeicons/primeicons.css'

// Configure Axios defaults
axios.defaults.baseURL = import.meta.env.VITE_API_URL || ''

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(PrimeVue, {
  theme: {
    preset: Aura,
    options: {
      darkModeSelector: '.dark' // If we have dark mode
    }
  }
})
app.use(ToastService)

app.mount('#app')
