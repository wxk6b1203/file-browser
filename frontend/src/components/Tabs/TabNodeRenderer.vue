<template>
  <TabGroup
    v-if="node.type === 'tabs'"
    :node="node"
    :enable-file-drop="enableFileDrop"
    :enable-panel-drag="enablePanelDrag"
    :key="node.id"
    :data-node-id="node.id"
    @panel-drop="onPanelDrop"
  />
  <SplitPane
    v-else
    :layout="splitNode.layout"
    :gap="4"
    :enable-panel-drag="enablePanelDrag"
    :key="node.id"
    :data-node-id="node.id"
    @resize-end="onSplitResizeEnd"
    @panel-drop="onPanelDrop"
  >
    <SplitPanePanel
      v-for="(child, i) in splitNode.children"
      :key="child.id"
      :size="splitNode.sizes?.[i] ?? `${100 / splitNode.children.length}%`"
    >
      <TabNodeRenderer
        :node="child"
        :enable-panel-drag="enablePanelDrag"
        :enable-file-drop="enableFileDrop"
        @panel-drop="onPanelDrop"
      />
    </SplitPanePanel>
  </SplitPane>
</template>

<script setup lang="ts">
import { computed, inject } from 'vue'
import type { TabNode, TabSplitNode } from './types'
import type { PanelDropEvent } from '../SplitPane'
import { tabsContextKey } from './types'
import TabGroup from './TabGroup.vue'
import { SplitPane, SplitPanePanel } from '../SplitPane'

const props = defineProps<{
  node: TabNode
  enablePanelDrag?: boolean
  enableFileDrop?: boolean
}>()

const emit = defineEmits<{
  panelDrop: [event: PanelDropEvent]
}>()

/** Type-safe access for split nodes (only used in v-else branch) */
const splitNode = computed(() => props.node as TabSplitNode)
const enablePanelDrag = computed(() => props.enablePanelDrag ?? false)
const enableFileDrop = computed(() => props.enableFileDrop ?? false)

const ctx = inject(tabsContextKey)!

function onSplitResizeEnd(_index: number, sizes: number[]) {
  const total = sizes.reduce((sum, n) => sum + n, 0)
  if (total <= 0) return
  // Persist as percentages so layout restores proportionally across container sizes.
  const percentSizes = sizes.map((n) => `${((n / total) * 100).toFixed(4)}%`)
  ctx.setSplitSizes(splitNode.value.id, percentSizes)
}

function onPanelDrop(event: PanelDropEvent) {
  emit('panelDrop', event)
}
</script>
