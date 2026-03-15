import './assets/main.css'

import { createApp } from 'vue'
import { createI18n } from 'vue-i18n'
import { createPinia } from 'pinia'
import messages from '@intlify/unplugin-vue-i18n/messages'

import App from './App.vue'
import router from './router'
import { useThemeStore } from './stores/theme'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)

// Initialize theme before mount so <html> gets the correct class immediately
useThemeStore()

const i18n = createI18n({
    legacy: false,
    locale: 'zh',
    fallbackLocale: 'en',
    messages: messages,
})
app.use(router)
app.use(i18n)

app.mount('#app')
