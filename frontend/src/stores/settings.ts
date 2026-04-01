import { ref } from 'vue'
import { defineStore } from 'pinia'
import { ElMessage } from 'element-plus'
import { GetAppConfig, SaveAppConfig } from '../../wailsjs/go/main/App'
import { config } from '../../wailsjs/go/models'
import { i18n } from '@/i18n'
import { BUILTIN_THEMES, SYSTEM_THEME, useThemeStore, type ThemeMode } from './theme'

const SUPPORTED_LOCALES = new Set(['zh', 'en'])

function cloneAppConfig(value: config.AppConfig | null | undefined) {
  if (!value) return null
  return JSON.parse(JSON.stringify(value)) as config.AppConfig
}

function normalizeLocale(value?: string) {
  const locale = String(value ?? '').trim().toLowerCase()
  return SUPPORTED_LOCALES.has(locale) ? locale : 'zh'
}

function normalizeTheme(value?: string) {
  const theme = String(value ?? '').trim()
  if (theme === SYSTEM_THEME) return SYSTEM_THEME
  if (BUILTIN_THEMES.some((item) => item.id === theme)) return theme
  return SYSTEM_THEME
}

export const useSettingsStore = defineStore('settings', () => {
  const theme = useThemeStore()
  const ready = ref(false)
  const loading = ref(false)
  const saving = ref(false)
  const appConfig = ref<config.AppConfig | null>(null)
  let pendingHydration: Promise<config.AppConfig | null> | null = null

  function applyUISettings(next: config.AppConfig | null | undefined) {
    if (!next) return

    const locale = normalizeLocale(next.ui?.locale || next.app?.locale)
    const themeMode = normalizeTheme(next.ui?.theme || next.app?.theme)

    i18n.global.locale.value = locale
    theme.setTheme(themeMode as ThemeMode)
  }

  async function hydrate(force = false) {
    if (ready.value && !force) {
      applyUISettings(appConfig.value)
      return cloneAppConfig(appConfig.value)
    }
    if (pendingHydration) {
      return pendingHydration
    }

    loading.value = true
    pendingHydration = (async () => {
      try {
        const next = cloneAppConfig(await GetAppConfig())
        appConfig.value = next
        ready.value = true
        applyUISettings(next)
        return cloneAppConfig(next)
      } finally {
        loading.value = false
        pendingHydration = null
      }
    })()

    return pendingHydration
  }

  async function saveConfig(next: config.AppConfig) {
    saving.value = true
    try {
      const saved = cloneAppConfig(await SaveAppConfig(next))
      appConfig.value = saved
      ready.value = true
      applyUISettings(saved)
      ElMessage.success(i18n.global.t('workspace.settings.saved'))
      return cloneAppConfig(saved)
    } finally {
      saving.value = false
    }
  }

  return {
    ready,
    loading,
    saving,
    appConfig,
    hydrate,
    saveConfig,
  }
})
