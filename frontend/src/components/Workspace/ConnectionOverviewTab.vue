<template>
  <div class="file-browser">
    <div class="file-browser__toolbar">
      <div class="file-browser__meta">
        <p class="file-browser__eyebrow">{{ definition?.driver ?? 'Connection' }}</p>
        <h2 class="file-browser__title">{{ definition?.name ?? t('workspace.fileBrowser.loading') }}</h2>
      </div>

      <div class="file-browser__controls">
        <el-button @click="createDirectory">
          <i-ep-folder-add />
          {{ t('workspace.fileBrowser.newFolder') }}
        </el-button>
        <el-button @click="reload">
          <i-ep-refresh-right />
        </el-button>
        <div class="file-browser__view-toggle">
          <button
            class="file-browser__view-btn"
            :class="{ 'file-browser__view-btn--active': viewMode === 'list' }"
            @click="viewMode = 'list'"
          >
            <i-ep-expand />
          </button>
          <button
            class="file-browser__view-btn"
            :class="{ 'file-browser__view-btn--active': viewMode === 'grid' }"
            @click="viewMode = 'grid'"
          >
            <i-ep-grid />
          </button>
        </div>
      </div>
    </div>

    <div class="file-browser__breadcrumbs">
      <button class="file-browser__crumb" @click="goRoot">
        {{ definition?.name ?? 'Root' }}
      </button>
      <template v-for="crumb in breadcrumbs" :key="crumb.path">
        <span class="file-browser__crumb-sep">/</span>
        <button class="file-browser__crumb" @click="openPath(crumb.path)">
          {{ crumb.label }}
        </button>
      </template>
    </div>

    <div class="file-browser__status">
      <span
        class="file-browser__status-dot"
        :class="{ 'file-browser__status-dot--online': connections.stateMap.get(connectionId)?.connected }"
      />
      <span>{{ connections.stateMap.get(connectionId)?.connected ? t('workspace.fileBrowser.connected') : t('workspace.fileBrowser.disconnected') }}</span>
      <span>{{ items.length }} {{ t('workspace.fileBrowser.items') }}</span>
    </div>

    <div v-if="loading" class="file-browser__empty">
      {{ t('workspace.fileBrowser.loading') }}
    </div>

    <div v-else-if="items.length === 0" class="file-browser__empty">
      <i-mdi-folder-outline class="file-browser__empty-icon" />
      <span>{{ t('workspace.fileBrowser.empty') }}</span>
    </div>

    <div v-else-if="viewMode === 'list'" class="file-browser__list">
      <el-dropdown
        v-for="item in items"
        :key="item.path"
        trigger="contextmenu"
        placement="bottom-start"
        @command="onItemCommand($event, item)"
      >
        <button
          class="file-browser__row"
          draggable="true"
          @dragstart="onItemDragStart(item, $event)"
          @dragend="onItemDragEnd"
          @dblclick="openItem(item)"
        >
          <span class="file-browser__row-icon">
            <component :is="resolveFileIcon(item)" />
          </span>
          <span class="file-browser__row-name">{{ item.name }}</span>
          <span class="file-browser__row-size">{{ isDirectory(item) ? '--' : formatSize(item.size) }}</span>
          <span class="file-browser__row-time">{{ formatTime(item.lastModified) }}</span>
        </button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="download">{{ t('workspace.fileBrowser.download') }}</el-dropdown-item>
            <el-dropdown-item command="rename">{{ t('workspace.fileBrowser.rename') }}</el-dropdown-item>
            <el-dropdown-item command="delete" divided>{{ t('workspace.fileBrowser.delete') }}</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <div v-else class="file-browser__grid">
      <el-dropdown
        v-for="item in items"
        :key="item.path"
        trigger="contextmenu"
        placement="bottom-start"
        @command="onItemCommand($event, item)"
      >
        <button
          class="file-browser__tile"
          draggable="true"
          @dragstart="onItemDragStart(item, $event)"
          @dragend="onItemDragEnd"
          @dblclick="openItem(item)"
        >
          <span class="file-browser__tile-icon">
            <component :is="resolveFileIcon(item, { opened: isDirectory(item) })" />
          </span>
          <span class="file-browser__tile-name">{{ item.name }}</span>
          <span class="file-browser__tile-meta">{{ isDirectory(item) ? t('workspace.fileBrowser.directory') : formatSize(item.size) }}</span>
        </button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="download">{{ t('workspace.fileBrowser.download') }}</el-dropdown-item>
            <el-dropdown-item command="rename">{{ t('workspace.fileBrowser.rename') }}</el-dropdown-item>
            <el-dropdown-item command="delete" divided>{{ t('workspace.fileBrowser.delete') }}</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { CreateConnectionDirectory, DeleteConnectionEntry, DownloadConnectionFileToTemp, ListConnectionDirectory, RenameConnectionEntry } from '../../../wailsjs/go/main/App'
import { folder } from '../../../wailsjs/go/models'
import { splitPanePanelKey } from '@/components/SplitPane/types'
import { CONNECTION_DIRECTORY_REFRESH_EVENT, type ConnectionDirectoryRefreshDetail } from '@/composables/useConnectionDirectoryRefresh'
import { DIRECTORY_ENTRY_TYPE, resolveFileIcon } from '@/composables/useFileIcons'
import { SPLITPANE_DRAG_TYPE, clearActiveInternalDrag, setActiveInternalDrag } from '@/composables/splitPaneDragState'
import { useConnectionsStore } from '@/stores/connections'
import { useWorkspaceStore } from '@/stores/workspace'

const props = defineProps<{
  connectionId: string
}>()

const { t, locale } = useI18n()
const connections = useConnectionsStore()
const workspace = useWorkspaceStore()
const splitPanePanel = inject(splitPanePanelKey, null)

const loading = ref(false)
const currentPath = ref('')
const items = ref<folder.FileInfo[]>([])
const viewMode = ref<'list' | 'grid'>('list')

const connectionId = computed(() => props.connectionId)
const definition = computed(() => connections.definitionMap.get(connectionId.value) ?? null)
const targetPath = computed(() => workspace.getConnectionPath(connectionId.value))
const breadcrumbs = computed(() => {
  if (!currentPath.value) return []
  const segments = currentPath.value.split('/').filter(Boolean)
  return segments.map((label, index) => ({
    label,
    path: segments.slice(0, index + 1).join('/'),
  }))
})

interface ConnectionEntryDragPayload {
  type: 'connection-entry'
  data: {
    sourceConnectionId: string
    sourcePath: string
    sourceName: string
    sourceIsDirectory: boolean
  }
}

function isDirectory(item: folder.FileInfo) {
  return item.type === DIRECTORY_ENTRY_TYPE
}

function formatSize(size: number) {
  if (!size) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = size
  let unitIndex = 0
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex++
  }
  return `${value.toFixed(value >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

function formatTime(value?: any) {
  if (!value) return '--'
  return new Date(value).toLocaleString(locale.value)
}

async function load(dir = '') {
  loading.value = true
  try {
    await connections.openConnection(connectionId.value)
    items.value = await ListConnectionDirectory(connectionId.value, dir)
    currentPath.value = dir
    workspace.setConnectionPath(connectionId.value, dir)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  } finally {
    loading.value = false
  }
}

function reload() {
  return load(currentPath.value)
}

function goRoot() {
  return load('')
}

function openPath(path: string) {
  return load(path)
}

function openItem(item: folder.FileInfo) {
  if (isDirectory(item)) {
    return load(item.path)
  }
  ElMessage.info(t('workspace.fileBrowser.fileOpenPending', { name: item.name }))
}

function onItemDragStart(item: folder.FileInfo, event: DragEvent) {
  if (!event.dataTransfer) return
  event.dataTransfer.setData(SPLITPANE_DRAG_TYPE, '')
  event.dataTransfer.effectAllowed = 'move'

  const payload: ConnectionEntryDragPayload = {
    type: 'connection-entry',
    data: {
      sourceConnectionId: connectionId.value,
      sourcePath: item.path,
      sourceName: item.name,
      sourceIsDirectory: isDirectory(item),
    },
  }

  setActiveInternalDrag({
    sourcePanelUid: splitPanePanel?.uid ?? -1,
    sourcePanelIndex: splitPanePanel?.index.value ?? -1,
    payload,
  })
}

function onItemDragEnd() {
  clearActiveInternalDrag()
}

async function createDirectory() {
  try {
    const { value } = await ElMessageBox.prompt(
      t('workspace.fileBrowser.newFolderPrompt'),
      t('workspace.fileBrowser.newFolder'),
      {
        inputPlaceholder: t('workspace.fileBrowser.newFolderPlaceholder'),
      },
    )
    if (!value) return
    await CreateConnectionDirectory(connectionId.value, currentPath.value, value)
    await reload()
    ElMessage.success(t('workspace.fileBrowser.newFolderSuccess'))
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error instanceof Error ? error.message : String(error))
    }
  }
}

async function renameItem(item: folder.FileInfo) {
  try {
    const { value } = await ElMessageBox.prompt(
      t('workspace.fileBrowser.renamePrompt', { name: item.name }),
      t('workspace.fileBrowser.rename'),
      {
        inputValue: item.name,
      },
    )
    if (!value || value === item.name) return
    await RenameConnectionEntry(connectionId.value, item.path, value)
    await reload()
    ElMessage.success(t('workspace.fileBrowser.renameSuccess'))
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error instanceof Error ? error.message : String(error))
    }
  }
}

async function deleteItem(item: folder.FileInfo) {
  try {
    await ElMessageBox.confirm(
      t('workspace.fileBrowser.deletePrompt', { name: item.name }),
      t('workspace.fileBrowser.delete'),
      { type: 'warning' },
    )
    await DeleteConnectionEntry(connectionId.value, item.path)
    await reload()
    ElMessage.success(t('workspace.fileBrowser.deleteSuccess'))
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error instanceof Error ? error.message : String(error))
    }
  }
}

async function downloadItem(item: folder.FileInfo) {
  try {
    const taskIds = await DownloadConnectionFileToTemp(connectionId.value, item.path)
    if (taskIds.length === 0) {
      ElMessage.success(t('workspace.fileBrowser.downloadPrepared', { name: item.name }))
      return
    }
    ElMessage.success(t('workspace.fileBrowser.downloadQueuedBatch', { count: taskIds.length }))
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  }
}

function onItemCommand(command: string | number | object, item: folder.FileInfo) {
  const action = String(command)
  if (action === 'download') {
    return downloadItem(item)
  }
  if (action === 'rename') {
    return renameItem(item)
  }
  if (action === 'delete') {
    return deleteItem(item)
  }
}

function onDirectoryRefresh(event: Event) {
  const detail = (event as CustomEvent<ConnectionDirectoryRefreshDetail>).detail
  if (!detail) return
  if (detail.connectionId !== connectionId.value) return
  if (detail.path !== currentPath.value) return
  void reload()
}

watch(targetPath, (path) => {
  if (path === currentPath.value) return
  void load(path)
})

onMounted(async () => {
  await connections.hydrate()
  window.addEventListener(CONNECTION_DIRECTORY_REFRESH_EVENT, onDirectoryRefresh as EventListener)
  await load(targetPath.value)
})

onBeforeUnmount(() => {
  window.removeEventListener(CONNECTION_DIRECTORY_REFRESH_EVENT, onDirectoryRefresh as EventListener)
})
</script>

<style scoped>
.file-browser {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 18px;
  overflow: auto;
  background:
    radial-gradient(circle at top left, color-mix(in srgb, var(--theme-color-primary) 16%, transparent), transparent 28%),
    var(--theme-color-bg-base);
}

.file-browser__toolbar {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;
}

.file-browser__eyebrow {
  margin: 0 0 6px;
  font-size: 12px;
  color: var(--theme-color-primary);
  text-transform: uppercase;
  letter-spacing: 0.14em;
}

.file-browser__title {
  margin: 0;
  color: var(--theme-color-text-base);
  font-size: 24px;
}

.file-browser__controls {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.file-browser__view-toggle {
  display: inline-flex;
  padding: 4px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 12px;
  background: color-mix(in srgb, var(--theme-color-bg-surface) 88%, transparent);
}

.file-browser__view-btn {
  width: 34px;
  height: 34px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--theme-color-text-secondary);
  cursor: pointer;
}

.file-browser__view-btn--active {
  background: var(--theme-color-primary-light);
  color: var(--theme-color-primary);
}

.file-browser__breadcrumbs {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 10px;
}

.file-browser__crumb {
  padding: 4px 8px;
  border: none;
  border-radius: 8px;
  background: color-mix(in srgb, var(--theme-color-bg-surface) 86%, transparent);
  color: var(--theme-color-text);
  cursor: pointer;
}

.file-browser__crumb:hover {
  background: var(--theme-color-bg-hover);
}

.file-browser__crumb-sep {
  color: var(--theme-color-text-secondary);
}

.file-browser__status {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 16px;
  color: var(--theme-color-text-secondary);
}

.file-browser__status-dot {
  width: 10px;
  height: 10px;
  border-radius: 999px;
  background: var(--theme-color-danger);
}

.file-browser__status-dot--online {
  background: var(--theme-color-success);
}

.file-browser__empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: var(--theme-color-text-secondary);
  text-align: center;
}

.file-browser__empty-icon {
  font-size: 28px;
}

.file-browser__list {
  display: grid;
  gap: 8px;
}

.file-browser__row {
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr) 100px 180px;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 14px;
  background: color-mix(in srgb, var(--theme-color-bg-surface) 86%, transparent);
  color: var(--theme-color-text);
  text-align: left;
  cursor: pointer;
}

.file-browser__row:hover {
  background: var(--theme-color-bg-hover);
}

.file-browser__row-icon {
  font-size: 18px;
  color: var(--theme-color-primary);
}

.file-browser__row-name {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-browser__row-size,
.file-browser__row-time {
  color: var(--theme-color-text-secondary);
  font-size: 12px;
}

.file-browser__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 12px;
}

.file-browser__tile {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
  min-height: 132px;
  padding: 16px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 16px;
  background: color-mix(in srgb, var(--theme-color-bg-surface) 88%, transparent);
  color: var(--theme-color-text);
  text-align: left;
  cursor: pointer;
}

.file-browser__tile:hover {
  background: var(--theme-color-bg-hover);
}

.file-browser__tile-icon {
  font-size: 28px;
  color: var(--theme-color-primary);
}

.file-browser__tile-name {
  font-weight: 600;
  color: var(--theme-color-text-base);
}

.file-browser__tile-meta {
  font-size: 12px;
  color: var(--theme-color-text-secondary);
}

@media (max-width: 900px) {
  .file-browser {
    padding: 14px;
  }

  .file-browser__toolbar {
    flex-direction: column;
  }

  .file-browser__row {
    grid-template-columns: 28px minmax(0, 1fr);
  }

  .file-browser__row-size,
  .file-browser__row-time {
    display: none;
  }
}
</style>
