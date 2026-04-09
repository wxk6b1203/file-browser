import { describe, expect, it } from 'vitest'
import { folder } from '../../../../wailsjs/go/models'
import { explorerNodeKey, removeLoadedEntriesFromExplorerTree } from '../treeMutation'

function fileInfo(overrides: Partial<folder.FileInfo>) {
  return folder.FileInfo.createFrom({
    name: '',
    path: '',
    type: 1,
    size: 0,
    ...overrides,
  })
}

describe('removeLoadedEntriesFromExplorerTree', () => {
  it('removes deleted files from the loaded parent node only', () => {
    const next = removeLoadedEntriesFromExplorerTree(
      {
        [explorerNodeKey('connection', 'conn')]: [
          fileInfo({ name: 'a.txt', path: 'a.txt', type: 1 }),
          fileInfo({ name: 'docs', path: 'docs', type: 2 }),
        ],
      },
      'conn',
      '',
      ['a.txt'],
    )

    expect(next[explorerNodeKey('connection', 'conn')]!.map((item) => item.path)).toEqual(['docs'])
  })

  it('drops deleted directory nodes and their cached subtree', () => {
    const next = removeLoadedEntriesFromExplorerTree(
      {
        [explorerNodeKey('connection', 'conn')]: [
          fileInfo({ name: 'docs', path: 'docs', type: 2 }),
          fileInfo({ name: 'notes.txt', path: 'notes.txt', type: 1 }),
        ],
        [explorerNodeKey('directory', 'conn', 'docs')]: [
          fileInfo({ name: 'nested.txt', path: 'docs/nested.txt', type: 1 }),
        ],
        [explorerNodeKey('directory', 'conn', 'docs/subdir')]: [
          fileInfo({ name: 'deep.txt', path: 'docs/subdir/deep.txt', type: 1 }),
        ],
      },
      'conn',
      '',
      ['docs'],
    )

    expect(next[explorerNodeKey('connection', 'conn')]!.map((item) => item.path)).toEqual(['notes.txt'])
    expect(next[explorerNodeKey('directory', 'conn', 'docs')]).toBeUndefined()
    expect(next[explorerNodeKey('directory', 'conn', 'docs/subdir')]).toBeUndefined()
  })
})
