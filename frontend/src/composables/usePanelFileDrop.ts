import { ref, watchEffect, onBeforeUnmount, onMounted, type Ref } from 'vue'
import {
  getActiveInternalDrag,
  isInternalDragEvent,
  type PanelDropEvent,
} from './splitPaneDragState'
import { PANEL_OS_FILE_DROP_RESET_EVENT } from './useFileDrop'

// ─── OS drop coordination ─────────────────────────────────────────────────────
// DOM `drop` fires before Wails `OnFileDrop` delivers actual paths.
// The panel records its identity here; useFileDrop reads it when paths arrive.

export interface PendingPanelDropInfo {
  groupId: string
  tabId: string
  x: number
  y: number
}

let pendingPanelDropInfo: PendingPanelDropInfo | null = null

function resolveGroupHost(root: HTMLElement | null): HTMLElement | null {
  if (!root) return null
  if (root.hasAttribute('data-group-id')) {
    return root
  }
  return root.querySelector<HTMLElement>('[data-group-id]')
}

/** Called by useFileDrop to consume panel context when OS drop paths arrive. */
export function consumePendingPanelDropInfo(): PendingPanelDropInfo | null {
  const info = pendingPanelDropInfo
  pendingPanelDropInfo = null
  return info
}

// ─── Composable ───────────────────────────────────────────────────────────────

/**
 * Per-panel drag-drop tracking for SplitPanePanel.
 *
 * Handles two drag sources:
 *   1. Internal panel-to-panel drags initiated via usePanelDraggable.
 *      On drop → calls `onInternalDrop` immediately with full payload.
 *   2. OS-level file / directory drops (from Finder / Explorer).
 *      On drop → records panel identity; useFileDrop combines it with
 *      the paths from Wails OnFileDrop and calls OnPanelFileDrop.
 *
 * @param panelEl          Root element ref of the panel.
 * @param panelUid         Stable UID of this panel (used to skip overlay on source).
 * @param panelIndex       Reactive index of this panel within SplitPane.
 * @param enableFileDrop   Gate for OS file/directory drop handling.
 * @param enablePanelDrag  Gate for internal panel-to-panel drag handling.
 * @param onInternalDrop   Called when an internal panel-to-panel drop completes.
 */
export function usePanelFileDrop(
  panelEl: Ref<HTMLElement | undefined>,
  panelUid: number,
  panelIndex: Ref<number>,
  enableFileDrop: Ref<boolean>,
  enablePanelDrag: Ref<boolean>,
  onInternalDrop: (event: PanelDropEvent) => void,
) {
  const isDragOver = ref(false)
  let dragCounter = 0
  let currentEl: HTMLElement | null = null

  function onDragEnter(e: DragEvent) {
    const isInternal = isInternalDragEvent(e)
    if (isInternal) {
      if (!enablePanelDrag.value) return
      // Don't show overlay on the panel that is the drag source
      const active = getActiveInternalDrag()
      if (active?.sourcePanelUid === panelUid) return
    } else {
      if (!enableFileDrop.value) return
    }
    dragCounter++
    if (dragCounter === 1) {
      isDragOver.value = true
    }
  }

  function onDragOver(e: DragEvent) {
    // preventDefault is required for the browser to allow a drop here.
    // Must be unconditional to support files, dirs, and internal panel drags.
    e.preventDefault()
  }

  function onDragLeave() {
    dragCounter = Math.max(0, dragCounter - 1)
    if (dragCounter === 0) {
      isDragOver.value = false
    }
  }

  function onDrop(e: DragEvent) {
    e.preventDefault()
    dragCounter = 0
    isDragOver.value = false

    if (import.meta.env.VITE_APP_ENV === 'internal') return
    if (!panelEl.value) return

    // ── Internal panel-to-panel drop ──────────────────────────────────────────
    const isInternal = isInternalDragEvent(e)
    if (isInternal) {
      if (enablePanelDrag.value) {
        const active = getActiveInternalDrag()
        if (active && active.sourcePanelUid !== panelUid) {
          const groupEl = resolveGroupHost(panelEl.value)
          onInternalDrop({
            sourcePanelIndex: active.sourcePanelIndex,
            targetPanelIndex: panelIndex.value,
            targetGroupId: groupEl?.getAttribute('data-group-id') ?? '',
            targetTabId: groupEl?.getAttribute('data-active-tab-id') ?? '',
            payload: active.payload,
            x: e.clientX,
            y: e.clientY,
          })
        }
      }
      return
    }

    // ── OS file / directory drop ──────────────────────────────────────────────
    if (!enableFileDrop.value) return
    // Innermost panel wins (drop bubbles inner → outer). Skip if already set.
    // Stale entries from non-Wails drops are cleared by useFileDrop on next dragenter.
    if (pendingPanelDropInfo !== null) return

    const groupEl = resolveGroupHost(panelEl.value)
    pendingPanelDropInfo = {
      groupId: groupEl?.getAttribute('data-group-id') ?? '',
      tabId: groupEl?.getAttribute('data-active-tab-id') ?? '',
      x: e.clientX,
      y: e.clientY,
    }
  }

  function attach(el: HTMLElement) {
    currentEl = el
    el.addEventListener('dragenter', onDragEnter)
    el.addEventListener('dragover', onDragOver)
    el.addEventListener('dragleave', onDragLeave)
    el.addEventListener('drop', onDrop)
  }

  function detach() {
    if (!currentEl) return
    currentEl.removeEventListener('dragenter', onDragEnter)
    currentEl.removeEventListener('dragover', onDragOver)
    currentEl.removeEventListener('dragleave', onDragLeave)
    currentEl.removeEventListener('drop', onDrop)
    currentEl = null
    isDragOver.value = false
    dragCounter = 0
  }

  function resetOverlayState() {
    isDragOver.value = false
    dragCounter = 0
  }

  watchEffect(() => {
    detach()
    if ((enableFileDrop.value || enablePanelDrag.value) && panelEl.value) {
      attach(panelEl.value)
    }
  })

  onMounted(() => {
    window.addEventListener(PANEL_OS_FILE_DROP_RESET_EVENT, resetOverlayState)
  })

  onBeforeUnmount(() => {
    window.removeEventListener(PANEL_OS_FILE_DROP_RESET_EVENT, resetOverlayState)
    detach()
  })

  return { isDragOver }
}
