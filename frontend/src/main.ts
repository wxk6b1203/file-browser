import './assets/main.css'

import { createApp } from 'vue'
import { createI18n } from 'vue-i18n'
import { createPinia } from 'pinia'
import messages from '@intlify/unplugin-vue-i18n/messages'


import App from './App.vue'
import router from './router'

const app = createApp(App)
const i18n = createI18n({
    legacy: false,
    locale: 'zh',
    fallbackLocale: 'en',
    messages: messages,
})
app.use(createPinia())
app.use(router)
app.use(i18n)

app.mount('#app')
