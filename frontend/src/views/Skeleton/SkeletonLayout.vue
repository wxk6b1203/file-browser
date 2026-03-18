<template>
  <div class="skeleton-layout">
    <!-- ═══ Left button column (always visible) ═══ -->
    <SidebarButtonColumn
      v-if="hasLeftSidebar"
      side="left"
      :top-buttons="leftConfig.topButtons"
      :bottom-buttons="leftConfig.bottomButtons"
      :width="buttonColumnWidth"
      :model-value="leftActiveId"
      @update:model-value="setLeftActive"
    />

    <!-- ═══ Main resizable area ═══ -->
    <SplitPane
      ref="splitRef"
      layout="horizontal"
      :gap="gap"
      class="skeleton-layout__split"
    >
      <!-- Left menu panel -->
      <SplitPanePanel
        v-if="hasLeftSidebar"
        v-model:minimized="leftMinimized"
        :size="leftSize"
        :min-size="leftMinSize"
        :max-size="leftMaxSize"
      >
        <div class="skeleton-layout__panel-content">
          <slot name="left-panel" :active-id="leftActiveId" />
        </div>
      </SplitPanePanel>

      <!-- Center content (always visible, takes remaining space) -->
      <SplitPanePanel>
        <div class="skeleton-layout__center">
          <slot name="center" />
        </div>
      </SplitPanePanel>

      <!-- Right menu panel -->
      <SplitPanePanel
        v-if="hasRightSidebar"
        v-model:minimized="rightMinimized"
        :size="rightSize"
        :min-size="rightMinSize"
        :max-size="rightMaxSize"
      >
        <div class="skeleton-layout__panel-content">
          <slot name="right-panel" :active-id="rightActiveId" />
        </div>
      </SplitPanePanel>
    </SplitPane>

    <!-- ═══ Right button column (always visible) ═══ -->
    <SidebarButtonColumn
      v-if="hasRightSidebar"
      side="right"
      :top-buttons="rightConfig.topButtons"
      :bottom-buttons="rightConfig.bottomButtons"
      :width="buttonColumnWidth"
      :model-value="rightActiveId"
      @update:model-value="setRightActive"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, provide, ref, watch } from 'vue'
import { SplitPane, SplitPanePanel } from '@/components/SplitPane'
import SidebarButtonColumn from './SidebarButtonColumn.vue'
import { skeletonContextKey, type SidebarConfig } from './types'

const props = withDefaults(
  defineProps<{
    /** Left sidebar configuration */
    leftSidebar?: SidebarConfig
    /** Right sidebar configuration */
    rightSidebar?: SidebarConfig
    /** Width of button columns in px */
    buttonColumnWidth?: number
    /** Gap between SplitPane panels */
    gap?: number
    /** Currently active left menu button id. Supports v-model:leftActive */
    leftActive?: string | null
    /** Currently active right menu button id. Supports v-model:rightActive */
    rightActive?: string | null
  }>(),
  {
    buttonColumnWidth: 40,
    gap: 1,
    leftActive: undefined,
    rightActive: undefined,
  },
)

const emit = defineEmits<{
  'update:leftActive': [value: string | null]
  'update:rightActive': [value: string | null]
}>()

const splitRef = ref<InstanceType<typeof SplitPane>>()

// ─── Derived config ─────────────────────────────────────────

const leftConfig = computed<SidebarConfig>(() =>
  props.leftSidebar ?? { topButtons: [], bottomButtons: [] },
)
const rightConfig = computed<SidebarConfig>(() =>
  props.rightSidebar ?? { topButtons: [], bottomButtons: [] },
)

const hasLeftSidebar = computed(() => {
  const cfg = leftConfig.value
  return cfg.topButtons.length > 0 || cfg.bottomButtons.length > 0
})
const hasRightSidebar = computed(() => {
  const cfg = rightConfig.value
  return cfg.topButtons.length > 0 || cfg.bottomButtons.length > 0
})

// ─── Panel sizes ────────────────────────────────────────────

const leftSize = computed(() => leftConfig.value.defaultSize ?? '240px')
const leftMinSize = computed(() => leftConfig.value.minSize ?? '120px')
const leftMaxSize = computed(() => leftConfig.value.maxSize ?? undefined)

const rightSize = computed(() => rightConfig.value.defaultSize ?? '240px')
const rightMinSize = computed(() => rightConfig.value.minSize ?? '120px')
const rightMaxSize = computed(() => rightConfig.value.maxSize ?? undefined)

// ─── Active ID state ────────────────────────────────────────

// Internal active IDs. If v-model is bound, sync with props; otherwise internal.
const _leftActiveId = ref<string | null>(props.leftActive ?? null)
const _rightActiveId = ref<string | null>(props.rightActive ?? null)

// Sync from props → internal
watch(
  () => props.leftActive,
  (val) => {
    if (val !== undefined) _leftActiveId.value = val
  },
)
watch(
  () => props.rightActive,
  (val) => {
    if (val !== undefined) _rightActiveId.value = val
  },
)

const leftActiveId = computed(() => _leftActiveId.value)
const rightActiveId = computed(() => _rightActiveId.value)

function setLeftActive(id: string | null) {
  _leftActiveId.value = id
  emit('update:leftActive', id)
}

function setRightActive(id: string | null) {
  _rightActiveId.value = id
  emit('update:rightActive', id)
}

// ─── Minimize ↔ Active ID sync ──────────────────────────────

const leftMinimized = ref(_leftActiveId.value === null)
const rightMinimized = ref(_rightActiveId.value === null)

// When active ID changes → minimize/restore panel
watch(_leftActiveId, (val) => {
  leftMinimized.value = val === null
})
watch(_rightActiveId, (val) => {
  rightMinimized.value = val === null
})

// When panel minimized state changes externally (e.g. via SplitPane double-click)
// → sync back to active ID
watch(leftMinimized, (min) => {
  if (min && _leftActiveId.value !== null) {
    setLeftActive(null)
  } else if (!min && _leftActiveId.value === null) {
    // Restored via SplitPane but no active id — pick the first menu button
    const firstMenu = leftConfig.value.topButtons.find((b) => b.type === 'menu')
      ?? leftConfig.value.bottomButtons.find((b) => b.type === 'menu')
    if (firstMenu) setLeftActive(firstMenu.id)
  }
})
watch(rightMinimized, (min) => {
  if (min && _rightActiveId.value !== null) {
    setRightActive(null)
  } else if (!min && _rightActiveId.value === null) {
    const firstMenu = rightConfig.value.topButtons.find((b) => b.type === 'menu')
      ?? rightConfig.value.bottomButtons.find((b) => b.type === 'menu')
    if (firstMenu) setRightActive(firstMenu.id)
  }
})

// ─── Provide context ────────────────────────────────────────

provide(skeletonContextKey, {
  leftActiveId: computed(() => _leftActiveId.value),
  rightActiveId: computed(() => _rightActiveId.value),
  setLeftActive,
  setRightActive,
})

// ─── Expose ─────────────────────────────────────────────────

defineExpose({
  splitRef,
  setLeftActive,
  setRightActive,
  leftActiveId,
  rightActiveId,
})
</script>

<style scoped>
.skeleton-layout {
  display: flex;
  width: 100%;
  height: 100%;
  overflow: hidden;
  background-color: var(--theme-color-bg-base);
}

.skeleton-layout__split {
  flex: 1;
  min-width: 0;
}

.skeleton-layout__panel-content {
  width: 100%;
  height: 100%;
  overflow: auto;
  background-color: var(--theme-color-bg-surface);
}

.skeleton-layout__center {
  width: 100%;
  height: 100%;
  overflow: auto;
}
</style>

