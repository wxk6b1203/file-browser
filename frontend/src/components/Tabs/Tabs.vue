<template>
  <div
    ref="rootEl"
    class="tabs-root"
    :style="{ width: '100%', height: '100%', position: 'relative' }"
  >
    <!-- Recursive tree renderer -->
    <TabNodeRenderer :node="treeRef" />

    <!-- Drag ghost: floating header that follows the pointer -->
    <Teleport to="body">
      <div
        v-if="drag.active && drag.tab"
        class="tab-drag-ghost"
        :style="ghostStyle"
      >
        <span class="tab-drag-ghost__label">{{ drag.tab.label }}</span>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import {
  ref,
  shallowRef,
  toRef,
  provide,
  reactive,
  computed,
  onMounted,
  onBeforeUnmount,
  watch,
  type CSSProperties,
} from 'vue'
import type { TabNode, TabItem, DragState, TabsContext, TabGroupNode, NodeRect } from './types'
import { tabsContextKey } from './types'
import { useTabTree } from './composables/useTabTree'
import { calcDropZone } from './composables/useDropZone'
import TabNodeRenderer from './TabNodeRenderer.vue'

// ─── Props ───────────────────────────────────────────────────

const props = withDefaults(
  defineProps<{
    modelValue: TabNode
    /** Background color of tab bars */
    barBackground?: string
    /** Background color of content panels */
    tabBackground?: string
    /** Opacity of split-zone overlay (0-1) */
    overlayOpacity?: number
    /** Minimum width (px) to allow vertical split */
    minSplitWidth?: number
    /** Minimum height (px) to allow horizontal split */
    minSplitHeight?: number
  }>(),
  {
    barBackground: undefined,
    tabBackground: undefined,
    overlayOpacity: 0.15,
    minSplitWidth: 100,
    minSplitHeight: 80,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: TabNode]
  tabDragStart: [tab: TabItem, groupId: string]
  tabDragEnd: [tab: TabItem, groupId: string]
  tabReorder: [groupId: string, oldIndex: number, newIndex: number]
  tabSplit: [tabId: string, zone: 'top' | 'bottom' | 'left' | 'right']
  /** Fired when a tab is activated (clicked / selected) */
  tabActivate: [tab: TabItem, groupId: string]
}>()

// ─── Tree state ──────────────────────────────────────────────

// Use shallowRef to avoid deep-proxying component objects inside TabItem
const treeRef = shallowRef<TabNode>(props.modelValue)

// Flag to prevent v-model feedback loop
let internalUpdate = false

// Sync modelValue -> treeRef (only for truly external changes)
watch(
  () => props.modelValue,
  (val) => {
    if (internalUpdate) return
    treeRef.value = val
  },
  { deep: true },
)

// Emit on internal changes (fires on shallowRef reassignment or triggerRef)
watch(
  treeRef,
  (val) => {
    internalUpdate = true
    emit('update:modelValue', val)
    Promise.resolve().then(() => { internalUpdate = false })
  },
)

const { setActive, moveTab, splitGroup, removeTab, setSplitSizes } = useTabTree(treeRef)

// ─── Drag state ──────────────────────────────────────────────

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

const rootEl = ref<HTMLElement>()

// Ghost style: follows pointer but clamped to viewport
const ghostStyle = computed<CSSProperties>(() => {
  if (!drag.active) return { display: 'none' }
  let left = drag.pointerX - drag.offsetX
  let top = drag.pointerY - drag.offsetY

  // Clamp to viewport
  const vw = window.innerWidth
  const vh = window.innerHeight
  left = Math.max(0, Math.min(left, vw - drag.headerWidth))
  top = Math.max(0, Math.min(top, vh - drag.headerHeight))

  return {
    position: 'fixed',
    left: `${left}px`,
    top: `${top}px`,
    width: `${drag.headerWidth}px`,
    height: `${drag.headerHeight}px`,
    zIndex: 99999,
    pointerEvents: 'none',
  }
})

// ─── Global pointer listeners for drag ──────────────────────

function onGlobalPointerMove(e: PointerEvent) {
  if (!drag.active) return
  drag.pointerX = e.clientX
  drag.pointerY = e.clientY
}

function endDrag(commitDrop: boolean) {
  if (!drag.active) return
  if (commitDrop) {
    performDrop()
  }
  const tab = drag.tab
  const sourceGroupId = drag.sourceGroupId
  drag.active = false
  drag.tab = null
  drag.sourceGroupId = null
  if (tab && sourceGroupId) {
    emit('tabDragEnd', tab, sourceGroupId)
  }
}

function onGlobalPointerUp(_e: PointerEvent) {
  endDrag(true)
}

function onGlobalPointerCancel(_e: PointerEvent) {
  // Pointer cancellation means the gesture was interrupted by the system.
  // End drag without applying a drop target.
  endDrag(false)
}

function onWindowBlur() {
  // Releasing pointer outside the app window may skip pointerup.
  // Blur is a reliable fallback to clear drag state.
  endDrag(false)
}

function onVisibilityChange() {
  if (document.visibilityState === 'hidden') {
    endDrag(false)
  }
}

function performDrop() {
  if (!drag.tab || !drag.sourceGroupId) return

  let targetGroup: TabGroupNode | null = null
  let zone: ReturnType<typeof calcDropZone> = null

  // First, try to find a content area under the pointer (for split detection)
  const contentEls = rootEl.value?.querySelectorAll('.tab-group__content')
  if (contentEls) {
    for (const el of contentEls) {
      const rect = el.getBoundingClientRect()
      if (
        drag.pointerX >= rect.left &&
        drag.pointerX <= rect.right &&
        drag.pointerY >= rect.top &&
        drag.pointerY <= rect.bottom
      ) {
        zone = calcDropZone(drag.pointerX, drag.pointerY, rect)
        const groupDiv = el.closest('.tab-group')
        const groupId = groupDiv?.getAttribute('data-group-id')
        if (groupId) {
          targetGroup = findGroupNodeById(treeRef.value, groupId)
        }
      }
    }
  }

  // If pointer is over the tab bar (not content), treat as center/reorder
  if (!targetGroup) {
    const groupEls = rootEl.value?.querySelectorAll('.tab-group')
    if (groupEls) {
      for (const el of groupEls) {
        const rect = el.getBoundingClientRect()
        if (
          drag.pointerX >= rect.left &&
          drag.pointerX <= rect.right &&
          drag.pointerY >= rect.top &&
          drag.pointerY <= rect.bottom
        ) {
          const groupId = el.getAttribute('data-group-id')
          if (groupId) {
            targetGroup = findGroupNodeById(treeRef.value, groupId)
            zone = 'center'
          }
        }
      }
    }
  }

  if (!targetGroup || !zone) return

  const tabId = drag.tab.id
  const sourceGroupId = drag.sourceGroupId

  if (zone === 'center') {
    // Reorder / move tab into this group
    if (targetGroup.id === sourceGroupId) {
      // Reorder within same group – find tab bar position
      // For simplicity, try to find target index from pointer position in the tab bar
      const barEl = rootEl.value?.querySelector(
        `.tab-group[data-group-id="${targetGroup.id}"] .tab-bar`,
      )
      if (barEl) {
        const headers = barEl.querySelectorAll('.tab-header')
        let targetIndex = targetGroup.tabs.length
        for (let i = 0; i < headers.length; i++) {
          const header = headers[i]
          if (!header) continue
          const headerRect = header.getBoundingClientRect()
          const center = headerRect.left + headerRect.width / 2
          if (drag.pointerX < center) {
            targetIndex = i
            break
          }
        }
        const oldIndex = targetGroup.tabs.findIndex((t) => t.id === tabId)
        if (oldIndex !== -1 && oldIndex !== targetIndex) {
          if (targetIndex > oldIndex) targetIndex--
          moveTab(sourceGroupId, tabId, targetGroup.id, targetIndex)
          emit('tabReorder', targetGroup.id, oldIndex, targetIndex)
        }
      }
    } else {
      // Move tab to the end of target group
      moveTab(sourceGroupId, tabId, targetGroup.id, targetGroup.tabs.length)
    }
  } else {
    // Split
    // Check min-size constraints
    const contentEl = rootEl.value?.querySelector(
      `.tab-group[data-group-id="${targetGroup.id}"] .tab-group__content`,
    )
    if (contentEl) {
      const rect = contentEl.getBoundingClientRect()
      const canSplitH = rect.height >= props.minSplitHeight * 2
      const canSplitV = rect.width >= props.minSplitWidth * 2

      if ((zone === 'top' || zone === 'bottom') && !canSplitH) {
        // Fall back to move
        moveTab(sourceGroupId, tabId, targetGroup.id, targetGroup.tabs.length)
        return
      }
      if ((zone === 'left' || zone === 'right') && !canSplitV) {
        moveTab(sourceGroupId, tabId, targetGroup.id, targetGroup.tabs.length)
        return
      }
    }

    // If dragging from same group with only 1 tab, don't split
    if (sourceGroupId === targetGroup.id && targetGroup.tabs.length <= 1) return

    // First move the tab out if it's from a different group
    if (sourceGroupId !== targetGroup.id) {
      moveTab(sourceGroupId, tabId, targetGroup.id, targetGroup.tabs.length)
    }

    splitGroup(targetGroup.id, tabId, zone as 'top' | 'bottom' | 'left' | 'right')
    emit('tabSplit', tabId, zone as 'top' | 'bottom' | 'left' | 'right')
  }
}

function findGroupNodeById(node: TabNode, id: string): TabGroupNode | null {
  if (node.type === 'tabs' && node.id === id) return node
  if (node.type === 'split') {
    for (const child of node.children) {
      const found = findGroupNodeById(child, id)
      if (found) return found
    }
  }
  return null
}

onMounted(() => {
  document.addEventListener('pointermove', onGlobalPointerMove)
  document.addEventListener('pointerup', onGlobalPointerUp)
  document.addEventListener('pointercancel', onGlobalPointerCancel)
  document.addEventListener('visibilitychange', onVisibilityChange)
  window.addEventListener('blur', onWindowBlur)
})

onBeforeUnmount(() => {
  document.removeEventListener('pointermove', onGlobalPointerMove)
  document.removeEventListener('pointerup', onGlobalPointerUp)
  document.removeEventListener('pointercancel', onGlobalPointerCancel)
  document.removeEventListener('visibilitychange', onVisibilityChange)
  window.removeEventListener('blur', onWindowBlur)
})

// ─── Context ─────────────────────────────────────────────────

const context: TabsContext = {
  tree: treeRef,
  drag,
  barBackground: toRef(props, 'barBackground'),
  tabBackground: toRef(props, 'tabBackground'),
  overlayOpacity: toRef(props, 'overlayOpacity'),
  minSplitWidth: toRef(props, 'minSplitWidth'),
  minSplitHeight: toRef(props, 'minSplitHeight'),
  setActive,
  moveTab,
  splitGroup,
  removeTab,
  setSplitSizes,
  emitTabDragStart: (tab, groupId) => emit('tabDragStart', tab, groupId),
  emitTabDragEnd: (tab, groupId) => emit('tabDragEnd', tab, groupId),
  emitTabReorder: (groupId, oldIndex, newIndex) => emit('tabReorder', groupId, oldIndex, newIndex),
  emitTabSplit: (tabId, zone) => emit('tabSplit', tabId, zone),
  emitTabActivate: (tab, groupId) => emit('tabActivate', tab, groupId),
}

provide(tabsContextKey, context)

// ─── Public imperative API ───────────────────────────────────

/**
 * Get the bounding rectangle of a tree node by its ID.
 * Works for both TabGroupNode (type 'tabs') and TabSplitNode (type 'split').
 * Returns null if the node is not found or not currently rendered.
 */
function getNodeRect(nodeId: string): NodeRect | null {
  if (!rootEl.value) return null
  const el = rootEl.value.querySelector<HTMLElement>(`[data-node-id="${nodeId}"]`)
  if (!el) return null
  const rect = el.getBoundingClientRect()
  return { x: rect.x, y: rect.y, width: rect.width, height: rect.height }
}

/**
 * Collect all node IDs in the tree and return a map of id → NodeRect.
 * Useful for getting the layout of the entire tree at once.
 */
function getAllNodeRects(): Map<string, NodeRect> {
  const result = new Map<string, NodeRect>()
  if (!rootEl.value) return result
  const els = rootEl.value.querySelectorAll<HTMLElement>('[data-node-id]')
  for (const el of els) {
    const id = el.getAttribute('data-node-id')
    if (!id) continue
    const rect = el.getBoundingClientRect()
    result.set(id, { x: rect.x, y: rect.y, width: rect.width, height: rect.height })
  }
  return result
}

/**
 * Recursively find a node by ID in the tree.
 */
function findNodeById(nodeId: string): TabNode | null {
  function walk(node: TabNode): TabNode | null {
    if (node.id === nodeId) return node
    if (node.type === 'split') {
      for (const child of node.children) {
        const found = walk(child)
        if (found) return found
      }
    }
    return null
  }
  return walk(treeRef.value)
}

defineExpose({
  /** Get bounding rect of a single tree node by ID */
  getNodeRect,
  /** Get bounding rects of all rendered tree nodes */
  getAllNodeRects,
  /** Find a tree node object by ID */
  findNodeById,
  /** The reactive tree ref */
  treeRef,
})
</script>

<style scoped>
.tabs-root {
  overflow: hidden;
}

.tab-drag-ghost {
  display: inline-flex;
  align-items: center;
  padding: 6px 14px;
  background: var(--theme-color-bg-overlay);
  border: 1px solid var(--theme-color-border-dark);
  border-radius: 6px;
  box-shadow: var(--theme-shadow-float);
  font-size: 13px;
  color: var(--theme-color-text-base);
  white-space: nowrap;
  user-select: none;
  opacity: 0.92;
}
</style>



