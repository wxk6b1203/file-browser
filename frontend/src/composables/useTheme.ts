import { useThemeStore, type ThemeMode, type ThemeDefinition, SYSTEM_THEME } from '@/stores/theme'
import { storeToRefs } from 'pinia'

/**
 * Composable for accessing the global multi-theme system.
 *
 * Usage:
 * ```ts
 * const { isDark, resolvedTheme, mode, themes, toggle, setTheme, next } = useTheme()
 * ```
 */
export function useTheme() {
  const store = useThemeStore()
  const { mode, themes, resolvedTheme, currentTheme, isDark } = storeToRefs(store)

  return {
    /** User's preference: a theme id or 'system' */
    mode,
    /** List of all registered themes */
    themes,
    /** Resolved concrete theme id (never 'system') */
    resolvedTheme,
    /** Full ThemeDefinition of the current resolved theme */
    currentTheme,
    /** Whether the active theme is a dark variant */
    isDark,
    /** Set theme by id (or 'system') */
    setTheme: (m: ThemeMode) => store.setTheme(m),
    /** Toggle between the first light ↔ first dark theme */
    toggle: () => store.toggle(),
    /** Cycle to the next theme in the list */
    next: () => store.next(),
    /** Constant for the special 'system' mode */
    SYSTEM_THEME,
  }
}

export type { ThemeMode, ThemeDefinition }
