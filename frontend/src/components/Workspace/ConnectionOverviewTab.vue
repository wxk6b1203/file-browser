<template>
  <div
    ref="fileBrowserRef"
    class="file-browser"
    :class="{ 'file-browser--busy': dropBusy }"
    @contextmenu="clearNativeSelection"
    @selectstart="onFileBrowserSelectStart"
    @keydown.capture="onFileBrowserShortcutKeydown"
  >
    <div v-if="dropBusy" class="file-browser__busy-mask">
      <div class="file-browser__busy-card">
        <i-ep-loading class="file-browser__busy-icon" />
        <span>{{ t('workspace.fileBrowser.dropWorking') }}</span>
      </div>
    </div>
    <div class="file-browser__toolbar">
      <div class="file-browser__meta">
        <p class="file-browser__eyebrow">{{ definition?.driver ?? 'Connection' }}</p>
        <h2 class="file-browser__title">{{ definition?.name ?? t('workspace.fileBrowser.loading') }}</h2>
      </div>

      <div class="file-browser__controls">
        <el-button-group class="file-browser__action-group">
          <el-button size="small" @click="createDirectory">
            <i-ep-folder-add />
            {{ t('workspace.fileBrowser.newFolder') }}
          </el-button>
          <el-button size="small" @click="uploadFile">
            {{ t('workspace.fileBrowser.upload') }}
          </el-button>
          <el-button size="small" @click="reload">
            <i-ep-refresh-right />
            {{ t('workspace.fileBrowser.refresh') }}
          </el-button>
          <el-popover placement="bottom-end" trigger="click" :width="220">
            <template #reference>
              <el-button size="small">
                <i-ep-operation />
                {{ t('workspace.fileBrowser.columns') }}
              </el-button>
            </template>
            <div class="file-browser__column-picker">
              <label class="file-browser__column-option file-browser__column-option--locked">
                <el-checkbox :model-value="true" disabled />
                <span>{{ t('workspace.fileBrowser.columnName') }}</span>
              </label>
              <label
                v-for="column in optionalColumns"
                :key="column.key"
                class="file-browser__column-option"
              >
                <el-checkbox
                  :model-value="visibleColumnSet.has(column.key)"
                  @change="setColumnVisible(column.key, Boolean($event))"
                />
                <span>{{ column.label }}</span>
              </label>
            </div>
          </el-popover>
        </el-button-group>
        <el-button-group v-if="selectedItems.length > 0" class="file-browser__action-group">
          <el-button size="small" @click="downloadSelected">
            <i-ep-download />
            {{ t('workspace.fileBrowser.downloadSelected') }}
          </el-button>
          <el-button size="small" @click="copySelectedPaths">
            <i-ep-document-copy />
            {{ t('workspace.fileBrowser.copySelectedPaths') }}
          </el-button>
          <el-button size="small" @click="deleteSelected">
            <i-ep-delete />
            {{ t('workspace.fileBrowser.deleteSelected') }}
          </el-button>
          <el-button size="small" @click="clearSelection">
            <i-ep-close />
            {{ t('workspace.fileBrowser.clearSelection') }}
          </el-button>
        </el-button-group>
        <div class="file-browser__view-toggle">
          <button
            type="button"
            class="file-browser__view-btn"
            :class="{ 'file-browser__view-btn--active': viewMode === 'list' }"
            @click="viewMode = 'list'"
          >
            <i-ep-expand />
          </button>
          <button
            type="button"
            class="file-browser__view-btn"
            :class="{ 'file-browser__view-btn--active': viewMode === 'grid' }"
            @click="viewMode = 'grid'"
          >
            <i-ep-grid />
          </button>
        </div>
      </div>
    </div>

    <div class="file-browser__subbar">
      <div class="file-browser__breadcrumbs">
        <button class="file-browser__crumb file-browser__crumb--root" @click="goRoot">
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
        <span class="file-browser__status-pill">
          <span
            class="file-browser__status-dot"
            :class="{ 'file-browser__status-dot--online': connections.stateMap.get(connectionId)?.connected }"
          />
          <span>{{ connections.stateMap.get(connectionId)?.connected ? t('workspace.fileBrowser.connected') : t('workspace.fileBrowser.disconnected') }}</span>
        </span>
        <span class="file-browser__status-pill">
          {{ items.length }} {{ t('workspace.fileBrowser.items') }}
        </span>
        <span v-if="selectedItems.length > 0" class="file-browser__status-pill file-browser__status-pill--accent">
          {{ t('workspace.fileBrowser.selectedCount', { count: selectedItems.length }) }}
        </span>
      </div>
    </div>

    <div v-if="loading" class="file-browser__empty">
      {{ t('workspace.fileBrowser.loading') }}
    </div>

    <div v-else class="file-browser__viewport-wrap">
      <div
        v-if="localSearchOpen"
        class="file-browser__local-search"
        @mousedown.stop
        @click.stop
        @keydown.stop
      >
        <input
          ref="localSearchInputRef"
          v-model="localSearchDraft"
          class="file-browser__local-search-input"
          :placeholder="t('workspace.fileBrowser.localSearchPlaceholder')"
          @keydown.enter.prevent.stop="commitLocalSearch"
          @keydown.esc.prevent.stop="closeLocalSearch"
        >
        <button
          type="button"
          class="file-browser__local-search-option"
          :class="{ 'file-browser__local-search-option--active': localSearchCaseSensitive }"
          :title="t('workspace.fileBrowser.localSearchCaseSensitive')"
          @click="localSearchCaseSensitive = !localSearchCaseSensitive"
        >
          Aa
        </button>
        <button
          type="button"
          class="file-browser__local-search-option"
          :class="{ 'file-browser__local-search-option--active': localSearchWholeText }"
          :title="t('workspace.fileBrowser.localSearchWholeText')"
          @click="localSearchWholeText = !localSearchWholeText"
        >
          =
        </button>
        <button
          type="button"
          class="file-browser__local-search-option"
          :class="{ 'file-browser__local-search-option--active': localSearchRegex }"
          :title="t('workspace.fileBrowser.localSearchRegex')"
          @click="localSearchRegex = !localSearchRegex"
        >
          .*
        </button>
        <span
          class="file-browser__local-search-count"
          :class="{ 'file-browser__local-search-count--error': Boolean(localSearchError) }"
        >
          {{ localSearchError || localSearchCountLabel }}
        </span>
        <button
          type="button"
          class="file-browser__local-search-close"
          :title="t('workspace.fileBrowser.localSearchClose')"
          @click="closeLocalSearch"
        >
          ×
        </button>
      </div>

      <div
        v-if="viewMode === 'list'"
        ref="browserViewportRef"
        class="file-browser__list-shell"
        :class="{ 'file-browser__drop-root': browserDropActive }"
        tabindex="0"
        @contextmenu.prevent="onViewportContextMenu($event)"
        @keydown="onViewportKeydown"
        @dragover="onBrowserDragOver"
        @dragleave="onBrowserDragLeave"
        @drop="onBrowserDrop"
      >
        <div class="file-browser__list-table">
          <div class="file-browser__list-header" :style="listGridStyle">
            <button
              v-for="(column, index) in visibleColumns"
              :key="column.key"
              type="button"
              class="file-browser__list-header-btn"
              :class="[`file-browser__list-header-btn--${column.align}`]"
              @click="onHeaderClick(column.key, $event)"
              @dragstart.prevent
            >
              <span>{{ column.label }}</span>
              <span class="file-browser__sort-indicator">{{ sortIndicator(column.key) }}</span>
              <span
                v-if="canResizeColumn(index)"
                class="file-browser__column-resizer"
                @mousedown.stop.prevent="startColumnResize(index, $event)"
                @click.stop
              />
            </button>
          </div>

        <div class="file-browser__list-body">
          <div
            v-if="inlineCreateActive"
            class="file-browser__row file-browser__row--inline-create"
            :style="listGridStyle"
            @click.stop
            @dblclick.stop
            @contextmenu.prevent.stop="clearNativeSelection"
          >
            <span class="file-browser__cell file-browser__cell--name file-browser__cell--left">
              <span class="file-browser__row-icon">
                <component :is="resolveFileIcon({ name: inlineCreateName || 'folder', type: DIRECTORY_ENTRY_TYPE })" />
              </span>
              <input
                ref="inlineCreateInputRef"
                v-model="inlineCreateName"
                class="file-browser__inline-create-input"
                :placeholder="t('workspace.fileBrowser.newFolderPlaceholder')"
                :disabled="inlineCreateBusy"
                @keydown.enter.prevent="commitInlineCreate"
                @keydown.esc.prevent="cancelInlineCreate"
                @blur="commitInlineCreate"
              >
            </span>
            <span
              v-for="column in visibleColumns.slice(1)"
              :key="`inline-create:${column.key}`"
              class="file-browser__cell"
              :class="[`file-browser__cell--${column.key}`, `file-browser__cell--${column.align}`]"
            />
          </div>
          <div v-if="searchedItems.length === 0 && !inlineCreateActive" class="file-browser__viewport-empty file-browser__viewport-empty--list">
            <i-mdi-folder-outline class="file-browser__empty-icon" />
            <span>{{ localSearchQuery ? t('workspace.fileBrowser.localSearchNoMatches') : t('workspace.fileBrowser.empty') }}</span>
          </div>
          <div
            v-for="item in searchedItems"
            :key="normalizeEntryPath(item.path)"
            role="button"
            tabindex="-1"
            class="file-browser__row"
            :class="{
              'file-browser__row--selected': isSelected(item),
              'file-browser__row--active': activePath === normalizeEntryPath(item.path),
              'file-browser__row--delete-pending': isInlineDeletePending(item),
              'file-browser__row--drop-target': isDirectory(item) && directoryDropPath === normalizeEntryPath(item.path),
            }"
            :style="listGridStyle"
            :data-entry-path="normalizeEntryPath(item.path)"
            draggable="true"
            @mousedown.left="onItemMouseDown(item, $event)"
            @click="onItemClick(item, $event)"
            @contextmenu.prevent.stop="onItemContextMenu(item, $event)"
            @dragenter="onDirectoryDragEnter(item, $event)"
            @dragover="onDirectoryDragOver(item, $event)"
            @dragleave="onDirectoryDragLeave(item, $event)"
            @drop="onDirectoryDrop(item, $event)"
            @dragstart="onItemDragStart(item, $event)"
            @dragend="onItemDragEnd"
            @dblclick.stop.prevent="onItemDoubleClick(item)"
          >
            <span
              v-for="column in visibleColumns"
              :key="`${normalizeEntryPath(item.path)}:${column.key}`"
              class="file-browser__cell"
              :class="[`file-browser__cell--${column.key}`, `file-browser__cell--${column.align}`]"
            >
              <template v-if="column.key === 'name'">
                <span class="file-browser__row-icon">
                  <component :is="resolveFileIcon(item)" />
                </span>
                <span class="file-browser__entry-name-wrap">
                  <span class="file-browser__row-name">{{ item.name }}</span>
                  <span
                    v-if="isInlineDeletePending(item)"
                    class="file-browser__inline-delete-actions"
                    @mousedown.stop
                    @click.stop
                    @dblclick.stop
                  >
                    <button
                      type="button"
                      class="file-browser__inline-delete-btn file-browser__inline-delete-btn--danger"
                      :disabled="isInlineDeleteBusy(item)"
                      @click.stop="confirmInlineDelete(item)"
                    >
                      {{ t('workspace.fileBrowser.delete') }}
                    </button>
                    <button
                      type="button"
                      class="file-browser__inline-delete-btn"
                      :disabled="isInlineDeleteBusy(item)"
                      @click.stop="cancelInlineDelete(item)"
                    >
                      {{ t('workspace.fileBrowser.cancelDelete') }}
                    </button>
                  </span>
                </span>
              </template>
              <template v-else-if="column.key === 'modified'">
                {{ formatTime(item.lastModified) }}
              </template>
              <template v-else-if="column.key === 'size'">
                {{ isDirectory(item) ? '--' : formatSize(item.size) }}
              </template>
              <template v-else>
                {{ formatType(item) }}
              </template>
            </span>
          </div>
        </div>
      </div>
    </div>

      <div
        v-else
        ref="browserViewportRef"
        class="file-browser__grid"
        :class="{ 'file-browser__drop-root': browserDropActive }"
        tabindex="0"
        @contextmenu.prevent="onViewportContextMenu($event)"
        @keydown="onViewportKeydown"
        @dragover="onBrowserDragOver"
        @dragleave="onBrowserDragLeave"
        @drop="onBrowserDrop"
      >
        <div
          v-if="inlineCreateActive"
          class="file-browser__tile file-browser__tile--inline-create"
          @click.stop
          @dblclick.stop
          @contextmenu.prevent.stop="clearNativeSelection"
        >
          <span class="file-browser__tile-icon">
            <component :is="resolveFileIcon({ name: inlineCreateName || 'folder', type: DIRECTORY_ENTRY_TYPE })" />
          </span>
          <input
            ref="inlineCreateInputRef"
            v-model="inlineCreateName"
            class="file-browser__inline-create-input file-browser__inline-create-input--tile"
            :placeholder="t('workspace.fileBrowser.newFolderPlaceholder')"
            :disabled="inlineCreateBusy"
            @keydown.enter.prevent="commitInlineCreate"
            @keydown.esc.prevent="cancelInlineCreate"
            @blur="commitInlineCreate"
          >
        </div>
        <div v-if="searchedItems.length === 0 && !inlineCreateActive" class="file-browser__viewport-empty file-browser__viewport-empty--grid">
          <i-mdi-folder-outline class="file-browser__empty-icon" />
          <span>{{ localSearchQuery ? t('workspace.fileBrowser.localSearchNoMatches') : t('workspace.fileBrowser.empty') }}</span>
        </div>
        <div
          v-for="item in searchedItems"
        :key="normalizeEntryPath(item.path)"
        role="button"
        tabindex="-1"
        class="file-browser__tile"
        :class="{
          'file-browser__tile--selected': isSelected(item),
          'file-browser__tile--active': activePath === normalizeEntryPath(item.path),
          'file-browser__tile--delete-pending': isInlineDeletePending(item),
          'file-browser__tile--drop-target': isDirectory(item) && directoryDropPath === normalizeEntryPath(item.path),
        }"
        :data-entry-path="normalizeEntryPath(item.path)"
        draggable="true"
        @mousedown.left="onItemMouseDown(item, $event)"
        @click="onItemClick(item, $event)"
        @contextmenu.prevent.stop="onItemContextMenu(item, $event)"
        @dragenter="onDirectoryDragEnter(item, $event)"
        @dragover="onDirectoryDragOver(item, $event)"
        @dragleave="onDirectoryDragLeave(item, $event)"
        @drop="onDirectoryDrop(item, $event)"
        @dragstart="onItemDragStart(item, $event)"
        @dragend="onItemDragEnd"
        @dblclick.stop.prevent="onItemDoubleClick(item)"
      >
        <span class="file-browser__tile-icon">
          <component :is="resolveFileIcon(item, { opened: isDirectory(item) })" />
        </span>
        <span class="file-browser__tile-name-wrap">
          <span class="file-browser__tile-name">{{ item.name }}</span>
          <span
            v-if="isInlineDeletePending(item)"
            class="file-browser__inline-delete-actions file-browser__inline-delete-actions--tile"
            @mousedown.stop
            @click.stop
            @dblclick.stop
          >
            <button
              type="button"
              class="file-browser__inline-delete-btn file-browser__inline-delete-btn--danger"
              :disabled="isInlineDeleteBusy(item)"
              @click.stop="confirmInlineDelete(item)"
            >
              {{ t('workspace.fileBrowser.delete') }}
            </button>
            <button
              type="button"
              class="file-browser__inline-delete-btn"
              :disabled="isInlineDeleteBusy(item)"
              @click.stop="cancelInlineDelete(item)"
            >
              {{ t('workspace.fileBrowser.cancelDelete') }}
            </button>
          </span>
        </span>
        <span class="file-browser__tile-meta">{{ isDirectory(item) ? t('workspace.fileBrowser.directory') : formatSize(item.size) }}</span>
        </div>
      </div>
    </div>

    <teleport to="body">
      <div
        v-if="contextMenu.visible"
        ref="contextMenuRef"
        class="file-browser__context-menu"
        :style="contextMenuStyle"
        @contextmenu.prevent.stop
      >
        <template v-if="contextMenuScope === 'item' && contextMenuItem">
          <button type="button" class="file-browser__context-menu-item" @click="executeContextMenuAction('open')">
            {{ t('workspace.fileBrowser.open') }}
          </button>
          <button type="button" class="file-browser__context-menu-item" @click="executeContextMenuAction('copy-path')">
            {{ t('workspace.fileBrowser.copyPath') }}
          </button>
          <button type="button" class="file-browser__context-menu-item" @click="executeContextMenuAction('download')">
            {{ t('workspace.fileBrowser.download') }}
          </button>
          <button type="button" class="file-browser__context-menu-item" @click="executeContextMenuAction('download-temp')">
            {{ t('workspace.fileBrowser.downloadTemp') }}
          </button>
          <button
            v-if="!isDirectory(contextMenuItem)"
            type="button"
            class="file-browser__context-menu-item"
            @click="executeContextMenuAction('save-as')"
          >
            {{ t('workspace.fileBrowser.saveAs') }}
          </button>
          <button type="button" class="file-browser__context-menu-item" @click="executeContextMenuAction('rename')">
            {{ t('workspace.fileBrowser.rename') }}
          </button>
          <div class="file-browser__context-menu-separator" />
          <button type="button" class="file-browser__context-menu-item file-browser__context-menu-item--danger" @click="executeContextMenuAction('delete')">
            <span>{{ t('workspace.fileBrowser.delete') }}</span>
            <span class="file-browser__context-menu-shortcut">⌫</span>
          </button>
        </template>
        <template v-else>
          <button type="button" class="file-browser__context-menu-item" @click="executeBlankContextMenuAction('new-folder')">
            {{ t('workspace.fileBrowser.newFolder') }}
          </button>
          <button type="button" class="file-browser__context-menu-item" @click="executeBlankContextMenuAction('upload-file')">
            {{ t('workspace.fileBrowser.upload') }}
          </button>
        </template>
      </div>
    </teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { CreateConnectionDirectory, DeleteConnectionEntry, DownloadConnectionFile, DownloadConnectionFileToTemp, ListConnectionDirectory, OpenConnectionFile, PickUploadFile, RenameConnectionEntry, SaveConnectionFileAs } from '../../../wailsjs/go/main/App'
import { folder } from '../../../wailsjs/go/models'
import { ClipboardSetText } from '../../../wailsjs/runtime/runtime'
import { splitPanePanelKey } from '@/components/SplitPane/types'
import { buildConnectionEntryDragPayload, buildConnectionEntryDropFeedback, canDropConnectionEntries, executeConnectionEntryDrop, labelForDropTarget, resolveConnectionEntryDragPayload } from '@/composables/connectionEntryDrop'
import { PANEL_OS_FILE_DROP_EVENT, type PanelOSFileDropDetail } from '@/composables/useFileDrop'
import { CONNECTION_CONFIG_REFRESH_EVENT, type ConnectionConfigRefreshDetail } from '@/composables/useConnectionConfigRefresh'
import { normalizeRemotePath } from '@/composables/remotePath'
import { CONNECTION_DIRECTORY_REFRESH_EVENT, emitConnectionDirectoryRefresh, type ConnectionDirectoryRefreshDetail } from '@/composables/useConnectionDirectoryRefresh'
import { CONNECTION_ENTRY_DROP_LIFECYCLE_EVENT, consumeLatestSuccessfulConnectionEntryDrop, type ConnectionEntryDropLifecycleDetail } from '@/composables/useConnectionEntryDropLifecycle'
import { DIRECTORY_ENTRY_TYPE, resolveFileIcon } from '@/composables/useFileIcons'
import { clearActiveInternalDrag, clearActiveInternalDragSoon, markInternalDragDataTransfer, setActiveInternalDrag } from '@/composables/splitPaneDragState'
import { useConnectionsStore } from '@/stores/connections'
import { useWorkspaceStore } from '@/stores/workspace'
import { buildInlineDeletePaths, removeInlineDeletePath } from './inlineDelete'
import {
  applyNavigationHistory,
  createNavigationHistory,
  moveNavigationHistoryIndex,
  navigationTarget,
  type NavigationHistoryMode,
  type NavigationHistoryState,
} from './navigationHistory'

const props = defineProps<{
  connectionId: string
}>()

type ListColumnKey = 'name' | 'modified' | 'size' | 'type'
type ContextMenuScope = 'item' | 'blank'

interface ListColumnDefinition {
  key: ListColumnKey
  label: string
  minWidth: number
  defaultWidth: number
  align: 'left' | 'right'
}

interface ContextMenuState {
  visible: boolean
  x: number
  y: number
  scope: ContextMenuScope
  item: folder.FileInfo | null
}

const LIST_COLUMNS_STORAGE_KEY = 'workspace:file-browser:list-columns'
const LIST_SORT_STORAGE_KEY = 'workspace:file-browser:list-sort'
const LIST_WIDTHS_STORAGE_KEY = 'workspace:file-browser:list-widths'
const LIST_COLUMN_DEFAULT_WIDTHS: Record<ListColumnKey, number> = {
  name: 320,
  modified: 168,
  size: 110,
  type: 120,
}
const LIST_COLUMN_MIN_WIDTHS: Record<ListColumnKey, number> = {
  name: 220,
  modified: 132,
  size: 88,
  type: 96,
}
const DOUBLE_CLICK_DRAG_SUPPRESS_MS = 420
const SINGLE_CLICK_SELECT_DELAY_MS = 180
const directoryRefreshOrigin = `file-browser:${Math.random().toString(36).slice(2)}`

const { t, locale } = useI18n()
const connections = useConnectionsStore()
const workspace = useWorkspaceStore()
const splitPanePanel = inject(splitPanePanelKey, null)

const loading = ref(false)
const currentPath = ref('')
const items = ref<folder.FileInfo[]>([])
const viewMode = ref<'list' | 'grid'>('list')
const visibleColumnKeys = ref<ListColumnKey[]>(['name', 'modified', 'size'])
const sortKey = ref<ListColumnKey>('name')
const sortDirection = ref<'asc' | 'desc'>('asc')
const fileBrowserRef = ref<HTMLElement | null>(null)
const browserViewportRef = ref<HTMLElement | null>(null)
const contextMenuRef = ref<HTMLElement | null>(null)
const selectedPaths = ref<string[]>([])
const selectionAnchorPath = ref<string | null>(null)
const activePath = ref<string | null>(null)
const columnWidths = ref<Record<ListColumnKey, number>>({ ...LIST_COLUMN_DEFAULT_WIDTHS })
const browserDropActive = ref(false)
const directoryDropPath = ref<string | null>(null)
const dropBusy = ref(false)
const dropBusyOperations = ref<string[]>([])
const pendingHiddenPaths = ref<string[]>([])
const contextMenu = ref<ContextMenuState>({
  visible: false,
  x: 0,
  y: 0,
  scope: 'blank',
  item: null,
})
const dragSession = ref<{
  sourceConnectionId: string
  sourceViewDir: string
  sourcePaths: string[]
} | null>(null)
const dragPreviewEl = ref<HTMLElement | null>(null)
const suppressedDrag = ref<{
  path: string
  expiresAt: number
} | null>(null)
const resizingColumn = ref<ListColumnKey | null>(null)
const pendingSingleClickTimer = ref<ReturnType<typeof window.setTimeout> | null>(null)
const inlineCreateActive = ref(false)
const inlineCreateName = ref('')
const inlineCreateBusy = ref(false)
const inlineCreateInputRef = ref<HTMLInputElement | null>(null)
const localSearchOpen = ref(false)
const localSearchDraft = ref('')
const localSearchQuery = ref('')
const localSearchCaseSensitive = ref(false)
const localSearchWholeText = ref(false)
const localSearchRegex = ref(false)
const localSearchInputRef = ref<HTMLInputElement | null>(null)
const inlineDeletePaths = ref<string[]>([])
const inlineDeleteBusyPath = ref<string | null>(null)
const navigationHistory = ref<NavigationHistoryState>(createNavigationHistory(''))

const connectionId = computed(() => props.connectionId)
const tabId = computed(() => `connection:${connectionId.value}`)
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
const allColumns = computed<ListColumnDefinition[]>(() => [
  {
    key: 'name',
    label: t('workspace.fileBrowser.columnName'),
    minWidth: LIST_COLUMN_MIN_WIDTHS.name,
    defaultWidth: LIST_COLUMN_DEFAULT_WIDTHS.name,
    align: 'left',
  },
  {
    key: 'modified',
    label: t('workspace.fileBrowser.columnModified'),
    minWidth: LIST_COLUMN_MIN_WIDTHS.modified,
    defaultWidth: LIST_COLUMN_DEFAULT_WIDTHS.modified,
    align: 'left',
  },
  {
    key: 'size',
    label: t('workspace.fileBrowser.columnSize'),
    minWidth: LIST_COLUMN_MIN_WIDTHS.size,
    defaultWidth: LIST_COLUMN_DEFAULT_WIDTHS.size,
    align: 'right',
  },
  {
    key: 'type',
    label: t('workspace.fileBrowser.columnType'),
    minWidth: LIST_COLUMN_MIN_WIDTHS.type,
    defaultWidth: LIST_COLUMN_DEFAULT_WIDTHS.type,
    align: 'left',
  },
])
const optionalColumns = computed(() => allColumns.value.filter((column) => column.key !== 'name'))
const visibleColumnSet = computed(() => new Set(visibleColumnKeys.value))
const visibleColumns = computed(() => allColumns.value.filter((column) => visibleColumnSet.value.has(column.key)))
const listGridStyle = computed(() => ({
  gridTemplateColumns: visibleColumns.value
    .map((column) => resolveColumnWidth(column))
    .join(' '),
}))
const selectedPathSet = computed(() => new Set(selectedPaths.value))
const inlineDeletePathSet = computed(() => new Set(inlineDeletePaths.value))
const pendingHiddenPathSet = computed(() => new Set(pendingHiddenPaths.value))
const visibleItems = computed(() =>
  normalizeFolderItems(items.value).filter((item) => !pendingHiddenPathSet.value.has(normalizeEntryPath(item.path))),
)
const sortedItems = computed(() => {
  const collator = new Intl.Collator(locale.value, {
    numeric: true,
    sensitivity: 'base',
  })

  return [...visibleItems.value].sort((left, right) => {
    const leftDir = isDirectory(left)
    const rightDir = isDirectory(right)
    if (leftDir !== rightDir) return leftDir ? -1 : 1

    const direction = sortDirection.value === 'asc' ? 1 : -1
    let result = 0

    if (sortKey.value === 'modified') {
      result = timestampOf(left.lastModified) - timestampOf(right.lastModified)
    } else if (sortKey.value === 'size') {
      result = (left.size || 0) - (right.size || 0)
    } else if (sortKey.value === 'type') {
      result = collator.compare(formatType(left), formatType(right))
    } else {
      result = collator.compare(left.name || left.path, right.name || right.path)
    }

    if (result === 0) {
      result = collator.compare(left.name || left.path, right.name || right.path)
    }

    return result * direction
  })
})
const localSearchPlan = computed(() => {
  const query = localSearchQuery.value.trim()
  if (!query) {
    return {
      error: '',
      matcher: null as ((item: folder.FileInfo) => boolean) | null,
    }
  }

  if (localSearchRegex.value) {
    try {
      const pattern = localSearchWholeText.value ? `^(?:${query})$` : query
      const regex = new RegExp(pattern, localSearchCaseSensitive.value ? 'u' : 'iu')
      return {
        error: '',
        matcher: (item: folder.FileInfo) => regex.test(localSearchText(item)),
      }
    } catch {
      return {
        error: t('workspace.fileBrowser.localSearchInvalidRegex'),
        matcher: null,
      }
    }
  }

  const needle = normalizeLocalSearchText(query)
  return {
    error: '',
    matcher: (item: folder.FileInfo) => {
      const haystack = normalizeLocalSearchText(localSearchText(item))
      return localSearchWholeText.value ? haystack === needle : haystack.includes(needle)
    },
  }
})
const localSearchError = computed(() => localSearchPlan.value.error)
const searchedItems = computed(() => {
  const matcher = localSearchPlan.value.matcher
  return matcher ? sortedItems.value.filter(matcher) : sortedItems.value
})
const localSearchCountLabel = computed(() => {
  if (!localSearchQuery.value.trim()) return ''
  return t('workspace.fileBrowser.localSearchCount', {
    count: searchedItems.value.length,
    total: sortedItems.value.length,
  })
})
const selectedItems = computed(() =>
  searchedItems.value.filter((item) => selectedPathSet.value.has(normalizeEntryPath(item.path))),
)
const contextMenuItem = computed(() => contextMenu.value.item)
const contextMenuScope = computed(() => contextMenu.value.scope)
const contextMenuStyle = computed(() => ({
  left: `${contextMenu.value.x}px`,
  top: `${contextMenu.value.y}px`,
}))

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

function formatType(item: folder.FileInfo) {
  if (isDirectory(item)) {
    return t('workspace.fileBrowser.directory')
  }

  const source = item.name || item.path || ''
  const ext = source.includes('.') ? source.split('.').pop()?.trim() : ''
  if (!ext) {
    return t('workspace.fileBrowser.file')
  }
  return ext.toUpperCase()
}

function timestampOf(value?: any) {
  if (!value) return 0
  const timestamp = new Date(value).getTime()
  return Number.isNaN(timestamp) ? 0 : timestamp
}

function localSearchText(item: folder.FileInfo) {
  return item.name || item.path || ''
}

function normalizeLocalSearchText(value: string) {
  return localSearchCaseSensitive.value ? value : value.toLocaleLowerCase(locale.value)
}

function normalizeEntryPath(value: string) {
  return normalizeRemotePath(value)
}

function normalizeFolderItems(input: folder.FileInfo[] | null | undefined): folder.FileInfo[] {
  return Array.isArray(input) ? input : []
}

function normalizeVisibleColumns(input: unknown): ListColumnKey[] {
  const knownKeys = new Set<ListColumnKey>(['name', 'modified', 'size', 'type'])
  const keys = Array.isArray(input)
    ? input.filter((value): value is ListColumnKey => typeof value === 'string' && knownKeys.has(value as ListColumnKey))
    : []

  const merged = new Set<ListColumnKey>(['name', ...keys])
  return ['name', 'modified', 'size', 'type'].filter((key): key is ListColumnKey => merged.has(key as ListColumnKey))
}

function normalizeColumnWidths(input: unknown): Record<ListColumnKey, number> {
  const next = { ...LIST_COLUMN_DEFAULT_WIDTHS }
  if (!input || typeof input !== 'object') {
    return next
  }

  for (const key of Object.keys(LIST_COLUMN_DEFAULT_WIDTHS) as ListColumnKey[]) {
    const value = (input as Record<string, unknown>)[key]
    if (typeof value !== 'number' || !Number.isFinite(value)) continue
    next[key] = Math.max(LIST_COLUMN_MIN_WIDTHS[key], Math.round(value))
  }

  return next
}

function resolveColumnWidth(column: ListColumnDefinition) {
  const width = resolveColumnPixelWidth(column.key)
  return `${width}px`
}

function resolveColumnPixelWidth(key: ListColumnKey) {
  const column = allColumns.value.find((item) => item.key === key)
  const defaultWidth = column?.defaultWidth ?? LIST_COLUMN_DEFAULT_WIDTHS[key]
  const minWidth = column?.minWidth ?? LIST_COLUMN_MIN_WIDTHS[key]
  return Math.max(columnWidths.value[key] ?? defaultWidth, minWidth)
}

function canResizeColumn(index: number) {
  return index >= 0 && index < visibleColumns.value.length - 1
}

function setColumnVisible(key: ListColumnKey, visible: boolean) {
  if (key === 'name') {
    visibleColumnKeys.value = normalizeVisibleColumns(visibleColumnKeys.value)
    return
  }

  const next = new Set(visibleColumnKeys.value)
  if (visible) {
    next.add(key)
  } else {
    next.delete(key)
  }
  visibleColumnKeys.value = normalizeVisibleColumns([...next])
}

function toggleSort(nextKey: ListColumnKey) {
  if (sortKey.value === nextKey) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
    return
  }
  sortKey.value = nextKey
  sortDirection.value = nextKey === 'modified' || nextKey === 'size' ? 'desc' : 'asc'
}

function onHeaderClick(nextKey: ListColumnKey, event: MouseEvent) {
  if (resizingColumn.value) {
    event.preventDefault()
    event.stopPropagation()
    return
  }
  toggleSort(nextKey)
}

function sortIndicator(key: ListColumnKey) {
  if (sortKey.value !== key) return '↕'
  return sortDirection.value === 'asc' ? '↑' : '↓'
}

function orderedPaths() {
  return searchedItems.value.map((item) => normalizeEntryPath(item.path))
}

function findItemByPath(path: string | null | undefined) {
  if (!path) return null
  const normalized = normalizeEntryPath(path)
  return searchedItems.value.find((item) => normalizeEntryPath(item.path) === normalized) ?? null
}

function focusBrowserViewport() {
  browserViewportRef.value?.focus()
}

function parentPath(path: string) {
  const segments = normalizeEntryPath(path).split('/').filter(Boolean)
  if (segments.length === 0) return ''
  return segments.slice(0, -1).join('/')
}

function recordNavigationPath(path: string, mode: NavigationHistoryMode) {
  navigationHistory.value = applyNavigationHistory(navigationHistory.value, normalizeEntryPath(path), mode)
}

function isEditableShortcutTarget(target: EventTarget | null) {
  if (!(target instanceof HTMLElement)) return false
  const tag = target.tagName
  return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || target.isContentEditable
}

function isMacPlatform() {
  if (typeof navigator === 'undefined') return false
  return /mac/i.test(navigator.platform)
}

function historyShortcutDelta(event: KeyboardEvent): -1 | 1 | null {
  const key = event.key.toLowerCase()
  const macBracketBack = isMacPlatform()
    && event.metaKey
    && !event.ctrlKey
    && !event.altKey
    && !event.shiftKey
    && key === '['
  const macBracketForward = isMacPlatform()
    && event.metaKey
    && !event.ctrlKey
    && !event.altKey
    && !event.shiftKey
    && key === ']'
  const altBack = event.altKey
    && !event.ctrlKey
    && !event.metaKey
    && !event.shiftKey
    && event.key === 'ArrowLeft'
  const altForward = event.altKey
    && !event.ctrlKey
    && !event.metaKey
    && !event.shiftKey
    && event.key === 'ArrowRight'

  if (macBracketBack || altBack) return -1
  if (macBracketForward || altForward) return 1
  return null
}

function isLocalSearchShortcut(event: KeyboardEvent) {
  return (event.metaKey || event.ctrlKey)
    && !event.shiftKey
    && !event.altKey
    && event.key.toLowerCase() === 'f'
}

function isFileBrowserShortcutScope(target: EventTarget | null) {
  if (target instanceof Node && fileBrowserRef.value?.contains(target)) {
    return true
  }

  const activeInWorkspace = workspace.isTabActiveInActiveGroup(tabId.value)
  if (!activeInWorkspace) {
    return false
  }

  if (target === document || target === document.body || target === document.documentElement) {
    return true
  }

  if (target instanceof HTMLElement) {
    return Boolean(target.closest('.shell-center'))
  }

  return false
}

function onFileBrowserShortcutKeydown(event: KeyboardEvent) {
  if (!isFileBrowserShortcutScope(event.target)) return

  if (isLocalSearchShortcut(event)) {
    event.preventDefault()
    event.stopPropagation()
    void toggleLocalSearch()
    return
  }

  if (event.key === 'Escape' && localSearchOpen.value) {
    event.preventDefault()
    event.stopPropagation()
    closeLocalSearch()
    return
  }

  if (isEditableShortcutTarget(event.target) || loading.value) return

  const delta = historyShortcutDelta(event)
  if (!delta) return
  if (navigationTarget(navigationHistory.value, delta) === null) return

  event.preventDefault()
  event.stopPropagation()
  void navigateHistory(delta)
}

async function openLocalSearch() {
  localSearchOpen.value = true
  localSearchDraft.value = localSearchQuery.value
  await nextTick()
  localSearchInputRef.value?.focus()
  localSearchInputRef.value?.select()
}

async function toggleLocalSearch() {
  if (localSearchOpen.value) {
    closeLocalSearch()
    focusBrowserViewport()
    return
  }
  await openLocalSearch()
}

function commitLocalSearch() {
  localSearchQuery.value = localSearchDraft.value.trim()
}

function closeLocalSearch() {
  localSearchOpen.value = false
  localSearchDraft.value = ''
  localSearchQuery.value = ''
}

async function navigateHistory(delta: -1 | 1) {
  const target = navigationTarget(navigationHistory.value, delta)
  if (target === null) return

  const loaded = await load(target, { history: 'none' })
  if (!loaded) return
  navigationHistory.value = moveNavigationHistoryIndex(navigationHistory.value, delta)
}

async function scrollPathIntoView(path: string | null) {
  if (!path) return
  await nextTick()
  const escapedPath = typeof CSS !== 'undefined' && typeof CSS.escape === 'function'
    ? CSS.escape(path)
    : path.replace(/["\\]/g, '\\$&')
  const target = browserViewportRef.value?.querySelector<HTMLElement>(`[data-entry-path="${escapedPath}"]`)
  target?.scrollIntoView({
    block: 'nearest',
    inline: 'nearest',
  })
}

function revealLoadedPath(path: string | null | undefined) {
  const normalized = normalizeEntryPath(path ?? '')
  if (!normalized) return

  const item = normalizeFolderItems(items.value).find((entry) => normalizeEntryPath(entry.path) === normalized)
  if (!item) return

  const itemPath = normalizeEntryPath(item.path)
  selectedPaths.value = [itemPath]
  selectionAnchorPath.value = itemPath
  activePath.value = itemPath
  void scrollPathIntoView(itemPath)
}

async function clampViewportScrollToContent() {
  await nextTick()
  const viewport = browserViewportRef.value
  if (!viewport) return
  const maxScrollTop = Math.max(0, viewport.scrollHeight - viewport.clientHeight)
  if (viewport.scrollTop > maxScrollTop) {
    viewport.scrollTop = maxScrollTop
  }
}

function isSelected(item: folder.FileInfo) {
  return selectedPathSet.value.has(normalizeEntryPath(item.path))
}

function isInlineDeletePending(item: folder.FileInfo) {
  return inlineDeletePathSet.value.has(normalizeEntryPath(item.path))
}

function isInlineDeleteBusy(item: folder.FileInfo) {
  return inlineDeleteBusyPath.value === normalizeEntryPath(item.path)
}

function clearSelection() {
  selectedPaths.value = []
  selectionAnchorPath.value = null
  activePath.value = null
  closeContextMenu()
  focusBrowserViewport()
}

function clearNativeSelection() {
  window.getSelection()?.removeAllRanges()
}

function onFileBrowserSelectStart(event: Event) {
  const target = event.target
  if (
    target instanceof HTMLInputElement
    || target instanceof HTMLTextAreaElement
    || (target instanceof HTMLElement && target.isContentEditable)
  ) {
    return
  }
  event.preventDefault()
  clearNativeSelection()
}

function selectSingle(item: folder.FileInfo) {
  cancelInlineDelete()
  const path = normalizeEntryPath(item.path)
  selectedPaths.value = [path]
  selectionAnchorPath.value = path
  activePath.value = path
  focusBrowserViewport()
}

function toggleSelection(targetPath: string) {
  cancelInlineDelete()
  const next = new Set(selectedPaths.value)
  if (next.has(targetPath)) {
    next.delete(targetPath)
  } else {
    next.add(targetPath)
  }

  selectedPaths.value = orderedPaths().filter((path) => next.has(path))
  selectionAnchorPath.value = targetPath
  activePath.value = targetPath
  focusBrowserViewport()
}

function toggleClickSelection(item: folder.FileInfo) {
  const path = normalizeEntryPath(item.path)
  if (!selectedPathSet.value.has(path)) {
    selectSingle(item)
    return
  }

  selectedPaths.value = selectedPaths.value.filter((itemPath) => itemPath !== path)
  if (activePath.value === path) {
    activePath.value = selectedPaths.value[0] ?? null
  }
  if (selectionAnchorPath.value === path) {
    selectionAnchorPath.value = selectedPaths.value[0] ?? null
  }
  focusBrowserViewport()
}

function clearPendingSingleClick() {
  if (!pendingSingleClickTimer.value) return
  window.clearTimeout(pendingSingleClickTimer.value)
  pendingSingleClickTimer.value = null
}

function selectRange(targetPath: string, additive = false) {
  cancelInlineDelete()
  const paths = orderedPaths()
  if (paths.length === 0) return

  const anchor = selectionAnchorPath.value ?? activePath.value ?? targetPath
  const anchorIndex = paths.indexOf(anchor)
  const targetIndex = paths.indexOf(targetPath)
  if (anchorIndex === -1 || targetIndex === -1) {
    const item = findItemByPath(targetPath)
    if (item) selectSingle(item)
    return
  }

  const start = Math.min(anchorIndex, targetIndex)
  const end = Math.max(anchorIndex, targetIndex)
  const range = paths.slice(start, end + 1)
  const next = additive ? new Set(selectedPaths.value) : new Set<string>()
  range.forEach((path) => next.add(path))
  selectedPaths.value = paths.filter((path) => next.has(path))
  activePath.value = targetPath
  if (!selectionAnchorPath.value) {
    selectionAnchorPath.value = anchor
  }
  focusBrowserViewport()
}

function onItemClick(item: folder.FileInfo, event: MouseEvent) {
  clearPendingSingleClick()
  const path = normalizeEntryPath(item.path)
  if (event.shiftKey) {
    selectRange(path, event.metaKey || event.ctrlKey)
    return
  }

  if (event.metaKey || event.ctrlKey) {
    toggleSelection(path)
    return
  }

  pendingSingleClickTimer.value = window.setTimeout(() => {
    pendingSingleClickTimer.value = null
    toggleClickSelection(item)
  }, SINGLE_CLICK_SELECT_DELAY_MS)
}

function onItemDoubleClick(item: folder.FileInfo) {
  clearPendingSingleClick()
  return openItem(item)
}

function onItemMouseDown(item: folder.FileInfo, event: MouseEvent) {
  if (event.button !== 0) return

  const current = suppressedDrag.value
  if (current && Date.now() > current.expiresAt) {
    suppressedDrag.value = null
  }

  if (event.detail >= 2) {
    event.preventDefault()
    clearNativeSelection()
    suppressedDrag.value = {
      path: normalizeEntryPath(item.path),
      expiresAt: Date.now() + DOUBLE_CLICK_DRAG_SUPPRESS_MS,
    }
  }
}

function moveActive(offset: number, extendSelection = false) {
  const paths = orderedPaths()
  if (paths.length === 0) return

  const currentIndex = activePath.value ? paths.indexOf(activePath.value) : -1
  const fallbackIndex = offset >= 0 ? 0 : paths.length - 1
  const nextIndex = currentIndex === -1
    ? fallbackIndex
    : Math.min(Math.max(currentIndex + offset, 0), paths.length - 1)
  const nextPath = paths[nextIndex]!
  const nextItem = findItemByPath(nextPath)
  if (!nextItem) return

  if (extendSelection) {
    if (!selectionAnchorPath.value) {
      selectionAnchorPath.value = activePath.value ?? nextPath
    }
    selectRange(nextPath)
    return
  }

  selectSingle(nextItem)
}

function moveActiveToBoundary(position: 'start' | 'end', extendSelection = false) {
  const paths = orderedPaths()
  if (paths.length === 0) return
  const nextPath = position === 'start' ? paths[0] : paths[paths.length - 1]
  if (!nextPath) return

  if (extendSelection) {
    if (!selectionAnchorPath.value) {
      selectionAnchorPath.value = activePath.value ?? nextPath
    }
    selectRange(nextPath)
    return
  }

  const nextItem = findItemByPath(nextPath)
  if (nextItem) {
    selectSingle(nextItem)
  }
}

function clearDropIndicators() {
  browserDropActive.value = false
  directoryDropPath.value = null
}

function removeDragPreview() {
  dragPreviewEl.value?.remove()
  dragPreviewEl.value = null
}

function buildDragPreviewNodes(dragItems: folder.FileInfo[]) {
  const sourceItems = dragItems.slice(0, Math.min(dragItems.length, 3))
  const nodes = sourceItems
    .map((entry) => {
      const path = normalizeEntryPath(entry.path)
      const escapedPath = typeof CSS !== 'undefined' && typeof CSS.escape === 'function'
        ? CSS.escape(path)
        : path.replace(/["\\]/g, '\\$&')
      return browserViewportRef.value?.querySelector<HTMLElement>(`[data-entry-path="${escapedPath}"]`) ?? null
    })
    .filter((node): node is HTMLElement => node instanceof HTMLElement)

  if (nodes.length === 0) return null

  const preview = document.createElement('div')
  preview.style.position = 'fixed'
  preview.style.left = '-9999px'
  preview.style.top = '-9999px'
  preview.style.pointerEvents = 'none'
  preview.style.zIndex = '9999'
  preview.style.width = viewMode.value === 'list' ? '320px' : '168px'
  preview.style.height = viewMode.value === 'list'
    ? `${56 + (nodes.length - 1) * 10}px`
    : `${152 + (nodes.length - 1) * 12}px`

  nodes.forEach((node, index) => {
    const clone = node.cloneNode(true) as HTMLElement
    clone.style.position = 'absolute'
    clone.style.left = `${index * 10}px`
    clone.style.top = `${index * 10}px`
    clone.style.width = viewMode.value === 'list' ? '300px' : '152px'
    clone.style.margin = '0'
    clone.style.pointerEvents = 'none'
    clone.style.boxShadow = '0 12px 24px rgba(0, 0, 0, 0.22)'
    clone.style.transform = 'none'
    clone.style.opacity = `${1 - index * 0.12}`
    clone.classList.remove(
      'file-browser__row--active',
      'file-browser__row--drop-target',
      'file-browser__tile--active',
      'file-browser__tile--drop-target',
    )
    preview.appendChild(clone)
  })

  if (dragItems.length > 1) {
    const badge = document.createElement('div')
    badge.textContent = String(dragItems.length)
    badge.style.position = 'absolute'
    badge.style.right = '0'
    badge.style.top = '0'
    badge.style.minWidth = '24px'
    badge.style.height = '24px'
    badge.style.padding = '0 7px'
    badge.style.borderRadius = '999px'
    badge.style.display = 'flex'
    badge.style.alignItems = 'center'
    badge.style.justifyContent = 'center'
    badge.style.background = 'var(--theme-color-primary)'
    badge.style.color = 'var(--theme-color-on-primary, #fff)'
    badge.style.fontSize = '12px'
    badge.style.fontWeight = '700'
    badge.style.boxShadow = '0 8px 18px rgba(0, 0, 0, 0.22)'
    preview.appendChild(badge)
  }

  document.body.appendChild(preview)
  dragPreviewEl.value = preview
  return preview
}

function addBusyOperation(operationId: string) {
  if (dropBusyOperations.value.includes(operationId)) return
  dropBusyOperations.value = [...dropBusyOperations.value, operationId]
  dropBusy.value = dropBusyOperations.value.length > 0
}

function removeBusyOperation(operationId: string) {
  dropBusyOperations.value = dropBusyOperations.value.filter((id) => id !== operationId)
  dropBusy.value = dropBusyOperations.value.length > 0
}

function addPendingHiddenPaths(paths: string[]) {
  const next = new Set(pendingHiddenPaths.value)
  paths.forEach((path) => next.add(normalizeEntryPath(path)))
  pendingHiddenPaths.value = [...next]
}

function removePendingHiddenPaths(paths: string[]) {
  const next = new Set(pendingHiddenPaths.value)
  paths.forEach((path) => next.delete(normalizeEntryPath(path)))
  pendingHiddenPaths.value = [...next]
}

function removeItemsByPath(paths: string[]) {
  const target = new Set(paths.map((path) => normalizeEntryPath(path)))
  items.value = normalizeFolderItems(items.value).filter((item) => !target.has(normalizeEntryPath(item.path)))
  workspace.setConnectionBrowserState(connectionId.value, currentPath.value, items.value)
  void clampViewportScrollToContent()
}

function matchesCurrentDirectory(connection: string, dir: string) {
  return connection === connectionId.value && normalizeEntryPath(dir) === normalizeEntryPath(currentPath.value)
}

function onEntryDropLifecycle(event: Event) {
  const detail = (event as CustomEvent<ConnectionEntryDropLifecycleDetail>).detail
  if (!detail) return

  const affectsSource = detail.mode === 'move'
    && detail.sourceConnectionId === connectionId.value
  const affectsTarget = matchesCurrentDirectory(detail.targetConnectionId, detail.targetDir)
  if (!affectsSource && !affectsTarget) return

  if (detail.phase === 'start') {
    if (affectsSource && detail.mode === 'move') {
      addBusyOperation(detail.operationId)
      addPendingHiddenPaths(detail.sourcePaths)
    }
    if (affectsTarget) {
      addBusyOperation(detail.operationId)
    }
    return
  }

  if (affectsSource || affectsTarget) {
    removeBusyOperation(detail.operationId)
  }
  if (affectsSource && detail.mode === 'move') {
    removePendingHiddenPaths(detail.sourcePaths)
    if (detail.success) {
      removeItemsByPath(detail.sourcePaths)
    }
  }
  if (affectsTarget) {
    void reload()
  }
}

function applyDropEffect(event: DragEvent, sameConnection: boolean) {
  if (!event.dataTransfer) return
  event.dataTransfer.dropEffect = sameConnection ? 'move' : 'copy'
}

function onBrowserDragOver(event: DragEvent) {
  if (resizingColumn.value) {
    event.preventDefault()
    event.stopPropagation()
    return
  }
  const target = event.target as HTMLElement | null
  if (target?.closest?.('[data-entry-path]')) {
    browserDropActive.value = false
    return
  }
  const payload = resolveConnectionEntryDragPayload(event)
  if (!canDropConnectionEntries(payload, connectionId.value, currentPath.value)) {
    browserDropActive.value = false
    return
  }

  event.preventDefault()
  browserDropActive.value = true
  directoryDropPath.value = null
  applyDropEffect(event, payload?.data.sourceConnectionId === connectionId.value)
}

function onBrowserDragLeave(event: DragEvent) {
  if (resizingColumn.value) {
    browserDropActive.value = false
    return
  }
  const relatedTarget = event.relatedTarget
  if (relatedTarget instanceof Node && (event.currentTarget as HTMLElement | null)?.contains(relatedTarget)) {
    return
  }
  browserDropActive.value = false
}

async function onBrowserDrop(event: DragEvent) {
  if (resizingColumn.value) {
    event.preventDefault()
    event.stopPropagation()
    clearDropIndicators()
    return
  }
  const target = event.target as HTMLElement | null
  if (target?.closest?.('[data-entry-path]')) {
    clearDropIndicators()
    return
  }
  const payload = resolveConnectionEntryDragPayload(event)
  clearDropIndicators()
  if (!payload || !canDropConnectionEntries(payload, connectionId.value, currentPath.value)) return

  event.preventDefault()
  event.stopPropagation()
  await executeDrop(payload, currentPath.value)
}

function onDirectoryDragEnter(item: folder.FileInfo, event: DragEvent) {
  if (!isDirectory(item)) return
  const itemPath = normalizeEntryPath(item.path)
  const payload = resolveConnectionEntryDragPayload(event)
  if (!canDropConnectionEntries(payload, connectionId.value, itemPath)) return

  event.preventDefault()
  event.stopPropagation()
  directoryDropPath.value = itemPath
  browserDropActive.value = false
}

function onDirectoryDragOver(item: folder.FileInfo, event: DragEvent) {
  if (!isDirectory(item)) return
  const itemPath = normalizeEntryPath(item.path)
  const payload = resolveConnectionEntryDragPayload(event)
  if (!canDropConnectionEntries(payload, connectionId.value, itemPath)) return

  event.preventDefault()
  event.stopPropagation()
  directoryDropPath.value = itemPath
  browserDropActive.value = false
  applyDropEffect(event, payload?.data.sourceConnectionId === connectionId.value)
}

function onDirectoryDragLeave(item: folder.FileInfo, event: DragEvent) {
  const relatedTarget = event.relatedTarget
  if (relatedTarget instanceof Node && (event.currentTarget as HTMLElement | null)?.contains(relatedTarget)) {
    return
  }
  if (directoryDropPath.value === normalizeEntryPath(item.path)) {
    directoryDropPath.value = null
  }
}

async function onDirectoryDrop(item: folder.FileInfo, event: DragEvent) {
  const itemPath = normalizeEntryPath(item.path)
  const payload = resolveConnectionEntryDragPayload(event)
  clearDropIndicators()
  if (!isDirectory(item) || !payload || !canDropConnectionEntries(payload, connectionId.value, itemPath)) return

  event.preventDefault()
  event.stopPropagation()
  await executeDrop(payload, itemPath)
}

async function executeDrop(
  payload: ReturnType<typeof resolveConnectionEntryDragPayload>,
  targetDir: string,
) {
  if (!payload) return

  try {
    const result = await executeConnectionEntryDrop(payload, connectionId.value, targetDir)
    const feedback = buildConnectionEntryDropFeedback(
      result,
      labelForDropTarget(targetDir, definition.value?.name ?? connectionId.value),
    )
    if (result.mode !== 'noop') {
      ElMessage.success(t(feedback.key, feedback.params))
    }

    if (result.mode === 'transfer' && normalizeEntryPath(targetDir) === normalizeEntryPath(currentPath.value)) {
      await reload()
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  }
}

async function downloadSelected() {
  const snapshot = [...selectedItems.value]
  if (snapshot.length === 0) return

  try {
    let taskCount = 0
    for (const item of snapshot) {
      const taskIds = await DownloadConnectionFile(connectionId.value, item.path)
      taskCount += taskIds.length
    }

    ElMessage.success(
      taskCount > 0
        ? t('workspace.fileBrowser.downloadQueuedBatch', { count: taskCount })
        : t('workspace.fileBrowser.downloadSelectedPrepared', { count: snapshot.length }),
    )
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  }
}

async function copySelectedPaths() {
  const snapshot = [...selectedItems.value]
  if (snapshot.length === 0) return

  try {
    await ClipboardSetText(snapshot.map((item) => item.path).join('\n'))
    ElMessage.success(t('workspace.fileBrowser.copySelectedPathsSuccess', { count: snapshot.length }))
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  }
}

async function deleteSelected() {
  const snapshot = [...selectedItems.value]
  if (snapshot.length === 0) return

  beginInlineDelete(snapshot)
}

function onViewportKeydown(event: KeyboardEvent) {
  if (event.target instanceof HTMLInputElement || event.target instanceof HTMLTextAreaElement) return
  if (loading.value) return

  const isMetaSelectAll = (event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'a'
  if (isMetaSelectAll) {
    event.preventDefault()
    selectedPaths.value = orderedPaths()
    selectionAnchorPath.value = activePath.value ?? selectedPaths.value[0] ?? null
    activePath.value = selectedPaths.value[0] ?? null
    return
  }

  if (event.key === 'Escape') {
    event.preventDefault()
    if (localSearchOpen.value) {
      closeLocalSearch()
      return
    }
    if (inlineDeletePaths.value.length > 0) {
      cancelInlineDelete()
    }
    clearSelection()
    return
  }

  if (event.key === 'ArrowDown' || event.key === 'ArrowRight') {
    event.preventDefault()
    moveActive(1, event.shiftKey)
    return
  }

  if (event.key === 'ArrowUp' || event.key === 'ArrowLeft') {
    event.preventDefault()
    moveActive(-1, event.shiftKey)
    return
  }

  if (event.key === 'Home') {
    event.preventDefault()
    moveActiveToBoundary('start', event.shiftKey)
    return
  }

  if (event.key === 'End') {
    event.preventDefault()
    moveActiveToBoundary('end', event.shiftKey)
    return
  }

  if (event.key === 'Enter') {
    const item = findItemByPath(activePath.value)
    if (!item) return
    event.preventDefault()
    void openItem(item)
    return
  }

  if (event.key === 'Delete') {
    return
  }

  if (event.key === 'F2') {
    const item = findItemByPath(activePath.value)
    if (!item) return
    event.preventDefault()
    void renameItem(item)
    return
  }

  if (event.key === 'Backspace') {
    if (selectedItems.value.length > 0) {
      event.preventDefault()
      void deleteSelected()
      return
    }

    if (!currentPath.value) return
    event.preventDefault()
    void openPath(parentPath(currentPath.value))
  }
}

function reconcileSelection(nextItems: folder.FileInfo[]) {
  const available = new Set(nextItems.map((item) => normalizeEntryPath(item.path)))
  selectedPaths.value = selectedPaths.value.filter((path) => available.has(path))

  if (activePath.value && !available.has(activePath.value)) {
    activePath.value = selectedPaths.value[0] ?? null
  }

  if (selectionAnchorPath.value && !available.has(selectionAnchorPath.value)) {
    selectionAnchorPath.value = activePath.value
  }
}

function hydrateListPreferences() {
  try {
    const storedColumns = localStorage.getItem(LIST_COLUMNS_STORAGE_KEY)
    if (storedColumns) {
      visibleColumnKeys.value = normalizeVisibleColumns(JSON.parse(storedColumns))
    }

    const storedSort = localStorage.getItem(LIST_SORT_STORAGE_KEY)
    if (!storedSort) return
    const parsed = JSON.parse(storedSort) as {
      key?: ListColumnKey
      direction?: 'asc' | 'desc'
    }
    if (parsed.key) {
      sortKey.value = parsed.key
    }
    if (parsed.direction === 'asc' || parsed.direction === 'desc') {
      sortDirection.value = parsed.direction
    }

    const storedWidths = localStorage.getItem(LIST_WIDTHS_STORAGE_KEY)
    if (storedWidths) {
      columnWidths.value = normalizeColumnWidths(JSON.parse(storedWidths))
    }
  } catch {
    visibleColumnKeys.value = ['name', 'modified', 'size']
    sortKey.value = 'name'
    sortDirection.value = 'asc'
    columnWidths.value = { ...LIST_COLUMN_DEFAULT_WIDTHS }
  }
}

type ColumnResizeState = {
  leftKey: ListColumnKey
  rightKey: ListColumnKey
  startX: number
  startLeftWidth: number
  startRightWidth: number
}

let activeColumnResize: ColumnResizeState | null = null

function onColumnResizeMove(event: MouseEvent) {
  if (!activeColumnResize) return
  event.preventDefault()
  event.stopPropagation()

  const minLeft = LIST_COLUMN_MIN_WIDTHS[activeColumnResize.leftKey]
  const minRight = LIST_COLUMN_MIN_WIDTHS[activeColumnResize.rightKey]
  const rawDelta = event.clientX - activeColumnResize.startX
  const minDelta = minLeft - activeColumnResize.startLeftWidth
  const maxDelta = activeColumnResize.startRightWidth - minRight
  const delta = Math.max(minDelta, Math.min(rawDelta, maxDelta))

  columnWidths.value = {
    ...columnWidths.value,
    [activeColumnResize.leftKey]: Math.round(activeColumnResize.startLeftWidth + delta),
    [activeColumnResize.rightKey]: Math.round(activeColumnResize.startRightWidth - delta),
  }
}

function stopColumnResize() {
  if (!activeColumnResize) return
  activeColumnResize = null
  resizingColumn.value = null
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
  window.removeEventListener('mousemove', onColumnResizeMove, true)
  window.removeEventListener('mouseup', stopColumnResize, true)
  window.removeEventListener('blur', stopColumnResize)
  document.removeEventListener('visibilitychange', stopColumnResize)
}

function startColumnResize(index: number, event: MouseEvent) {
  const leftColumn = visibleColumns.value[index]
  const rightColumn = visibleColumns.value[index + 1]
  if (!leftColumn || !rightColumn) return

  event.preventDefault()
  event.stopPropagation()
  closeContextMenu()
  clearDropIndicators()
  suppressedDrag.value = null
  clearActiveInternalDrag()
  activeColumnResize = {
    leftKey: leftColumn.key,
    rightKey: rightColumn.key,
    startX: event.clientX,
    startLeftWidth: resolveColumnPixelWidth(leftColumn.key),
    startRightWidth: resolveColumnPixelWidth(rightColumn.key),
  }
  resizingColumn.value = leftColumn.key

  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
  window.addEventListener('mousemove', onColumnResizeMove, true)
  window.addEventListener('mouseup', stopColumnResize, true)
  window.addEventListener('blur', stopColumnResize)
  document.addEventListener('visibilitychange', stopColumnResize)
}

async function load(dir = '', options: { history?: NavigationHistoryMode } = {}) {
  resetTransientInteractionState()
  closeLocalSearch()
  loading.value = true
  try {
    await connections.openConnection(connectionId.value)
    const nextDir = normalizeEntryPath(dir)
    items.value = normalizeFolderItems(await ListConnectionDirectory(connectionId.value, nextDir))
    currentPath.value = nextDir
    workspace.setConnectionPath(connectionId.value, nextDir)
    workspace.setConnectionBrowserState(connectionId.value, nextDir, items.value)
    recordNavigationPath(nextDir, options.history ?? 'push')
    revealLoadedPath(workspace.consumeConnectionRevealPath(connectionId.value))
    return true
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
    return false
  } finally {
    loading.value = false
  }
}

function reload() {
  return load(currentPath.value, { history: 'reset' })
}

function goRoot() {
  return load('')
}

function openPath(path: string) {
  return load(path)
}

function openItem(item: folder.FileInfo) {
  resetTransientInteractionState()
  if (isDirectory(item)) {
    return load(item.path)
  }
  return openFile(item)
}

function onItemDragStart(item: folder.FileInfo, event: DragEvent) {
  closeContextMenu()
  const suppression = suppressedDrag.value
  if (suppression) {
    if (Date.now() > suppression.expiresAt) {
      suppressedDrag.value = null
    } else if (suppression.path === normalizeEntryPath(item.path)) {
      event.preventDefault()
      resetTransientInteractionState()
      return
    }
  }
  if (!event.dataTransfer) return
  markInternalDragDataTransfer(event.dataTransfer)
  const dragItems = isSelected(item) ? selectedItems.value : [item]
  removeDragPreview()
  const preview = buildDragPreviewNodes(dragItems)
  if (preview) {
    const offsetX = viewMode.value === 'list' ? 48 : 36
    const offsetY = 18
    event.dataTransfer.setDragImage(preview, offsetX, offsetY)
  }
  const payload = buildConnectionEntryDragPayload(connectionId.value, currentPath.value, dragItems)
  dragSession.value = {
    sourceConnectionId: connectionId.value,
    sourceViewDir: currentPath.value,
    sourcePaths: payload.data.entries.map((entry) => entry.path),
  }

  setActiveInternalDrag({
    sourcePanelUid: splitPanePanel?.uid ?? -1,
    sourcePanelIndex: splitPanePanel?.index.value ?? -1,
    payload,
  })
}

function onItemDragEnd() {
  const completedDrop = dragSession.value
    ? consumeLatestSuccessfulConnectionEntryDrop(dragSession.value)
    : null
  if (completedDrop) {
    removeItemsByPath(completedDrop.sourcePaths)
    removePendingHiddenPaths(completedDrop.sourcePaths)
  }
  dragSession.value = null
  removeDragPreview()
  suppressedDrag.value = null
  clearDropIndicators()
  clearActiveInternalDragSoon()
}

function resetTransientInteractionState() {
  clearPendingSingleClick()
  cancelInlineCreate()
  cancelInlineDelete()
  closeContextMenu()
  clearDropIndicators()
  dragSession.value = null
  removeDragPreview()
  suppressedDrag.value = null
  clearActiveInternalDrag()
  stopColumnResize()
}

function closeContextMenu() {
  contextMenu.value = {
    visible: false,
    x: 0,
    y: 0,
    scope: 'blank',
    item: null,
  }
}

async function positionContextMenu() {
  await nextTick()
  const menuEl = contextMenuRef.value
  if (!menuEl || !contextMenu.value.visible) return

  const width = menuEl.offsetWidth
  const height = menuEl.offsetHeight
  const margin = 8
  const maxX = window.innerWidth - width - margin
  const maxY = window.innerHeight - height - margin

  contextMenu.value = {
    ...contextMenu.value,
    x: Math.max(margin, Math.min(contextMenu.value.x, maxX)),
    y: Math.max(margin, Math.min(contextMenu.value.y, maxY)),
  }
}

function onItemContextMenu(item: folder.FileInfo, event: MouseEvent) {
  clearNativeSelection()
  clearPendingSingleClick()
  selectSingle(item)
  contextMenu.value = {
    visible: true,
    x: event.clientX,
    y: event.clientY,
    scope: 'item',
    item,
  }
  void positionContextMenu()
}

function onViewportContextMenu(event: MouseEvent) {
  const target = event.target
  if (target instanceof HTMLElement && target.closest('[data-entry-path]')) return

  clearNativeSelection()
  clearPendingSingleClick()
  cancelInlineDelete()
  selectedPaths.value = []
  selectionAnchorPath.value = null
  activePath.value = null
  contextMenu.value = {
    visible: true,
    x: event.clientX,
    y: event.clientY,
    scope: 'blank',
    item: null,
  }
  void positionContextMenu()
}

function executeContextMenuAction(command: string) {
  const item = contextMenuItem.value
  closeContextMenu()
  if (!item) return
  return onItemCommand(command, item)
}

function emitPanelUploadEvent(paths: string[], position?: { x: number; y: number }) {
  if (paths.length === 0) return
  window.dispatchEvent(new CustomEvent<PanelOSFileDropDetail>(PANEL_OS_FILE_DROP_EVENT, {
    detail: {
      groupId: '',
      tabId: tabId.value,
      x: position?.x ?? 0,
      y: position?.y ?? 0,
      paths,
    },
  }))
}

async function uploadFile(position?: { x: number; y: number }) {
  try {
    const localPath = (await PickUploadFile()).trim()
    if (!localPath) return
    emitPanelUploadEvent([localPath], position)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  }
}

function executeBlankContextMenuAction(command: 'new-folder' | 'upload-file') {
  const position = {
    x: contextMenu.value.x,
    y: contextMenu.value.y,
  }
  closeContextMenu()
  if (command === 'new-folder') {
    return createDirectory()
  }
  if (command === 'upload-file') {
    return uploadFile(position)
  }
}

function onWindowPointerDown(event: PointerEvent) {
  if (!contextMenu.value.visible) return
  const target = event.target
  if (target instanceof Node && contextMenuRef.value?.contains(target)) return
  closeContextMenu()
}

function onWindowKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    stopColumnResize()
    closeContextMenu()
  }
}

function onWindowDragComplete(event: DragEvent) {
  const target = event.target
  const eventInsidePanel = target instanceof Node && Boolean(fileBrowserRef.value?.contains(target))
  const activeInWorkspace = workspace.isTabActiveInActiveGroup(tabId.value)
  if (!eventInsidePanel && !activeInWorkspace && !dragSession.value) return

  window.setTimeout(() => {
    resetTransientInteractionState()
  }, 120)
}

async function createDirectory() {
  if (inlineCreateActive.value) {
    await nextTick()
    inlineCreateInputRef.value?.focus()
    inlineCreateInputRef.value?.select()
    return
  }

  clearPendingSingleClick()
  closeContextMenu()
  clearNativeSelection()
  cancelInlineDelete()
  selectedPaths.value = []
  selectionAnchorPath.value = null
  activePath.value = null
  inlineCreateName.value = ''
  inlineCreateActive.value = true
  inlineCreateBusy.value = false

  await nextTick()
  browserViewportRef.value?.scrollTo({ top: 0, left: 0 })
  await nextTick()
  inlineCreateInputRef.value?.focus()
  inlineCreateInputRef.value?.select()
}

function cancelInlineCreate() {
  if (!inlineCreateActive.value && !inlineCreateName.value && !inlineCreateBusy.value) return
  inlineCreateActive.value = false
  inlineCreateName.value = ''
  inlineCreateBusy.value = false
}

async function commitInlineCreate() {
  if (!inlineCreateActive.value || inlineCreateBusy.value) return

  const name = inlineCreateName.value.trim()
  if (!name) {
    cancelInlineCreate()
    return
  }

  inlineCreateBusy.value = true
  try {
    await CreateConnectionDirectory(connectionId.value, currentPath.value, name)
    inlineCreateActive.value = false
    inlineCreateName.value = ''
    await reload()
    emitConnectionDirectoryRefresh({
      connectionId: connectionId.value,
      path: currentPath.value,
      source: 'mutation',
      taskId: 'create-directory',
      origin: directoryRefreshOrigin,
    })
    ElMessage.success(t('workspace.fileBrowser.newFolderSuccess'))
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
    await nextTick()
    inlineCreateInputRef.value?.focus()
    inlineCreateInputRef.value?.select()
  } finally {
    inlineCreateBusy.value = false
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
  beginInlineDelete([item])
}

function beginInlineDelete(targetItems: folder.FileInfo[]) {
  const nextPaths = targetItems
    .map((item) => normalizeEntryPath(item.path))
    .filter(Boolean)
  if (nextPaths.length === 0) return

  clearPendingSingleClick()
  cancelInlineCreate()
  closeContextMenu()
  clearNativeSelection()
  inlineDeletePaths.value = buildInlineDeletePaths(nextPaths, orderedPaths())
  selectionAnchorPath.value = inlineDeletePaths.value[0] ?? null
  activePath.value = inlineDeletePaths.value[0] ?? null
  selectedPaths.value = inlineDeletePaths.value
  focusBrowserViewport()
}

function cancelInlineDelete(item?: folder.FileInfo) {
  if (inlineDeleteBusyPath.value) return
  if (item) {
    const path = normalizeEntryPath(item.path)
    inlineDeletePaths.value = removeInlineDeletePath(inlineDeletePaths.value, path)
    return
  }
  inlineDeletePaths.value = []
}

async function confirmInlineDelete(item: folder.FileInfo) {
  const targetPath = normalizeEntryPath(item.path)
  if (inlineDeleteBusyPath.value || !inlineDeletePathSet.value.has(targetPath)) return

  const snapshot = normalizeFolderItems(items.value)
  const targetItem = snapshot.find((entry) => normalizeEntryPath(entry.path) === targetPath)
  if (!targetItem) {
    cancelInlineDelete(item)
    return
  }

  inlineDeleteBusyPath.value = targetPath
  try {
    await DeleteConnectionEntry(connectionId.value, targetItem.path)
    inlineDeletePaths.value = removeInlineDeletePath(inlineDeletePaths.value, targetPath)
    selectedPaths.value = selectedPaths.value.filter((itemPath) => itemPath !== targetPath)
    if (activePath.value === targetPath) {
      activePath.value = selectedPaths.value[0] ?? inlineDeletePaths.value[0] ?? null
    }
    if (selectionAnchorPath.value === targetPath) {
      selectionAnchorPath.value = activePath.value
    }
    removeItemsByPath([targetPath])
    emitConnectionDirectoryRefresh({
      connectionId: connectionId.value,
      path: currentPath.value,
      source: 'mutation',
      taskId: `delete:${targetPath}`,
      origin: directoryRefreshOrigin,
      mutation: 'delete',
      paths: [targetPath],
    })
    ElMessage.success(t('workspace.fileBrowser.deleteSuccess'))
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  } finally {
    inlineDeleteBusyPath.value = null
  }
}

async function downloadItem(item: folder.FileInfo) {
  try {
    const taskIds = await DownloadConnectionFile(connectionId.value, item.path)
    if (taskIds.length === 0) {
      ElMessage.success(t('workspace.fileBrowser.downloadPrepared', { name: item.name }))
      return
    }
    ElMessage.success(t('workspace.fileBrowser.downloadQueuedBatch', { count: taskIds.length }))
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  }
}

async function downloadItemToTemp(item: folder.FileInfo) {
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

async function saveItemAs(item: folder.FileInfo) {
  try {
    const taskIds = await SaveConnectionFileAs(connectionId.value, item.path)
    if (taskIds.length === 0) return
    ElMessage.success(t('workspace.fileBrowser.downloadQueuedBatch', { count: taskIds.length }))
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  }
}

async function copyItemPath(item: folder.FileInfo) {
  try {
    await ClipboardSetText(item.path)
    ElMessage.success(t('workspace.fileBrowser.copyPathSuccess'))
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  }
}

async function openFile(item: folder.FileInfo) {
  try {
    const taskIds = await OpenConnectionFile(connectionId.value, item.path)
    ElMessage.success(t('workspace.fileBrowser.openQueued', {
      count: taskIds.length || 1,
      name: item.name,
    }))
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  }
}

function onItemCommand(command: string | number | object, item: folder.FileInfo) {
  closeContextMenu()
  const action = String(command)
  if (action === 'open') {
    return openItem(item)
  }
  if (action === 'copy-path') {
    return copyItemPath(item)
  }
  if (action === 'download') {
    return downloadItem(item)
  }
  if (action === 'download-temp') {
    return downloadItemToTemp(item)
  }
  if (action === 'save-as') {
    return saveItemAs(item)
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
  if (detail.origin === directoryRefreshOrigin) return
  if (detail.connectionId !== connectionId.value) return
  if (normalizeEntryPath(detail.path) !== normalizeEntryPath(currentPath.value)) return
  void reload()
}

function onConnectionConfigRefresh(event: Event) {
  const detail = (event as CustomEvent<ConnectionConfigRefreshDetail>).detail
  if (!detail || detail.connectionId !== connectionId.value) return

  workspace.resetConnectionBrowserState(connectionId.value, detail.resetToRoot ? '' : currentPath.value)

  if (detail.resetToRoot) {
    void load('', { history: 'reset' })
    return
  }

  void load(currentPath.value, { history: 'reset' })
}

watch(targetPath, (path) => {
  if (normalizeEntryPath(path) === normalizeEntryPath(currentPath.value)) return
  void load(path)
})

watch(searchedItems, (nextItems) => {
  reconcileSelection(nextItems)
}, { immediate: true })

watch(activePath, (path) => {
  void scrollPathIntoView(path)
})

watch(viewMode, () => {
  void scrollPathIntoView(activePath.value)
})

watch(visibleColumnKeys, (value) => {
  localStorage.setItem(LIST_COLUMNS_STORAGE_KEY, JSON.stringify(normalizeVisibleColumns(value)))
}, { deep: true })

watch([sortKey, sortDirection], ([nextKey, nextDirection]) => {
  localStorage.setItem(LIST_SORT_STORAGE_KEY, JSON.stringify({
    key: nextKey,
    direction: nextDirection,
  }))
})

watch(columnWidths, (value) => {
  localStorage.setItem(LIST_WIDTHS_STORAGE_KEY, JSON.stringify(normalizeColumnWidths(value)))
}, { deep: true })

onMounted(async () => {
  hydrateListPreferences()
  await connections.hydrate()
  window.addEventListener(CONNECTION_CONFIG_REFRESH_EVENT, onConnectionConfigRefresh as EventListener)
  window.addEventListener(CONNECTION_DIRECTORY_REFRESH_EVENT, onDirectoryRefresh as EventListener)
  window.addEventListener(CONNECTION_ENTRY_DROP_LIFECYCLE_EVENT, onEntryDropLifecycle as EventListener)
  window.addEventListener('pointerdown', onWindowPointerDown)
  window.addEventListener('keydown', onWindowKeydown)
  window.addEventListener('keydown', onFileBrowserShortcutKeydown)
  window.addEventListener('dragend', onWindowDragComplete, true)
  window.addEventListener('drop', onWindowDragComplete, true)
  window.addEventListener('resize', closeContextMenu)
  window.addEventListener('scroll', closeContextMenu, true)
  const cachedState = workspace.getConnectionBrowserState(connectionId.value)
  if (cachedState && normalizeEntryPath(cachedState.path) === normalizeEntryPath(targetPath.value)) {
    currentPath.value = normalizeEntryPath(cachedState.path)
    items.value = normalizeFolderItems(cachedState.items)
    workspace.setConnectionPath(connectionId.value, currentPath.value)
    recordNavigationPath(currentPath.value, 'reset')
    revealLoadedPath(workspace.consumeConnectionRevealPath(connectionId.value))
    return
  }
  await load(targetPath.value, { history: 'reset' })
})

onBeforeUnmount(() => {
  resetTransientInteractionState()
  dropBusyOperations.value = []
  pendingHiddenPaths.value = []
  stopColumnResize()
  window.removeEventListener(CONNECTION_CONFIG_REFRESH_EVENT, onConnectionConfigRefresh as EventListener)
  window.removeEventListener(CONNECTION_DIRECTORY_REFRESH_EVENT, onDirectoryRefresh as EventListener)
  window.removeEventListener(CONNECTION_ENTRY_DROP_LIFECYCLE_EVENT, onEntryDropLifecycle as EventListener)
  window.removeEventListener('pointerdown', onWindowPointerDown)
  window.removeEventListener('keydown', onWindowKeydown)
  window.removeEventListener('keydown', onFileBrowserShortcutKeydown)
  window.removeEventListener('dragend', onWindowDragComplete, true)
  window.removeEventListener('drop', onWindowDragComplete, true)
  window.removeEventListener('resize', closeContextMenu)
  window.removeEventListener('scroll', closeContextMenu, true)
})
</script>

<style scoped>
.file-browser {
  position: relative;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 18px;
  overflow: auto;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--theme-color-bg-overlay) 35%, transparent), transparent 18%),
    var(--theme-color-bg-base);
  user-select: none;
  -webkit-user-select: none;
}

.file-browser :not(input):not(textarea):not([contenteditable]) {
  user-select: none;
  -webkit-user-select: none;
}

.file-browser input,
.file-browser textarea,
.file-browser [contenteditable] {
  user-select: text;
  -webkit-user-select: text;
}

.file-browser--busy {
  overflow: hidden;
}

.file-browser__busy-mask {
  position: absolute;
  inset: 0;
  z-index: 5;
  display: flex;
  align-items: center;
  justify-content: center;
  background: color-mix(in srgb, var(--theme-color-bg-overlay) 82%, transparent);
  backdrop-filter: blur(2px);
}

.file-browser__busy-card {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 14px 18px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 14px;
  background: color-mix(in srgb, var(--theme-color-bg-surface) 96%, transparent);
  color: var(--theme-color-text-base);
  box-shadow: 0 14px 32px color-mix(in srgb, var(--theme-color-shadow) 20%, transparent);
}

.file-browser__busy-icon {
  font-size: 16px;
  color: var(--theme-color-primary);
}

.file-browser__toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 14px;
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
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  flex-wrap: wrap;
}

.file-browser__action-group {
  flex: 0 0 auto;
}

:deep(.file-browser__action-group > .el-button),
:deep(.file-browser__action-group > .el-tooltip__trigger > .el-button) {
  min-height: 30px;
  padding: 0 11px;
  font-size: 12px;
}

.file-browser__controls :deep(.el-button [class^='i-']),
.file-browser__controls :deep(.el-button [class*=' i-']) {
  font-size: 14px;
}

.file-browser__view-toggle {
  display: inline-flex;
  align-items: center;
  flex: 0 0 auto;
  padding: 2px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 10px;
  background: color-mix(in srgb, var(--theme-color-bg-surface) 88%, transparent);
}

.file-browser__view-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--theme-color-text-secondary);
  cursor: pointer;
  line-height: 1;
  flex: 0 0 30px;
}

.file-browser__view-btn > * {
  font-size: 15px;
}

.file-browser__view-btn--active {
  background: var(--theme-color-primary-light);
  color: var(--theme-color-primary);
}

.file-browser__subbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.file-browser__breadcrumbs {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
  min-width: 0;
  padding: 4px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 12px;
  background: color-mix(in srgb, var(--theme-color-bg-surface) 88%, transparent);
}

.file-browser__crumb {
  min-height: 26px;
  padding: 0 8px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--theme-color-text);
  font-size: 12px;
  cursor: pointer;
  white-space: nowrap;
}

.file-browser__crumb:hover {
  background: color-mix(in srgb, var(--theme-color-primary) 9%, transparent);
}

.file-browser__crumb--root {
  font-weight: 600;
  color: var(--theme-color-text-base);
}

.file-browser__crumb-sep {
  color: var(--theme-color-text-secondary);
  font-size: 12px;
  user-select: none;
}

.file-browser__status {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 6px;
  color: var(--theme-color-text-secondary);
}

.file-browser__status-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-height: 26px;
  padding: 0 9px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 999px;
  background: color-mix(in srgb, var(--theme-color-bg-surface) 88%, transparent);
  font-size: 12px;
  white-space: nowrap;
}

.file-browser__status-pill--accent {
  border-color: color-mix(in srgb, var(--theme-color-primary) 22%, var(--theme-color-border-light));
  color: var(--theme-color-primary);
}

.file-browser__status-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: var(--theme-color-danger);
  flex: 0 0 8px;
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

.file-browser__column-picker {
  display: grid;
  gap: 8px;
}

.file-browser__column-option {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--theme-color-text);
}

.file-browser__column-option--locked {
  color: var(--theme-color-text-secondary);
}

.file-browser__viewport-wrap {
  position: relative;
  flex: 1;
  min-height: 0;
  display: flex;
}

.file-browser__local-search {
  position: absolute;
  top: 10px;
  right: 12px;
  z-index: 6;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  max-width: min(560px, calc(100% - 24px));
  padding: 4px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 10px;
  background: color-mix(in srgb, var(--theme-color-bg-overlay) 96%, var(--theme-color-bg-surface));
  box-shadow: 0 12px 28px color-mix(in srgb, var(--theme-color-shadow) 16%, transparent);
}

.file-browser__local-search-input {
  width: min(220px, 32vw);
  min-width: 120px;
  height: 28px;
  padding: 0 8px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 7px;
  background: var(--theme-color-bg-base);
  color: var(--theme-color-text-base);
  font: inherit;
  font-size: 12px;
  outline: none;
  user-select: text;
}

.file-browser__local-search-input:focus {
  border-color: var(--theme-color-primary);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--theme-color-primary) 18%, transparent);
}

.file-browser__local-search-option,
.file-browser__local-search-close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 7px;
  background: transparent;
  color: var(--theme-color-text-secondary);
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}

.file-browser__local-search-option:hover,
.file-browser__local-search-close:hover {
  background: var(--theme-color-bg-hover);
  color: var(--theme-color-text-base);
}

.file-browser__local-search-option--active {
  background: var(--theme-color-primary-light);
  color: var(--theme-color-primary);
}

.file-browser__local-search-count {
  min-width: 52px;
  color: var(--theme-color-text-secondary);
  font-size: 11px;
  text-align: center;
  white-space: nowrap;
}

.file-browser__local-search-count--error {
  color: var(--theme-color-danger);
}

.file-browser__list-shell {
  flex: 1;
  min-height: 0;
  overflow: auto;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 14px;
  background: color-mix(in srgb, var(--theme-color-bg-surface) 94%, transparent);
}

.file-browser__list-shell:focus-visible,
.file-browser__grid:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--theme-color-primary) 38%, transparent);
  outline-offset: 2px;
}

.file-browser__drop-root {
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--theme-color-primary) 32%, transparent);
  background: color-mix(in srgb, var(--theme-color-primary) 6%, var(--theme-color-bg-surface));
}

.file-browser__list-table {
  min-width: 100%;
  width: max-content;
}

.file-browser__list-header,
.file-browser__row {
  display: grid;
  align-items: center;
}

.file-browser__list-header {
  position: sticky;
  top: 0;
  z-index: 2;
  min-width: 100%;
  background: color-mix(in srgb, var(--theme-color-bg-overlay) 92%, var(--theme-color-bg-surface));
}

.file-browser__list-header-btn {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 36px;
  padding: 0 16px 0 12px;
  border: none;
  background: transparent;
  color: var(--theme-color-text-secondary);
  font-size: clamp(11px, calc(var(--ui-file-list-font-size, 13px) - 1px), 17px);
  font-weight: 600;
  text-align: left;
  cursor: pointer;
}

.file-browser__list-header-btn--right {
  justify-content: flex-end;
  text-align: right;
}

.file-browser__list-header-btn:hover {
  color: var(--theme-color-text-base);
}

.file-browser__sort-indicator {
  width: 10px;
  color: var(--theme-color-primary);
  font-size: 11px;
}

.file-browser__column-resizer {
  position: absolute;
  top: 6px;
  right: -4px;
  width: 10px;
  height: calc(100% - 12px);
  border-radius: 999px;
  background: transparent;
  cursor: col-resize;
}

.file-browser__list-body {
  min-width: 100%;
}

.file-browser__viewport-empty {
  min-height: 220px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: var(--theme-color-text-secondary);
  text-align: center;
  user-select: none;
}

.file-browser__viewport-empty--list {
  min-width: 100%;
  padding: 24px 16px 32px;
}

.file-browser__viewport-empty--grid {
  grid-column: 1 / -1;
  min-height: 240px;
  padding: 24px 16px 32px;
}

.file-browser__row {
  width: 100%;
  min-height: calc(var(--ui-file-list-font-size, 13px) + 25px);
  padding: 0;
  border: none;
  border-bottom: 1px solid var(--theme-color-border-light);
  background: transparent;
  color: var(--theme-color-text);
  text-align: left;
  cursor: pointer;
}

.file-browser__row:hover {
  background: color-mix(in srgb, var(--theme-color-primary) 7%, transparent);
}

.file-browser__row--selected {
  background: color-mix(in srgb, var(--theme-color-primary) 12%, var(--theme-color-bg-surface));
}

.file-browser__row--active {
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--theme-color-primary) 42%, transparent);
}

.file-browser__row--drop-target {
  background: color-mix(in srgb, var(--theme-color-success) 12%, var(--theme-color-bg-surface));
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--theme-color-success) 38%, transparent);
}

.file-browser__row--delete-pending {
  background: color-mix(in srgb, var(--theme-color-danger) 8%, var(--theme-color-bg-surface));
}

.file-browser__row--inline-create {
  cursor: default;
  background: color-mix(in srgb, var(--theme-color-primary) 7%, var(--theme-color-bg-surface));
}

.file-browser__row--inline-create:hover {
  background: color-mix(in srgb, var(--theme-color-primary) 7%, var(--theme-color-bg-surface));
}

.file-browser__cell {
  min-width: 0;
  padding: 0 12px;
  color: var(--theme-color-text-secondary);
  font-size: var(--ui-file-list-font-size, 13px);
}

.file-browser__cell--left {
  text-align: left;
}

.file-browser__cell--right {
  text-align: right;
}

.file-browser__cell--name {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--theme-color-text-base);
}

.file-browser__row-icon {
  flex: 0 0 auto;
  font-size: calc(var(--ui-file-list-font-size, 13px) + 5px);
  color: var(--theme-color-primary);
}

.file-browser__entry-name-wrap {
  position: relative;
  z-index: 1;
  flex: 1 1 auto;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-browser__row-name {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-browser__inline-delete-actions {
  position: relative;
  z-index: 3;
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-left: auto;
  padding-left: 8px;
  background: linear-gradient(90deg, transparent, color-mix(in srgb, var(--theme-color-bg-surface) 92%, transparent) 16%);
}

.file-browser__inline-delete-actions--tile {
  margin-left: 0;
  padding-left: 0;
  background: transparent;
}

.file-browser__inline-delete-btn {
  min-height: 24px;
  padding: 0 8px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 7px;
  background: var(--theme-color-bg-card);
  color: var(--theme-color-text-secondary);
  font-size: 11px;
  line-height: 1;
  cursor: pointer;
  white-space: nowrap;
}

.file-browser__inline-delete-btn:hover:not(:disabled) {
  background: var(--theme-color-bg-hover);
  color: var(--theme-color-text-base);
}

.file-browser__inline-delete-btn:disabled {
  opacity: 0.62;
  cursor: default;
}

.file-browser__inline-delete-btn--danger {
  border-color: color-mix(in srgb, var(--theme-color-danger) 38%, var(--theme-color-border-light));
  color: var(--theme-color-danger);
}

.file-browser__inline-create-input {
  min-width: 0;
  width: 100%;
  height: 26px;
  padding: 0 8px;
  border: 1px solid color-mix(in srgb, var(--theme-color-primary) 38%, var(--theme-color-border-light));
  border-radius: 7px;
  background: var(--theme-color-bg-base);
  color: var(--theme-color-text-base);
  font: inherit;
  outline: none;
  user-select: text;
}

.file-browser__inline-create-input:focus {
  border-color: var(--theme-color-primary);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--theme-color-primary) 18%, transparent);
}

.file-browser__grid {
  flex: 1;
  min-height: 0;
  overflow: auto;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(168px, 1fr));
  gap: 14px;
  align-content: start;
  padding: 2px;
}

.file-browser__tile {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  width: 100%;
  height: 152px;
  padding: 16px 14px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 16px;
  background: color-mix(in srgb, var(--theme-color-bg-surface) 88%, transparent);
  color: var(--theme-color-text);
  text-align: center;
  cursor: pointer;
}

.file-browser__tile:hover {
  background: var(--theme-color-bg-hover);
}

.file-browser__tile--selected {
  background: color-mix(in srgb, var(--theme-color-primary) 10%, var(--theme-color-bg-surface));
  border-color: color-mix(in srgb, var(--theme-color-primary) 24%, var(--theme-color-border-light));
}

.file-browser__tile--active {
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--theme-color-primary) 42%, transparent);
}

.file-browser__tile--drop-target {
  background: color-mix(in srgb, var(--theme-color-success) 10%, var(--theme-color-bg-surface));
  border-color: color-mix(in srgb, var(--theme-color-success) 36%, var(--theme-color-border-light));
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--theme-color-success) 34%, transparent);
}

.file-browser__tile--delete-pending {
  background: color-mix(in srgb, var(--theme-color-danger) 8%, var(--theme-color-bg-surface));
  border-color: color-mix(in srgb, var(--theme-color-danger) 28%, var(--theme-color-border-light));
}

.file-browser__tile--inline-create {
  cursor: default;
  background: color-mix(in srgb, var(--theme-color-primary) 7%, var(--theme-color-bg-surface));
}

.file-browser__tile--inline-create:hover {
  background: color-mix(in srgb, var(--theme-color-primary) 7%, var(--theme-color-bg-surface));
}

.file-browser__tile-icon {
  font-size: calc(var(--ui-file-list-font-size, 13px) + 15px);
  color: var(--theme-color-primary);
  line-height: 1;
}

.file-browser__tile-name {
  width: 100%;
  font-size: var(--ui-file-list-font-size, 13px);
  font-weight: 600;
  color: var(--theme-color-text-base);
  display: -webkit-box;
  overflow: hidden;
  text-align: center;
  text-wrap: balance;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.file-browser__tile-name-wrap {
  position: relative;
  z-index: 1;
  width: 100%;
  min-width: 0;
  display: grid;
  justify-items: center;
  gap: 7px;
}

.file-browser__tile-meta {
  font-size: clamp(11px, calc(var(--ui-file-list-font-size, 13px) - 1px), 17px);
  color: var(--theme-color-text-secondary);
  text-align: center;
}

.file-browser__inline-create-input--tile {
  width: min(128px, 100%);
  text-align: center;
}

.file-browser__context-menu {
  position: fixed;
  z-index: 2000;
  min-width: 188px;
  padding: 6px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 12px;
  background: color-mix(in srgb, var(--theme-color-bg-surface) 96%, transparent);
  box-shadow:
    0 16px 32px color-mix(in srgb, var(--theme-color-shadow) 22%, transparent),
    0 0 0 1px color-mix(in srgb, var(--theme-color-bg-overlay) 35%, transparent);
  backdrop-filter: blur(8px);
  user-select: none;
}

.file-browser__context-menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  width: 100%;
  min-height: 34px;
  padding: 0 10px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--theme-color-text-base);
  font-size: 13px;
  text-align: left;
  cursor: pointer;
}

.file-browser__context-menu-item:hover {
  background: color-mix(in srgb, var(--theme-color-primary) 10%, transparent);
}

.file-browser__context-menu-item--danger {
  color: var(--theme-color-danger);
}

.file-browser__context-menu-shortcut {
  flex: 0 0 auto;
  min-width: 22px;
  padding: 1px 6px;
  border: 1px solid color-mix(in srgb, currentColor 22%, var(--theme-color-border-light));
  border-radius: 6px;
  color: var(--theme-color-text-secondary);
  font-size: 11px;
  line-height: 1.3;
  text-align: center;
}

.file-browser__context-menu-separator {
  height: 1px;
  margin: 6px 2px;
  background: var(--theme-color-border-light);
}

@media (max-width: 900px) {
  .file-browser {
    padding: 14px;
  }

  .file-browser__toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .file-browser__subbar {
    flex-direction: column;
    align-items: stretch;
  }

  .file-browser__controls {
    justify-content: flex-start;
  }

  .file-browser__status {
    justify-content: flex-start;
  }

  .file-browser__list-shell {
    border-radius: 12px;
  }
}
</style>
