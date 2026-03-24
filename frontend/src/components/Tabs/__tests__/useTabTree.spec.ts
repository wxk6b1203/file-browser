import { ref } from 'vue'
import { describe, expect, it } from 'vitest'

import { useTabTree } from '../composables/useTabTree'
import type { TabNode } from '../types'

describe('useTabTree', () => {
  it('keeps parent split sizes when replacing a child group with a nested split', () => {
    const tree = ref<TabNode>({
      type: 'split',
      id: 'root-split',
      layout: 'horizontal',
      sizes: ['65%', '35%'],
      children: [
        {
          type: 'tabs',
          id: 'group-left',
          activeId: 'tab-1',
          tabs: [
            { id: 'tab-1', label: 'Tab 1' },
          ],
        },
        {
          type: 'tabs',
          id: 'group-right',
          activeId: 'tab-2',
          tabs: [
            { id: 'tab-2', label: 'Tab 2' },
            { id: 'tab-3', label: 'Tab 3' },
          ],
        },
      ],
    })

    const { splitGroup } = useTabTree(tree)

    // Split inside the right child group (equivalent to user splitting again after resizing parent divider)
    splitGroup('group-right', 'tab-3', 'right')

    expect(tree.value.type).toBe('split')
    if (tree.value.type !== 'split') return

    expect(tree.value.sizes).toEqual(['65%', '35%'])
    expect(tree.value.children).toHaveLength(2)
    expect(tree.value.children[1]?.type).toBe('split')
  })
})

