import { ref, onMounted, onBeforeUnmount } from 'vue'
import { OnFileDrop, OnFileDropOff } from '../../wailsjs/runtime/runtime'
import { OnDragSignal } from '../../wailsjs/go/render/Manager'
import { render } from '../../wailsjs/go/models'

/**
 * Coordinates OS-level file drag-and-drop between the frontend and the Go backend.
 *
 * Architecture:
 *   - DOM dragenter / dragleave (with counter) → visual mask + backend "enter"/"leave" signals
 *   - DOM dragover → preventDefault() so the browser allows the drop
 *   - Wails OnFileDrop callback → actual OS file paths → backend "drop" signal
 *
 * The counter pattern is needed because dragenter/dragleave fire for every child
 * element the pointer crosses, not just the window boundary.
 */
export function useFileDrop() {
  const isDragging = ref(false)
  let dragCounter = 0

  function onDragEnter(e: DragEvent) {
    // Only react to OS files, not internal element drag
    if (!e.dataTransfer?.types.includes('Files')) return
    e.preventDefault()
    dragCounter++
    if (dragCounter === 1) {
      isDragging.value = true
      OnDragSignal(new render.DragSignal({ type: 'enter', x: e.clientX, y: e.clientY }))
    }
  }

  function onDragOver(e: DragEvent) {
    // Must preventDefault here to allow the drop. Without this the cursor
    // shows "no drop" and OnFileDrop will never fire on macOS/WebKit.
    if (e.dataTransfer?.types.includes('Files')) {
      e.preventDefault()
    }
  }

  function onDragLeave(e: DragEvent) {
    dragCounter = Math.max(0, dragCounter - 1)
    if (dragCounter === 0) {
      isDragging.value = false
      OnDragSignal(new render.DragSignal({ type: 'leave', x: e.clientX, y: e.clientY }))
    }
  }

  function onDrop(e: DragEvent) {
    // Prevent browser from navigating to the dropped file.
    // Actual paths arrive via the Wails OnFileDrop callback below.
    e.preventDefault()
    dragCounter = 0
    isDragging.value = false
  }

  onMounted(() => {
    document.addEventListener('dragenter', onDragEnter)
    document.addEventListener('dragover', onDragOver)
    document.addEventListener('dragleave', onDragLeave)
    document.addEventListener('drop', onDrop)

    // Wails runtime callback — fires after a successful drop with real OS paths.
    // useDropTarget=false means we accept drops anywhere in the window (not just
    // elements with the --wails-drop-target CSS custom property).
    OnFileDrop((x: number, y: number, paths: string[]) => {
      isDragging.value = false
      OnDragSignal(new render.DragSignal({ type: 'drop', x, y, paths }))
    }, false)
  })

  onBeforeUnmount(() => {
    document.removeEventListener('dragenter', onDragEnter)
    document.removeEventListener('dragover', onDragOver)
    document.removeEventListener('dragleave', onDragLeave)
    document.removeEventListener('drop', onDrop)
    // Unregister the Wails file drop listener
    OnFileDropOff()
  })

  return { isDragging }
}
