import { computed, markRaw, ref } from 'vue'
import { defineStore } from 'pinia'
import type { TabGroupNode, TabItem, TabNode } from '@/components/Tabs'
import WelcomeTab from '@/components/Workspace/WelcomeTab.vue'
import ConnectionFormTab from '@/components/Workspace/ConnectionFormTab.vue'
import SettingsTab from '@/components/Workspace/SettingsTab.vue'
import ConnectionOverviewTab from '@/components/Workspace/ConnectionOverviewTab.vue'

const ROOT_GROUP_ID = 'workspace-root'

interface ConnectionTabState {
  path: string
}

function createEmptyGroup(): TabGroupNode {
  return {
    type: 'tabs',
    id: ROOT_GROUP_ID,
    tabs: [],
    activeId: '',
  }
}

function findGroupById(node: TabNode, id: string): TabGroupNode | null {
  if (node.type === 'tabs') {
    return node.id === id ? node : null
  }
  for (const child of node.children) {
    const found = findGroupById(child, id)
    if (found) return found
  }
  return null
}

function findFirstGroup(node: TabNode): TabGroupNode | null {
  if (node.type === 'tabs') return node
  for (const child of node.children) {
    const found = findFirstGroup(child)
    if (found) return found
  }
  return null
}

function hasTabs(node: TabNode): boolean {
  if (node.type === 'tabs') return node.tabs.length > 0
  return node.children.some((child) => hasTabs(child))
}

function findTabGroup(node: TabNode, tabId: string): TabGroupNode | null {
  if (node.type === 'tabs') {
    return node.tabs.some((tab) => tab.id === tabId) ? node : null
  }
  for (const child of node.children) {
    const found = findTabGroup(child, tabId)
    if (found) return found
  }
  return null
}

function cloneLayout(node: TabNode): TabNode {
  if (node.type === 'tabs') {
    return {
      ...node,
      tabs: node.tabs.map((tab) => ({
        ...tab,
        props: tab.props ? { ...tab.props } : undefined,
      })),
    }
  }

  return {
    ...node,
    children: node.children.map((child) => cloneLayout(child)),
    sizes: node.sizes ? [...node.sizes] : undefined,
  }
}

export const useWorkspaceStore = defineStore('workspace', () => {
  const layout = ref<TabNode>(createEmptyGroup())
  const activeGroupId = ref(ROOT_GROUP_ID)
  const connectionTabState = ref<Record<string, ConnectionTabState>>({})

  const rootGroup = computed(() => findGroupById(layout.value, activeGroupId.value) ?? findFirstGroup(layout.value))

  function ensureWelcomeTab() {
    if (hasTabs(layout.value)) return
    layout.value = createEmptyGroup()
    openTab({
      id: 'welcome',
      label: 'Welcome',
      closable: false,
      component: markRaw(WelcomeTab),
    }, ROOT_GROUP_ID)
  }

  function setLayout(next: TabNode) {
    layout.value = next
    if (!hasTabs(layout.value)) {
      ensureWelcomeTab()
      return
    }

    const currentGroup = findGroupById(layout.value, activeGroupId.value)
    if (!currentGroup) {
      activeGroupId.value = findFirstGroup(layout.value)?.id ?? ROOT_GROUP_ID
    }
  }

  function setActiveGroup(groupId: string) {
    activeGroupId.value = groupId
  }

  function setConnectionPath(connectionId: string, path = '') {
    connectionTabState.value = {
      ...connectionTabState.value,
      [connectionId]: { path },
    }
  }

  function getConnectionPath(connectionId: string) {
    return connectionTabState.value[connectionId]?.path ?? ''
  }

  function removeWelcomePlaceholder(group: TabGroupNode) {
    if (group.tabs.length === 1 && group.tabs[0]?.id === 'welcome') {
      group.tabs = []
      group.activeId = ''
    }
  }

  function openTab(tab: TabItem, targetGroupId?: string) {
    const existingGroup = findTabGroup(layout.value, tab.id)
    if (existingGroup) {
      existingGroup.activeId = tab.id
      activeGroupId.value = existingGroup.id
      layout.value = cloneLayout(layout.value)
      return
    }

    const targetGroup = findGroupById(layout.value, targetGroupId ?? activeGroupId.value)
      ?? findFirstGroup(layout.value)

    if (!targetGroup) {
      layout.value = createEmptyGroup()
      return openTab(tab, ROOT_GROUP_ID)
    }

    removeWelcomePlaceholder(targetGroup)
    targetGroup.tabs.push(tab)
    targetGroup.activeId = tab.id
    activeGroupId.value = targetGroup.id
    layout.value = cloneLayout(layout.value)
  }

  function openWelcome() {
    ensureWelcomeTab()
  }

  function openSettings() {
    openTab({
      id: 'settings',
      label: 'Settings',
      closable: true,
      component: markRaw(SettingsTab),
    })
  }

  function openNewConnection() {
    openTab({
      id: 'connection-form:new',
      label: 'New Connection',
      closable: true,
      component: markRaw(ConnectionFormTab),
      props: {
        mode: 'create',
      },
    })
  }

  function openEditConnection(connectionId: string, name: string) {
    openTab({
      id: `connection-form:${connectionId}`,
      label: `Edit ${name}`,
      closable: true,
      component: markRaw(ConnectionFormTab),
      props: {
        mode: 'edit',
        connectionId,
      },
    })
  }

  function openConnection(connectionId: string, name: string, path?: string) {
    if (typeof path === 'string') {
      setConnectionPath(connectionId, path)
    } else if (!(connectionId in connectionTabState.value)) {
      setConnectionPath(connectionId, '')
    }

    openTab({
      id: `connection:${connectionId}`,
      label: name,
      closable: true,
      component: markRaw(ConnectionOverviewTab),
      props: {
        connectionId,
      },
    })
  }

  return {
    layout,
    activeGroupId,
    connectionTabState,
    rootGroup,
    setLayout,
    setActiveGroup,
    setConnectionPath,
    getConnectionPath,
    openWelcome,
    openSettings,
    openNewConnection,
    openEditConnection,
    openConnection,
    ensureWelcomeTab,
  }
})
