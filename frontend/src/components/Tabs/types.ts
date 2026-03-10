import type { Component, InjectionKey, Ref } from 'vue'

// ─── Tab Item ────────────────────────────────────────────────

/** A single tab definition */
export interface TabItem {
  /** Unique id */
  id: string
  /** Display label */
  label: string
  /** Whether the tab can be closed */
  closable?: boolean
  /**
   * Content component.
   * Can be an imported component object or a globally registered component name.
   */
  component?: Component | string
  /** Props forwarded to the content component */
  props?: Record<string, any>
}

// ─── Tree nodes ──────────────────────────────────────────────

/** A leaf node: a group of tabs with a tab bar */
export interface TabGroupNode {
  type: 'tabs'
  id: string
  tabs: TabItem[]
  /** Currently active tab id */
  activeId: string
}

/** A split node: two or more children laid out via SplitPane */
export interface TabSplitNode {
  type: 'split'
  id: string
  /** Layout direction of the split */
  layout: 'horizontal' | 'vertical'
  /** Children – can be tab groups or nested splits */
  children: TabNode[]
  /** Initial sizes forwarded to SplitPanePanel (e.g. ['50%','50%']) */
  sizes?: (string | number)[]
}

/** Recursive union – the tree */
export type TabNode = TabGroupNode | TabSplitNode

// ─── Drop zone ───────────────────────────────────────────────

export type DropZone = 'center' | 'top' | 'bottom' | 'left' | 'right' | null

// ─── Drag state (shared across components) ───────────────────

export interface DragState {
  /** Is a header currently being dragged? */
  active: boolean
  /** The tab being dragged */
  tab: TabItem | null
  /** Group id the tab originated from */
  sourceGroupId: string | null
  /** Current pointer position relative to viewport */
  pointerX: number
  pointerY: number
  /** Offset from pointer to the top-left corner of the ghost */
  offsetX: number
  offsetY: number
  /** Size of the original header element */
  headerWidth: number
  headerHeight: number
}

// ─── Context injected by root Tabs ──────────────────────────

export interface TabsContext {
  /** Root tree data (reactive) */
  tree: Ref<TabNode>
  /** Shared drag state */
  drag: DragState

  // ── customisation ──
  barBackground: Ref<string | undefined>
  tabBackground: Ref<string | undefined>
  overlayOpacity: Ref<number>
  minSplitWidth: Ref<number>
  minSplitHeight: Ref<number>

  // ── tree mutations ──
  setActive: (groupId: string, tabId: string) => void
  moveTab: (fromGroupId: string, tabId: string, toGroupId: string, newIndex: number) => void
  splitGroup: (groupId: string, tabId: string, zone: 'top' | 'bottom' | 'left' | 'right') => void
  removeTab: (groupId: string, tabId: string) => void

  // ── events ──
  emitTabDragStart: (tab: TabItem, groupId: string) => void
  emitTabDragEnd: (tab: TabItem, groupId: string) => void
  emitTabReorder: (groupId: string, oldIndex: number, newIndex: number) => void
  emitTabSplit: (tabId: string, zone: 'top' | 'bottom' | 'left' | 'right') => void
}

export const tabsContextKey: InjectionKey<TabsContext> = Symbol('tabsContext')

// ─── Utility: generate IDs ──────────────────────────────────

let _uid = 0
export function genId(prefix = 'tab'): string {
  return `${prefix}-${++_uid}-${Date.now().toString(36)}`
}

