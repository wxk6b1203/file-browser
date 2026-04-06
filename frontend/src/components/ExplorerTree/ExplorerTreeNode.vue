<template>
  <div class="explorer-node">
    <div
      class="explorer-node__row"
      :class="{
        'explorer-node__row--drop-target': dropActive,
        'explorer-node__row--drop-expand': dropExpandPending,
      }"
      :style="{ paddingLeft: `${8 + node.level * 16}px` }"
      @dblclick="onOpen(node)"
      @dragenter="onDragEnter"
      @dragover="onDragOver"
      @dragleave="onDragLeave"
      @drop="onDrop"
    >
      <button
        v-if="canToggle"
        type="button"
        class="explorer-node__toggle"
        @click.stop="emit('toggle', node)"
      >
        <i-ep-arrow-right
          class="explorer-node__toggle-icon"
          :class="{ 'explorer-node__toggle-icon--open': expanded }"
        />
      </button>
      <span v-else class="explorer-node__toggle explorer-node__toggle--placeholder" />

      <button type="button" class="explorer-node__main" @click="onActivate(node)">
        <span class="explorer-node__icon">
          <i-mdi-server-network-outline v-if="node.kind === 'connection'" />
          <component :is="explorerIcon" v-else />
        </span>

        <span class="explorer-node__label">{{ node.label }}</span>

        <span v-if="node.kind === 'connection' && node.driver" class="explorer-node__driver">
          {{ node.driver }}
        </span>

        <span
          v-if="node.kind === 'connection' && !isDeletePending"
          class="explorer-node__status"
          :class="{ 'explorer-node__status--online': node.connected }"
        >
          {{ node.connected ? t('shell.explorer.connected') : t('shell.explorer.saved') }}
        </span>
      </button>

      <div v-if="isDeletePending" class="explorer-node__delete-actions">
        <button
          type="button"
          class="explorer-node__action-btn explorer-node__action-btn--danger"
          :disabled="deleteBusy"
          @click.stop="emit('command', 'confirm-delete', node)"
        >
          {{ t('shell.explorer.confirmDelete') }}
        </button>
        <button
          type="button"
          class="explorer-node__action-btn"
          :disabled="deleteBusy"
          @click.stop="emit('command', 'cancel-delete', node)"
        >
          {{ t('shell.explorer.cancelDelete') }}
        </button>
      </div>

      <el-dropdown
        v-else-if="node.kind === 'connection'"
        trigger="click"
        @command="emit('command', String($event), node)"
      >
        <button type="button" class="explorer-node__menu-btn" @click.stop>
          <i-mdi-dots-horizontal />
        </button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="open">{{ t('shell.explorer.open') }}</el-dropdown-item>
            <el-dropdown-item command="edit">{{ t('shell.explorer.edit') }}</el-dropdown-item>
            <el-dropdown-item command="close">{{ t('shell.explorer.close') }}</el-dropdown-item>
            <el-dropdown-item command="delete" divided>{{ t('shell.explorer.delete') }}</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <div v-if="expanded" class="explorer-node__children">
      <div v-if="loading" class="explorer-node__loading">
        {{ t('shell.explorer.loadingChildren') }}
      </div>

      <div v-else-if="node.children.length === 0" class="explorer-node__loading explorer-node__loading--muted">
        {{ t('shell.explorer.emptyDirectory') }}
      </div>

      <template v-else>
        <ExplorerTreeNode
          v-for="child in node.children"
          :key="child.key"
          :node="child"
          :expanded-keys="expandedKeys"
          :loading-keys="loadingKeys"
          :delete-pending-connection-id="deletePendingConnectionId"
          :delete-busy-connection-id="deleteBusyConnectionId"
          @toggle="forwardToggle"
          @activate="forwardActivate"
          @open="forwardOpen"
          @command="forwardCommand"
          @hover-expand="forwardHoverExpand"
          @drop-entries="forwardDropEntries"
        />
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { canDropConnectionEntries, resolveConnectionEntryDragPayload, type ConnectionEntryPanelDragPayload } from '@/composables/connectionEntryDrop'
import { DIRECTORY_ENTRY_TYPE, resolveFileIcon } from '@/composables/useFileIcons'
import type { ExplorerNode } from './types'

const props = defineProps<{
  node: ExplorerNode
  expandedKeys: Set<string>
  loadingKeys: Set<string>
  deletePendingConnectionId?: string
  deleteBusyConnectionId?: string
}>()

const emit = defineEmits<{
  toggle: [node: ExplorerNode]
  activate: [node: ExplorerNode]
  open: [node: ExplorerNode]
  command: [command: string, node: ExplorerNode]
  hoverExpand: [node: ExplorerNode]
  dropEntries: [node: ExplorerNode, payload: ConnectionEntryPanelDragPayload]
}>()

const { t } = useI18n()

const expanded = computed(() => props.expandedKeys.has(props.node.key))
const loading = computed(() => props.loadingKeys.has(props.node.key))
const canToggle = computed(() => props.node.kind !== 'file')
const explorerIcon = computed(() => resolveFileIcon(
  props.node.entry ?? {
    name: props.node.label,
    path: props.node.path,
    type: props.node.kind === 'directory' ? DIRECTORY_ENTRY_TYPE : undefined,
  },
  { opened: props.node.kind === 'directory' && expanded.value },
))
const isDeletePending = computed(() => (
  props.node.kind === 'connection' && props.deletePendingConnectionId === props.node.connectionId
))
const deleteBusy = computed(() => (
  props.node.kind === 'connection' && props.deleteBusyConnectionId === props.node.connectionId
))
const dropActive = ref(false)
const dropExpandPending = ref(false)
let hoverExpandTimer: ReturnType<typeof setTimeout> | null = null

function onOpen(node: ExplorerNode) {
  emit('open', node)
}

function onActivate(node: ExplorerNode) {
  if (node.kind === 'connection') {
    emit('toggle', node)
    return
  }
  if (node.kind === 'file') {
    return
  }
  emit('activate', node)
}

function forwardToggle(node: ExplorerNode) {
  emit('toggle', node)
}

function forwardActivate(node: ExplorerNode) {
  emit('activate', node)
}

function forwardOpen(node: ExplorerNode) {
  emit('open', node)
}

function forwardCommand(command: string, node: ExplorerNode) {
  emit('command', command, node)
}

function scheduleHoverExpand() {
  if (!canToggle.value || expanded.value || loading.value || hoverExpandTimer) return
  hoverExpandTimer = setTimeout(() => {
    hoverExpandTimer = null
    dropExpandPending.value = false
    emit('hoverExpand', props.node)
  }, 550)
  dropExpandPending.value = true
}

function clearHoverExpandSchedule() {
  if (hoverExpandTimer) {
    clearTimeout(hoverExpandTimer)
    hoverExpandTimer = null
  }
  dropExpandPending.value = false
}

function dropTargetPath() {
  if (props.node.kind === 'file') return null
  return props.node.kind === 'connection' ? '' : props.node.path
}

function onDragEnter(event: DragEvent) {
  const payload = resolveConnectionEntryDragPayload(event)
  const targetPath = dropTargetPath()
  if (targetPath === null || !canDropConnectionEntries(payload, props.node.connectionId, targetPath)) return

  event.preventDefault()
  event.stopPropagation()
  dropActive.value = true
  scheduleHoverExpand()
}

function onDragOver(event: DragEvent) {
  const payload = resolveConnectionEntryDragPayload(event)
  const targetPath = dropTargetPath()
  if (targetPath === null || !canDropConnectionEntries(payload, props.node.connectionId, targetPath)) return

  event.preventDefault()
  event.stopPropagation()
  dropActive.value = true
  scheduleHoverExpand()
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = payload?.data.sourceConnectionId === props.node.connectionId ? 'move' : 'copy'
  }
}

function onDragLeave(event: DragEvent) {
  const relatedTarget = event.relatedTarget
  if (relatedTarget instanceof Node && (event.currentTarget as HTMLElement | null)?.contains(relatedTarget)) {
    return
  }
  dropActive.value = false
  clearHoverExpandSchedule()
}

function onDrop(event: DragEvent) {
  const payload = resolveConnectionEntryDragPayload(event)
  const targetPath = dropTargetPath()
  dropActive.value = false
  clearHoverExpandSchedule()
  if (targetPath === null || !payload || !canDropConnectionEntries(payload, props.node.connectionId, targetPath)) return

  event.preventDefault()
  event.stopPropagation()
  emit('dropEntries', props.node, payload)
}

function forwardHoverExpand(node: ExplorerNode) {
  emit('hoverExpand', node)
}

function forwardDropEntries(node: ExplorerNode, payload: ConnectionEntryPanelDragPayload) {
  emit('dropEntries', node, payload)
}

onBeforeUnmount(() => {
  clearHoverExpandSchedule()
})
</script>

<style scoped>
.explorer-node__row {
  display: flex;
  align-items: center;
  gap: 4px;
  min-height: 34px;
  border-radius: 10px;
  font-size: var(--ui-explorer-font-size, 13px);
  user-select: none;
}

.explorer-node__row:hover {
  background: var(--theme-color-bg-hover);
}

.explorer-node__row--drop-target {
  background: color-mix(in srgb, var(--theme-color-success) 12%, transparent);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--theme-color-success) 34%, transparent);
}

.explorer-node__row--drop-expand {
  box-shadow:
    inset 0 0 0 1px color-mix(in srgb, var(--theme-color-success) 34%, transparent),
    inset 3px 0 0 color-mix(in srgb, var(--theme-color-primary) 48%, transparent);
}

.explorer-node__toggle {
  width: 18px;
  height: 18px;
  border: none;
  background: transparent;
  color: var(--theme-color-text-secondary);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.explorer-node__toggle--placeholder {
  cursor: default;
}

.explorer-node__toggle-icon {
  transition: transform 0.14s ease;
}

.explorer-node__toggle-icon--open {
  transform: rotate(90deg);
}

.explorer-node__main {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 8px;
  border: none;
  background: transparent;
  color: inherit;
  text-align: left;
  cursor: pointer;
}

.explorer-node__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--theme-color-primary);
  font-size: 16px;
}

.explorer-node__label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--theme-color-text-base);
}

.explorer-node__driver {
  padding: 1px 6px;
  border-radius: 999px;
  background: var(--theme-color-bg-overlay);
  color: var(--theme-color-text-secondary);
  font-size: clamp(11px, calc(var(--ui-explorer-font-size, 13px) - 2px), 16px);
}

.explorer-node__status {
  margin-left: auto;
  color: var(--theme-color-text-secondary);
  font-size: clamp(11px, calc(var(--ui-explorer-font-size, 13px) - 2px), 16px);
}

.explorer-node__status--online {
  color: var(--theme-color-success);
}

.explorer-node__menu-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--theme-color-text-secondary);
  cursor: pointer;
}

.explorer-node__menu-btn:hover {
  background: var(--theme-color-bg-hover);
  color: var(--theme-color-text-base);
}

.explorer-node__delete-actions {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-right: 2px;
}

.explorer-node__action-btn {
  min-height: 24px;
  padding: 0 8px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 7px;
  background: var(--theme-color-bg-card);
  color: var(--theme-color-text-secondary);
  font-size: 11px;
  line-height: 1;
  cursor: pointer;
}

.explorer-node__action-btn:hover:not(:disabled) {
  background: var(--theme-color-bg-hover);
  color: var(--theme-color-text-base);
}

.explorer-node__action-btn:disabled {
  opacity: 0.6;
  cursor: default;
}

.explorer-node__action-btn--danger {
  border-color: color-mix(in srgb, var(--theme-color-danger) 34%, var(--theme-color-border-light));
  color: var(--theme-color-danger);
}

.explorer-node__children {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.explorer-node__loading {
  padding: 6px 8px 6px 32px;
  color: var(--theme-color-text-secondary);
  font-size: clamp(11px, calc(var(--ui-explorer-font-size, 13px) - 1px), 17px);
}

.explorer-node__loading--muted {
  opacity: 0.7;
}
</style>
