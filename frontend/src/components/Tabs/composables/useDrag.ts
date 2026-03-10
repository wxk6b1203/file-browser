import { reactive } from 'vue'
import type { DragState, TabItem } from '../types'

export function useDrag() {
  const drag: DragState = reactive({
    active: false,
    tab: null,
    sourceGroupId: null,
    pointerX: 0,
    pointerY: 0,
    offsetX: 0,
    offsetY: 0,
    headerWidth: 0,
    headerHeight: 0,
  })

  function startDrag(
    tab: TabItem,
    groupId: string,
    event: PointerEvent,
    headerEl: HTMLElement,
  ) {
    const rect = headerEl.getBoundingClientRect()
    drag.active = true
    drag.tab = tab
    drag.sourceGroupId = groupId
    drag.pointerX = event.clientX
    drag.pointerY = event.clientY
    drag.offsetX = event.clientX - rect.left
    drag.offsetY = event.clientY - rect.top
    drag.headerWidth = rect.width
    drag.headerHeight = rect.height
  }

  function updatePointer(x: number, y: number) {
    drag.pointerX = x
    drag.pointerY = y
  }

  function endDrag() {
    drag.active = false
    drag.tab = null
    drag.sourceGroupId = null
  }

  return { drag, startDrag, updatePointer, endDrag }
}

