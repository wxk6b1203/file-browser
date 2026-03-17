<template>
  <div
    ref="panelEl"
    :class="['split-pane-panel', props.customClass, { 'split-pane-panel--minimized': panelState.minimized }]"
    :style="panelStyle"
  >
    <slot />
  </div>
  <SplitPaneDivider
    v-if="showDivider"
    :index="panelIndex"
    :layout="context.layout.value"
    :resizable="isResizable"
    :gap="context.gap.value"
    :inset-start="dividerInset[0]"
    :inset-end="dividerInset[1]"
    :indicator-width="context.indicatorSize.value[0]"
    :indicator-height="context.indicatorSize.value[1]"
    @move-start="context.onMoveStart"
    @moving="context.onMoving"
    @move-end="context.onMoveEnd"
    @dblclick="context.onDblClick"
  />
</template>

<script setup lang="ts">
import { computed, inject, onBeforeUnmount, onMounted, reactive, ref, watch, type CSSProperties } from 'vue'
import { splitPaneContextKey, type PanelState, type SplitPanePanelProps } from './types'
import SplitPaneDivider from './SplitPaneDivider.vue'

const props = withDefaults(defineProps<SplitPanePanelProps>(), {
  resizable: true,
  minimized: false,
})

const emit = defineEmits<{
  'update:size': [value: number]
  'update:minimized': [value: boolean]
}>()

const context = inject(splitPaneContextKey)!
if (!context) {
  throw new Error('[SplitPanePanel] must be used inside <SplitPane>')
}

const panelEl = ref<HTMLElement>()
const panelIndex = ref(0)
const uid = Date.now() + Math.random()

// Create the reactive panel state for registration
const panelState: PanelState = reactive({
  uid,
  index: 0,
  size: props.size,
  minSize: props.minSize,
  maxSize: props.maxSize,
  resizable: props.resizable ?? true,
  borderRadius: props.borderRadius,
  minimized: props.minimized ?? false,
})

// Sync props to panel state
watch(
  () => props.size,
  (val) => {
    panelState.size = val
  },
)
watch(
  () => props.minSize,
  (val) => {
    panelState.minSize = val
  },
)
watch(
  () => props.maxSize,
  (val) => {
    panelState.maxSize = val
  },
)
watch(
  () => props.resizable,
  (val) => {
    panelState.resizable = val ?? true
  },
)
watch(
  () => props.borderRadius,
  (val) => {
    panelState.borderRadius = val
  },
)

// Sync minimized prop → call context minimize/restore
watch(
  () => props.minimized,
  (val) => {
    if (val && !panelState.minimized) {
      context.minimizePanel(panelState.uid)
    } else if (!val && panelState.minimized) {
      context.restorePanel(panelState.uid)
    }
  },
)

// Sync internal minimized state → emit update:minimized (for v-model support)
watch(
  () => panelState.minimized,
  (val) => {
    if (val !== (props.minimized ?? false)) {
      emit('update:minimized', val)
    }
  },
)

// Keep local index in sync
watch(
  () => panelState.index,
  (val) => {
    panelIndex.value = val
  },
  { immediate: true },
)

// Watch size changes from drag and emit update:size
watch(
  () => panelState.size,
  (val) => {
    if (val !== undefined && val !== props.size) {
      emit('update:size', typeof val === 'number' ? val : parseFloat(String(val)))
    }
  },
)

// Register / unregister
onMounted(() => {
  panelState.el = panelEl.value
  context.registerPanel(panelState)
})
onBeforeUnmount(() => context.unregisterPanel(panelState))

// Computed: this panel's pixel size
const pxSize = computed(() => {
  return context.pxSizes.value[panelIndex.value] ?? 0
})

/** Find the next non-minimized panel after this one */
function findNextVisible(): PanelState | null {
  for (let i = panelIndex.value + 1; i < context.panels.value.length; i++) {
    const p = context.panels.value[i]
    if (p && !p.minimized) return p
  }
  return null
}

// Whether to show divider after this panel.
// A non-minimized panel shows a divider when there is at least one
// non-minimized panel somewhere after it (skipping minimized ones).
const showDivider = computed(() => {
  if (panelIndex.value >= context.panels.value.length - 1) return false
  if (panelState.minimized) return false
  return findNextVisible() !== null
})

// Whether resize is allowed (both this and the effective next visible panel must be resizable)
const isResizable = computed(() => {
  if (panelState.minimized) return false
  const nextVisible = findNextVisible()
  if (!nextVisible) return false
  return (props.resizable ?? true) && nextVisible.resizable
})

/**
 * Parse CSS border-radius shorthand into [topLeft, topRight, bottomRight, bottomLeft] in px.
 * Supports: '8px', '8px 4px', '8px 4px 2px', '8px 4px 2px 0'.
 * Percentage values are ignored (treated as 0).
 */
function parseBorderRadius(val?: string): [number, number, number, number] {
  if (!val) return [0, 0, 0, 0]
  const parts = val.trim().split(/\s+/)
  const parse = (s: string | undefined) => {
    if (!s) return 0
    if (s.endsWith('px')) return parseFloat(s) || 0
    const n = parseFloat(s)
    return Number.isNaN(n) ? 0 : n // bare number treated as px
  }
  switch (parts.length) {
    case 1: {
      const v = parse(parts[0])
      return [v, v, v, v]
    }
    case 2: {
      const tl_br = parse(parts[0])
      const tr_bl = parse(parts[1])
      return [tl_br, tr_bl, tl_br, tr_bl]
    }
    case 3: {
      const tl = parse(parts[0])
      const tr_bl = parse(parts[1])
      const br = parse(parts[2])
      return [tl, tr_bl, br, tr_bl]
    }
    default: {
      return [parse(parts[0]), parse(parts[1]), parse(parts[2]), parse(parts[3])]
    }
  }
}

// Compute inset for the divider to avoid overlapping rounded corners.
// For horizontal: insetStart = top inset, insetEnd = bottom inset
// For vertical:   insetStart = left inset, insetEnd = right inset
const dividerInset = computed<[number, number]>(() => {
  const nextPanel = findNextVisible()
  if (!nextPanel) return [0, 0]

  const isH = context.layout.value === 'horizontal'
  // corners: [topLeft, topRight, bottomRight, bottomLeft]
  const leftCorners = parseBorderRadius(panelState.borderRadius)
  const rightCorners = parseBorderRadius(nextPanel.borderRadius)

  if (isH) {
    // Divider sits between left panel's right edge and right panel's left edge
    // top inset = max(leftPanel topRight, rightPanel topLeft)
    // bottom inset = max(leftPanel bottomRight, rightPanel bottomLeft)
    const top = Math.max(leftCorners[1], rightCorners[0])
    const bottom = Math.max(leftCorners[2], rightCorners[3])
    return [top, bottom]
  } else {
    // Divider sits between top panel's bottom edge and bottom panel's top edge
    // left inset = max(topPanel bottomLeft, bottomPanel topLeft)
    // right inset = max(topPanel bottomRight, bottomPanel topRight)
    const left = Math.max(leftCorners[3], rightCorners[0])
    const right = Math.max(leftCorners[2], rightCorners[1])
    return [left, right]
  }
})

// Merged style for the panel
const panelStyle = computed<CSSProperties>(() => {
  const base: CSSProperties = {
    flexBasis: panelState.minimized ? '0px' : `${pxSize.value}px`,
    flexGrow: 0,
    flexShrink: 0,
    overflow: 'hidden',
  }
  if (panelState.minimized) {
    // Ensure the panel can fully collapse even if content has min-width/height
    base.minWidth = '0px'
    base.minHeight = '0px'
  }
  if (props.borderRadius) {
    base.borderRadius = props.borderRadius
  }
  if (props.backgroundColor) {
    base.backgroundColor = props.backgroundColor
  }
  if (props.customStyle) {
    Object.assign(base, props.customStyle)
  }
  return base
})

defineExpose({ panelEl })
</script>

<style scoped>
.split-pane-panel {
  position: relative;
  box-sizing: border-box;
}

.split-pane-panel--minimized {
  pointer-events: none;
}
</style>

