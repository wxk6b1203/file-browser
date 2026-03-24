import { nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

vi.mock('@vueuse/core', async () => {
  const vue = await import('vue')
  return {
    useElementSize: () => ({
      width: vue.ref(1000),
      height: vue.ref(600),
    }),
  }
})

import Tabs from '../Tabs.vue'
import SplitPane from '../../SplitPane/SplitPane.vue'
import TabHeader from '../TabHeader.vue'
import TabBar from '../TabBar.vue'
import type { TabNode } from '../types'

function makeSplitTree(): TabNode {
  return {
    type: 'split',
    id: 'root-split',
    layout: 'horizontal',
    sizes: ['50%', '50%'],
    children: [
      {
        type: 'tabs',
        id: 'left-group',
        activeId: 'left-1',
        tabs: [
          { id: 'left-1', label: 'Left 1' },
        ],
      },
      {
        type: 'tabs',
        id: 'right-group',
        activeId: 'right-1',
        tabs: [
          { id: 'right-1', label: 'Right 1' },
        ],
      },
    ],
  }
}

describe('Tabs', () => {
  it('persists split pane sizes into tree model on resize-end', async () => {
    const wrapper = mount(Tabs, {
      props: {
        modelValue: makeSplitTree(),
      },
    })

    const split = wrapper.findComponent(SplitPane)
    expect(split.exists()).toBe(true)

    split.vm.$emit('resizeEnd', 0, [300, 700])
    await nextTick()
    await nextTick()

    const updates = wrapper.emitted('update:modelValue')
    expect(updates).toBeTruthy()

    const latest = updates![updates!.length - 1]![0] as TabNode
    expect(latest.type).toBe('split')
    if (latest.type === 'split') {
      expect(latest.sizes).toEqual(['30.0000%', '70.0000%'])
    }
  })

  it('emits tabActivate and updates activeId when a tab is selected', async () => {
    const model: TabNode = {
      type: 'tabs',
      id: 'root-group',
      activeId: 'tab-1',
      tabs: [
        { id: 'tab-1', label: 'Tab 1' },
        { id: 'tab-2', label: 'Tab 2' },
      ],
    }

    const wrapper = mount(Tabs, {
      props: {
        modelValue: model,
      },
    })

    const headers = wrapper.findAllComponents(TabHeader)
    expect(headers).toHaveLength(2)

    headers[1]!.vm.$emit('select', { id: 'tab-2', label: 'Tab 2' })
    await nextTick()
    await nextTick()

    const activateEvents = wrapper.emitted('tabActivate')
    expect(activateEvents).toBeTruthy()
    expect(activateEvents![0]![0]).toMatchObject({ id: 'tab-2', label: 'Tab 2' })
    expect(activateEvents![0]![1]).toBe('root-group')

    const updates = wrapper.emitted('update:modelValue')
    expect(updates).toBeTruthy()
    const latest = updates![updates!.length - 1]![0] as TabNode
    expect(latest.type).toBe('tabs')
    if (latest.type === 'tabs') {
      expect(latest.activeId).toBe('tab-2')
    }
  })

  it('cleans up drag state on pointercancel and emits tabDragEnd', async () => {
    const model: TabNode = {
      type: 'tabs',
      id: 'root-group',
      activeId: 'tab-1',
      tabs: [
        { id: 'tab-1', label: 'Tab 1' },
        { id: 'tab-2', label: 'Tab 2' },
      ],
    }

    const wrapper = mount(Tabs, {
      attachTo: document.body,
      props: {
        modelValue: model,
      },
    })

    const bar = wrapper.findComponent(TabBar)
    expect(bar.exists()).toBe(true)

    const fakeHeaderEl = document.createElement('div')
    fakeHeaderEl.getBoundingClientRect = () =>
      ({
        left: 80,
        top: 10,
        width: 120,
        height: 32,
        right: 200,
        bottom: 42,
        x: 80,
        y: 10,
      }) as DOMRect

    bar.vm.$emit(
      'dragStart',
      { id: 'tab-1', label: 'Tab 1' },
      { clientX: 100, clientY: 24 } as PointerEvent,
      fakeHeaderEl,
    )
    await nextTick()

    expect(document.body.querySelector('.tab-drag-ghost')).not.toBeNull()

    document.dispatchEvent(new Event('pointercancel'))
    await nextTick()

    expect(document.body.querySelector('.tab-drag-ghost')).toBeNull()

    const endEvents = wrapper.emitted('tabDragEnd')
    expect(endEvents).toBeTruthy()
    expect(endEvents![0]![0]).toMatchObject({ id: 'tab-1', label: 'Tab 1' })
    expect(endEvents![0]![1]).toBe('root-group')

    wrapper.unmount()
  })
})
