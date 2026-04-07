export type NavigationHistoryMode = 'push' | 'replace' | 'reset' | 'none'

export interface NavigationHistoryState {
  entries: string[]
  index: number
}

export const NAVIGATION_HISTORY_LIMIT = 80

export function createNavigationHistory(initialPath = ''): NavigationHistoryState {
  return {
    entries: [initialPath],
    index: 0,
  }
}

export function applyNavigationHistory(
  state: NavigationHistoryState,
  path: string,
  mode: NavigationHistoryMode,
  limit = NAVIGATION_HISTORY_LIMIT,
): NavigationHistoryState {
  if (mode === 'none') {
    return cloneNavigationHistory(state)
  }

  if (mode === 'reset') {
    return createNavigationHistory(path)
  }

  const normalizedLimit = Math.max(1, Math.floor(limit))
  const entries = normalizeEntries(state)
  const index = normalizeIndex(state.index, entries)

  if (mode === 'replace') {
    const nextEntries = entries.slice(0, index + 1)
    nextEntries[index] = path
    return trimHistory({
      entries: nextEntries,
      index,
    }, normalizedLimit)
  }

  if (entries[index] === path) {
    return {
      entries,
      index,
    }
  }

  return trimHistory({
    entries: [...entries.slice(0, index + 1), path],
    index: index + 1,
  }, normalizedLimit)
}

export function navigationTarget(state: NavigationHistoryState, delta: -1 | 1): string | null {
  const entries = normalizeEntries(state)
  const index = normalizeIndex(state.index, entries)
  const nextIndex = index + delta
  if (nextIndex < 0 || nextIndex >= entries.length) {
    return null
  }
  return entries[nextIndex] ?? null
}

export function moveNavigationHistoryIndex(
  state: NavigationHistoryState,
  delta: -1 | 1,
): NavigationHistoryState {
  const entries = normalizeEntries(state)
  const index = normalizeIndex(state.index, entries)
  const nextIndex = Math.max(0, Math.min(index + delta, entries.length - 1))
  return {
    entries,
    index: nextIndex,
  }
}

function cloneNavigationHistory(state: NavigationHistoryState): NavigationHistoryState {
  const entries = normalizeEntries(state)
  return {
    entries,
    index: normalizeIndex(state.index, entries),
  }
}

function normalizeEntries(state: NavigationHistoryState) {
  return state.entries.length > 0 ? [...state.entries] : ['']
}

function normalizeIndex(index: number, entries: string[]) {
  if (!Number.isFinite(index)) return entries.length - 1
  return Math.max(0, Math.min(Math.floor(index), entries.length - 1))
}

function trimHistory(state: NavigationHistoryState, limit: number): NavigationHistoryState {
  if (state.entries.length <= limit) {
    return state
  }

  const removeCount = state.entries.length - limit
  return {
    entries: state.entries.slice(removeCount),
    index: Math.max(0, state.index - removeCount),
  }
}
