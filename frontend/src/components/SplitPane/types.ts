import type { CSSProperties, InjectionKey, Ref } from 'vue'

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
}

/** Context provided by SplitPane to children */
export interface SplitPaneContext {
  layout: Ref<SplitLayout>
  lazy: Ref<boolean>
  gap: Ref<number>
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
}

/** Injection key for provide/inject */
export const splitPaneContextKey: InjectionKey<SplitPaneContext> = Symbol('splitPaneContext')

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
}

