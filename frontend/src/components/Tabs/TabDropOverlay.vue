<template>
  <div
    v-if="visible"
    class="tab-drop-overlay"
  >
    <!-- Zone highlight -->
    <div
      v-if="zone === 'top'"
      class="tab-drop-overlay__zone tab-drop-overlay__zone--top"
      :style="{ opacity: overlayOpacity }"
    />
    <div
      v-else-if="zone === 'bottom'"
      class="tab-drop-overlay__zone tab-drop-overlay__zone--bottom"
      :style="{ opacity: overlayOpacity }"
    />
    <div
      v-else-if="zone === 'left'"
      class="tab-drop-overlay__zone tab-drop-overlay__zone--left"
      :style="{ opacity: overlayOpacity }"
    />
    <div
      v-else-if="zone === 'right'"
      class="tab-drop-overlay__zone tab-drop-overlay__zone--right"
      :style="{ opacity: overlayOpacity }"
    />
    <div
      v-else-if="zone === 'center'"
      class="tab-drop-overlay__zone tab-drop-overlay__zone--center"
      :style="{ opacity: overlayOpacity }"
    />
  </div>
</template>

<script setup lang="ts">
import { inject, computed } from 'vue'
import type { DropZone } from './types'
import { tabsContextKey } from './types'

defineProps<{
  visible: boolean
  zone: DropZone
}>()

const ctx = inject(tabsContextKey)!

const overlayOpacity = computed(() => ctx.overlayOpacity.value)
</script>

<style scoped>
.tab-drop-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 100;
  pointer-events: none;
}

.tab-drop-overlay__zone {
  position: absolute;
  background-color: var(--theme-color-primary);
  border-radius: 4px;
  transition: all 0.15s ease;
}

.tab-drop-overlay__zone--top {
  top: 0;
  left: 0;
  right: 0;
  height: 50%;
}

.tab-drop-overlay__zone--bottom {
  bottom: 0;
  left: 0;
  right: 0;
  height: 50%;
}

.tab-drop-overlay__zone--left {
  top: 0;
  left: 0;
  bottom: 0;
  width: 50%;
}

.tab-drop-overlay__zone--right {
  top: 0;
  right: 0;
  bottom: 0;
  width: 50%;
}

.tab-drop-overlay__zone--center {
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}
</style>

