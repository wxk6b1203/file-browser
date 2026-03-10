import { useElementSize } from '@vueuse/core'
import { computed, ref, type Ref } from 'vue'
import type { SplitLayout } from '../types'

export function useContainer(layout: Ref<SplitLayout>) {
  const containerEl = ref<HTMLElement>()
  const { width, height } = useElementSize(containerEl)

  const containerSize = computed(() => {
    return layout.value === 'horizontal' ? width.value : height.value
  })

  return {
    containerEl,
    containerSize,
  }
}

