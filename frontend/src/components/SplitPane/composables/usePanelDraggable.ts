import { inject } from 'vue'
import { splitPanePanelKey } from '../types'
import {
  setActiveInternalDrag,
  clearActiveInternalDrag,
  SPLITPANE_DRAG_TYPE,
  type PanelDragPayload,
} from '@/composables/splitPaneDragState'

/**
 * Makes an element inside a SplitPanePanel draggable in the SplitPane framework.
 *
 * Usage:
 * ```vue
 * <script setup>
 * import { usePanelDraggable } from '@/components/SplitPane'
 * const { dragProps } = usePanelDraggable(() => ({ type: 'file', data: myFile }))
 * </script>
 *
 * <template>
 *   <div v-bind="dragProps">Drag me</div>
 * </template>
 * ```
 *
 * The parent SplitPane will emit a `panelDrop` event when this element is
 * dropped onto a different panel, containing source / target panel indices
 * and the payload returned by `getPayload`.
 *
 * @param getPayload  Lazy factory — called on dragstart to capture current state.
 */
export function usePanelDraggable(getPayload: () => PanelDragPayload) {
  // Injected by SplitPanePanel; undefined if used outside a SplitPane (graceful degradation)
  const panel = inject(splitPanePanelKey)

  function onDragStart(e: DragEvent) {
    if (!e.dataTransfer) return
    // Mark this as a SplitPane-internal drag so panels can detect it in dragenter.types
    e.dataTransfer.setData(SPLITPANE_DRAG_TYPE, '')
    e.dataTransfer.effectAllowed = 'move'
    // Store source info in shared module state — readable synchronously in dragenter
    setActiveInternalDrag({
      sourcePanelUid: panel?.uid ?? -1,
      sourcePanelIndex: panel?.index.value ?? -1,
      payload: getPayload(),
    })
  }

  function onDragEnd() {
    clearActiveInternalDrag()
  }

  return {
    /**
     * Spread onto the draggable element:
     * `<div v-bind="dragProps">…</div>`
     */
    dragProps: {
      draggable: true,
      onDragstart: onDragStart,
      onDragend: onDragEnd,
    },
  }
}
