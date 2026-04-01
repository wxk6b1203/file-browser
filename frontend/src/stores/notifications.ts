import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

export type NotificationLevel = 'info' | 'success' | 'warning' | 'error'

export interface NotificationItem {
  id: string
  level: NotificationLevel
  title: string
  message: string
  source: string
  createdAt: number
}

function createNotificationId() {
  return `notification:${Date.now()}:${Math.random().toString(36).slice(2, 10)}`
}

export const useNotificationsStore = defineStore('notifications', () => {
  const items = ref<NotificationItem[]>([])

  const unreadCount = computed(() => items.value.length)

  function push(input: Omit<NotificationItem, 'id' | 'createdAt'>) {
    items.value = [
      {
        id: createNotificationId(),
        createdAt: Date.now(),
        ...input,
      },
      ...items.value,
    ].slice(0, 200)
  }

  function remove(id: string) {
    items.value = items.value.filter((item) => item.id !== id)
  }

  function clear() {
    items.value = []
  }

  return {
    items,
    unreadCount,
    push,
    remove,
    clear,
  }
})
