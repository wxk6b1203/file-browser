<template>
  <div class="explorer-panel">
    <div class="explorer-panel__header">
      <div>
        <p class="explorer-panel__eyebrow">{{ t('shell.explorer.eyebrow') }}</p>
        <h2 class="explorer-panel__title">{{ t('shell.explorer.title') }}</h2>
      </div>
      <div class="explorer-panel__actions">
        <el-button text @click="refreshAll">
          <i-ep-refresh-right />
        </el-button>
        <el-button text @click="workspace.openNewConnection()">
          <i-ep-plus />
        </el-button>
      </div>
    </div>

    <div v-if="connections.loading && !connections.ready" class="explorer-panel__empty">
      {{ t('shell.explorer.loading') }}
    </div>

    <div v-else-if="treeNodes.length === 0" class="explorer-panel__empty">
      <i-mdi-lan-disconnect class="explorer-panel__empty-icon" />
      <span>{{ t('shell.explorer.empty') }}</span>
      <el-button type="primary" size="small" @click="workspace.openNewConnection()">
        {{ t('actions.newConnection') }}
      </el-button>
    </div>

    <div v-else ref="explorerListRef" class="explorer-panel__list">
      <ExplorerTreeNode
        v-for="node in treeNodes"
        :key="node.key"
        :node="node"
        :expanded-keys="expandedKeys"
        :loading-keys="loadingKeys"
        :delete-pending-connection-id="pendingDeleteConnectionId"
        :delete-busy-connection-id="deleteBusyConnectionId"
        @toggle="toggleNode"
        @activate="activateNode"
        @open="openNode"
        @command="onNodeCommand"
        @hover-expand="onNodeHoverExpand"
        @drop-entries="onNodeDropEntries"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { ListConnectionDirectory } from '../../../wailsjs/go/main/App'
import { folder } from '../../../wailsjs/go/models'
import { buildConnectionEntryDropFeedback, executeConnectionEntryDrop, labelForDropTarget, type ConnectionEntryPanelDragPayload } from '@/composables/connectionEntryDrop'
import { CONNECTION_CONFIG_REFRESH_EVENT, type ConnectionConfigRefreshDetail } from '@/composables/useConnectionConfigRefresh'
import { normalizeRemotePath } from '@/composables/remotePath'
import { CONNECTION_DIRECTORY_REFRESH_EVENT, type ConnectionDirectoryRefreshDetail } from '@/composables/useConnectionDirectoryRefresh'
import { CONNECTION_ENTRY_DROP_LIFECYCLE_EVENT, type ConnectionEntryDropLifecycleDetail } from '@/composables/useConnectionEntryDropLifecycle'
import { useConnectionsStore } from '@/stores/connections'
import { useWorkspaceStore } from '@/stores/workspace'
import ExplorerTreeNode from './ExplorerTreeNode.vue'
import type { ExplorerNode } from './types'

const DIRECTORY_TYPE = 2

const { t } = useI18n()
const connections = useConnectionsStore()
const workspace = useWorkspaceStore()

const expandedKeys = ref(new Set<string>())
const loadingKeys = ref(new Set<string>())
const childrenByKey = ref<Record<string, folder.FileInfo[]>>({})
const explorerListRef = ref<HTMLElement | null>(null)
const pendingDeleteConnectionId = ref('')
const deleteBusyConnectionId = ref('')

const treeNodes = computed(() => {
  return connections.definitions.map((item) => buildConnectionNode(item.id))
})

function normalizeFolderItems(input: folder.FileInfo[] | null | undefined): folder.FileInfo[] {
  return Array.isArray(input) ? input : []
}

function nodeKey(kind: 'connection' | 'directory', connectionId: string, path = '') {
  return kind === 'connection'
    ? `connection:${connectionId}`
    : `directory:${connectionId}:${normalizeRemotePath(path)}`
}

function buildConnectionNode(connectionId: string): ExplorerNode {
  const item = connections.definitionMap.get(connectionId)
  const key = nodeKey('connection', connectionId)
  return {
    key,
    kind: 'connection',
    connectionId,
    label: item?.name ?? connectionId,
    path: '',
    level: 0,
    driver: item?.driver,
    connected: connections.stateMap.get(connectionId)?.connected ?? false,
    children: buildDirectoryChildren(key, connectionId, 1),
  }
}

function buildDirectoryChildren(parentKey: string, connectionId: string, level: number): ExplorerNode[] {
  const items = childrenByKey.value[parentKey] ?? []
  return items
    .filter((item) => item.type === DIRECTORY_TYPE)
    .map((item) => {
      const cleanPath = normalizeRemotePath(item.path)
      const key = nodeKey('directory', connectionId, cleanPath)
      return {
        key,
        kind: 'directory',
        connectionId,
        label: item.name,
        path: cleanPath,
        level,
        children: buildDirectoryChildren(key, connectionId, level + 1),
      }
    })
}

function setExpanded(key: string, expanded: boolean) {
  const next = new Set(expandedKeys.value)
  if (expanded) {
    next.add(key)
  } else {
    next.delete(key)
  }
  expandedKeys.value = next
}

function setLoading(key: string, loading: boolean) {
  const next = new Set(loadingKeys.value)
  if (loading) {
    next.add(key)
  } else {
    next.delete(key)
  }
  loadingKeys.value = next
}

function setChildren(key: string, items: folder.FileInfo[]) {
  childrenByKey.value = {
    ...childrenByKey.value,
    [key]: normalizeFolderItems(items),
  }
}

async function withPreservedExplorerScroll<T>(runner: () => Promise<T>) {
  const scrollTop = explorerListRef.value?.scrollTop ?? null
  const scrollLeft = explorerListRef.value?.scrollLeft ?? null
  try {
    return await runner()
  } finally {
    if (scrollTop !== null || scrollLeft !== null) {
      await nextTick()
      if (explorerListRef.value) {
        if (scrollTop !== null) explorerListRef.value.scrollTop = scrollTop
        if (scrollLeft !== null) explorerListRef.value.scrollLeft = scrollLeft
      }
    }
  }
}

async function loadChildren(node: ExplorerNode, options?: {
  preserveScroll?: boolean
}) {
  const runner = async () => {
    setLoading(node.key, true)
    try {
      if (!connections.stateMap.get(node.connectionId)?.connected) {
        await connections.openConnection(node.connectionId)
      }
      const children = normalizeFolderItems(await ListConnectionDirectory(node.connectionId, node.path))
      setChildren(node.key, children)
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : String(error))
    } finally {
      setLoading(node.key, false)
    }
  }

  if (options?.preserveScroll) {
    await withPreservedExplorerScroll(runner)
    return
  }

  await runner()
}

async function toggleNode(node: ExplorerNode) {
  const expanded = expandedKeys.value.has(node.key)
  if (expanded) {
    setExpanded(node.key, false)
    return
  }

  if (!(node.key in childrenByKey.value)) {
    await loadChildren(node)
  }
  setExpanded(node.key, true)
}

async function openNode(node: ExplorerNode) {
  if (node.kind === 'connection') {
    return onOpenConnection(node.connectionId, node.label)
  }
  return navigateToNode(node)
}

async function activateNode(node: ExplorerNode) {
  if (node.kind === 'connection') {
    return toggleNode(node)
  }
  return navigateToNode(node)
}

async function onOpenConnection(id: string, name: string) {
  try {
    await connections.openConnection(id)
    workspace.openConnection(id, name, '')
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  }
}

async function navigateToNode(node: ExplorerNode) {
  const name = connections.definitionMap.get(node.connectionId)?.name ?? node.connectionId
  try {
    await connections.openConnection(node.connectionId)
    workspace.openConnection(node.connectionId, name, node.path)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  }
}

async function onCommand(command: string, node: ExplorerNode) {
  if (node.kind !== 'connection') return
  const id = node.connectionId
  const name = node.label

  if (command === 'delete') {
    pendingDeleteConnectionId.value = id
    return
  }
  if (command === 'confirm-delete') {
    await confirmDeleteConnection(id)
    return
  }
  if (command === 'cancel-delete') {
    pendingDeleteConnectionId.value = ''
    return
  }
  if (command === 'open') {
    pendingDeleteConnectionId.value = ''
    await onOpenConnection(id, name)
    return
  }
  if (command === 'edit') {
    pendingDeleteConnectionId.value = ''
    workspace.openEditConnection(id, name)
    return
  }
  if (command === 'close') {
    pendingDeleteConnectionId.value = ''
    try {
      await connections.closeConnection(id)
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : String(error))
    }
    return
  }
}

function onNodeCommand(command: string, node: ExplorerNode) {
  return onCommand(command, node)
}

async function confirmDeleteConnection(connectionId: string) {
  if (!connectionId || deleteBusyConnectionId.value) return

  deleteBusyConnectionId.value = connectionId
  try {
    await connections.deleteConnection(connectionId)
    workspace.closeConnectionTabs(connectionId)
    resetConnectionTree(connectionId)
    pendingDeleteConnectionId.value = ''
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  } finally {
    deleteBusyConnectionId.value = ''
  }
}

async function onNodeHoverExpand(node: ExplorerNode) {
  if (expandedKeys.value.has(node.key) || loadingKeys.value.has(node.key)) return
  if (!(node.key in childrenByKey.value)) {
    await loadChildren(node)
  }
  setExpanded(node.key, true)
}

async function onNodeDropEntries(node: ExplorerNode, payload: ConnectionEntryPanelDragPayload) {
  const targetDir = node.kind === 'connection' ? '' : normalizeRemotePath(node.path)
  const targetName = connections.definitionMap.get(node.connectionId)?.name ?? node.connectionId

  try {
    const result = await executeConnectionEntryDrop(payload, node.connectionId, targetDir)
    const feedback = buildConnectionEntryDropFeedback(
      result,
      labelForDropTarget(targetDir, targetName),
    )
    if (result.mode !== 'noop') {
      ElMessage.success(t(feedback.key, feedback.params))
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  }
}

function resetConnectionTree(connectionId?: string) {
  if (!connectionId) {
    childrenByKey.value = {}
    expandedKeys.value = new Set<string>()
    loadingKeys.value = new Set<string>()
    return
  }

  const prefixA = `connection:${connectionId}`
  const prefixB = `directory:${connectionId}:`

  const nextChildren: Record<string, folder.FileInfo[]> = {}
  for (const [key, value] of Object.entries(childrenByKey.value)) {
    if (key !== prefixA && !key.startsWith(prefixB)) {
      nextChildren[key] = value
    }
  }
  childrenByKey.value = nextChildren

  expandedKeys.value = new Set([...expandedKeys.value].filter((key) => key !== prefixA && !key.startsWith(prefixB)))
  loadingKeys.value = new Set([...loadingKeys.value].filter((key) => key !== prefixA && !key.startsWith(prefixB)))
}

async function refreshAll() {
  resetConnectionTree()
  await connections.hydrate()
}

function buildRefreshNode(connectionId: string, targetPath = ''): ExplorerNode {
  const cleanTargetPath = normalizeRemotePath(targetPath)
  if (!cleanTargetPath) {
    return buildConnectionNode(connectionId)
  }

  const segments = cleanTargetPath.split('/').filter(Boolean)
  return {
    key: nodeKey('directory', connectionId, cleanTargetPath),
    kind: 'directory',
    connectionId,
    label: segments[segments.length - 1] ?? cleanTargetPath,
    path: cleanTargetPath,
    level: 0,
    children: [],
  }
}

function onDirectoryRefresh(event: Event) {
  const detail = (event as CustomEvent<ConnectionDirectoryRefreshDetail>).detail
  if (!detail) return

  const targetPath = normalizeRemotePath(detail.path)
  const targetKey = targetPath
    ? nodeKey('directory', detail.connectionId, targetPath)
    : nodeKey('connection', detail.connectionId)

  if (!(targetKey in childrenByKey.value)) {
    return
  }

  void loadChildren(buildRefreshNode(detail.connectionId, targetPath), { preserveScroll: true })
}

function dropDirectorySubtree(connectionId: string, sourcePath: string) {
  const cleanSourcePath = normalizeRemotePath(sourcePath)
  if (!cleanSourcePath) return

  const subtreeRootKey = `directory:${connectionId}:${cleanSourcePath}`
  const subtreePrefix = `${subtreeRootKey}/`
  const nextChildren: Record<string, folder.FileInfo[]> = {}
  for (const [key, value] of Object.entries(childrenByKey.value)) {
    if (key !== subtreeRootKey && !key.startsWith(subtreePrefix)) {
      nextChildren[key] = value
    }
  }
  childrenByKey.value = nextChildren
  expandedKeys.value = new Set(
    [...expandedKeys.value].filter((key) => key !== subtreeRootKey && !key.startsWith(subtreePrefix)),
  )
  loadingKeys.value = new Set(
    [...loadingKeys.value].filter((key) => key !== subtreeRootKey && !key.startsWith(subtreePrefix)),
  )
}

function removeMovedDirectoriesFromLoadedParents(detail: ConnectionEntryDropLifecycleDetail) {
  const directoryPaths = detail.sourceDirectoryPaths
    .map((path) => normalizeRemotePath(path))
    .filter(Boolean)

  if (directoryPaths.length === 0) return

  const directoryPathSet = new Set(directoryPaths)
  const nextChildren: Record<string, folder.FileInfo[]> = { ...childrenByKey.value }

  for (const sourceDir of new Set(detail.sourceDirs.map((path) => normalizeRemotePath(path)))) {
    const parentKey = sourceDir
      ? nodeKey('directory', detail.sourceConnectionId, sourceDir)
      : nodeKey('connection', detail.sourceConnectionId)
    const currentItems = nextChildren[parentKey]
    if (!currentItems) continue

    const filteredItems = currentItems.filter((item) => !directoryPathSet.has(normalizeRemotePath(item.path)))
    if (filteredItems.length !== currentItems.length) {
      nextChildren[parentKey] = filteredItems
    }
  }

  childrenByKey.value = nextChildren
  directoryPaths.forEach((path) => dropDirectorySubtree(detail.sourceConnectionId, path))
}

function onEntryDropLifecycle(event: Event) {
  const detail = (event as CustomEvent<ConnectionEntryDropLifecycleDetail>).detail
  if (!detail || detail.phase !== 'finish' || !detail.success || detail.mode !== 'move') return
  if (detail.sourceConnectionId !== detail.targetConnectionId) return
  if (detail.sourceDirectoryPaths.length === 0) return

  removeMovedDirectoriesFromLoadedParents(detail)

  const targetPath = normalizeRemotePath(detail.targetDir)
  const targetKey = targetPath
    ? nodeKey('directory', detail.targetConnectionId, targetPath)
    : nodeKey('connection', detail.targetConnectionId)

  if (!(targetKey in childrenByKey.value)) {
    return
  }

  void loadChildren(buildRefreshNode(detail.targetConnectionId, targetPath), { preserveScroll: true })
}

function onConnectionConfigRefresh(event: Event) {
  const detail = (event as CustomEvent<ConnectionConfigRefreshDetail>).detail
  if (!detail?.connectionId) return

  const connectionKey = nodeKey('connection', detail.connectionId)
  const wasExpanded = expandedKeys.value.has(connectionKey)

  resetConnectionTree(detail.connectionId)

  if (!wasExpanded) {
    return
  }

  setExpanded(connectionKey, true)
  void loadChildren(buildConnectionNode(detail.connectionId), { preserveScroll: true })
}

onMounted(() => {
  window.addEventListener(CONNECTION_CONFIG_REFRESH_EVENT, onConnectionConfigRefresh as EventListener)
  window.addEventListener(CONNECTION_DIRECTORY_REFRESH_EVENT, onDirectoryRefresh as EventListener)
  window.addEventListener(CONNECTION_ENTRY_DROP_LIFECYCLE_EVENT, onEntryDropLifecycle as EventListener)
})

watch(() => connections.definitions, (items) => {
  if (pendingDeleteConnectionId.value && !items.some((item) => item.id === pendingDeleteConnectionId.value)) {
    pendingDeleteConnectionId.value = ''
  }
})

onBeforeUnmount(() => {
  window.removeEventListener(CONNECTION_CONFIG_REFRESH_EVENT, onConnectionConfigRefresh as EventListener)
  window.removeEventListener(CONNECTION_DIRECTORY_REFRESH_EVENT, onDirectoryRefresh as EventListener)
  window.removeEventListener(CONNECTION_ENTRY_DROP_LIFECYCLE_EVENT, onEntryDropLifecycle as EventListener)
})
</script>

<style scoped>
.explorer-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--theme-color-bg-overlay) 50%, transparent), transparent 32%);
}

.explorer-panel__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 14px 10px;
  border-bottom: 1px solid var(--theme-color-border-light);
}

.explorer-panel__eyebrow {
  margin: 0 0 4px;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.14em;
  color: var(--theme-color-text-secondary);
}

.explorer-panel__title {
  margin: 0;
  font-size: 15px;
  color: var(--theme-color-text-base);
}

.explorer-panel__actions {
  display: flex;
  gap: 4px;
}

.explorer-panel__empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 18px;
  color: var(--theme-color-text-secondary);
  text-align: center;
}

.explorer-panel__empty-icon {
  font-size: 26px;
}

.explorer-panel__list {
  flex: 1;
  overflow: auto;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
</style>
