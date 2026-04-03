import { ref } from 'vue'
import { defineStore } from 'pinia'

export type LeftSidebarPanel = 'explorer' | 'search' | null
export type RightSidebarPanel = 'tasks' | 'notifications' | null

export const useShellStore = defineStore('shell', () => {
  const leftActive = ref<LeftSidebarPanel>('explorer')
  const rightActive = ref<RightSidebarPanel>(null)

  function showExplorer() {
    leftActive.value = 'explorer'
  }

  function showSearch() {
    leftActive.value = 'search'
  }

  function showTasks() {
    rightActive.value = 'tasks'
  }

  function showNotifications() {
    rightActive.value = 'notifications'
  }

  return {
    leftActive,
    rightActive,
    showExplorer,
    showSearch,
    showTasks,
    showNotifications,
  }
})
