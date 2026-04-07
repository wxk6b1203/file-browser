import { MoveConnectionEntry, TransferConnectionEntry } from '../../wailsjs/go/main/App'
import type { folder } from '../../wailsjs/go/models'
import { emitConnectionDirectoryRefresh } from './useConnectionDirectoryRefresh'
import { emitConnectionEntryDropLifecycle } from './useConnectionEntryDropLifecycle'
import { getActiveInternalDrag, isInternalDragEvent, type PanelDragPayload } from './splitPaneDragState'
import {
  baseRemotePath,
  isRemotePathWithin,
  joinRemotePath,
  normalizeRemotePath,
  parentRemotePath,
  remotePathKey,
} from './remotePath'

export interface ConnectionDragEntry {
  path: string
  name: string
  isDirectory: boolean
}

export interface ConnectionEntryPanelDragPayload extends PanelDragPayload {
  type: 'connection-entry'
  data: {
    sourceConnectionId: string
    sourceViewDir: string
    entries: ConnectionDragEntry[]
  }
}

interface PlannedSameConnectionMove {
  sourcePath: string
  sourceParentDir: string
  isDirectory: boolean
}

interface PlannedCrossConnectionTransfer {
  sourcePath: string
  isDirectory: boolean
}

interface PlannedConnectionEntryDrop {
  targetDir: string
  sourceConnectionId: string
  sameConnectionMoves: PlannedSameConnectionMove[]
  crossConnectionTransfers: PlannedCrossConnectionTransfer[]
  skippedCount: number
}

export interface ConnectionEntryDropResult {
  mode: 'move' | 'transfer' | 'noop'
  requestedEntryCount: number
  movedCount: number
  transferredEntryCount: number
  transferTaskCount: number
  skippedCount: number
}

export interface ConnectionEntryDropFeedback {
  key: string
  params: Record<string, string | number>
}

export function compactConnectionDragEntries(entries: ConnectionDragEntry[]): ConnectionDragEntry[] {
  const uniqueEntries: ConnectionDragEntry[] = []
  const seen = new Set<string>()

  for (const entry of entries) {
    const cleanPath = normalizeRemotePath(entry.path)
    if (!cleanPath || seen.has(cleanPath)) continue
    seen.add(cleanPath)
    uniqueEntries.push({
      ...entry,
      path: cleanPath,
    })
  }

  const selectedDirectoryPaths = new Set(
    uniqueEntries
      .filter((entry) => entry.isDirectory)
      .map((entry) => entry.path),
  )

  return uniqueEntries.filter((entry) => {
    const parentPath = parentRemotePath(entry.path)
    let cursor = parentPath
    while (cursor) {
      if (selectedDirectoryPaths.has(cursor)) {
        return false
      }
      cursor = parentRemotePath(cursor)
    }
    return true
  })
}

export function buildConnectionEntryDragPayload(
  sourceConnectionId: string,
  sourceViewDir: string,
  items: folder.FileInfo[],
): ConnectionEntryPanelDragPayload {
  const entries = compactConnectionDragEntries(items.map((item) => ({
    path: item.path,
    name: item.name,
    isDirectory: item.type === 2,
  })))

  return {
    type: 'connection-entry',
    data: {
      sourceConnectionId: sourceConnectionId.trim(),
      sourceViewDir: normalizeRemotePath(sourceViewDir),
      entries,
    },
  }
}

export function resolveConnectionEntryDragPayload(event: DragEvent): ConnectionEntryPanelDragPayload | null {
  if (!isInternalDragEvent(event)) return null
  const payload = getActiveInternalDrag()?.payload
  return isConnectionEntryDragPayload(payload) ? payload : null
}

export function canDropConnectionEntries(
  payload: ConnectionEntryPanelDragPayload | null | undefined,
  targetConnectionId: string,
  targetDir: string,
): boolean {
  return buildDropPlan(payload, targetConnectionId, targetDir) !== null
}

export function labelForDropTarget(targetDir: string, fallback: string) {
  const cleanTargetDir = normalizeRemotePath(targetDir)
  if (!cleanTargetDir) return fallback
  const segments = cleanTargetDir.split('/').filter(Boolean)
  return segments[segments.length - 1] ?? fallback
}

export function buildConnectionEntryDropFeedback(
  result: ConnectionEntryDropResult,
  targetLabel: string,
): ConnectionEntryDropFeedback {
  if (result.mode === 'move') {
    return result.skippedCount > 0
      ? {
          key: 'workspace.fileBrowser.dropMovedSummarySkipped',
          params: {
            entryCount: result.movedCount,
            target: targetLabel,
            skipped: result.skippedCount,
          },
        }
      : {
          key: 'workspace.fileBrowser.dropMovedSummary',
          params: {
            entryCount: result.movedCount,
            target: targetLabel,
          },
        }
  }

  if (result.mode === 'transfer') {
    if (result.transferTaskCount > 0) {
      return result.skippedCount > 0
        ? {
            key: 'workspace.fileBrowser.dropQueuedSummarySkipped',
            params: {
              entryCount: result.transferredEntryCount,
              taskCount: result.transferTaskCount,
              target: targetLabel,
              skipped: result.skippedCount,
            },
          }
        : {
            key: 'workspace.fileBrowser.dropQueuedSummary',
            params: {
              entryCount: result.transferredEntryCount,
              taskCount: result.transferTaskCount,
              target: targetLabel,
            },
          }
    }

    return result.skippedCount > 0
      ? {
          key: 'workspace.fileBrowser.dropPreparedSummarySkipped',
          params: {
            entryCount: result.transferredEntryCount,
            target: targetLabel,
            skipped: result.skippedCount,
          },
        }
      : {
          key: 'workspace.fileBrowser.dropPreparedSummary',
          params: {
            entryCount: result.transferredEntryCount,
            target: targetLabel,
          },
        }
  }

  return {
    key: 'workspace.fileBrowser.dropNoop',
    params: {
      skipped: result.skippedCount || result.requestedEntryCount,
    },
  }
}

export async function executeConnectionEntryDrop(
  payload: ConnectionEntryPanelDragPayload,
  targetConnectionId: string,
  targetDir: string,
): Promise<ConnectionEntryDropResult> {
  const plan = buildDropPlan(payload, targetConnectionId, targetDir)
  if (!plan) {
    return {
      mode: 'noop',
      requestedEntryCount: payload.data.entries.length,
      movedCount: 0,
      transferredEntryCount: 0,
      transferTaskCount: 0,
      skippedCount: payload.data.entries.length,
    }
  }

  const refreshTargets = new Set<string>()
  const sourcePaths = plan.sameConnectionMoves.map((item) => item.sourcePath)
  const sourceDirs = Array.from(new Set(plan.sameConnectionMoves.map((item) => item.sourceParentDir)))
  const sourceDirectoryPaths = plan.sameConnectionMoves
    .filter((item) => item.isDirectory)
    .map((item) => item.sourcePath)
  let movedCount = 0
  let transferredEntryCount = 0
  let transferTaskCount = 0
  let success = false
  const operationId = `drop-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
  const mode: 'move' | 'transfer' = plan.sameConnectionMoves.length > 0 ? 'move' : 'transfer'

  emitConnectionEntryDropLifecycle({
    operationId,
    phase: 'start',
    mode,
    sourceConnectionId: plan.sourceConnectionId,
    sourceViewDir: payload.data.sourceViewDir,
    sourcePaths,
    sourceDirs,
    sourceDirectoryPaths,
    targetConnectionId: targetConnectionId.trim(),
    targetDir: plan.targetDir,
    success: false,
  })

  try {
    if (plan.sameConnectionMoves.length > 0) {
      for (const item of plan.sameConnectionMoves) {
        await MoveConnectionEntry(plan.sourceConnectionId, item.sourcePath, plan.targetDir)
        movedCount++
        if (item.isDirectory) {
          refreshTargets.add(remotePathKey(plan.sourceConnectionId, item.sourceParentDir))
          refreshTargets.add(remotePathKey(plan.sourceConnectionId, plan.targetDir))
        }
      }
    }

    if (plan.crossConnectionTransfers.length > 0) {
      for (const item of plan.crossConnectionTransfers) {
        const taskIDs = await TransferConnectionEntry(
          plan.sourceConnectionId,
          item.sourcePath,
          targetConnectionId.trim(),
          plan.targetDir,
        )
        transferredEntryCount++
        transferTaskCount += taskIDs.length
        if (item.isDirectory) {
          refreshTargets.add(remotePathKey(targetConnectionId, plan.targetDir))
        }
      }
    }

    for (const key of refreshTargets) {
      const [connectionId, path] = key.split('::')
      emitConnectionDirectoryRefresh({
        connectionId: connectionId ?? '',
        path: path ?? '',
        source: 'transfer',
        taskId: 'drag-drop',
      })
    }

    success = true

    if (movedCount > 0) {
      return {
        mode: 'move',
        requestedEntryCount: payload.data.entries.length,
        movedCount,
        transferredEntryCount: 0,
        transferTaskCount: 0,
        skippedCount: plan.skippedCount,
      }
    }

    if (transferredEntryCount > 0) {
      return {
        mode: 'transfer',
        requestedEntryCount: payload.data.entries.length,
        movedCount: 0,
        transferredEntryCount,
        transferTaskCount,
        skippedCount: plan.skippedCount,
      }
    }

    return {
      mode: 'noop',
      requestedEntryCount: payload.data.entries.length,
      movedCount: 0,
      transferredEntryCount: 0,
      transferTaskCount: 0,
      skippedCount: plan.skippedCount,
    }
  } finally {
    emitConnectionEntryDropLifecycle({
      operationId,
      phase: 'finish',
      mode,
      sourceConnectionId: plan.sourceConnectionId,
      sourceViewDir: payload.data.sourceViewDir,
      sourcePaths,
      sourceDirs,
      sourceDirectoryPaths,
      targetConnectionId: targetConnectionId.trim(),
      targetDir: plan.targetDir,
      success,
    })
  }
}

function isConnectionEntryDragPayload(value: unknown): value is ConnectionEntryPanelDragPayload {
  if (!value || typeof value !== 'object') return false
  const payload = value as Partial<ConnectionEntryPanelDragPayload>
  if (payload.type !== 'connection-entry') return false
  if (!payload.data || typeof payload.data !== 'object') return false
  if (typeof payload.data.sourceConnectionId !== 'string') return false
  if (typeof payload.data.sourceViewDir !== 'string') return false
  if (!Array.isArray(payload.data.entries)) return false
  return payload.data.entries.every((entry) =>
    entry &&
    typeof entry === 'object' &&
    typeof entry.path === 'string' &&
    typeof entry.name === 'string' &&
    typeof entry.isDirectory === 'boolean',
  )
}

function buildDropPlan(
  payload: ConnectionEntryPanelDragPayload | null | undefined,
  targetConnectionId: string,
  targetDir: string,
): PlannedConnectionEntryDrop | null {
  if (!payload) return null

  const sourceConnectionId = payload.data.sourceConnectionId.trim()
  const cleanTargetConnectionId = targetConnectionId.trim()
  if (!sourceConnectionId || !cleanTargetConnectionId) return null

  const cleanTargetDir = normalizeRemotePath(targetDir)
  const entries = compactConnectionDragEntries(payload.data.entries)
  const sameConnectionMoves: PlannedSameConnectionMove[] = []
  const crossConnectionTransfers: PlannedCrossConnectionTransfer[] = []
  let skippedCount = 0

  for (const entry of entries) {
    const cleanSourcePath = normalizeRemotePath(entry.path)
    if (!cleanSourcePath) {
      skippedCount++
      continue
    }

    if (sourceConnectionId === cleanTargetConnectionId) {
      const nextPath = joinRemotePath(cleanTargetDir, baseRemotePath(cleanSourcePath))
      if (!nextPath || nextPath === cleanSourcePath) {
        skippedCount++
        continue
      }
      if (entry.isDirectory && isRemotePathWithin(cleanTargetDir, cleanSourcePath)) {
        skippedCount++
        continue
      }
      sameConnectionMoves.push({
        sourcePath: cleanSourcePath,
        sourceParentDir: parentRemotePath(cleanSourcePath),
        isDirectory: entry.isDirectory,
      })
      continue
    }

    crossConnectionTransfers.push({
      sourcePath: cleanSourcePath,
      isDirectory: entry.isDirectory,
    })
  }

  if (sameConnectionMoves.length === 0 && crossConnectionTransfers.length === 0) {
    return null
  }

  return {
    targetDir: cleanTargetDir,
    sourceConnectionId,
    sameConnectionMoves,
    crossConnectionTransfers,
    skippedCount,
  }
}

export const __test__ = {
  buildDropPlan,
}
