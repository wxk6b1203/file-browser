<template>
  <div
    ref="barEl"
    :class="['tab-bar']"
    :style="barStyle"
  >
    <TabHeader
      v-for="tab in tabs"
      :key="tab.id"
      :ref="(el: any) => setHeaderRef(tab.id, el)"
      :tab="tab"
      :group-id="groupId"
      :is-active="tab.id === activeId"
      :translate-x="getTranslateX(tab.id)"
      @drag-start="onHeaderDragStart"
      @close="onCloseTab"
      @select="onSelectTab"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, inject, watch, type CSSProperties, onBeforeUpdate } from 'vue'
import TabHeader from './TabHeader.vue'
import type { TabItem } from './types'
import { tabsContextKey } from './types'

const props = defineProps<{
  tabs: TabItem[]
  activeId: string
  groupId: string
}>()

const emit = defineEmits<{
  activate: [tabId: string]
  close: [tab: TabItem]
  dragStart: [tab: TabItem, event: PointerEvent, el: HTMLElement]
  reorder: [fromIndex: number, toIndex: number]
}>()

const ctx = inject(tabsContextKey)!
const barEl = ref<HTMLElement>()

// Track translate offsets for reorder animation
const translateMap = ref<Map<string, number>>(new Map())

// Track header component refs for measuring positions
const headerRefs = new Map<string, any>()

onBeforeUpdate(() => {
  headerRefs.clear()
})

function setHeaderRef(tabId: string, el: any) {
  if (el) headerRefs.set(tabId, el)
}

function getTranslateX(tabId: string): number {
  return translateMap.value.get(tabId) ?? 0
}

const barStyle = computed<CSSProperties>(() => {
  const s: CSSProperties = {}
  if (ctx.barBackground.value) {
    s.backgroundColor = ctx.barBackground.value
  }
  return s
})

// ─── Reorder animation during drag ─────────────────────────

/**
 * While dragging, watch the pointer position and determine where the
 * dragged tab would land. Shift other headers with CSS transforms.
 */
watch(
  () => [ctx.drag.pointerX, ctx.drag.pointerY, ctx.drag.active] as const,
  ([px, _py, active]) => {
    if (!active || !barEl.value) {
      // Clear transforms when drag ends
      if (translateMap.value.size > 0) {
        translateMap.value = new Map()
      }
      return
    }
    // Only animate reorder within the same group
    if (ctx.drag.sourceGroupId !== props.groupId) {
      if (translateMap.value.size > 0) {
        translateMap.value = new Map()
      }
      return
    }
    // Check if pointer is within the tab bar area
    const barRect = barEl.value.getBoundingClientRect()
    if (px < barRect.left || px > barRect.right || _py < barRect.top || _py > barRect.bottom) {
      if (translateMap.value.size > 0) {
        translateMap.value = new Map()
      }
      return
    }

    const dragTabId = ctx.drag.tab?.id
    if (!dragTabId) return

    // Measure all header positions
    const headerEls = barEl.value.querySelectorAll('.tab-header')
    if (!headerEls.length) return

    const rects: { id: string; left: number; right: number; center: number; width: number }[] = []
    for (let i = 0; i < props.tabs.length; i++) {
      const tab = props.tabs[i]
      if (!tab) continue
      const el = headerEls[i]
      if (!el) continue
      const rect = el.getBoundingClientRect()
      rects.push({
        id: tab.id,
        left: rect.left,
        right: rect.right,
        center: rect.left + rect.width / 2,
        width: rect.width,
      })
    }

    const dragIndex = rects.findIndex((r) => r.id === dragTabId)
    if (dragIndex === -1) return

    const dragRect = rects[dragIndex]
    if (!dragRect) return

    // Determine insertion index: where the pointer sits relative to other tab centers
    let insertIndex = rects.length
    for (let i = 0; i < rects.length; i++) {
      const r = rects[i]
      if (!r) continue
      if (r.id === dragTabId) continue
      if (px < r.center) {
        insertIndex = i
        break
      }
    }

    // Build translate map: shift tabs that need to move
    const newMap = new Map<string, number>()
    for (let i = 0; i < rects.length; i++) {
      const r = rects[i]
      if (!r || r.id === dragTabId) continue

      if (dragIndex < insertIndex) {
        // Dragging right: tabs between dragIndex+1..insertIndex-1 shift left
        if (i > dragIndex && i < insertIndex) {
          newMap.set(r.id, -dragRect.width)
        }
      } else if (dragIndex > insertIndex) {
        // Dragging left: tabs between insertIndex..dragIndex-1 shift right
        if (i >= insertIndex && i < dragIndex) {
          newMap.set(r.id, dragRect.width)
        }
      }
    }

    translateMap.value = newMap
  },
)

// ─── Event handlers ─────────────────────────────────────────

function onSelectTab(tab: TabItem) {
  emit('activate', tab.id)
}

function onCloseTab(tab: TabItem) {
  emit('close', tab)
}

function onHeaderDragStart(tab: TabItem, event: PointerEvent, el: HTMLElement) {
  emit('dragStart', tab, event, el)
}

defineExpose({ barEl })
</script>

<style scoped>
.tab-bar {
  display: flex;
  align-items: stretch;
  overflow-x: auto;
  overflow-y: hidden;
  background-color: var(--theme-color-bg-surface);
  border-bottom: 1px solid var(--theme-color-border);
  flex-shrink: 0;
  position: relative;
  transition: background-color 0.2s ease, border-color 0.2s ease;
}

.tab-bar::-webkit-scrollbar {
  height: var(--ui-scrollbar-size);
}

.tab-bar::-webkit-scrollbar-thumb {
  border: calc((var(--ui-scrollbar-size) - var(--ui-scrollbar-thumb-size)) / 2) solid transparent;
  border-radius: 999px;
  background-color: var(--theme-color-scrollbar);
  background-clip: content-box;
}
</style>
