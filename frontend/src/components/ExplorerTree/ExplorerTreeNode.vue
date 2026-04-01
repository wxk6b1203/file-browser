<template>
  <div class="explorer-node">
    <div
      class="explorer-node__row"
      :style="{ paddingLeft: `${8 + node.level * 16}px` }"
      @dblclick="onOpen(node)"
    >
      <button class="explorer-node__toggle" @click.stop="emit('toggle', node)">
        <i-ep-arrow-right
          class="explorer-node__toggle-icon"
          :class="{ 'explorer-node__toggle-icon--open': expanded }"
        />
      </button>

      <button class="explorer-node__main" @click="onActivate(node)">
        <span class="explorer-node__icon">
          <i-mdi-server-network-outline v-if="node.kind === 'connection'" />
          <i-mdi-folder-open v-else-if="expanded" />
          <i-mdi-folder v-else />
        </span>

        <span class="explorer-node__label">{{ node.label }}</span>

        <span v-if="node.kind === 'connection' && node.driver" class="explorer-node__driver">
          {{ node.driver }}
        </span>

        <span
          v-if="node.kind === 'connection'"
          class="explorer-node__status"
          :class="{ 'explorer-node__status--online': node.connected }"
        >
          {{ node.connected ? t('shell.explorer.connected') : t('shell.explorer.saved') }}
        </span>
      </button>

      <el-dropdown
        v-if="node.kind === 'connection'"
        trigger="click"
        @command="emit('command', String($event), node)"
      >
        <button class="explorer-node__menu-btn">
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

      <ExplorerTreeNode
        v-for="child in node.children"
        v-else
        :key="child.key"
        :node="child"
        :expanded-keys="expandedKeys"
        :loading-keys="loadingKeys"
        @toggle="forwardToggle"
        @activate="forwardActivate"
        @open="forwardOpen"
        @command="forwardCommand"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ExplorerNode } from './types'

const props = defineProps<{
  node: ExplorerNode
  expandedKeys: Set<string>
  loadingKeys: Set<string>
}>()

const emit = defineEmits<{
  toggle: [node: ExplorerNode]
  activate: [node: ExplorerNode]
  open: [node: ExplorerNode]
  command: [command: string, node: ExplorerNode]
}>()

const { t } = useI18n()

const expanded = computed(() => props.expandedKeys.has(props.node.key))
const loading = computed(() => props.loadingKeys.has(props.node.key))

function onOpen(node: ExplorerNode) {
  emit('open', node)
}

function onActivate(node: ExplorerNode) {
  if (node.kind === 'connection') {
    emit('toggle', node)
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
</script>

<style scoped>
.explorer-node__row {
  display: flex;
  align-items: center;
  gap: 4px;
  min-height: 34px;
  border-radius: 10px;
}

.explorer-node__row:hover {
  background: var(--theme-color-bg-hover);
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
  font-size: 11px;
}

.explorer-node__status {
  margin-left: auto;
  color: var(--theme-color-text-secondary);
  font-size: 11px;
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

.explorer-node__children {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.explorer-node__loading {
  padding: 6px 8px 6px 32px;
  color: var(--theme-color-text-secondary);
  font-size: 12px;
}

.explorer-node__loading--muted {
  opacity: 0.7;
}
</style>
