import { describe, expect, it } from 'vitest'
import {
  applyNavigationHistory,
  createNavigationHistory,
  moveNavigationHistoryIndex,
  navigationTarget,
} from '../navigationHistory'

describe('navigation history', () => {
  it('records paths and exposes back/forward targets', () => {
    let state = createNavigationHistory('')
    state = applyNavigationHistory(state, 'docs', 'push')
    state = applyNavigationHistory(state, 'docs/archive', 'push')

    expect(navigationTarget(state, -1)).toBe('docs')
    expect(navigationTarget(state, 1)).toBeNull()

    state = moveNavigationHistoryIndex(state, -1)
    expect(navigationTarget(state, -1)).toBe('')
    expect(navigationTarget(state, 1)).toBe('docs/archive')
  })

  it('clears forward entries when a new path is pushed after going back', () => {
    let state = createNavigationHistory('')
    state = applyNavigationHistory(state, 'a', 'push')
    state = applyNavigationHistory(state, 'b', 'push')
    state = moveNavigationHistoryIndex(state, -1)
    state = applyNavigationHistory(state, 'c', 'push')

    expect(state.entries).toEqual(['', 'a', 'c'])
    expect(state.index).toBe(2)
    expect(navigationTarget(state, 1)).toBeNull()
  })

  it('replaces the current entry without adding a history item', () => {
    let state = createNavigationHistory('')
    state = applyNavigationHistory(state, 'a', 'push')
    state = applyNavigationHistory(state, 'b', 'push')
    state = moveNavigationHistoryIndex(state, -1)
    state = applyNavigationHistory(state, 'root-after-config-change', 'replace')

    expect(state.entries).toEqual(['', 'root-after-config-change'])
    expect(state.index).toBe(1)
  })

  it('resets the stack when navigation state is invalidated', () => {
    let state = createNavigationHistory('')
    state = applyNavigationHistory(state, 'a', 'push')
    state = applyNavigationHistory(state, 'b', 'push')
    state = applyNavigationHistory(state, 'fresh', 'reset')

    expect(state.entries).toEqual(['fresh'])
    expect(state.index).toBe(0)
    expect(navigationTarget(state, -1)).toBeNull()
  })

  it('does not change history when mode is none', () => {
    const state = applyNavigationHistory(createNavigationHistory('a'), 'b', 'none')

    expect(state.entries).toEqual(['a'])
    expect(state.index).toBe(0)
  })

  it('enforces the configured history limit', () => {
    let state = createNavigationHistory('0')
    state = applyNavigationHistory(state, '1', 'push', 3)
    state = applyNavigationHistory(state, '2', 'push', 3)
    state = applyNavigationHistory(state, '3', 'push', 3)

    expect(state.entries).toEqual(['1', '2', '3'])
    expect(state.index).toBe(2)
  })
})
