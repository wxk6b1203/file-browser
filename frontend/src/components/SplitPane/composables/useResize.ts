import { ref, computed, type Ref } from 'vue'
import type { PanelState } from '../types'
import { parseSizeToPx } from './useSize'

export function useResize(
  panels: Ref<PanelState[]>,
  containerSize: Ref<number>,
  pxSizes: Ref<number[]>,
  lazy: Ref<boolean>,
) {
  const lazyOffset = ref(0)
  const movingIndex = ref<{ index: number; confirmed: boolean } | null>(null)
  let cachePxSizes: number[] = []
  let pendingUpdate: (() => void) | null = null

  const limitSizes = computed(() =>
    panels.value.map((item) => [item.minSize, item.maxSize] as const),
  )

  function getLimitSize(val: number | string | undefined, defaultLimit: number): number {
    return parseSizeToPx(val, containerSize.value, defaultLimit)
  }

  const onMoveStart = (index: number) => {
    lazyOffset.value = 0
    movingIndex.value = { index, confirmed: false }
    cachePxSizes = [...pxSizes.value]
  }

  const onMoving = (index: number, offset: number) => {
    let confirmedIndex: number | null = null

    // Determine the actual panel being resized based on drag direction
    if ((!movingIndex.value || !movingIndex.value.confirmed) && offset !== 0) {
      if (offset > 0) {
        // Find the first non-minimized, non-zero-sized panel at or before this index
        for (let i = index; i >= 0; i--) {
          if (!panels.value[i]?.minimized && (cachePxSizes[i] ?? 0) >= 0) {
            confirmedIndex = i
            movingIndex.value = { index: i, confirmed: true }
            break
          }
        }
      } else {
        // Find the first non-minimized, non-zero-sized panel at or before this index
        for (let i = index; i >= 0; i--) {
          if (!panels.value[i]?.minimized && (cachePxSizes[i] ?? 0) > 0) {
            confirmedIndex = i
            movingIndex.value = { index: i, confirmed: true }
            break
          }
        }
      }
    }

    const mergedIndex = confirmedIndex ?? movingIndex.value?.index ?? index
    const numSizes = [...cachePxSizes]

    // Find the next non-minimized panel after mergedIndex
    let nextIndex = -1
    for (let i = mergedIndex + 1; i < panels.value.length; i++) {
      if (!panels.value[i]?.minimized) {
        nextIndex = i
        break
      }
    }

    if (nextIndex < 0 || nextIndex >= numSizes.length) return

    const currentSize = numSizes[mergedIndex] ?? 0
    const nextSize = numSizes[nextIndex] ?? 0

    // Get limits
    const startMinSize = getLimitSize(limitSizes.value[mergedIndex]?.[0], 0)
    const endMinSize = getLimitSize(limitSizes.value[nextIndex]?.[0], 0)
    const startMaxSize = getLimitSize(limitSizes.value[mergedIndex]?.[1], containerSize.value || 0)
    const endMaxSize = getLimitSize(limitSizes.value[nextIndex]?.[1], containerSize.value || 0)

    // Clamp offset
    let mergedOffset = offset
    if (currentSize + mergedOffset < startMinSize)
      mergedOffset = startMinSize - currentSize
    if (nextSize - mergedOffset < endMinSize)
      mergedOffset = nextSize - endMinSize
    if (currentSize + mergedOffset > startMaxSize)
      mergedOffset = startMaxSize - currentSize
    if (nextSize - mergedOffset > endMaxSize)
      mergedOffset = nextSize - endMaxSize

    numSizes[mergedIndex] = currentSize + mergedOffset
    numSizes[nextIndex] = nextSize - mergedOffset

    lazyOffset.value = mergedOffset

    pendingUpdate = () => {
      panels.value.forEach((panel, i) => {
        panel.size = numSizes[i]
      })
      pendingUpdate = null
    }

    if (!lazy.value) {
      pendingUpdate()
    }
  }

  const onMoveEnd = () => {
    if (lazy.value && pendingUpdate) {
      pendingUpdate()
    }
    lazyOffset.value = 0
    movingIndex.value = null
    cachePxSizes = []
  }

  return {
    lazyOffset,
    movingIndex,
    onMoveStart,
    onMoving,
    onMoveEnd,
  }
}



