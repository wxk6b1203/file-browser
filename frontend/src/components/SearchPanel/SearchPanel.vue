<template>
  <div class="search-panel">
    <div class="search-panel__header">
      <div>
        <p class="search-panel__eyebrow">{{ t('shell.search.eyebrow') }}</p>
        <h2 class="search-panel__title">{{ t('shell.search.title') }}</h2>
      </div>
      <div class="search-panel__actions">
        <el-button text @click="clearSearch">
          <i-ep-delete />
        </el-button>
      </div>
    </div>

    <div class="search-panel__controls">
      <el-input
        v-model="search.query"
        :placeholder="t('shell.search.queryPlaceholder')"
        clearable
        @keyup.enter="runSearch"
      >
        <template #prefix>
          <i-ep-search />
        </template>
      </el-input>

      <el-select
        v-model="search.selectedConnectionIds"
        multiple
        clearable
        collapse-tags
        collapse-tags-tooltip
        :placeholder="t('shell.search.scopePlaceholder')"
      >
        <el-option
          v-for="item in connections.definitions"
          :key="item.id"
          :label="item.name"
          :value="item.id"
        />
      </el-select>

      <div class="search-panel__buttons">
        <el-button type="primary" :loading="search.running" @click="runSearch">
          {{ t('shell.search.run') }}
        </el-button>
        <el-button :disabled="!search.running" @click="search.cancel()">
          {{ t('shell.search.cancel') }}
        </el-button>
      </div>
    </div>

    <div class="search-panel__summary">
      <span v-if="search.running">{{ t('shell.search.running') }}</span>
      <template v-else-if="search.summary">
        <span>{{ t('shell.search.summary.matched', { count: search.summary.matchedCount }) }}</span>
        <span>{{ t('shell.search.summary.scanned', { count: search.summary.scannedCount }) }}</span>
        <span>{{ t('shell.search.summary.connections', { count: search.summary.connectionCount }) }}</span>
        <span>{{ t('shell.search.summary.duration', { count: search.summary.durationMs }) }}</span>
      </template>
      <span v-else>{{ t('shell.search.idle') }}</span>
    </div>

    <div v-if="search.hasErrors" class="search-panel__errors">
      <div v-for="(item, index) in search.errors" :key="`${item.connectionId ?? 'global'}-${index}`" class="search-panel__error">
        <span class="search-panel__error-title">
          {{ item.connectionName || t('shell.search.globalError') }}
        </span>
        <span>{{ item.message }}</span>
      </div>
    </div>

    <div v-if="search.hasResults" class="search-panel__results">
      <button
        v-for="item in search.results"
        :key="`${item.connectionId}:${item.file?.path}`"
        class="search-result"
        @click="openResult(item)"
      >
        <span class="search-result__icon">
          <component :is="resolveFileIcon(item.file, { opened: item.file?.type === DIRECTORY_ENTRY_TYPE })" />
        </span>
        <span class="search-result__main">
          <span class="search-result__name">{{ item.file?.name || item.file?.path }}</span>
          <span class="search-result__path">{{ item.file?.path }}</span>
        </span>
        <span class="search-result__meta">
          <span class="search-result__connection">{{ item.connectionName }}</span>
          <span class="search-result__driver">{{ item.driver }}</span>
        </span>
      </button>
    </div>

    <div v-else class="search-panel__empty">
      <i-ep-search class="search-panel__empty-icon" />
      <h3>{{ t('shell.search.emptyTitle') }}</h3>
      <p>{{ emptyMessage }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { DIRECTORY_ENTRY_TYPE, resolveFileIcon } from '@/composables/useFileIcons'
import { useConnectionsStore } from '@/stores/connections'
import { useSearchStore, type SearchResultItem } from '@/stores/search'
import { useWorkspaceStore } from '@/stores/workspace'

const { t } = useI18n()
const connections = useConnectionsStore()
const search = useSearchStore()
const workspace = useWorkspaceStore()

const emptyMessage = computed(() => {
  if (search.running) return t('shell.search.placeholder')
  if (search.ready && !search.hasResults) return t('shell.search.emptyResult')
  return t('shell.search.placeholder')
})

function parentDir(path: string) {
  const normalized = path.replace(/\\/g, '/').replace(/^\/+|\/+$/g, '')
  if (!normalized.includes('/')) return ''
  return normalized.split('/').slice(0, -1).join('/')
}

function openResult(item: SearchResultItem) {
  if (!item.file) return
  const targetPath = item.file.type === DIRECTORY_ENTRY_TYPE ? item.file.path : parentDir(item.file.path)
  workspace.setConnectionRevealPath(item.connectionId, item.file.path)
  workspace.openConnection(item.connectionId, item.connectionName, targetPath)
}

function clearSearch() {
  search.query = ''
  search.selectedConnectionIds = []
  search.reset()
}

function runSearch() {
  return search.search()
}

onMounted(async () => {
  await connections.hydrate()
  search.ensureSubscribed()
})
</script>

<style scoped>
.search-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--theme-color-bg-overlay) 50%, transparent), transparent 32%);
}

.search-panel__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 14px 10px;
  border-bottom: 1px solid var(--theme-color-border-light);
}

.search-panel__eyebrow {
  margin: 0 0 4px;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.14em;
  color: var(--theme-color-text-secondary);
}

.search-panel__title {
  margin: 0;
  font-size: 15px;
  color: var(--theme-color-text-base);
}

.search-panel__controls {
  display: grid;
  gap: 10px;
  padding: 12px 14px;
  border-bottom: 1px solid var(--theme-color-border-light);
}

.search-panel__buttons {
  display: flex;
  gap: 8px;
}

.search-panel__summary {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
  padding: 10px 14px;
  color: var(--theme-color-text-secondary);
  font-size: 12px;
  border-bottom: 1px solid var(--theme-color-border-light);
}

.search-panel__errors {
  display: grid;
  gap: 8px;
  padding: 12px 14px 0;
}

.search-panel__error {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 10px 12px;
  border-radius: 12px;
  background: color-mix(in srgb, var(--theme-color-danger) 10%, transparent);
  color: var(--theme-color-text-base);
  font-size: 12px;
}

.search-panel__error-title {
  font-weight: 700;
}

.search-panel__results {
  flex: 1;
  overflow: auto;
  padding: 10px 8px 14px;
}

.search-result {
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr) auto;
  align-items: center;
  gap: 10px;
  width: 100%;
  min-width: 0;
  margin-bottom: 6px;
  padding: 10px 12px;
  border: 1px solid var(--theme-color-border-light);
  border-radius: 12px;
  background: color-mix(in srgb, var(--theme-color-bg-surface) 88%, transparent);
  color: var(--theme-color-text);
  text-align: left;
  cursor: pointer;
}

.search-result:hover {
  background: var(--theme-color-bg-hover);
}

.search-result__icon {
  font-size: 18px;
  color: var(--theme-color-primary);
}

.search-result__main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.search-result__name,
.search-result__path {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-result__name {
  color: var(--theme-color-text-base);
  font-weight: 600;
}

.search-result__path {
  color: var(--theme-color-text-secondary);
  font-size: 12px;
}

.search-result__meta {
  display: flex;
  min-width: 72px;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
  color: var(--theme-color-text-secondary);
  font-size: 11px;
}

.search-result__driver {
  padding: 1px 6px;
  border-radius: 999px;
  background: var(--theme-color-bg-overlay);
}

.search-panel__empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 18px;
  color: var(--theme-color-text-secondary);
  text-align: center;
}

.search-panel__empty h3 {
  margin: 0;
  color: var(--theme-color-text-base);
  font-size: 15px;
}

.search-panel__empty p {
  margin: 0;
  max-width: 240px;
}

.search-panel__empty-icon {
  font-size: 22px;
  color: var(--theme-color-primary);
}
</style>
