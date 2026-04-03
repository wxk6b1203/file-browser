import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { CancelTransferTask, ClearFinishedTransferTasks, ListTransferTasks, RemoveTransferTask } from '../../wailsjs/go/main/App'
import { folder } from '../../wailsjs/go/models'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { emitConnectionDirectoryRefresh } from '@/composables/useConnectionDirectoryRefresh'
import { useNotificationsStore } from './notifications'

const TRANSFER_EVENT = 'transfer:event'
const UPLOAD_DIRECTION = 1
const COMPLETED_STATUS = 2

interface TransferEventPayload {
  type?: 'upsert' | 'remove' | 'error'
  taskId?: string
  task?: folder.TransferTask | null
  message?: string
}

let unsubscribeTransferEvents: (() => void) | null = null

export const useTasksStore = defineStore('tasks', () => {
  const notifications = useNotificationsStore()
  const ready = ref(false)
  const loading = ref(false)
  const tasks = ref<folder.TransferTask[]>([])
  const notifiedTaskIds = ref<Record<string, true>>({})
  const refreshedUploadTaskIds = ref<Record<string, true>>({})

  const activeTasks = computed(() => tasks.value.filter((item) => item.status === 0 || item.status === 1))
  const finishedTasks = computed(() => tasks.value.filter((item) => item.status >= 2))

  function ensureSubscribed() {
    if (unsubscribeTransferEvents) return
    unsubscribeTransferEvents = EventsOn(TRANSFER_EVENT, (payload: TransferEventPayload) => {
      handleEvent(payload)
    })
    void refresh(false)
  }

  async function refresh(shouldNotify = true) {
    if (loading.value) return
    loading.value = true
    try {
      const previousReady = ready.value
      const nextTasks = sortTasks(await ListTransferTasks())
      if (shouldNotify && previousReady) {
        notifyTaskTransitions(nextTasks)
      }
      tasks.value = nextTasks
      ready.value = true
    } finally {
      loading.value = false
    }
  }

  async function cancelTask(taskId: string) {
    await CancelTransferTask(taskId)
  }

  async function removeTask(taskId: string) {
    await RemoveTransferTask(taskId)
  }

  async function clearFinished() {
    await ClearFinishedTransferTasks()
  }

  function handleEvent(payload?: TransferEventPayload) {
    if (!payload?.type) return

    ready.value = true

    if (payload.type === 'remove' && payload.taskId) {
      tasks.value = tasks.value.filter((item) => item.id !== payload.taskId)
      return
    }

    if (payload.type === 'error' && payload.message) {
      notifications.push({
        level: 'error',
        source: 'Transfer',
        title: 'Transfer Follow-up Failed',
        message: payload.message,
        action: {
          kind: 'open-task-panel',
          taskId: payload.taskId,
        },
      })
      return
    }

    if (payload.type === 'upsert' && payload.task) {
      const nextTask = folder.TransferTask.createFrom(payload.task)
      const existingIndex = tasks.value.findIndex((item) => item.id === nextTask.id)
      const nextTasks = [...tasks.value]

      if (existingIndex >= 0) {
        nextTasks[existingIndex] = nextTask
      } else {
        nextTasks.push(nextTask)
      }

      tasks.value = sortTasks(nextTasks)
      triggerDirectoryRefresh(nextTask)
      notifyTaskTransitions([nextTask])
    }
  }

  function triggerDirectoryRefresh(task: folder.TransferTask) {
    if (!task?.id || refreshedUploadTaskIds.value[task.id]) return
    if (task.status !== COMPLETED_STATUS || task.direction !== UPLOAD_DIRECTION) return
    if (!task.instanceName) return

    refreshedUploadTaskIds.value = {
      ...refreshedUploadTaskIds.value,
      [task.id]: true,
    }

    emitConnectionDirectoryRefresh({
      connectionId: task.instanceName,
      path: parentRemoteDir(task.remotePath),
      source: 'transfer',
      taskId: task.id,
    })
  }

  function notifyTaskTransitions(nextTasks: folder.TransferTask[]) {
    for (const task of nextTasks) {
      if (!task?.id || notifiedTaskIds.value[task.id]) continue

      if (task.status === 3) {
        notifications.push({
          level: 'error',
          source: task.instanceName || task.driverName || 'Transfer',
          title: 'Transfer Failed',
          message: task.error || `${task.remotePath} failed`,
          action: {
            kind: 'open-task-panel',
            taskId: task.id,
          },
        })
        notifiedTaskIds.value = {
          ...notifiedTaskIds.value,
          [task.id]: true,
        }
        continue
      }

      if (task.status === 4) {
        notifications.push({
          level: 'warning',
          source: task.instanceName || task.driverName || 'Transfer',
          title: 'Transfer Cancelled',
          message: task.remotePath || task.localPath,
          action: {
            kind: 'open-task-panel',
            taskId: task.id,
          },
        })
        notifiedTaskIds.value = {
          ...notifiedTaskIds.value,
          [task.id]: true,
        }
      }
    }
  }

  function sortTasks(items: folder.TransferTask[]) {
    return [...items].sort((left, right) => toTimestamp(right.createdAt) - toTimestamp(left.createdAt))
  }

  function toTimestamp(value: unknown) {
    if (value instanceof Date) return value.getTime()
    if (typeof value === 'string' || typeof value === 'number') {
      const timestamp = new Date(value).getTime()
      return Number.isNaN(timestamp) ? 0 : timestamp
    }
    return 0
  }

  function parentRemoteDir(remotePath?: string) {
    const normalized = String(remotePath ?? '').replace(/\\/g, '/').trim().replace(/^\/+|\/+$/g, '')
    if (!normalized) return ''

    const segments = normalized.split('/').filter(Boolean)
    if (segments.length <= 1) return ''
    return segments.slice(0, -1).join('/')
  }

  return {
    ready,
    loading,
    tasks,
    activeTasks,
    finishedTasks,
    ensureSubscribed,
    refresh,
    cancelTask,
    removeTask,
    clearFinished,
  }
})
