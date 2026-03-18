import type { Component, InjectionKey, Ref } from 'vue'

/** A single button in the sidebar button column */
export interface SidebarButton {
  /** Unique identifier for this button */
  id: string
  /** Icon component (e.g. an unplugin-icons component like IEpFolder) */
  icon: Component
  /** Tooltip text shown on hover */
  tooltip?: string
  /**
   * 'menu' — clicking toggles the associated sidebar menu panel.
   * 'action' — clicking triggers onClick without affecting the panel.
   */
  type: 'menu' | 'action'
  /** Callback for action-type buttons, or additional handler for menu-type buttons */
  onClick?: () => void
}

/** Configuration for one sidebar (left or right) */
export interface SidebarConfig {
  /** Buttons aligned to the top of the column */
  topButtons: SidebarButton[]
  /** Buttons aligned to the bottom of the column */
  bottomButtons: SidebarButton[]
  /** Default width of the menu panel, e.g. '240px' or '20%' */
  defaultSize?: string | number
  /** Minimum width of the menu panel */
  minSize?: string | number
  /** Maximum width of the menu panel */
  maxSize?: string | number
}

/** Position of a sidebar */
export type SidebarSide = 'left' | 'right'

/** Context provided by SkeletonLayout for deeply nested children */
export interface SkeletonContext {
  /** Currently active menu button id for the left sidebar (null = panel collapsed) */
  leftActiveId: Ref<string | null>
  /** Currently active menu button id for the right sidebar (null = panel collapsed) */
  rightActiveId: Ref<string | null>
  /** Toggle the left sidebar to a specific menu button id, or collapse if same */
  setLeftActive: (id: string | null) => void
  /** Toggle the right sidebar to a specific menu button id, or collapse if same */
  setRightActive: (id: string | null) => void
}

/** Injection key for SkeletonLayout context */
export const skeletonContextKey: InjectionKey<SkeletonContext> = Symbol('skeletonContext')

