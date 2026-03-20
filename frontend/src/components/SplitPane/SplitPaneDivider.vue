<template>
  <div
    :class="[
      'split-pane-divider',
      `split-pane-divider--${layout}`,
      { 'split-pane-divider--active': isDragging },
      { 'split-pane-divider--disabled': !resizable },
    ]"
    :style="wrapStyle"
  >
    <div
      :class="['split-pane-divider__dragger', `split-pane-divider__dragger--${layout}`]"
      :style="draggerStyle"
      @pointerdown="onPointerDown"
    >
      <div
        ref="indicatorRef"
        :class="['split-pane-divider__indicator', `split-pane-divider__indicator--${layout}`]"
        :style="indicatorStyle"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { SplitLayout } from './types'

const props = withDefaults(
  defineProps<{
    index: number
    layout: SplitLayout
    resizable?: boolean
    gap?: number
    /** Inset from the start edge (top for horizontal, left for vertical) in px */
    insetStart?: number
    /** Inset from the end edge (bottom for horizontal, right for vertical) in px */
    insetEnd?: number
    /** Indicator thickness (short side), CSS value */
    indicatorWidth?: string
    /** Indicator length (long side), CSS value */
    indicatorHeight?: string
  }>(),
  {
    resizable: true,
    gap: 0,
    insetStart: 0,
    insetEnd: 0,
    indicatorWidth: '2px',
    indicatorHeight: '24px',
  },
)

const emit = defineEmits<{
  moveStart: [index: number]
  moving: [index: number, offset: number]
  moveEnd: [index: number]
  dblclick: [index: number]
}>()

const isHorizontal = computed(() => props.layout === 'horizontal')

// Indicator style: for horizontal, width=short side, height=long side; for vertical, swap
const indicatorStyle = computed(() => {
  if (isHorizontal.value) {
    return { width: props.indicatorWidth, height: props.indicatorHeight }
  } else {
    return { width: props.indicatorHeight, height: props.indicatorWidth }
  }
})

const wrapStyle = computed(() => {
  return isHorizontal.value
    ? { width: `${props.gap}px` }
    : { height: `${props.gap}px` }
})

const draggerStyle = computed(() => {
  const minHit = 12
  const size = Math.max(props.gap, minHit)
  const overflows = size > props.gap
  const style: Record<string, string> = {
    cursor: !props.resizable ? 'default' : isHorizontal.value ? 'col-resize' : 'row-resize',
  }

  if (isHorizontal.value) {
    style.width = `${size}px`
    // Use absolute positioning to apply top/bottom insets for rounded corners
    if (props.insetStart || props.insetEnd) {
      style.position = 'absolute'
      style.top = `${props.insetStart}px`
      style.bottom = `${props.insetEnd}px`
      // Center horizontally within the wrap
      style.left = '50%'
      style.transform = 'translateX(-50%)'
    } else {
      style.height = '100%'
      // Parent flex (justify-content:center) handles centering when dragger overflows gap
    }
  } else {
    style.height = `${size}px`
    // Use absolute positioning to apply left/right insets for rounded corners
    if (props.insetStart || props.insetEnd) {
      style.position = 'absolute'
      style.left = `${props.insetStart}px`
      style.right = `${props.insetEnd}px`
      // Center vertically within the wrap
      style.top = '50%'
      style.transform = 'translateY(-50%)'
    } else {
      style.width = '100%'
      // Parent flex (align-items:center) handles centering when dragger overflows gap
    }
  }

  return style
})

const indicatorRef = ref<HTMLElement | null>(null)
const isDragging = ref(false)
let startPos: [number, number] | null = null

// ─── Manual double-click detection ─────────────────────────────────────────
// WebKit on macOS suppresses native dblclick when pointerdown calls
// preventDefault() + setPointerCapture(). We detect double-click ourselves
// by tracking the timestamp and position of the last pointerup.
// Only the indicator (the visible bar) supports double-click-to-center;
// clicking the wider dragger hit-area does NOT trigger it.
const DBLCLICK_INTERVAL = 300 // ms
const DBLCLICK_DISTANCE = 7  // px – max movement between two clicks
let lastUpTime = 0
let lastUpPos: [number, number] = [0, 0]

function onPointerDown(e: PointerEvent) {
  if (!props.resizable) return
  e.preventDefault()

  const isOnIndicator = indicatorRef.value?.contains(e.target as Node) ?? false

  // Check for double-click: only when clicking on the indicator
  if (isOnIndicator) {
    const now = performance.now()
    const dx = e.pageX - lastUpPos[0]
    const dy = e.pageY - lastUpPos[1]
    const dist = Math.sqrt(dx * dx + dy * dy)

    if (now - lastUpTime < DBLCLICK_INTERVAL && dist < DBLCLICK_DISTANCE) {
      // Treat as double-click — reset lastUpTime so a third tap doesn't re-trigger
      lastUpTime = 0
      emit('dblclick', props.index)
      return
    }
  }

  ;(e.target as HTMLElement).setPointerCapture(e.pointerId)
  isDragging.value = true
  startPos = [e.pageX, e.pageY]
  let hasMoved = false
  emit('moveStart', props.index)

  const target = e.target as HTMLElement
  const onPointerMove = (ev: PointerEvent) => {
    if (!startPos) return
    const offset = isHorizontal.value
      ? ev.pageX - startPos[0]
      : ev.pageY - startPos[1]
    if (Math.abs(offset) > 2) hasMoved = true
    emit('moving', props.index, offset)
  }

  const onPointerUp = (ev: PointerEvent) => {
    target.releasePointerCapture(ev.pointerId)
    isDragging.value = false
    startPos = null
    // Only record for double-click detection if no actual dragging occurred
    // and the original pointerdown was on the indicator
    if (!hasMoved && isOnIndicator) {
      lastUpTime = performance.now()
      lastUpPos = [ev.pageX, ev.pageY]
    } else {
      lastUpTime = 0
    }
    emit('moveEnd', props.index)
    target.removeEventListener('pointermove', onPointerMove)
    target.removeEventListener('pointerup', onPointerUp)
    target.removeEventListener('pointercancel', onPointerUp)
  }

  target.addEventListener('pointermove', onPointerMove)
  target.addEventListener('pointerup', onPointerUp)
  target.addEventListener('pointercancel', onPointerUp)
}
</script>

<style scoped>
.split-pane-divider {
  position: relative;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1;
}

.split-pane-divider--horizontal {
  /* width set by inline style (gap) */
}

.split-pane-divider--vertical {
  /* height set by inline style (gap) */
}

.split-pane-divider__dragger {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  touch-action: none;
  user-select: none;
  transition: background-color 0.15s ease;
}

.split-pane-divider__dragger:hover,
.split-pane-divider--active .split-pane-divider__dragger {
  background-color: var(--theme-color-divider-hover-bg);
}

.split-pane-divider--disabled .split-pane-divider__dragger {
  cursor: default !important;
}

.split-pane-divider--disabled .split-pane-divider__dragger:hover {
  background-color: transparent;
}

.split-pane-divider__indicator {
  flex-shrink: 0;
  border-radius: 1px;
  background-color: var(--theme-color-divider);
  transition: background-color 0.15s ease;
}


.split-pane-divider__dragger:hover .split-pane-divider__indicator,
.split-pane-divider--active .split-pane-divider__indicator {
  background-color: var(--theme-color-divider-active);
}

.split-pane-divider--disabled .split-pane-divider__dragger:hover .split-pane-divider__indicator {
  background-color: var(--theme-color-divider);
}
</style>

