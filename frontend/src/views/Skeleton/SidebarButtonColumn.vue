<template>
  <div
    :class="[
      'sidebar-btn-col',
      `sidebar-btn-col--${side}`,
    ]"
    :style="columnStyle"
  >
    <!-- Top-aligned buttons -->
    <div class="sidebar-btn-col__top">
      <button
        v-for="btn in topButtons"
        :key="btn.id"
        :class="[
          'sidebar-btn-col__btn',
          { 'sidebar-btn-col__btn--active': btn.type === 'menu' && modelValue === btn.id },
        ]"
        :title="btn.tooltip"
        @click="onButtonClick(btn)"
      >
        <component :is="btn.icon" class="sidebar-btn-col__icon" />
      </button>
    </div>

    <!-- Bottom-aligned buttons -->
    <div class="sidebar-btn-col__bottom">
      <button
        v-for="btn in bottomButtons"
        :key="btn.id"
        :class="[
          'sidebar-btn-col__btn',
          { 'sidebar-btn-col__btn--active': btn.type === 'menu' && modelValue === btn.id },
        ]"
        :title="btn.tooltip"
        @click="onButtonClick(btn)"
      >
        <component :is="btn.icon" class="sidebar-btn-col__icon" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { SidebarButton, SidebarSide } from './types'

const props = withDefaults(
  defineProps<{
    /** Which side this column belongs to */
    side: SidebarSide
    /** Top-aligned buttons */
    topButtons?: SidebarButton[]
    /** Bottom-aligned buttons */
    bottomButtons?: SidebarButton[]
    /** Width of the button column in px */
    width?: number
    /** Currently active menu button id (null = no panel open) */
    modelValue?: string | null
  }>(),
  {
    topButtons: () => [],
    bottomButtons: () => [],
    width: 40,
    modelValue: null,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string | null]
}>()

const columnStyle = {
  width: `${props.width}px`,
  minWidth: `${props.width}px`,
}

function onButtonClick(btn: SidebarButton) {
  if (btn.type === 'menu') {
    // Toggle: click the active one → collapse (null), click a different one → switch
    const next = props.modelValue === btn.id ? null : btn.id
    emit('update:modelValue', next)
  }
  // Always call the handler if present
  btn.onClick?.()
}
</script>

<style scoped>
.sidebar-btn-col {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  flex-shrink: 0;
  background-color: var(--theme-color-bg-surface);
  border-color: var(--theme-color-border);
  box-sizing: border-box;
  overflow: hidden;
  user-select: none;
  z-index: 2;
}

.sidebar-btn-col--left {
  border-right: 1px solid var(--theme-color-border);
}

.sidebar-btn-col--right {
  border-left: 1px solid var(--theme-color-border);
}

.sidebar-btn-col__top,
.sidebar-btn-col__bottom {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 4px 0;
}

.sidebar-btn-col__btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: var(--theme-radius-sm);
  background: transparent;
  color: var(--theme-color-text-secondary);
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease;
  outline: none;
  padding: 0;
}

.sidebar-btn-col__btn:hover {
  background-color: var(--theme-color-bg-hover);
  color: var(--theme-color-text);
}

.sidebar-btn-col__btn--active {
  background-color: var(--theme-color-primary-light);
  color: var(--theme-color-primary);
}

.sidebar-btn-col__btn--active:hover {
  background-color: var(--theme-color-primary-light);
  color: var(--theme-color-primary);
}

.sidebar-btn-col__icon {
  width: 20px;
  height: 20px;
}
</style>

