<template>
  <TabGroup
    v-if="node.type === 'tabs'"
    :node="node"
    :key="node.id"
    :data-node-id="node.id"
  />
  <SplitPane
    v-else
    :layout="splitNode.layout"
    :gap="4"
    :key="node.id"
    :data-node-id="node.id"
  >
    <SplitPanePanel
      v-for="(child, i) in splitNode.children"
      :key="child.id"
      :size="splitNode.sizes?.[i] ?? `${100 / splitNode.children.length}%`"
    >
      <TabNodeRenderer :node="child" />
    </SplitPanePanel>
  </SplitPane>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { TabNode, TabSplitNode } from './types'
import TabGroup from './TabGroup.vue'
import { SplitPane, SplitPanePanel } from '../SplitPane'

const props = defineProps<{
  node: TabNode
}>()

/** Type-safe access for split nodes (only used in v-else branch) */
const splitNode = computed(() => props.node as TabSplitNode)
</script>
