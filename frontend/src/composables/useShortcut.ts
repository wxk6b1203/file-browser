import { onMounted, onUnmounted } from 'vue'
import { EventsEmit, EventsOn } from '../../wailsjs/runtime/runtime'

// ---------------------------------------------------------------------------
// Event constants — mirror Go shortcut.EventShortcutFired / EventShortcutTrigger
// ---------------------------------------------------------------------------

/** Emitted by the frontend to notify Go that a shortcut was pressed. */
export const SHORTCUT_FIRED_EVENT = 'shortcut:fired'

/**
 * Emitted by Go to programmatically trigger a shortcut on the frontend.
 * Kept separate from SHORTCUT_FIRED_EVENT so the frontend's own
 * EventsEmit does NOT echo back through EventsOn and double-fire handlers.
 */
export const SHORTCUT_TRIGGER_EVENT = 'shortcut:trigger'

// ---------------------------------------------------------------------------
// Shortcut definition type
// ---------------------------------------------------------------------------

export interface ShortcutDefinition {
  /** Unique identifier, e.g. "save", "new-folder" */
  id: string
  /** Human-readable label (can be an i18n key) */
  label: string
  /** Key combo in Electron-style format, e.g. "CmdOrCtrl+S", "F2" */
  accelerator: string
  /** Whether this shortcut is active. Default: true */
  enabled?: boolean
}

// ---------------------------------------------------------------------------
// Default preset
// ---------------------------------------------------------------------------

export const DEFAULT_SHORTCUTS: ShortcutDefinition[] = [
  { id: 'new-connection', label: 'New Connection', accelerator: 'CmdOrCtrl+Shift+N' },
  { id: 'open-settings',  label: 'Open Settings',  accelerator: 'CmdOrCtrl+.' },
  { id: 'search',         label: 'Search',         accelerator: 'CmdOrCtrl+Shift+F' },
  { id: 'close-tab',      label: 'Close Tab',      accelerator: 'CmdOrCtrl+W' },
  { id: 'refresh',        label: 'Refresh',        accelerator: 'CmdOrCtrl+R' },
  { id: 'rename',         label: 'Rename',         accelerator: 'F2' },
]

// ---------------------------------------------------------------------------
// Accelerator parser
// ---------------------------------------------------------------------------

interface ParsedAccelerator {
  ctrl: boolean
  meta: boolean
  shift: boolean
  alt: boolean
  key: string // lowercased
}

const IS_MAC =
  typeof navigator !== 'undefined' && /mac/i.test(navigator.platform)

const KEY_ALIAS: Record<string, string> = {
  delete: 'delete',
  backspace: 'backspace',
  enter: 'enter',
  return: 'enter',
  tab: 'tab',
  escape: 'escape',
  esc: 'escape',
  space: ' ',
  up: 'arrowup',
  down: 'arrowdown',
  left: 'arrowleft',
  right: 'arrowright',
  plus: '+',
}

function parseAccelerator(raw: string): ParsedAccelerator | null {
  const parts = raw.split('+')
  const parsed: ParsedAccelerator = {
    ctrl: false, meta: false, shift: false, alt: false, key: '',
  }
  for (const part of parts) {
    const p = part.trim().toLowerCase()
    switch (p) {
      case 'cmdorctrl':
        IS_MAC ? (parsed.meta = true) : (parsed.ctrl = true); break
      case 'ctrl': case 'control':
        parsed.ctrl = true; break
      case 'shift':
        parsed.shift = true; break
      case 'alt': case 'option': case 'optionoralt':
        parsed.alt = true; break
      case 'cmd': case 'command': case 'meta': case 'super':
        parsed.meta = true; break
      default:
        parsed.key = p
    }
  }
  return parsed.key ? parsed : null
}

function matchesEvent(e: KeyboardEvent, p: ParsedAccelerator): boolean {
  if (e.ctrlKey !== p.ctrl) return false
  if (e.metaKey !== p.meta) return false
  if (e.shiftKey !== p.shift) return false
  if (e.altKey !== p.alt) return false
  const target = KEY_ALIAS[p.key] ?? p.key
  return e.key.toLowerCase() === target
}

// ---------------------------------------------------------------------------
// Singleton keydown engine
// ---------------------------------------------------------------------------

type HandlerFn = () => void

interface ShortcutEntry {
  def: ShortcutDefinition
  parsed: ParsedAccelerator
}

/** Frontend handler registry: shortcutID → Set of callbacks */
const handlers = new Map<string, Set<HandlerFn>>()

/** Parsed shortcut definitions */
let entries: ShortcutEntry[] = []
let listenerInstalled = false

function isEditable(e: KeyboardEvent): boolean {
  const el = e.target as HTMLElement | null
  if (!el) return false
  const tag = el.tagName
  if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return true
  return el.isContentEditable
}

function hasModifier(p: ParsedAccelerator): boolean {
  return p.ctrl || p.meta || p.alt
}

function onKeyDown(e: KeyboardEvent) {
  const editable = isEditable(e)

  for (const entry of entries) {
    if (entry.def.enabled === false) continue
    // In editable elements, skip modifier-less shortcuts (Delete, F2, etc.)
    if (editable && !hasModifier(entry.parsed)) continue

    if (matchesEvent(e, entry.parsed)) {
      const fns = handlers.get(entry.def.id)
      if (fns && fns.size > 0) {
        e.preventDefault()
        e.stopPropagation()
        fns.forEach((fn) => fn())
      }
      // Notify Go so it can also react (fire-and-forget)
      EventsEmit(SHORTCUT_FIRED_EVENT, entry.def.id)
      return // first match wins
    }
  }
}

function installListener() {
  if (listenerInstalled) return
  listenerInstalled = true
  document.addEventListener('keydown', onKeyDown, true)

  // Also listen for Go-triggered shortcuts (Go → frontend via Emit)
  // Note: we listen on SHORTCUT_TRIGGER_EVENT, NOT SHORTCUT_FIRED_EVENT.
  // In the Wails runtime, EventsEmit from the frontend is broadcast to ALL
  // listeners including frontend EventsOn callbacks. If we listened on
  // SHORTCUT_FIRED_EVENT here, every keydown would fire handlers twice:
  // once directly in onKeyDown, and once via the echo-back event.
  // SHORTCUT_TRIGGER_EVENT is only ever emitted by Go (Dispatcher.Emit),
  // so this callback is exclusively for Go-originated programmatic triggers.
  if (import.meta.env.VITE_APP_ENV !== 'internal') {
    EventsOn(SHORTCUT_TRIGGER_EVENT, (id: string) => {
      const fns = handlers.get(id)
      if (fns) fns.forEach((fn) => fn())
    })
  }
}

// ---------------------------------------------------------------------------
// Public: shortcut registration (call once, typically in App.vue setup)
// ---------------------------------------------------------------------------

/**
 * Register an array of shortcut definitions and start the keydown listener.
 *
 * @example
 * ```ts
 * import { defineShortcuts, DEFAULT_SHORTCUTS } from '@/composables/useShortcut'
 * defineShortcuts(DEFAULT_SHORTCUTS)
 * ```
 */
export function defineShortcuts(shortcuts: ShortcutDefinition[]) {
  for (const s of shortcuts) {
    defineShortcut(s)
  }
  installListener()
}

/**
 * Register a single shortcut definition.
 */
export function defineShortcut(s: ShortcutDefinition) {
  const parsed = parseAccelerator(s.accelerator)
  if (!parsed) {
    console.warn(`[shortcut] invalid accelerator "${s.accelerator}" for "${s.id}", skipped`)
    return
  }
  // Replace existing entry with the same id
  entries = entries.filter((e) => e.def.id !== s.id)
  entries.push({ def: { ...s, enabled: s.enabled !== false }, parsed })
  installListener()
}

/**
 * Remove a shortcut definition by ID.
 */
export function removeShortcut(id: string) {
  entries = entries.filter((e) => e.def.id !== id)
}

/**
 * Get a snapshot of all registered shortcut definitions.
 */
export function getShortcuts(): ShortcutDefinition[] {
  return entries.map((e) => ({ ...e.def }))
}

// ---------------------------------------------------------------------------
// Public composables — call inside <script setup>
// ---------------------------------------------------------------------------

/**
 * Listen for a single shortcut by its ID.
 *
 * Handler is automatically attached on mount and removed on unmount.
 *
 * @example
 * ```ts
 * useShortcut('save', () => {
 *   console.log('Ctrl+S pressed!')
 * })
 * ```
 */
export function useShortcut(shortcutId: string, callback: HandlerFn) {
  onMounted(() => {
    let set = handlers.get(shortcutId)
    if (!set) {
      set = new Set()
      handlers.set(shortcutId, set)
    }
    set.add(callback)
  })

  onUnmounted(() => {
    const set = handlers.get(shortcutId)
    if (set) {
      set.delete(callback)
      if (set.size === 0) handlers.delete(shortcutId)
    }
  })
}

/**
 * Listen for multiple shortcuts at once using `{ id: handler }`.
 *
 * @example
 * ```ts
 * useShortcutMap({
 *   save:        () => handleSave(),
 *   'new-file':  () => handleNewFile(),
 *   refresh:     () => handleRefresh(),
 * })
 * ```
 */
export function useShortcutMap(map: Record<string, HandlerFn>) {
  const pairs = Object.entries(map)

  onMounted(() => {
    for (const [id, fn] of pairs) {
      let set = handlers.get(id)
      if (!set) {
        set = new Set()
        handlers.set(id, set)
      }
      set.add(fn)
    }
  })

  onUnmounted(() => {
    for (const [id, fn] of pairs) {
      const set = handlers.get(id)
      if (set) {
        set.delete(fn)
        if (set.size === 0) handlers.delete(id)
      }
    }
  })
}
