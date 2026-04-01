/**
 * Shared module-level state for SplitPane internal panel-to-panel drag operations.
 *
 * Lives outside Vue reactivity so it can be read synchronously during `dragenter`
 * (where dataTransfer.getData() is blocked by the browser for security reasons).
 *
 * Both usePanelDraggable (sets state on dragstart) and usePanelFileDrop (reads state
 * on dragenter / drop) import from this module.
 */

/** User-defined payload attached to a draggable element via usePanelDraggable. */
export interface PanelDragPayload {
  /** Drag category — allows drop targets to decide whether to accept the drag. */
  type: string
  /** Arbitrary data the source wants to pass to the drop target. */
  data: unknown
}

/** Event emitted by SplitPane when an element is dropped from one panel onto another. */
export interface PanelDropEvent {
  /** Index of the panel where the drag originated. */
  sourcePanelIndex: number
  /** Index of the panel that received the drop. */
  targetPanelIndex: number
  /** Group id rendered inside the target panel. */
  targetGroupId?: string
  /** Active tab id rendered inside the target panel. */
  targetTabId?: string
  /** The payload registered by usePanelDraggable on the dragged element. */
  payload: PanelDragPayload
  /** Drop coordinates relative to the viewport. */
  x: number
  y: number
}

/** Internal — carries source panel identity alongside the user payload. */
export interface ActiveInternalDrag {
  sourcePanelUid: number
  sourcePanelIndex: number
  payload: PanelDragPayload
}

/** DataTransfer MIME type used to flag a SplitPane-internal drag. */
export const SPLITPANE_DRAG_TYPE = 'application/x-splitpane-drag'

let activeInternalDrag: ActiveInternalDrag | null = null

export function setActiveInternalDrag(drag: ActiveInternalDrag): void {
  activeInternalDrag = drag
}

export function clearActiveInternalDrag(): void {
  activeInternalDrag = null
}

export function getActiveInternalDrag(): ActiveInternalDrag | null {
  return activeInternalDrag
}
