import { defineComponent, h, nextTick } from 'vue'
import { createI18n } from 'vue-i18n'
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

import SplitPane from '../SplitPane.vue'
import SplitPanePanel from '../SplitPanePanel.vue'

function readFlexBasis(el: Element): number {
  return parseFloat((el as HTMLElement).style.flexBasis || '0')
}

function mountWithI18n(component: ReturnType<typeof defineComponent>) {
  const i18n = createI18n({
    legacy: false,
    locale: 'en',
    messages: {
      en: {
        splitPane: {
          dropToPanel: 'Drop to This Panel',
        },
      },
    },
  })

  return mount(component, {
    global: {
      plugins: [i18n],
    },
  })
}

describe('SplitPane', () => {
  it('applies initial minimized state and redistributes space', async () => {
    const Host = defineComponent({
      setup() {
        return () =>
          h(
            SplitPane,
            { layout: 'horizontal' },
            {
              default: () => [
                h(SplitPanePanel, { size: '30%', minimized: true }),
                h(SplitPanePanel, { size: '70%' }),
              ],
            },
          )
      },
    })

    const wrapper = mountWithI18n(Host)
    await nextTick()
    await nextTick()

    const panels = wrapper.findAll('.split-pane-panel')
    expect(panels).toHaveLength(2)

    const first = readFlexBasis(panels[0]!.element)
    const second = readFlexBasis(panels[1]!.element)

    expect(first).toBeLessThanOrEqual(0.5)
    expect(second).toBeGreaterThan(990)
  })

  it('keeps all panel sizes within max constraints after redistribution', async () => {
    const Host = defineComponent({
      setup() {
        return () =>
          h(
            SplitPane,
            { layout: 'horizontal' },
            {
              default: () => [
                h(SplitPanePanel, { size: '80%', maxSize: '60%' }),
                h(SplitPanePanel, { size: '20%', maxSize: '25%' }),
                h(SplitPanePanel, { size: '0%', maxSize: '50%' }),
              ],
            },
          )
      },
    })

    const wrapper = mountWithI18n(Host)
    await nextTick()
    await nextTick()

    const panels = wrapper.findAll('.split-pane-panel')
    expect(panels).toHaveLength(3)

    const first = readFlexBasis(panels[0]!.element)
    const second = readFlexBasis(panels[1]!.element)
    const third = readFlexBasis(panels[2]!.element)

    expect(first).toBeLessThanOrEqual(600.5)
    expect(second).toBeLessThanOrEqual(250.5)
    expect(third).toBeLessThanOrEqual(500.5)
    expect(first + second + third).toBeCloseTo(1000, 0)
  })
})
