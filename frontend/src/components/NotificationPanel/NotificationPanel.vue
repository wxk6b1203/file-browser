<template>
  <div class="notification-panel">
    <div class="notification-panel__header">
      <div>
        <p class="notification-panel__eyebrow">{{ t('shell.notifications.eyebrow') }}</p>
        <h2 class="notification-panel__title">{{ t('shell.notifications.title') }}</h2>
      </div>
      <div class="notification-panel__actions">
        <el-button text :disabled="notifications.items.length === 0" @click="notifications.clear()">
          <i-ep-delete />
        </el-button>
      </div>
    </div>

    <div class="notification-panel__summary">
      <span>{{ t('shell.notifications.summary.total', { count: notifications.items.length }) }}</span>
    </div>

    <div v-if="notifications.items.length === 0" class="notification-panel__empty">
      <i-ep-bell class="notification-panel__empty-icon" />
      <h3>{{ t('shell.notifications.emptyTitle') }}</h3>
      <p>{{ t('shell.notifications.placeholder') }}</p>
    </div>

    <div v-else class="notification-panel__list">
      <article v-for="item in notifications.items" :key="item.id" class="notification-card">
        <div class="notification-card__top">
          <div class="notification-card__meta">
            <span class="notification-card__level" :class="levelClass(item.level)">
              {{ levelLabel(item.level) }}
            </span>
            <span class="notification-card__source">{{ item.source }}</span>
          </div>
          <button class="notification-card__remove" @click="notifications.remove(item.id)">
            <i-ep-close />
          </button>
        </div>

        <h3 class="notification-card__title">{{ item.title }}</h3>
        <p class="notification-card__message">{{ item.message }}</p>
        <div v-if="item.action" class="notification-card__actions">
          <el-button size="small" @click="runAction(item)">
            {{ actionLabel(item.action) }}
          </el-button>
        </div>
        <time class="notification-card__time">{{ formatTime(item.createdAt) }}</time>
      </article>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useConnectionsStore } from '@/stores/connections'
import { useNotificationsStore, type NotificationAction, type NotificationLevel, type NotificationItem } from '@/stores/notifications'
import { useShellStore } from '@/stores/shell'
import { useWorkspaceStore } from '@/stores/workspace'

const { t, locale } = useI18n()
const connections = useConnectionsStore()
const notifications = useNotificationsStore()
const shell = useShellStore()
const workspace = useWorkspaceStore()

function levelLabel(level: NotificationLevel) {
  if (level === 'success') return t('shell.notifications.level.success')
  if (level === 'warning') return t('shell.notifications.level.warning')
  if (level === 'error') return t('shell.notifications.level.error')
  return t('shell.notifications.level.info')
}

function levelClass(level: NotificationLevel) {
  return {
    'notification-card__level--info': level === 'info',
    'notification-card__level--success': level === 'success',
    'notification-card__level--warning': level === 'warning',
    'notification-card__level--error': level === 'error',
  }
}

function actionLabel(action: NotificationAction) {
  if (action.kind === 'open-task-panel') return t('shell.notifications.actions.openTasks')
  return t('shell.notifications.actions.openDirectory')
}

async function runAction(item: NotificationItem) {
  const action = item.action
  if (!action) return

  try {
    if (action.kind === 'open-task-panel') {
      shell.showTasks()
      return
    }

    const connectionId = action.connectionId
    const definition = connections.definitionMap.get(connectionId)
    const connectionName = action.connectionName || definition?.name
    if (!connectionName) {
      throw new Error(`connection ${connectionId} not found`)
    }

    shell.showExplorer()
    workspace.openConnection(connectionId, connectionName, action.path || '')
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : String(error))
  }
}

function formatTime(value: number) {
  return new Date(value).toLocaleString(locale.value)
}
</script>

<style scoped>
.notification-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--theme-color-bg-overlay) 50%, transparent), transparent 32%);
}

.notification-panel__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 14px 10px;
  border-bottom: 1px solid var(--theme-color-border-light);
}

.notification-panel__eyebrow {
  margin: 0 0 4px;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.14em;
  color: var(--theme-color-text-secondary);
}

.notification-panel__title {
  margin: 0;
  font-size: 15px;
  color: var(--theme-color-text-base);
}

.notification-panel__summary {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
  padding: 10px 14px;
  color: var(--theme-color-text-secondary);
  font-size: 12px;
  border-bottom: 1px solid var(--theme-color-border-light);
}

.notification-panel__empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 18px;
  text-align: center;
  color: var(--theme-color-text-secondary);
}

.notification-panel__empty h3 {
  margin: 0;
  color: var(--theme-color-text-base);
  font-size: 15px;
}

.notification-panel__empty p {
  margin: 0;
  max-width: 220px;
}

.notification-panel__empty-icon {
  font-size: 22px;
  color: var(--theme-color-primary);
}

.notification-panel__list {
  flex: 1;
  overflow: auto;
  padding: 10px 8px 14px;
  display: grid;
  gap: 8px;
}

.notification-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 14px;
  background: color-mix(in srgb, var(--theme-color-bg-surface) 88%, transparent);
}

.notification-card__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.notification-card__meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.notification-card__level {
  padding: 1px 6px;
  border-radius: 999px;
  font-size: 11px;
  background: var(--theme-color-bg-overlay);
  color: var(--theme-color-text-secondary);
}

.notification-card__level--info {
  color: var(--theme-color-primary);
}

.notification-card__level--success {
  color: var(--theme-color-success);
}

.notification-card__level--warning {
  color: var(--theme-color-warning);
}

.notification-card__level--error {
  color: var(--theme-color-danger);
}

.notification-card__source {
  font-size: 11px;
  color: var(--theme-color-text-secondary);
}

.notification-card__remove {
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--theme-color-text-secondary);
  cursor: pointer;
}

.notification-card__remove:hover {
  background: var(--theme-color-bg-hover);
}

.notification-card__title {
  margin: 0;
  color: var(--theme-color-text-base);
  font-size: 14px;
}

.notification-card__message {
  margin: 0;
  color: var(--theme-color-text);
  font-size: 12px;
  line-height: 1.5;
  word-break: break-word;
}

.notification-card__actions {
  display: flex;
  justify-content: flex-start;
}

.notification-card__time {
  color: var(--theme-color-text-secondary);
  font-size: 11px;
}
</style>
