import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import { i18n } from './i18n'
import router from './router'
import { useThemeStore } from './stores/theme'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)

// Initialize theme before mount so <html> gets the correct class immediately
useThemeStore()

app.use(router)
app.use(i18n)

app.mount('#app')
