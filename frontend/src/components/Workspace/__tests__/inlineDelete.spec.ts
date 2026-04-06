import { describe, expect, it } from 'vitest'
import { DEFAULT_SHORTCUTS } from '@/composables/useShortcut'
import { buildInlineDeletePaths, removeInlineDeletePath } from '../inlineDelete'

describe('inline delete planning', () => {
  it('keeps pending delete entries in current list order', () => {
    expect(buildInlineDeletePaths(
      ['b.txt', 'a.txt'],
      ['a.txt', 'b.txt', 'c.txt'],
    )).toEqual(['a.txt', 'b.txt'])
  })

  it('falls back to target order when the target is not in the current ordered list', () => {
    expect(buildInlineDeletePaths(
      ['detached.txt'],
      ['a.txt', 'b.txt'],
    )).toEqual(['detached.txt'])
  })

  it('removes only the requested pending entry', () => {
    expect(removeInlineDeletePath(['a.txt', 'b.txt', 'c.txt'], 'b.txt')).toEqual(['a.txt', 'c.txt'])
  })

  it('does not register Delete as a global shortcut', () => {
    expect(DEFAULT_SHORTCUTS.some((item) => item.id === 'delete' || item.accelerator === 'Delete')).toBe(false)
  })
})
