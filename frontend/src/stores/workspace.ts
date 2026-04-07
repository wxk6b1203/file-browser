import { computed, defineAsyncComponent, markRaw, ref } from 'vue'
import { defineStore } from 'pinia'
import type { TabGroupNode, TabItem, TabNode } from '@/components/Tabs'
import { folder } from '../../wailsjs/go/models'

const WelcomeTab = markRaw(defineAsyncComponent(() => import('@/components/Workspace/WelcomeTab.vue')))
const ConnectionFormTab = markRaw(defineAsyncComponent(() => import('@/components/Workspace/ConnectionFormTab.vue')))
const SettingsTab = markRaw(defineAsyncComponent(() => import('@/components/Workspace/SettingsTab.vue')))
const ConnectionOverviewTab = markRaw(defineAsyncComponent(() => import('@/components/Workspace/ConnectionOverviewTab.vue')))

const ROOT_GROUP_ID = 'workspace-root'

interface ConnectionTabState {
  path: string
  items?: folder.FileInfo[]
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

function removeTabsByPredicate(node: TabNode, predicate: (tab: TabItem) => boolean): TabNode | null {
  if (node.type === 'tabs') {
    const tabs = node.tabs.filter((tab) => !predicate(tab))
    if (tabs.length === 0) return null

    const activeId = tabs.some((tab) => tab.id === node.activeId)
      ? node.activeId
      : (tabs[0]?.id ?? '')

    return {
      ...node,
      tabs,
      activeId,
    }
  }

  const children = node.children
    .map((child) => removeTabsByPredicate(child, predicate))
    .filter((child): child is TabNode => child !== null)

  if (children.length === 0) return null
  if (children.length === 1) return children[0]!

  return {
    ...node,
    children,
    sizes: node.sizes ? node.sizes.slice(0, children.length) : undefined,
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
      closable: true,
      component: markRaw(WelcomeTab),
    }, ROOT_GROUP_ID)
  }

  function setLayout(next: TabNode) {
    layout.value = next
    const currentGroup = findGroupById(layout.value, activeGroupId.value)
    if (!currentGroup) {
      activeGroupId.value = findFirstGroup(layout.value)?.id ?? ROOT_GROUP_ID
    }
  }

  function setActiveGroup(groupId: string) {
    activeGroupId.value = groupId
  }

  function isTabActiveInActiveGroup(tabId: string) {
    const group = findGroupById(layout.value, activeGroupId.value)
    return group?.activeId === tabId
  }

  function setConnectionPath(connectionId: string, path = '') {
    connectionTabState.value = {
      ...connectionTabState.value,
      [connectionId]: {
        ...connectionTabState.value[connectionId],
        path,
      },
    }
  }

  function getConnectionPath(connectionId: string) {
    return connectionTabState.value[connectionId]?.path ?? ''
  }

  function setConnectionBrowserState(connectionId: string, path: string, items: folder.FileInfo[]) {
    connectionTabState.value = {
      ...connectionTabState.value,
      [connectionId]: {
        path,
        items: items.map((item) => folder.FileInfo.createFrom(JSON.parse(JSON.stringify(item)))),
      },
    }
  }

  function getConnectionBrowserState(connectionId: string) {
    const state = connectionTabState.value[connectionId]
    if (!state?.items) return null

    return {
      path: state.path,
      items: state.items.map((item) => folder.FileInfo.createFrom(JSON.parse(JSON.stringify(item)))),
    }
  }

  function resetConnectionBrowserState(connectionId: string, path = '') {
    const nextState = { ...connectionTabState.value }
    delete nextState[connectionId]
    connectionTabState.value = nextState
    setConnectionPath(connectionId, path)
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

  function closeTabById(tabId: string) {
    const nextLayout = removeTabsByPredicate(layout.value, (tab) => tab.id === tabId)
    layout.value = nextLayout ?? createEmptyGroup()

    const currentGroup = findGroupById(layout.value, activeGroupId.value)
    if (!currentGroup) {
      activeGroupId.value = findFirstGroup(layout.value)?.id ?? ROOT_GROUP_ID
    }
  }

  function closeConnectionTabs(connectionId: string) {
    const overviewTabId = `connection:${connectionId}`
    const formTabId = `connection-form:${connectionId}`
    const nextLayout = removeTabsByPredicate(layout.value, (tab) => (
      tab.id === overviewTabId || tab.id === formTabId
    ))

    layout.value = nextLayout ?? createEmptyGroup()

    const nextState = { ...connectionTabState.value }
    delete nextState[connectionId]
    connectionTabState.value = nextState

    const currentGroup = findGroupById(layout.value, activeGroupId.value)
    if (!currentGroup) {
      activeGroupId.value = findFirstGroup(layout.value)?.id ?? ROOT_GROUP_ID
    }
  }

  return {
    layout,
    activeGroupId,
    connectionTabState,
    rootGroup,
    setLayout,
    setActiveGroup,
    isTabActiveInActiveGroup,
    setConnectionPath,
    getConnectionPath,
    setConnectionBrowserState,
    getConnectionBrowserState,
    resetConnectionBrowserState,
    openWelcome,
    openSettings,
    openNewConnection,
    openEditConnection,
    openConnection,
    closeTabById,
    closeConnectionTabs,
    ensureWelcomeTab,
  }
})
