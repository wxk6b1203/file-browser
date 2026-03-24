<template>
  <div
    ref="containerEl"
    :class="[
      'split-pane',
      `split-pane--${layout}`,
      { 'split-pane--dragging': movingIndex !== null },
    ]"
  >
    <slot />
    <!-- Mask to prevent iframe from capturing pointer events during drag -->
    <div
      v-if="movingIndex !== null"
      :class="['split-pane__mask', `split-pane__mask--${layout}`]"
    />
  </div>
</template>

<script setup lang="ts">
import { provide, toRef, watch, nextTick, computed } from 'vue'
import { splitPaneContextKey, type PanelState, type PanelDropEvent, type SplitLayout } from './types'
import { useContainer } from './composables/useContainer'
import { useSize, parseSizeToPx } from './composables/useSize'
import { useResize } from './composables/useResize'
import { ref } from 'vue'

const props = withDefaults(
  defineProps<{
    /** Layout direction: 'horizontal' (left-right) or 'vertical' (top-bottom) */
    layout?: SplitLayout
    /** If true, resize only applies after drag ends */
    lazy?: boolean
    /** Gap between panels in px (default 0) */
    gap?: number
    /** Indicator thickness (the short side), CSS value e.g. '2px', '0.2em'. Default '2px' */
    indicatorWidth?: string
    /** Indicator length (the long side), CSS value e.g. '24px', '50%'. Default '24px' */
    indicatorHeight?: string
    /** Global max panel size in the drag direction. CSS value e.g. '400px', '60%', or number (px). Each panel cannot exceed this during drag or double-click reset. */
    maxPanelSize?: number | string
    /** Enable OS file drag-drop overlays on panels (default false). */
    enableFileDrop?: boolean
    /** Enable internal panel-to-panel drag via usePanelDraggable (default false). */
    enablePanelDrag?: boolean
  }>(),
  {
    layout: 'horizontal',
    lazy: false,
    gap: 0,
    indicatorWidth: '2px',
    indicatorHeight: '24px',
    enableFileDrop: false,
    enablePanelDrag: false,
  },
)

const emit = defineEmits<{
  /** Fired when drag starts. Args: divider index, current sizes in px */
  resizeStart: [index: number, sizes: number[]]
  /** Fired during drag (if not lazy). Args: divider index, current sizes in px */
  resize: [index: number, sizes: number[]]
  /** Fired when drag ends. Args: divider index, final sizes in px */
  resizeEnd: [index: number, sizes: number[]]
  /** Fired when divider indicator is double-clicked to reset to center. Args: divider index, final sizes in px */
  resetCenter: [index: number, sizes: number[]]
  /** Fired when a panel is minimized. Args: panel index */
  panelMinimized: [index: number]
  /** Fired when a panel is restored from minimized state. Args: panel index */
  panelRestored: [index: number]
  /**
   * Fired when an element (dragged via usePanelDraggable) is dropped onto a different panel.
   * Contains source panel index, target panel index, the drag payload, and drop coordinates.
   */
  panelDrop: [event: PanelDropEvent]
}>()

const layout = toRef(props, 'layout')
const lazy = toRef(props, 'lazy')
const gap = toRef(props, 'gap')
const enableFileDrop = toRef(props, 'enableFileDrop')
const enablePanelDrag = toRef(props, 'enablePanelDrag')

function onPanelDrop(event: PanelDropEvent) {
  emit('panelDrop', event)
}

// Indicator size: [width(short side), height(long side)]
const indicatorSize = computed<[string, string]>(() => [props.indicatorWidth, props.indicatorHeight])

// Container size tracking
const { containerEl, containerSize } = useContainer(layout)

// Panel registry
const panels = ref<PanelState[]>([])

// Count visible dividers = gaps between non-minimized panels
const visibleDividerCount = computed(() => {
  const nonMinCount = panels.value.filter((p) => !p.minimized).length
  return Math.max(nonMinCount - 1, 0)
})

// Available size = container size minus total gap for visible dividers
const availableSize = computed(() => {
  return Math.max(containerSize.value - visibleDividerCount.value * gap.value, 0)
})

function registerPanel(panel: PanelState) {
  panels.value.push(panel)
  sortPanelsByDom()
}

function unregisterPanel(panel: PanelState) {
  const idx = panels.value.findIndex((p) => p.uid === panel.uid)
  if (idx !== -1) {
    panels.value.splice(idx, 1)
    minimizeStore.delete(panel.uid)
    reindex()
  }
}

/** Sort panels array to match actual DOM order of their elements */
function sortPanelsByDom() {
  panels.value.sort((a, b) => {
    if (!a.el || !b.el) return 0
    // Node.DOCUMENT_POSITION_FOLLOWING = 4 means a is before b
    const pos = a.el.compareDocumentPosition(b.el)
    if (pos & Node.DOCUMENT_POSITION_FOLLOWING) return -1
    if (pos & Node.DOCUMENT_POSITION_PRECEDING) return 1
    return 0
  })
  reindex()
}

function reindex() {
  panels.value.forEach((p, i) => {
    p.index = i
  })
}

// Watch panel changes to reset moving state
watch(panels, () => {
  movingIndex.value = null
  reindex()
})

// Size calculation (raw — may violate min/max after container resize)
const { pxSizes: rawPxSizes } = useSize(panels, availableSize)

// Global max panel size in px
const maxPanelSizePx = computed(() => {
  if (props.maxPanelSize === undefined || props.maxPanelSize === null) return Infinity
  return parseSizeToPx(props.maxPanelSize, availableSize.value, Infinity)
})

// ─── Clamped sizes (always respect min/max) ─────────────────────────────────
// useSize stores ratios → when the container grows, px sizes scale proportionally
// and can exceed per-panel maxSize or global maxPanelSize. Instead of trying to
// fix this reactively via a watcher (which has timing races), we wrap the raw
// pxSizes in a computed that clamps and redistributes on every read.

const pxSizes = computed(() => {
  const raw = [...rawPxSizes.value]
  const avail = availableSize.value
  if (avail <= 0) return raw

  const nonMin = panels.value.filter((p) => !p.minimized)
  if (nonMin.length < 2) return raw

  // EPS 是 “epsilon（极小容差）” 的缩写，在这里表示比较浮点数时允许的误差阈值（当前设为 0.5 像素）
  const EPS = 0.5
  const globalMax = maxPanelSizePx.value
  const indices = nonMin.map((p) => p.index)
  const minLimits: number[] = []
  const maxLimits: number[] = []
  let targetTotal = 0

  // Pass 1 — clamp each panel to its own bounds.
  for (const p of nonMin) {
    const i = p.index
    const current = raw[i] ?? 0
    const minPx = Math.max(parseSizeToPx(p.minSize, avail, 0), 0)
    const maxPx = Math.min(
      parseSizeToPx(p.maxSize, avail, Infinity),
      isFinite(globalMax) ? globalMax : Infinity,
    )
    const effectiveMax = Math.max(maxPx, minPx)
    minLimits[i] = minPx
    maxLimits[i] = effectiveMax
    targetTotal += current
    if (current < minPx) {
      raw[i] = minPx
    } else if (current > effectiveMax) {
      raw[i] = effectiveMax
    } else {
      raw[i] = current
    }
  }

  // Pass 2 — iteratively redistribute while keeping all panels within bounds.
  // This avoids single-pass redistribution from pushing another panel past min/max.
  let currentTotal = indices.reduce((sum, i) => sum + (raw[i] ?? 0), 0)
  let delta = targetTotal - currentTotal
  let guard = 0
  const maxGuard = indices.length * 8

  while (Math.abs(delta) > EPS && guard < maxGuard) {
    guard++
    if (delta > 0) {
      const expandable = indices.filter((i) => (maxLimits[i] ?? Infinity) - (raw[i] ?? 0) > EPS)
      if (expandable.length === 0) break
      const infiniteExpandable = expandable.filter((i) =>
        !isFinite((maxLimits[i] ?? Infinity) - (raw[i] ?? 0)),
      )
      if (infiniteExpandable.length > 0) {
        const apply = delta
        const totalWeight = infiniteExpandable.reduce((sum, i) => sum + (raw[i] ?? 0), 0)
        infiniteExpandable.forEach((i) => {
          const share =
            totalWeight > 0 ? (raw[i] ?? 0) / totalWeight : 1 / infiniteExpandable.length
          raw[i] = (raw[i] ?? 0) + apply * share
        })
      } else {
        const totalRoom = expandable.reduce(
          (sum, i) => sum + ((maxLimits[i] ?? Infinity) - (raw[i] ?? 0)),
          0,
        )
        if (totalRoom <= EPS) break
        const apply = Math.min(delta, totalRoom)
        expandable.forEach((i) => {
          const room = (maxLimits[i] ?? Infinity) - (raw[i] ?? 0)
          raw[i] = (raw[i] ?? 0) + apply * (room / totalRoom)
        })
      }
    } else {
      const shrinkable = indices.filter((i) => (raw[i] ?? 0) - (minLimits[i] ?? 0) > EPS)
      if (shrinkable.length === 0) break
      const totalRoom = shrinkable.reduce(
        (sum, i) => sum + ((raw[i] ?? 0) - (minLimits[i] ?? 0)),
        0,
      )
      if (totalRoom <= EPS) break
      const apply = Math.min(-delta, totalRoom)
      shrinkable.forEach((i) => {
        const room = (raw[i] ?? 0) - (minLimits[i] ?? 0)
        raw[i] = (raw[i] ?? 0) - apply * (room / totalRoom)
      })
    }
    currentTotal = indices.reduce((sum, i) => sum + (raw[i] ?? 0), 0)
    delta = targetTotal - currentTotal
  }

  return raw
})

// Clamped percent sizes derived from clamped px sizes
const percentSizes = computed(() => {
  const avail = availableSize.value
  return pxSizes.value.map((px) => (avail > 0 ? px / avail : 0))
})

// Resize logic (uses clamped pxSizes so drag always starts from a valid state)
const { lazyOffset, movingIndex, onMoveStart, onMoving, onMoveEnd } = useResize(
  panels,
  availableSize,
  pxSizes,
  lazy,
  maxPanelSizePx,
)

// Event wrappers
const onResizeStart = (index: number) => {
  onMoveStart(index)
  emit('resizeStart', index, [...pxSizes.value])
}

const onResize = (index: number, offset: number) => {
  onMoving(index, offset)
  if (!lazy.value) {
    emit('resize', index, [...pxSizes.value])
  }
}

const onResizeEnd = async (index: number) => {
  onMoveEnd()
  await nextTick()
  emit('resizeEnd', index, [...pxSizes.value])
}

// Double-click: reset the two adjacent visible panels to equal (center) size
const onResetCenter = (index: number) => {
  const leftPanel = panels.value[index]
  if (!leftPanel || leftPanel.minimized) return

  // Find the next non-minimized panel (skip minimized ones)
  let rightPanel: PanelState | undefined
  for (let i = index + 1; i < panels.value.length; i++) {
    if (!panels.value[i]?.minimized) {
      rightPanel = panels.value[i]
      break
    }
  }
  if (!rightPanel) return
  if (!leftPanel.resizable || !rightPanel.resizable) return

  const leftSize = pxSizes.value[leftPanel.index] ?? 0
  const rightSize = pxSizes.value[rightPanel.index] ?? 0
  const total = leftSize + rightSize
  let half = total / 2

  // Respect min/max constraints
  const cSize = availableSize.value || 0
  const leftMin = parseSizeToPx(leftPanel.minSize, cSize, 0)
  const leftMax = parseSizeToPx(leftPanel.maxSize, cSize, total)
  const rightMin = parseSizeToPx(rightPanel.minSize, cSize, 0)
  const rightMax = parseSizeToPx(rightPanel.maxSize, cSize, total)

  // Clamp left panel
  if (half < leftMin) half = leftMin
  if (half > leftMax) half = leftMax
  // Clamp right panel (total - half)
  if (total - half < rightMin) half = total - rightMin
  if (total - half > rightMax) half = total - rightMax

  // Clamp by global maxPanelSize
  const mps = maxPanelSizePx.value
  if (isFinite(mps)) {
    if (half > mps) half = mps
    if (total - half > mps) half = total - mps
  }

  leftPanel.size = half
  rightPanel.size = total - half

  nextTick(() => {
    emit('resetCenter', index, [...pxSizes.value])
  })
}

// ─── Minimize / Restore ─────────────────────────────────────────────────────

/** Internal store: saved ratio & containerSize at the moment of minimization */
const minimizeStore = new Map<number, { ratio: number; containerSize: number }>()

/** Minimize a panel by its uid (called from context) */
function minimizePanelByUid(uid: number) {
  const panel = panels.value.find((p) => p.uid === uid)
  if (!panel) return

  const idx = panel.index
  const currentPx = pxSizes.value[idx] ?? 0
  const alreadyCollapsed = panel.minimized && currentPx <= 0.5
  if (alreadyCollapsed) return

  // Prevent minimizing the last visible panel
  // If this panel is flagged minimized but not actually collapsed yet (initial state),
  // count it as visible for this guard.
  const nonMinCount =
    panels.value.filter((p) => !p.minimized).length + (panel.minimized ? 1 : 0)
  if (nonMinCount <= 1) return

  const ratio = availableSize.value > 0 ? currentPx / availableSize.value : 0

  minimizeStore.set(uid, {
    ratio,
    containerSize: containerSize.value,
  })

  panel.minimized = true
  panel.size = 0

  nextTick(() => {
    emit('panelMinimized', idx)
  })
}

/** Restore a minimized panel by its uid (called from context) */
function restorePanelByUid(uid: number) {
  const panel = panels.value.find((p) => p.uid === uid)
  if (!panel || !panel.minimized) return

  const saved = minimizeStore.get(uid)
  minimizeStore.delete(uid)

  const containerChanged =
    saved != null && Math.abs(saved.containerSize - containerSize.value) > 1

  panel.minimized = false

  if (containerChanged || !saved) {
    // Container size changed while minimized → center all non-minimized panels
    centerNonMinimizedPanels()
  } else {
    // Restore to the saved ratio
    const restoredPx = saved.ratio * availableSize.value
    panel.size = restoredPx

    // Proportionally shrink other non-minimized panels to make room
    const otherNonMin = panels.value.filter((p) => p.uid !== uid && !p.minimized)
    const totalOtherPx = otherNonMin.reduce(
      (sum, p) => sum + (pxSizes.value[p.index] ?? 0),
      0,
    )
    const targetOtherTotal = availableSize.value - restoredPx

    if (totalOtherPx > 0 && targetOtherTotal > 0) {
      const scale = targetOtherTotal / totalOtherPx
      otherNonMin.forEach((p) => {
        p.size = (pxSizes.value[p.index] ?? 0) * scale
      })
    }
  }

  nextTick(() => {
    emit('panelRestored', panel.index)
  })
}

/** Toggle minimize state by uid */
function togglePanelByUid(uid: number) {
  const panel = panels.value.find((p) => p.uid === uid)
  if (!panel) return
  if (panel.minimized) {
    restorePanelByUid(uid)
  } else {
    minimizePanelByUid(uid)
  }
}

/** Distribute all non-minimized panels equally */
function centerNonMinimizedPanels() {
  const nonMin = panels.value.filter((p) => !p.minimized)
  if (nonMin.length === 0) return
  const equalPx = availableSize.value / nonMin.length
  nonMin.forEach((p) => {
    p.size = equalPx
  })
  panels.value
    .filter((p) => p.minimized)
    .forEach((p) => {
      p.size = 0
    })
}

// ─── External imperative API (by panel index) ──────────────────────────────

/** Minimize a panel by its index (0-based). Exposed for external use. */
function minimizePanel(index: number) {
  const panel = panels.value.find((p) => p.index === index)
  if (panel) minimizePanelByUid(panel.uid)
}

/** Restore a minimized panel by its index. Exposed for external use. */
function restorePanel(index: number) {
  const panel = panels.value.find((p) => p.index === index)
  if (panel) restorePanelByUid(panel.uid)
}

/** Toggle minimize state by index. Exposed for external use. */
function togglePanel(index: number) {
  const panel = panels.value.find((p) => p.index === index)
  if (panel) togglePanelByUid(panel.uid)
}

/** Check if a panel at the given index is minimized */
function isPanelMinimized(index: number): boolean {
  const panel = panels.value.find((p) => p.index === index)
  return panel?.minimized ?? false
}

defineExpose({
  minimizePanel,
  restorePanel,
  togglePanel,
  isPanelMinimized,
  panels,
  pxSizes,
  percentSizes,
})

// Provide context to child panels
provide(splitPaneContextKey, {
  layout,
  lazy,
  gap,
  indicatorSize,
  enableFileDrop,
  enablePanelDrag,
  panels,
  pxSizes,
  percentSizes,
  containerSize,
  movingIndex,
  registerPanel,
  unregisterPanel,
  onMoveStart: onResizeStart,
  onMoving: onResize,
  onMoveEnd: onResizeEnd,
  onDblClick: onResetCenter,
  minimizePanel: minimizePanelByUid,
  restorePanel: restorePanelByUid,
  togglePanel: togglePanelByUid,
  onPanelDrop,
})
</script>

<style scoped>
.split-pane {
  display: flex;
  width: 100%;
  height: 100%;
  overflow: hidden;
  position: relative;
  box-sizing: border-box;
}

.split-pane--horizontal {
  flex-direction: row;
}

.split-pane--vertical {
  flex-direction: column;
}

.split-pane__mask {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
}

.split-pane__mask--horizontal {
  cursor: col-resize;
}

.split-pane__mask--vertical {
  cursor: row-resize;
}
</style>
