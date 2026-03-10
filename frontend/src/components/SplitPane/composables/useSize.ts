import { computed, ref, watch, type Ref } from 'vue'
import type { PanelState } from '../types'

function isPct(val: unknown): val is string {
  return typeof val === 'string' && val.endsWith('%')
}

function isPx(val: unknown): val is string {
  return typeof val === 'string' && val.endsWith('px')
}

function getPct(str: string): number {
  return Number(str.slice(0, -1)) / 100
}

function getPx(str: string): number {
  return Number(str.slice(0, -2))
}

export function parseSizeToRatio(
  val: number | string | undefined,
  containerSize: number,
): number | undefined {
  if (val === undefined || val === null || val === '') return undefined
  if (isPct(val)) return getPct(val)
  if (isPx(val)) return containerSize > 0 ? getPx(val) / containerSize : 0
  const num = Number(val)
  if (!Number.isNaN(num)) return containerSize > 0 ? num / containerSize : 0
  return undefined
}

export function parseSizeToPx(
  val: number | string | undefined,
  containerSize: number,
  defaultVal: number,
): number {
  if (val === undefined || val === null || val === '') return defaultVal
  if (isPct(val)) return getPct(val) * containerSize
  if (isPx(val)) return getPx(val)
  const num = Number(val)
  if (!Number.isNaN(num)) return num
  return defaultVal
}

export function useSize(panels: Ref<PanelState[]>, containerSize: Ref<number>) {
  const percentSizes = ref<number[]>([])

  watch(
    [() => panels.value.map((p) => p.size), () => panels.value.length, containerSize],
    () => {
      const count = panels.value.length
      if (count === 0) {
        percentSizes.value = []
        return
      }

      let ptgList: (number | undefined)[] = []
      let emptyCount = 0

      for (let i = 0; i < count; i++) {
        const itemSize = panels.value[i]?.size
        const ratio = parseSizeToRatio(itemSize, containerSize.value)
        if (ratio !== undefined) {
          ptgList[i] = ratio
        } else {
          emptyCount++
          ptgList[i] = undefined
        }
      }

      const totalPtg = ptgList.reduce<number>((acc, ptg) => acc + (ptg || 0), 0)

      if (totalPtg > 1 || !emptyCount) {
        // Scale all to fit within 1
        const scale = totalPtg > 0 ? 1 / totalPtg : 1 / count
        ptgList = ptgList.map((ptg) => (ptg === undefined ? 0 : ptg * scale))
      } else {
        // Distribute remaining space evenly among panels without size
        const avgRest = (1 - totalPtg) / emptyCount
        ptgList = ptgList.map((ptg) => (ptg === undefined ? avgRest : ptg))
      }

      percentSizes.value = ptgList as number[]
    },
    { immediate: true, deep: true },
  )

  const pxSizes = computed(() => percentSizes.value.map((ptg) => ptg * containerSize.value))

  return {
    percentSizes,
    pxSizes,
  }
}

