import type { folder } from '../../../wailsjs/go/models'
import { normalizeRemotePath } from '@/composables/remotePath'

export function explorerNodeKey(kind: 'connection' | 'directory' | 'file', connectionId: string, path = '') {
  if (kind === 'connection') {
    return `connection:${connectionId}`
  }
  return `${kind}:${connectionId}:${normalizeRemotePath(path)}`
}

function normalizeDeletedPaths(paths: string[]) {
  return [...new Set(paths.map((path) => normalizeRemotePath(path)).filter(Boolean))]
}

function dropDeletedSubtrees(
  childrenByKey: Record<string, folder.FileInfo[]>,
  connectionId: string,
  deletedPaths: string[],
) {
  if (deletedPaths.length === 0) return childrenByKey

  const deletedPrefixes = deletedPaths.map((path) => ({
    root: explorerNodeKey('directory', connectionId, path),
    prefix: `${explorerNodeKey('directory', connectionId, path)}/`,
  }))

  const nextChildren: Record<string, folder.FileInfo[]> = {}
  for (const [key, value] of Object.entries(childrenByKey)) {
    if (deletedPrefixes.some(({ root, prefix }) => key === root || key.startsWith(prefix))) {
      continue
    }
    nextChildren[key] = value
  }
  return nextChildren
}

export function removeLoadedEntriesFromExplorerTree(
  childrenByKey: Record<string, folder.FileInfo[]>,
  connectionId: string,
  parentPath: string,
  deletedPaths: string[],
) {
  const normalizedPaths = normalizeDeletedPaths(deletedPaths)
  if (normalizedPaths.length === 0) return childrenByKey

  const parentKey = normalizeRemotePath(parentPath)
    ? explorerNodeKey('directory', connectionId, parentPath)
    : explorerNodeKey('connection', connectionId)

  let nextChildren = childrenByKey
  const currentItems = childrenByKey[parentKey]
  if (currentItems) {
    const deletedPathSet = new Set(normalizedPaths)
    const filteredItems = currentItems.filter((item) => !deletedPathSet.has(normalizeRemotePath(item.path)))
    if (filteredItems.length !== currentItems.length) {
      nextChildren = {
        ...childrenByKey,
        [parentKey]: filteredItems,
      }
    }
  }

  return dropDeletedSubtrees(nextChildren, connectionId, normalizedPaths)
}
