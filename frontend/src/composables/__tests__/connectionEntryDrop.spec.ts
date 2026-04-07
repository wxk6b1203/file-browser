import { afterEach, describe, expect, it } from 'vitest'
import { __test__, compactConnectionDragEntries, resolveConnectionEntryDragPayload, type ConnectionEntryPanelDragPayload } from '../connectionEntryDrop'
import { clearActiveInternalDrag, setActiveInternalDrag } from '../splitPaneDragState'

describe('connectionEntryDrop', () => {
  afterEach(() => {
    clearActiveInternalDrag()
  })

  it('compacts nested selections under an already-selected directory', () => {
    const entries = compactConnectionDragEntries([
      { path: 'docs', name: 'docs', isDirectory: true },
      { path: 'docs/readme.md', name: 'readme.md', isDirectory: false },
      { path: 'docs/guides', name: 'guides', isDirectory: true },
      { path: 'src/main.ts', name: 'main.ts', isDirectory: false },
    ])

    expect(entries).toEqual([
      { path: 'docs', name: 'docs', isDirectory: true },
      { path: 'src/main.ts', name: 'main.ts', isDirectory: false },
    ])
  })

  it('normalizes drag entry paths before planning selection and drop', () => {
    const entries = compactConnectionDragEntries([
      { path: '/docs/', name: 'docs', isDirectory: true },
      { path: '\\docs\\guide.md', name: 'guide.md', isDirectory: false },
      { path: 'assets\\\\logo.png', name: 'logo.png', isDirectory: false },
    ])

    expect(entries).toEqual([
      { path: 'docs', name: 'docs', isDirectory: true },
      { path: 'assets/logo.png', name: 'logo.png', isDirectory: false },
    ])
  })

  it('skips only invalid descendant targets for same-connection multi-move', () => {
    const payload: ConnectionEntryPanelDragPayload = {
      type: 'connection-entry',
      data: {
        sourceConnectionId: 'conn-1',
        sourceViewDir: '',
        entries: [
          { path: 'docs', name: 'docs', isDirectory: true },
          { path: 'assets/logo.png', name: 'logo.png', isDirectory: false },
        ],
      },
    }

    const plan = __test__.buildDropPlan(payload, 'conn-1', 'docs/archive')
    expect(plan).not.toBeNull()
    expect(plan?.sameConnectionMoves).toEqual([
      {
        sourcePath: 'assets/logo.png',
        sourceParentDir: 'assets',
        isDirectory: false,
      },
    ])
    expect(plan?.skippedCount).toBe(1)
  })

  it('compacts nested entries for cross-connection transfer planning', () => {
    const payload: ConnectionEntryPanelDragPayload = {
      type: 'connection-entry',
      data: {
        sourceConnectionId: 'conn-1',
        sourceViewDir: '',
        entries: [
          { path: 'photos', name: 'photos', isDirectory: true },
          { path: 'photos/2025/a.jpg', name: 'a.jpg', isDirectory: false },
          { path: 'notes.txt', name: 'notes.txt', isDirectory: false },
        ],
      },
    }

    const plan = __test__.buildDropPlan(payload, 'conn-2', 'backup')
    expect(plan).not.toBeNull()
    expect(plan?.crossConnectionTransfers).toEqual([
      { sourcePath: 'photos', isDirectory: true },
      { sourcePath: 'notes.txt', isDirectory: false },
    ])
  })

  it('resolves active internal drag payload when WebView2 omits the custom dataTransfer type', () => {
    const payload: ConnectionEntryPanelDragPayload = {
      type: 'connection-entry',
      data: {
        sourceConnectionId: 'conn-1',
        sourceViewDir: '',
        entries: [
          { path: 'docs/readme.md', name: 'readme.md', isDirectory: false },
        ],
      },
    }

    setActiveInternalDrag({
      sourcePanelUid: 10,
      sourcePanelIndex: 0,
      payload,
    })

    const event = {
      dataTransfer: {
        types: [],
      },
    } as unknown as DragEvent

    expect(resolveConnectionEntryDragPayload(event)).toBe(payload)
  })

  it('does not treat external file drags as internal payloads even if stale state exists', () => {
    const payload: ConnectionEntryPanelDragPayload = {
      type: 'connection-entry',
      data: {
        sourceConnectionId: 'conn-1',
        sourceViewDir: '',
        entries: [
          { path: 'docs/readme.md', name: 'readme.md', isDirectory: false },
        ],
      },
    }

    setActiveInternalDrag({
      sourcePanelUid: 10,
      sourcePanelIndex: 0,
      payload,
    })

    const event = {
      dataTransfer: {
        types: ['Files'],
      },
    } as unknown as DragEvent

    expect(resolveConnectionEntryDragPayload(event)).toBeNull()
  })
})
