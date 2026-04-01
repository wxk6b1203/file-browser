<template>
  <div class="task-panel">
    <div class="task-panel__header">
      <div>
        <p class="task-panel__eyebrow">{{ t('shell.tasks.eyebrow') }}</p>
        <h2 class="task-panel__title">{{ t('shell.tasks.title') }}</h2>
      </div>
      <div class="task-panel__actions">
        <el-button text @click="tasks.refresh()">
          <i-ep-refresh-right />
        </el-button>
        <el-button text :disabled="tasks.finishedTasks.length === 0" @click="tasks.clearFinished()">
          <i-ep-delete />
        </el-button>
      </div>
    </div>

    <div class="task-panel__summary">
      <span>{{ t('shell.tasks.summary.active', { count: tasks.activeTasks.length }) }}</span>
      <span>{{ t('shell.tasks.summary.finished', { count: tasks.finishedTasks.length }) }}</span>
      <span>{{ t('shell.tasks.summary.total', { count: tasks.tasks.length }) }}</span>
    </div>

    <div v-if="tasks.tasks.length === 0" class="task-panel__empty">
      <i-mdi-tray-arrow-up class="task-panel__empty-icon" />
      <h3>{{ t('shell.tasks.emptyTitle') }}</h3>
      <p>{{ t('shell.tasks.placeholder') }}</p>
    </div>

    <div v-else class="task-panel__list">
      <article v-for="task in tasks.tasks" :key="task.id" class="task-card">
        <div class="task-card__top">
          <div class="task-card__meta">
            <span class="task-card__direction">
              {{ task.direction === DOWNLOAD_DIRECTION ? t('shell.tasks.download') : t('shell.tasks.upload') }}
            </span>
            <span class="task-card__driver">{{ task.driverName }}</span>
          </div>
          <span class="task-card__status" :class="statusClass(task.status)">
            {{ statusLabel(task.status) }}
          </span>
        </div>

        <div class="task-card__path">
          <span class="task-card__path-label">{{ t('shell.tasks.remotePath') }}</span>
          <span class="task-card__path-value">{{ task.remotePath }}</span>
        </div>

        <div class="task-card__path">
          <span class="task-card__path-label">{{ t('shell.tasks.localPath') }}</span>
          <span class="task-card__path-value">{{ task.localPath }}</span>
        </div>

        <div class="task-card__progress">
          <el-progress :percentage="progressPercent(task)" :stroke-width="8" :show-text="false" />
          <div class="task-card__progress-meta">
            <span>{{ formatTransferred(task) }}</span>
            <span>{{ formatSpeed(task.bytesPerSecond) }}</span>
          </div>
        </div>

        <p v-if="task.error" class="task-card__error">{{ task.error }}</p>

        <div class="task-card__actions">
          <el-button
            v-if="task.status === PENDING_STATUS || task.status === RUNNING_STATUS"
            size="small"
            @click="tasks.cancelTask(task.id)"
          >
            {{ t('shell.tasks.cancel') }}
          </el-button>
          <el-button
            v-else
            size="small"
            @click="tasks.removeTask(task.id)"
          >
            {{ t('shell.tasks.remove') }}
          </el-button>
        </div>
      </article>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { folder } from '../../../wailsjs/go/models'
import { useTasksStore } from '@/stores/tasks'

const PENDING_STATUS = 0
const RUNNING_STATUS = 1
const COMPLETED_STATUS = 2
const FAILED_STATUS = 3
const CANCELLED_STATUS = 4
const DOWNLOAD_DIRECTION = 2

const { t } = useI18n()
const tasks = useTasksStore()

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

function progressPercent(task: folder.TransferTask) {
  if (!task.totalBytes || task.totalBytes <= 0) return 0
  return Math.max(0, Math.min(100, Math.round((task.bytesTransferred / task.totalBytes) * 100)))
}

function formatTransferred(task: folder.TransferTask) {
  if (!task.totalBytes || task.totalBytes <= 0) {
    return formatSize(task.bytesTransferred)
  }
  return `${formatSize(task.bytesTransferred)} / ${formatSize(task.totalBytes)}`
}

function formatSpeed(bytesPerSecond: number) {
  if (!bytesPerSecond) return t('shell.tasks.waiting')
  return `${formatSize(bytesPerSecond)}/s`
}

function statusLabel(status: number) {
  if (status === PENDING_STATUS) return t('shell.tasks.status.pending')
  if (status === RUNNING_STATUS) return t('shell.tasks.status.running')
  if (status === COMPLETED_STATUS) return t('shell.tasks.status.completed')
  if (status === FAILED_STATUS) return t('shell.tasks.status.failed')
  if (status === CANCELLED_STATUS) return t('shell.tasks.status.cancelled')
  return t('shell.tasks.status.unknown')
}

function statusClass(status: number) {
  return {
    'task-card__status--running': status === PENDING_STATUS || status === RUNNING_STATUS,
    'task-card__status--success': status === COMPLETED_STATUS,
    'task-card__status--failed': status === FAILED_STATUS,
    'task-card__status--cancelled': status === CANCELLED_STATUS,
  }
}
</script>

<style scoped>
.task-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--theme-color-bg-overlay) 50%, transparent), transparent 32%);
}

.task-panel__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 14px 10px;
  border-bottom: 1px solid var(--theme-color-border-light);
}

.task-panel__eyebrow {
  margin: 0 0 4px;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.14em;
  color: var(--theme-color-text-secondary);
}

.task-panel__title {
  margin: 0;
  font-size: 15px;
  color: var(--theme-color-text-base);
}

.task-panel__summary {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
  padding: 10px 14px;
  color: var(--theme-color-text-secondary);
  font-size: 12px;
  border-bottom: 1px solid var(--theme-color-border-light);
}

.task-panel__empty {
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

.task-panel__empty h3 {
  margin: 0;
  font-size: 15px;
  color: var(--theme-color-text-base);
}

.task-panel__empty p {
  margin: 0;
  max-width: 220px;
}

.task-panel__empty-icon {
  font-size: 24px;
  color: var(--theme-color-primary);
}

.task-panel__list {
  flex: 1;
  overflow: auto;
  padding: 10px 8px 14px;
  display: grid;
  gap: 8px;
}

.task-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 14px;
  background: color-mix(in srgb, var(--theme-color-bg-surface) 88%, transparent);
}

.task-card__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.task-card__meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.task-card__direction {
  font-weight: 700;
  color: var(--theme-color-text-base);
}

.task-card__driver {
  padding: 1px 6px;
  border-radius: 999px;
  background: var(--theme-color-bg-overlay);
  color: var(--theme-color-text-secondary);
  font-size: 11px;
}

.task-card__status {
  font-size: 12px;
  color: var(--theme-color-text-secondary);
}

.task-card__status--running {
  color: var(--theme-color-primary);
}

.task-card__status--success {
  color: var(--theme-color-success);
}

.task-card__status--failed {
  color: var(--theme-color-danger);
}

.task-card__status--cancelled {
  color: var(--theme-color-warning);
}

.task-card__path {
  display: grid;
  gap: 4px;
}

.task-card__path-label {
  font-size: 11px;
  color: var(--theme-color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.task-card__path-value {
  word-break: break-all;
  color: var(--theme-color-text-base);
  font-size: 12px;
}

.task-card__progress {
  display: grid;
  gap: 6px;
}

.task-card__progress-meta {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  font-size: 12px;
  color: var(--theme-color-text-secondary);
}

.task-card__error {
  margin: 0;
  color: var(--theme-color-danger);
  font-size: 12px;
}

.task-card__actions {
  display: flex;
  justify-content: flex-end;
}
</style>
