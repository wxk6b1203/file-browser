import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import { i18n } from './i18n'
import router from './router'
import { useThemeStore } from './stores/theme'

function blockNativeContextMenu(event: MouseEvent) {
  const target = event.target as HTMLElement | null
  if (target?.closest('[data-allow-native-contextmenu="true"]')) {
    return
  }
  event.preventDefault()
}

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)

// Initialize theme before mount so <html> gets the correct class immediately
useThemeStore()

app.use(router)
app.use(i18n)

window.addEventListener('contextmenu', blockNativeContextMenu, { capture: true })

app.mount('#app')
