import type { CSSProperties, InjectionKey, Ref } from 'vue'
import type { PanelDragPayload, PanelDropEvent } from '@/composables/splitPaneDragState'

export type { PanelDragPayload, PanelDropEvent }

/** Layout direction */
export type SplitLayout = 'horizontal' | 'vertical'

/** Registered panel instance state */
export interface PanelState {
  uid: number
  index: number
  size?: number | string
  minSize?: number | string
  maxSize?: number | string
  resizable: boolean
  borderRadius?: string
  /** Whether the panel is minimized (collapsed to 0 size) */
  minimized: boolean
  /** DOM element reference for DOM-order sorting */
  el?: HTMLElement
}

/** Context provided by SplitPane to children */
export interface SplitPaneContext {
  layout: Ref<SplitLayout>
  lazy: Ref<boolean>
  gap: Ref<number>
  indicatorSize: Ref<[string, string]>
  /** Whether OS file drag-drop overlays are enabled on panels (default false) */
  enableFileDrop: Ref<boolean>
  /** Whether internal panel-to-panel drag (via usePanelDraggable) is enabled (default false) */
  enablePanelDrag: Ref<boolean>
  panels: Ref<PanelState[]>
  pxSizes: Ref<number[]>
  percentSizes: Ref<number[]>
  containerSize: Ref<number>
  movingIndex: Ref<{ index: number; confirmed: boolean } | null>
  registerPanel: (panel: PanelState) => void
  unregisterPanel: (panel: PanelState) => void
  onMoveStart: (index: number) => void
  onMoving: (index: number, offset: number) => void
  onMoveEnd: (index: number) => void
  onDblClick: (index: number) => void
  /** Minimize a panel by uid */
  minimizePanel: (uid: number) => void
  /** Restore a minimized panel by uid */
  restorePanel: (uid: number) => void
  /** Toggle minimize state by uid */
  togglePanel: (uid: number) => void
  /**
   * Called by SplitPanePanel when an internal panel-to-panel drop completes.
   * SplitPane re-emits this as the public `panelDrop` event.
   */
  onPanelDrop: (event: PanelDropEvent) => void
}

/** Injection key for provide/inject */
export const splitPaneContextKey: InjectionKey<SplitPaneContext> = Symbol('splitPaneContext')

/**
 * Injection key provided by SplitPanePanel to its slot children.
 * Consumed by usePanelDraggable so draggable elements can tag their source panel.
 */
export const splitPanePanelKey: InjectionKey<{
  readonly uid: number
  readonly index: Ref<number>
}> = Symbol('splitPanePanel')

/** Props for SplitPanePanel */
export interface SplitPanePanelProps {
  /** Initial size: '30%', '200px', or number (px) */
  size?: number | string
  /** Minimum size: '10%', '100px', or number (px) */
  minSize?: number | string
  /** Maximum size: '80%', '500px', or number (px) */
  maxSize?: number | string
  /** Whether the panel can be resized (default true) */
  resizable?: boolean
  /** Border radius CSS value, e.g. '8px' or '0 8px 8px 0' */
  borderRadius?: string
  /** Background color CSS value */
  backgroundColor?: string
  /** Custom inline styles */
  customStyle?: CSSProperties
  /** Custom CSS class */
  customClass?: string
  /** Whether the panel is minimized (collapsed to 0). Supports v-model:minimized */
  minimized?: boolean
}

