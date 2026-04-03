export const CONNECTION_ENTRY_DROP_LIFECYCLE_EVENT = 'workspace:connection-entry-drop-lifecycle'

export interface ConnectionEntryDropLifecycleDetail {
  operationId: string
  phase: 'start' | 'finish'
  mode: 'move' | 'transfer'
  sourceConnectionId: string
  sourceViewDir: string
  sourcePaths: string[]
  sourceDirs: string[]
  sourceDirectoryPaths: string[]
  targetConnectionId: string
  targetDir: string
  success: boolean
}

let latestSuccessfulDrop: ConnectionEntryDropLifecycleDetail | null = null

export function emitConnectionEntryDropLifecycle(detail: ConnectionEntryDropLifecycleDetail) {
  if (detail.phase === 'finish' && detail.success) {
    latestSuccessfulDrop = detail
  }
  window.dispatchEvent(new CustomEvent<ConnectionEntryDropLifecycleDetail>(
    CONNECTION_ENTRY_DROP_LIFECYCLE_EVENT,
    { detail },
  ))
}

export function consumeLatestSuccessfulConnectionEntryDrop(match: {
  sourceConnectionId: string
  sourceViewDir: string
  sourcePaths: string[]
}): ConnectionEntryDropLifecycleDetail | null {
  if (!latestSuccessfulDrop) return null
  if (latestSuccessfulDrop.mode !== 'move') return null
  if (latestSuccessfulDrop.sourceConnectionId !== match.sourceConnectionId) return null
  if (latestSuccessfulDrop.sourceViewDir !== match.sourceViewDir) return null

  const expected = new Set(match.sourcePaths)
  const actual = new Set(latestSuccessfulDrop.sourcePaths)
  if (expected.size !== actual.size) return null
  for (const path of expected) {
    if (!actual.has(path)) return null
  }

  const detail = latestSuccessfulDrop
  latestSuccessfulDrop = null
  return detail
}
