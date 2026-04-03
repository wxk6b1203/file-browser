import { ref, computed, watch } from 'vue'
import { defineStore } from 'pinia'

// ─── Theme metadata ──────────────────────────────────────────

/** Describes a single registered theme */
export interface ThemeDefinition {
  /** Unique identifier, matches [data-theme="<id>"] in CSS */
  id: string
  /** Display name (can be i18n key) */
  label: string
  /** Whether the theme is a dark variant (drives Element Plus dark helpers & class) */
  dark: boolean
}

/** All built-in themes. To add one: create the CSS file, then add an entry here. */
export const BUILTIN_THEMES: ThemeDefinition[] = [
  { id: 'light',         label: '浅色',           dark: false },
  { id: 'dark',          label: '深色',           dark: true  },
  { id: '2026-dark',     label: '2026 Dark',      dark: true  },
  { id: 'vscode-dark',   label: 'VS Code Dark',   dark: true  },
  { id: 'islands-light', label: 'Islands Light',   dark: false },
  { id: 'islands-dark',  label: 'Islands Dark',    dark: true  },
]

/**
 * Special value: follow system `prefers-color-scheme`.
 * When system is "dark" → uses the first built-in dark theme ('dark').
 * When system is "light" → uses 'light'.
 */
export const SYSTEM_THEME = 'system' as const

export type ThemeMode = typeof SYSTEM_THEME | string   // theme id or 'system'

// ─── Helpers ─────────────────────────────────────────────────

const STORAGE_KEY = 'app-theme'

function resolveSystemPreference(): 'light' | 'dark' {
  if (typeof window === 'undefined') return 'light'
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function readStoredMode(): ThemeMode {
  if (typeof window === 'undefined') return 'light'
  const stored = localStorage.getItem(STORAGE_KEY)
  if (!stored) return 'light'
  // Validate the stored value is a known theme or 'system'
  if (stored === SYSTEM_THEME) return SYSTEM_THEME
  if (BUILTIN_THEMES.some((t) => t.id === stored)) return stored
  return 'light'
}

// ─── Store ───────────────────────────────────────────────────

export const useThemeStore = defineStore('theme', () => {
  // ── State ────────────────────────────────────────────────
  const mode = ref<ThemeMode>(readStoredMode())
  const systemPreference = ref<'light' | 'dark'>(resolveSystemPreference())

  /** All available themes */
  const themes = ref<ThemeDefinition[]>([...BUILTIN_THEMES])

  // ── Getters ──────────────────────────────────────────────

  /**
   * The concrete theme id actually applied.
   * 'system' → resolves to 'light' or 'dark' based on OS preference.
   */
  const resolvedTheme = computed<string>(() => {
    if (mode.value === SYSTEM_THEME) {
      return systemPreference.value // 'light' or 'dark'
    }
    return mode.value
  })

  /** The ThemeDefinition object for the current resolved theme */
  const currentTheme = computed<ThemeDefinition>(() => {
    return themes.value.find((t) => t.id === resolvedTheme.value)
      ?? themes.value[0]! // fallback to first
  })

  /** Is the active theme a dark variant? */
  const isDark = computed(() => currentTheme.value.dark)

  // ── Actions ──────────────────────────────────────────────

  function setTheme(newMode: ThemeMode) {
    mode.value = newMode
  }

  /** Cycle: current → next theme in list, skipping 'system' */
  function next() {
    const ids = themes.value.map((t) => t.id)
    const idx = ids.indexOf(resolvedTheme.value)
    const nextIdx = (idx + 1) % ids.length
    mode.value = ids[nextIdx]!
  }

  /** Toggle between light ↔ dark (picks first dark / first light theme) */
  function toggle() {
    if (isDark.value) {
      mode.value = themes.value.find((t) => !t.dark)?.id ?? 'light'
    } else {
      mode.value = themes.value.find((t) => t.dark)?.id ?? 'dark'
    }
  }

  // ── Side Effects ─────────────────────────────────────────

  function applyTheme(themeId: string, dark: boolean) {
    const root = document.documentElement
    root.setAttribute('data-theme', themeId)
    // Toggle .dark class (useful for Tailwind dark: variant & Element Plus)
    root.classList.toggle('dark', dark)
  }

  watch(
    [resolvedTheme, isDark] as const,
    ([id, dark]) => applyTheme(id, dark),
    { immediate: true },
  )

  watch(mode, (val) => localStorage.setItem(STORAGE_KEY, val))

  // Listen for OS preference changes
  if (typeof window !== 'undefined') {
    window.matchMedia('(prefers-color-scheme: dark)')
      .addEventListener('change', (e) => {
        systemPreference.value = e.matches ? 'dark' : 'light'
      })
  }

  return {
    mode,
    themes,
    resolvedTheme,
    currentTheme,
    isDark,
    setTheme,
    toggle,
    next,
  }
})
