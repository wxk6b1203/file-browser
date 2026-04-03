<template>
  <SkeletonLayout
    v-model:left-active="leftActive"
    v-model:right-active="rightActive"
    :left-sidebar="leftSidebar"
    :right-sidebar="rightSidebar"
    :gap="1"
  >
    <template #left-panel="{ activeId }">
      <ExplorerPanel v-if="activeId === 'explorer'" />
      <SearchPanel v-else />
    </template>

    <template #center>
      <div class="shell-center">
        <Tabs
          :model-value="workspace.layout"
          :overlay-opacity="0.15"
          :min-split-width="220"
          :min-split-height="160"
          enable-file-drop
          enable-panel-drag
          @update:model-value="workspace.setLayout"
          @tab-activate="(_tab, groupId) => workspace.setActiveGroup(groupId)"
          @panel-drop="onPanelDrop"
        />
      </div>
    </template>

    <template #right-panel="{ activeId }">
      <TaskPanel v-if="activeId === 'tasks'" />
      <NotificationPanel v-else />
    </template>
  </SkeletonLayout>
</template>

<script setup lang="ts">
import { computed, markRaw, onBeforeUnmount, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { UploadConnectionLocalPath } from '../../../wailsjs/go/main/App'
import { Tabs } from '@/components/Tabs'
import type { PanelDropEvent } from '@/components/SplitPane'
import { buildConnectionEntryDropFeedback, executeConnectionEntryDrop, labelForDropTarget, type ConnectionEntryPanelDragPayload } from '@/composables/connectionEntryDrop'
import { SkeletonLayout, type SidebarConfig } from '@/views/Skeleton'
import ExplorerPanel from '@/components/ExplorerTree/ExplorerPanel.vue'
import SearchPanel from '@/components/SearchPanel/SearchPanel.vue'
import TaskPanel from '@/components/TaskPanel/TaskPanel.vue'
import NotificationPanel from '@/components/NotificationPanel/NotificationPanel.vue'
import { emitConnectionDirectoryRefresh } from '@/composables/useConnectionDirectoryRefresh'
import { PANEL_OS_FILE_DROP_EVENT, useFileDrop, type PanelOSFileDropDetail } from '@/composables/useFileDrop'
import { useConnectionsStore } from '@/stores/connections'
import { useNotificationsStore } from '@/stores/notifications'
import { useSettingsStore } from '@/stores/settings'
import { useShellStore } from '@/stores/shell'
import { useTasksStore } from '@/stores/tasks'
import { useWorkspaceStore } from '@/stores/workspace'
import { useShortcutMap } from '@/composables/useShortcut'
import IEpSearch from '~icons/ep/search'
import IMdiFolder from '~icons/mdi/folder'
import IEpSetting from '~icons/ep/setting'
import IMdiTrayArrowUp from '~icons/mdi/tray-arrow-up'
import IEpBell from '~icons/ep/bell'

const { t } = useI18n()
const connections = useConnectionsStore()
const notifications = useNotificationsStore()
const settings = useSettingsStore()
const shell = useShellStore()
const tasks = useTasksStore()
const workspace = useWorkspaceStore()
const { leftActive, rightActive } = storeToRefs(shell)

useFileDrop()

const leftSidebar: SidebarConfig = {
  topButtons: [
    {
      id: 'explorer',
      icon: markRaw(IMdiFolder),
      tooltip: t('shell.sidebar.explorer'),
      type: 'menu',
    },
    {
      id: 'search',
      icon: markRaw(IEpSearch),
      tooltip: t('shell.sidebar.search'),
      type: 'menu',
    },
  ],
  bottomButtons: [
    {
      id: 'settings',
      icon: markRaw(IEpSetting),
      tooltip: t('shell.sidebar.settings'),
      type: 'action',
      onClick: () => workspace.openSettings(),
    },
  ],
  defaultSize: '280px',
  minSize: '220px',
  maxSize: '420px',
}

const rightSidebar = computed<SidebarConfig>(() => ({
  topButtons: [
    {
      id: 'tasks',
      icon: markRaw(IMdiTrayArrowUp),
      tooltip: t('shell.sidebar.tasks'),
      type: 'menu',
    },
    {
      id: 'notifications',
      icon: markRaw(IEpBell),
      tooltip: t('shell.sidebar.notifications'),
      type: 'menu',
      indicator: notifications.unreadCount > 0,
    },
  ],
  bottomButtons: [],
  defaultSize: '260px',
  minSize: '220px',
  maxSize: '360px',
}))

useShortcutMap({
  'new-connection': () => workspace.openNewConnection(),
  'open-settings': () => workspace.openSettings(),
})

function connectionIdFromTabId(tabId: string) {
  return tabId.startsWith('connection:') ? tabId.slice('connection:'.length) : ''
}

async function onPanelOSFileDrop(event: Event) {
  const detail = (event as CustomEvent<PanelOSFileDropDetail>).detail
  if (!detail) return

  const connectionId = connectionIdFromTabId(detail.tabId)
  if (!connectionId) return

  const definition = connections.definitionMap.get(connectionId)
  if (!definition) return

  const remoteDir = workspace.getConnectionPath(connectionId)
  const results = await Promise.allSettled(
    detail.paths.map((localPath) => UploadConnectionLocalPath(connectionId, remoteDir, localPath)),
  )

  const successCount = results.filter((item) => item.status === 'fulfilled').length
  const failed = results.filter((item) => item.status === 'rejected')
  const queuedTaskCount = results.reduce((count, item) => {
    if (item.status !== 'fulfilled') return count
    return count + item.value.length
  }, 0)

  if (successCount > 0) {
    emitConnectionDirectoryRefresh({
      connectionId,
      path: remoteDir,
      source: 'transfer',
      taskId: 'queued-upload',
    })
    ElMessage.success(t('workspace.fileBrowser.uploadQueued', {
      count: queuedTaskCount || successCount,
      name: definition.name,
    }))
  }

  const firstError = failed[0]
  if (firstError) {
    const message = firstError.status === 'rejected'
      ? (firstError.reason instanceof Error ? firstError.reason.message : String(firstError.reason))
      : t('workspace.fileBrowser.uploadFailed')
    notifications.push({
      level: 'error',
      source: definition.name,
      title: 'Upload Failed',
      message,
      action: {
        kind: 'open-connection',
        connectionId,
        connectionName: definition.name,
        path: remoteDir,
      },
    })
    ElMessage.error(message)
  }
}

async function onPanelDrop(event: PanelDropEvent) {
  const payload = event.payload as ConnectionEntryPanelDragPayload | undefined
  if (!payload || payload.type !== 'connection-entry') return

  const targetConnectionId = connectionIdFromTabId(event.targetTabId ?? '')
  if (!targetConnectionId) return

  const targetDefinition = connections.definitionMap.get(targetConnectionId)
  if (!targetDefinition) return

  const remoteDir = workspace.getConnectionPath(targetConnectionId)

  try {
    const result = await executeConnectionEntryDrop(payload, targetConnectionId, remoteDir)
    const feedback = buildConnectionEntryDropFeedback(
      result,
      labelForDropTarget(remoteDir, targetDefinition.name),
    )
    if (result.mode !== 'noop') {
      ElMessage.success(t(feedback.key, feedback.params))
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    notifications.push({
      level: 'error',
      source: targetDefinition.name,
      title: 'Transfer Failed',
      message,
      action: {
        kind: 'open-connection',
        connectionId: targetConnectionId,
        connectionName: targetDefinition.name,
        path: remoteDir,
      },
    })
    ElMessage.error(message)
  }
}

onMounted(async () => {
  await settings.hydrate()
  await connections.hydrate()
  tasks.ensureSubscribed()
  workspace.ensureWelcomeTab()
  window.addEventListener(PANEL_OS_FILE_DROP_EVENT, onPanelOSFileDrop as EventListener)
})

onBeforeUnmount(() => {
  window.removeEventListener(PANEL_OS_FILE_DROP_EVENT, onPanelOSFileDrop as EventListener)
})
</script>

<style scoped>
.shell-center {
  width: 100%;
  height: 100%;
  overflow: hidden;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--theme-color-bg-overlay) 45%, transparent), transparent 18%),
    var(--theme-color-bg-base);
}

.shell-panel-placeholder {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 18px;
  text-align: center;
  color: var(--theme-color-text-secondary);
}

.shell-panel-placeholder h3 {
  margin: 0;
  color: var(--theme-color-text-base);
  font-size: 15px;
}

.shell-panel-placeholder p {
  margin: 0;
  max-width: 220px;
}

.shell-panel-placeholder__icon {
  font-size: 22px;
  color: var(--theme-color-primary);
}
</style>
