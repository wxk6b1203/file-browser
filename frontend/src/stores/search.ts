import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { CancelSearch, StartSearch } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { useNotificationsStore } from './notifications'

const SEARCH_EVENT = 'search:event'

interface SearchFileInfo {
  name: string
  path: string
  type: number
  size: number
  lastModified?: any
}

export interface SearchResultItem {
  connectionId: string
  connectionName: string
  driver: string
  file: SearchFileInfo | null
}

export interface SearchErrorItem {
  connectionId?: string
  connectionName?: string
  message: string
}

export interface SearchSummary {
  query: string
  connectionCount: number
  completedConnections: number
  failedConnections: number
  scannedCount: number
  matchedCount: number
  resultLimit: number
  limitReached: boolean
  cancelled: boolean
  durationMs: number
}

interface SearchEventPayload {
  requestId: string
  type: 'started' | 'result' | 'error' | 'completed'
  query: string
  result?: SearchResultItem
  error?: SearchErrorItem
  summary?: SearchSummary
}

let unsubscribeSearchEvents: (() => void) | null = null

export const useSearchStore = defineStore('search', () => {
  const notifications = useNotificationsStore()
  const ready = ref(false)
  const running = ref(false)
  const query = ref('')
  const requestId = ref('')
  const selectedConnectionIds = ref<string[]>([])
  const results = ref<SearchResultItem[]>([])
  const errors = ref<SearchErrorItem[]>([])
  const summary = ref<SearchSummary | null>(null)

  const hasResults = computed(() => results.value.length > 0)
  const hasErrors = computed(() => errors.value.length > 0)

  function ensureSubscribed() {
    if (unsubscribeSearchEvents) return
    unsubscribeSearchEvents = EventsOn(SEARCH_EVENT, (payload: SearchEventPayload) => {
      handleEvent(payload)
    })
  }

  function reset() {
    running.value = false
    requestId.value = ''
    results.value = []
    errors.value = []
    summary.value = null
  }

  async function search() {
    ensureSubscribed()

    const trimmedQuery = query.value.trim()
    if (!trimmedQuery) {
      reset()
      return
    }

    if (running.value && requestId.value) {
      await cancel()
    }

    results.value = []
    errors.value = []
    summary.value = null

    requestId.value = await StartSearch({
      query: trimmedQuery,
      connectionIds: [...selectedConnectionIds.value],
    })
    running.value = true
    ready.value = true
  }

  async function cancel() {
    if (!requestId.value || !running.value) return
    await CancelSearch(requestId.value)
  }

  function handleEvent(payload?: SearchEventPayload) {
    if (!payload) return
    if (!payload.requestId) return
    if (requestId.value && payload.requestId !== requestId.value) return

    ready.value = true

    if (payload.type === 'started') {
      running.value = true
      summary.value = payload.summary ?? null
      return
    }

    if (payload.type === 'result' && payload.result) {
      results.value = [...results.value, payload.result]
      return
    }

    if (payload.type === 'error' && payload.error) {
      errors.value = [...errors.value, payload.error]
      notifications.push({
        level: 'error',
        source: payload.error.connectionName || 'Search',
        title: `Search: ${payload.query}`,
        message: payload.error.message,
      })
      return
    }

    if (payload.type === 'completed') {
      running.value = false
      summary.value = payload.summary ?? summary.value
    }
  }

  return {
    ready,
    running,
    query,
    requestId,
    selectedConnectionIds,
    results,
    errors,
    summary,
    hasResults,
    hasErrors,
    ensureSubscribed,
    search,
    cancel,
    reset,
  }
})
