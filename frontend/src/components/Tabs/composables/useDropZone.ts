import type { DropZone } from '../types'

/**
 * Given a pointer position and a container rect, determine which drop zone
 * the pointer is in.
 *
 * Zones:
 * - top:    upper 1/3
 * - bottom: lower 1/3
 * - left:   left 1/4 (within the middle horizontal band)
 * - right:  right 1/4 (within the middle horizontal band)
 * - center: everything else
 */
export function calcDropZone(
  pointerX: number,
  pointerY: number,
  rect: DOMRect,
): DropZone {
  const relX = pointerX - rect.left
  const relY = pointerY - rect.top
  const w = rect.width
  const h = rect.height

  if (w === 0 || h === 0) return null

  const fracX = relX / w
  const fracY = relY / h

  if (fracY < 1 / 3) return 'top'
  if (fracY > 2 / 3) return 'bottom'
  if (fracX < 1 / 4) return 'left'
  if (fracX > 3 / 4) return 'right'
  return 'center'
}

