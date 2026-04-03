import { defineComponent, h, nextTick, ref } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, it } from 'vitest'
import { createI18n } from 'vue-i18n'

import Tabs from './src/components/Tabs/Tabs.vue'
import SplitPane from './src/components/SplitPane/SplitPane.vue'
import TabBar from './src/components/Tabs/TabBar.vue'
import type { TabNode } from './src/components/Tabs/types'

const i18n = createI18n({
  legacy: false,
  locale: 'zh',
  messages: { zh: { splitPane: { dropToPanel: '释放到此面板' } } },
})

function makeSplitTree(): TabNode {
  return {
    type: 'split',
    id: 'root-split',
    layout: 'horizontal',
    sizes: ['50%', '50%'],
    children: [
      { type: 'tabs', id: 'left-group', activeId: 'left-1', tabs: [{ id: 'left-1', label: 'Left 1' }] },
      { type: 'tabs', id: 'right-group', activeId: 'right-1', tabs: [{ id: 'right-1', label: 'Right 1' }] },
    ],
  }
}

describe('debug', () => {
  it('debug close left split', async () => {
    const Host = defineComponent({
      setup() {
        const model = ref<TabNode>(makeSplitTree())
        return () => h(Tabs, {
          modelValue: model.value,
          'onUpdate:modelValue': (next: TabNode) => {
            model.value = next
            console.log('UPDATE', JSON.stringify(next))
          },
        })
      },
    })

    const wrapper = mount(Host, { attachTo: document.body, global: { plugins: [i18n] } })
    await nextTick(); await nextTick(); await nextTick()
    console.log('BEFORE split exists', wrapper.findComponent(SplitPane).exists())
    console.log('BEFORE html', wrapper.html())
    const bars = wrapper.findAllComponents(TabBar)
    bars[0]!.vm.$emit('close', { id: 'left-1', label: 'Left 1' })
    await nextTick(); await nextTick(); await nextTick(); await new Promise(r => setTimeout(r, 0)); await nextTick()
    console.log('AFTER split exists', wrapper.findComponent(SplitPane).exists())
    console.log('AFTER html', wrapper.html())
    const tabsComp = wrapper.findComponent(Tabs)
    console.log('TREE', JSON.stringify((tabsComp.vm as any).treeRef))
    wrapper.unmount()
  })
})
