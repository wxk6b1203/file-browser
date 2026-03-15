<template>
  <div
    ref="headerEl"
    :class="[
      'tab-header',
      { 'tab-header--active': isActive },
      { 'tab-header--dragging': isDragging },
      { 'tab-header--placeholder': isPlaceholder },
    ]"
    :style="headerStyle"
    @pointerdown="onPointerDown"
  >
    <span class="tab-header__label">{{ tab.label }}</span>
    <span
      v-if="tab.closable !== false"
      class="tab-header__close"
      @pointerdown.stop
      @click.stop="onClose"
    >
      ×
    </span>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, inject, type CSSProperties } from 'vue'
import type { TabItem } from './types'
import { tabsContextKey } from './types'

/** Minimum pointer movement (px) before a press is treated as drag instead of click */
const DRAG_THRESHOLD = 5

const props = defineProps<{
  tab: TabItem
  groupId: string
  isActive: boolean
  /** External transform to apply (for reorder animation) */
  translateX?: number
}>()

const emit = defineEmits<{
  close: [tab: TabItem]
  dragStart: [tab: TabItem, event: PointerEvent, el: HTMLElement]
  select: [tab: TabItem]
}>()

const ctx = inject(tabsContextKey)!
const headerEl = ref<HTMLElement>()

const isDragging = computed(() => {
  return ctx.drag.active && ctx.drag.tab?.id === props.tab.id
})

const isPlaceholder = computed(() => {
  return isDragging.value
})

const headerStyle = computed<CSSProperties>(() => {
  const s: CSSProperties = {}
  if (props.translateX) {
    s.transform = `translateX(${props.translateX}px)`
    s.transition = 'transform 0.2s ease'
  }
  if (isPlaceholder.value) {
    s.opacity = 0.3
  }
  return s
})

function onPointerDown(e: PointerEvent) {
  if (e.button !== 0) return
  e.preventDefault()

  const el = headerEl.value
  if (!el) return

  const startX = e.clientX
  const startY = e.clientY
  // Keep the original PointerEvent for dragStart so offset calculations are correct
  const startEvent = e

  const onMove = (ev: PointerEvent) => {
    const dx = ev.clientX - startX
    const dy = ev.clientY - startY
    if (Math.abs(dx) + Math.abs(dy) >= DRAG_THRESHOLD) {
      // Exceeded threshold → start drag
      cleanup()
      emit('dragStart', props.tab, startEvent, el)
    }
  }

  const onUp = () => {
    // Released before threshold → treat as click (select tab)
    cleanup()
    emit('select', props.tab)
  }

  const cleanup = () => {
    document.removeEventListener('pointermove', onMove)
    document.removeEventListener('pointerup', onUp)
    document.removeEventListener('pointercancel', onUp)
  }

  document.addEventListener('pointermove', onMove)
  document.addEventListener('pointerup', onUp)
  document.addEventListener('pointercancel', onUp)
}

function onClose() {
  emit('close', props.tab)
}
</script>

<style scoped>
.tab-header {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  cursor: default;
  user-select: none;
  white-space: nowrap;
  font-size: 13px;
  color: var(--theme-color-text);
  border-bottom: 2px solid transparent;
  transition: background-color 0.15s ease, border-color 0.15s ease, opacity 0.15s ease, color 0.15s ease;
  flex-shrink: 0;
  position: relative;
}

.tab-header:hover {
  background-color: var(--theme-color-bg-hover);
}

.tab-header--active {
  color: var(--theme-color-primary);
  border-bottom-color: var(--theme-color-primary);
}

.tab-header--dragging {
  opacity: 0.3;
}

.tab-header__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  font-size: 14px;
  line-height: 1;
  border-radius: 3px;
  cursor: pointer;
  color: var(--theme-color-text-secondary);
  transition: background-color 0.15s, color 0.15s;
}

.tab-header__close:hover {
  background-color: var(--theme-color-close-hover);
  color: var(--theme-color-text-base);
}
</style>

