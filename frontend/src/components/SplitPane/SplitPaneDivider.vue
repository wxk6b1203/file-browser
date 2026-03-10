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
        :class="['split-pane-divider__indicator', `split-pane-divider__indicator--${layout}`]"
        :style="indicatorStyle"
        @dblclick.stop="onIndicatorDblClick"
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
      if (overflows) {
        style.transform = 'translateX(-50%)'
      }
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
      if (overflows) {
        style.transform = 'translateY(-50%)'
      }
    }
  }

  return style
})

const isDragging = ref(false)
let startPos: [number, number] | null = null

function onPointerDown(e: PointerEvent) {
  if (!props.resizable) return
  e.preventDefault()
  ;(e.target as HTMLElement).setPointerCapture(e.pointerId)
  isDragging.value = true
  startPos = [e.pageX, e.pageY]
  emit('moveStart', props.index)

  const target = e.target as HTMLElement
  const onPointerMove = (ev: PointerEvent) => {
    if (!startPos) return
    const offset = isHorizontal.value
      ? ev.pageX - startPos[0]
      : ev.pageY - startPos[1]
    emit('moving', props.index, offset)
  }

  const onPointerUp = (ev: PointerEvent) => {
    target.releasePointerCapture(ev.pointerId)
    isDragging.value = false
    startPos = null
    emit('moveEnd', props.index)
    target.removeEventListener('pointermove', onPointerMove)
    target.removeEventListener('pointerup', onPointerUp)
    target.removeEventListener('pointercancel', onPointerUp)
  }

  target.addEventListener('pointermove', onPointerMove)
  target.addEventListener('pointerup', onPointerUp)
  target.addEventListener('pointercancel', onPointerUp)
}

function onIndicatorDblClick() {
  if (!props.resizable) return
  emit('dblclick', props.index)
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
  background-color: rgba(59, 130, 246, 0.12);
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
  background-color: #d1d5db;
  transition: background-color 0.15s ease;
}


.split-pane-divider__dragger:hover .split-pane-divider__indicator,
.split-pane-divider--active .split-pane-divider__indicator {
  background-color: #3b82f6;
}

.split-pane-divider--disabled .split-pane-divider__dragger:hover .split-pane-divider__indicator {
  background-color: #d1d5db;
}
</style>

