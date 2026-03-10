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
import { splitPaneContextKey, type PanelState, type SplitLayout } from './types'
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
  }>(),
  {
    layout: 'horizontal',
    lazy: false,
    gap: 0,
    indicatorWidth: '2px',
    indicatorHeight: '24px',
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
}>()

const layout = toRef(props, 'layout')
const lazy = toRef(props, 'lazy')
const gap = toRef(props, 'gap')

// Indicator size: [width(short side), height(long side)]
const indicatorSize = computed<[string, string]>(() => [props.indicatorWidth, props.indicatorHeight])

// Container size tracking
const { containerEl, containerSize } = useContainer(layout)

// Panel registry
const panels = ref<PanelState[]>([])

// Available size = container size minus total gap between panels
const availableSize = computed(() => {
  const dividerCount = Math.max(panels.value.length - 1, 0)
  return Math.max(containerSize.value - dividerCount * gap.value, 0)
})

function registerPanel(panel: PanelState) {
  panels.value.push(panel)
  sortPanelsByDom()
}

function unregisterPanel(panel: PanelState) {
  const idx = panels.value.findIndex((p) => p.uid === panel.uid)
  if (idx !== -1) {
    panels.value.splice(idx, 1)
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

// Size calculation
const { percentSizes, pxSizes } = useSize(panels, availableSize)

// Resize logic
const { lazyOffset, movingIndex, onMoveStart, onMoving, onMoveEnd } = useResize(
  panels,
  availableSize,
  pxSizes,
  lazy,
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

// Double-click: reset the two adjacent panels to equal (center) size
const onResetCenter = (index: number) => {
  const leftPanel = panels.value[index]
  const rightPanel = panels.value[index + 1]
  if (!leftPanel || !rightPanel) return
  if (!leftPanel.resizable || !rightPanel.resizable) return

  const leftSize = pxSizes.value[index] ?? 0
  const rightSize = pxSizes.value[index + 1] ?? 0
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

  leftPanel.size = half
  rightPanel.size = total - half

  nextTick(() => {
    emit('resetCenter', index, [...pxSizes.value])
  })
}

// Provide context to child panels
provide(splitPaneContextKey, {
  layout,
  lazy,
  gap,
  indicatorSize,
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

