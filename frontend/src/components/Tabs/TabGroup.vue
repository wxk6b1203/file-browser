<template>
  <div ref="groupEl" class="tab-group" :data-group-id="node.id" :data-active-tab-id="node.activeId" :style="groupStyle">
    <TabBar
      :tabs="node.tabs"
      :active-id="node.activeId"
      :group-id="node.id"
      @activate="onActivate"
      @close="onClose"
      @drag-start="onDragStart"
    />
    <div ref="contentEl" class="tab-group__content" :style="contentStyle">
      <component
        v-if="activeTab && activeTab.component"
        :is="activeTab.component"
        v-bind="activeTab.props"
      />
      <TabDropOverlay
        :visible="showOverlay"
        :zone="currentZone"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, inject, watch, type CSSProperties } from 'vue'
import TabBar from './TabBar.vue'
import TabDropOverlay from './TabDropOverlay.vue'
import type { TabGroupNode, DropZone, TabItem } from './types'
import { tabsContextKey } from './types'
import { calcDropZone } from './composables/useDropZone'

const props = defineProps<{
  node: TabGroupNode
}>()

const ctx = inject(tabsContextKey)!
const groupEl = ref<HTMLElement>()
const contentEl = ref<HTMLElement>()
const currentZone = ref<DropZone>(null)

const activeTab = computed(() => {
  return props.node.tabs.find((t) => t.id === props.node.activeId) ?? props.node.tabs[0] ?? null
})

const showOverlay = computed(() => {
  if (!ctx.drag.active) return false
  // Don't show overlay on the source group if it only has one tab
  if (ctx.drag.sourceGroupId === props.node.id && props.node.tabs.length <= 1) return false
  return currentZone.value !== null && currentZone.value !== 'center'
})

const groupStyle = computed<CSSProperties>(() => {
  return {
    width: '100%',
    height: '100%',
  }
})

const contentStyle = computed<CSSProperties>(() => {
  const s: CSSProperties = {}
  if (ctx.tabBackground.value) {
    s.backgroundColor = ctx.tabBackground.value
  }
  return s
})

function onActivate(tabId: string) {
  ctx.setActive(props.node.id, tabId)
  const tab = props.node.tabs.find((t) => t.id === tabId)
  if (tab) {
    ctx.emitTabActivate(tab, props.node.id)
  }
}

function onClose(tab: TabItem) {
  ctx.removeTab(props.node.id, tab.id)
}

function onDragStart(tab: TabItem, event: PointerEvent, el: HTMLElement) {
  ctx.drag.active = true
  ctx.drag.tab = tab
  ctx.drag.sourceGroupId = props.node.id

  const rect = el.getBoundingClientRect()
  ctx.drag.offsetX = event.clientX - rect.left
  ctx.drag.offsetY = event.clientY - rect.top
  ctx.drag.headerWidth = rect.width
  ctx.drag.headerHeight = rect.height
  ctx.drag.pointerX = event.clientX
  ctx.drag.pointerY = event.clientY

  ctx.emitTabDragStart(tab, props.node.id)
  ctx.setActive(props.node.id, tab.id)

  // Start global listeners (managed by root Tabs.vue)
}

// Watch drag state to compute drop zones
watch(
  () => [ctx.drag.pointerX, ctx.drag.pointerY, ctx.drag.active] as const,
  ([px, py, active]) => {
    if (!active || !contentEl.value) {
      currentZone.value = null
      return
    }
    const rect = contentEl.value.getBoundingClientRect()
    // Check if the pointer is within the content panel
    if (px >= rect.left && px <= rect.right && py >= rect.top && py <= rect.bottom) {
      // Check min-size constraints before allowing split
      const canSplitH = rect.height >= ctx.minSplitHeight.value * 2
      const canSplitV = rect.width >= ctx.minSplitWidth.value * 2
      const rawZone = calcDropZone(px, py, rect)

      if (rawZone === 'top' || rawZone === 'bottom') {
        currentZone.value = canSplitH ? rawZone : 'center'
      } else if (rawZone === 'left' || rawZone === 'right') {
        currentZone.value = canSplitV ? rawZone : 'center'
      } else {
        currentZone.value = rawZone
      }
    } else {
      currentZone.value = null
    }
  },
)

defineExpose({
  groupEl,
  contentEl,
  currentZone,
  node: props.node,
})
</script>

<style scoped>
.tab-group {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.tab-group__content {
  flex: 1;
  overflow: auto;
  position: relative;
}
</style>


