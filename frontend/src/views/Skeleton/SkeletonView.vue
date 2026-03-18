<template>
  <SkeletonLayout
    v-model:left-active="leftActive"
    v-model:right-active="rightActive"
    :left-sidebar="leftSidebar"
    :right-sidebar="rightSidebar"
    :gap="1"
  >
    <!-- ═══ Left Panel Content ═══ -->
    <template #left-panel="{ activeId }">
      <div class="panel-content">
        <div v-if="activeId === 'explorer'" class="panel-section">
          <div class="panel-header">
            <i-mdi-file-tree class="panel-header__icon" />
            <span>资源管理器</span>
          </div>
          <div class="panel-body">
            <div class="tree-item" v-for="i in 12" :key="i">
              <i-mdi-folder-outline class="tree-icon" />
              <span>文件夹 {{ i }}</span>
            </div>
          </div>
        </div>

        <div v-else-if="activeId === 'search'" class="panel-section">
          <div class="panel-header">
            <i-ep-search class="panel-header__icon" />
            <span>搜索</span>
          </div>
          <div class="panel-body">
            <el-input placeholder="搜索文件..." size="small" style="margin-bottom: 8px" />
            <div class="search-result" v-for="i in 5" :key="i">
              搜索结果 {{ i }}
            </div>
          </div>
        </div>

        <div v-else-if="activeId === 'git'" class="panel-section">
          <div class="panel-header">
            <i-mdi-source-branch class="panel-header__icon" />
            <span>版本控制</span>
          </div>
          <div class="panel-body">
            <div class="git-item">
              <i-mdi-file-document-edit-outline style="color: var(--theme-color-warning)" />
              <span>已修改: app.go</span>
            </div>
            <div class="git-item">
              <i-mdi-file-plus-outline style="color: var(--theme-color-success)" />
              <span>新增: config.go</span>
            </div>
            <div class="git-item">
              <i-mdi-file-remove-outline style="color: var(--theme-color-danger)" />
              <span>已删除: old.go</span>
            </div>
          </div>
        </div>

        <div v-else class="panel-empty">
          <i-mdi-arrow-left style="font-size: 24px" />
          <span>点击左侧按钮打开面板</span>
        </div>
      </div>
    </template>

    <!-- ═══ Center Content ═══ -->
    <template #center>
      <div class="center-content">
        <div class="center-content__header">
          <span class="center-content__title">骨架布局演示</span>
          <span class="center-content__subtitle">
            JetBrains IDEA / GoLand 风格
          </span>
        </div>

        <div class="center-content__info">
          <div class="info-card">
            <div class="info-card__icon">🖥️</div>
            <div class="info-card__title">整体布局</div>
            <div class="info-card__desc">
              左右两侧各有一个纵向按钮列 + 可展开/折叠的菜单面板，中间为主内容区。
            </div>
          </div>

          <div class="info-card">
            <div class="info-card__icon">🔘</div>
            <div class="info-card__title">按钮列</div>
            <div class="info-card__desc">
              按钮分为上方和下方两组，支持"菜单按钮"（点击切换面板）和"动作按钮"（点击触发事件）。
            </div>
          </div>

          <div class="info-card">
            <div class="info-card__icon">↔️</div>
            <div class="info-card__title">拖拽调整</div>
            <div class="info-card__desc">
              菜单面板支持拖拽分割线调整宽度，支持最大宽度限制，双击分割线居中。
            </div>
          </div>

          <div class="info-card">
            <div class="info-card__icon">📌</div>
            <div class="info-card__title">最小化/还原</div>
            <div class="info-card__desc">
              点击已激活的菜单按钮可折叠面板，再次点击或点击其他菜单按钮可还原。
            </div>
          </div>
        </div>

        <div class="center-content__status">
          <span>左侧面板: <strong>{{ leftActive ?? '已折叠' }}</strong></span>
          <span>·</span>
          <span>右侧面板: <strong>{{ rightActive ?? '已折叠' }}</strong></span>
        </div>
      </div>
    </template>

    <!-- ═══ Right Panel Content ═══ -->
    <template #right-panel="{ activeId }">
      <div class="panel-content">
        <div v-if="activeId === 'notifications'" class="panel-section">
          <div class="panel-header">
            <i-ep-bell class="panel-header__icon" />
            <span>通知</span>
          </div>
          <div class="panel-body">
            <div class="notif-item" v-for="i in 4" :key="i">
              <div class="notif-dot" />
              <div>
                <div class="notif-title">通知 {{ i }}</div>
                <div class="notif-desc">这是一条示例通知消息</div>
              </div>
            </div>
          </div>
        </div>

        <div v-else-if="activeId === 'outline'" class="panel-section">
          <div class="panel-header">
            <i-mdi-format-list-bulleted class="panel-header__icon" />
            <span>大纲</span>
          </div>
          <div class="panel-body">
            <div class="outline-item" v-for="i in 8" :key="i" :style="{ paddingLeft: `${(i % 3) * 16 + 8}px` }">
              <i-mdi-function style="color: var(--theme-color-primary)" />
              <span>func_{{ i }}()</span>
            </div>
          </div>
        </div>

        <div v-else class="panel-empty">
          <span>点击右侧按钮打开面板</span>
          <i-mdi-arrow-right style="font-size: 24px" />
        </div>
      </div>
    </template>
  </SkeletonLayout>
</template>

<script setup lang="ts">
import { ref, markRaw } from 'vue'
import SkeletonLayout from './SkeletonLayout.vue'
import type { SidebarConfig } from './types'

// ─── Dynamic icon imports ───────────────────────────────────
// Template auto-import (e.g. <i-mdi-file-tree />) only works for static template usage.
// For data-driven dynamic rendering (<component :is="btn.icon" />), explicit ~icons/* imports
// are the documented unplugin-icons approach.
import IEpSearch from '~icons/ep/search'
import IMdiFileTree from '~icons/mdi/file-tree'
import IMdiSourceBranch from '~icons/mdi/source-branch'
import IEpSetting from '~icons/ep/setting'
import IEpBell from '~icons/ep/bell'
import IMdiFormatListBulleted from '~icons/mdi/format-list-bulleted'
import IMdiInformationOutline from '~icons/mdi/information-outline'

// ─── Left sidebar config ────────────────────────────────────

const leftSidebar: SidebarConfig = {
  topButtons: [
    {
      id: 'explorer',
      icon: markRaw(IMdiFileTree),
      tooltip: '资源管理器',
      type: 'menu',
    },
    {
      id: 'search',
      icon: markRaw(IEpSearch),
      tooltip: '搜索',
      type: 'menu',
    },
    {
      id: 'git',
      icon: markRaw(IMdiSourceBranch),
      tooltip: '版本控制',
      type: 'menu',
    },
  ],
  bottomButtons: [
    {
      id: 'settings',
      icon: markRaw(IEpSetting),
      tooltip: '设置',
      type: 'action',
      onClick: () => {
        console.log('Settings clicked')
      },
    },
  ],
  defaultSize: '240px',
  minSize: '150px',
  maxSize: '400px',
}

// ─── Right sidebar config ───────────────────────────────────

const rightSidebar: SidebarConfig = {
  topButtons: [
    {
      id: 'notifications',
      icon: markRaw(IEpBell),
      tooltip: '通知',
      type: 'menu',
    },
    {
      id: 'outline',
      icon: markRaw(IMdiFormatListBulleted),
      tooltip: '大纲',
      type: 'menu',
    },
  ],
  bottomButtons: [
    {
      id: 'info',
      icon: markRaw(IMdiInformationOutline),
      tooltip: '信息',
      type: 'action',
      onClick: () => {
        console.log('Info clicked')
      },
    },
  ],
  defaultSize: '220px',
  minSize: '120px',
  maxSize: '360px',
}

// ─── State ──────────────────────────────────────────────────

const leftActive = ref<string | null>('explorer')
const rightActive = ref<string | null>(null)
</script>

<style scoped>
/* ─── Panel common ─────────────────────────────────── */

.panel-content {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.panel-section {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.panel-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 12px;
  font-size: 13px;
  font-weight: 600;
  color: var(--theme-color-text-base);
  border-bottom: 1px solid var(--theme-color-border-light);
  flex-shrink: 0;
}

.panel-header__icon {
  width: 16px;
  height: 16px;
  color: var(--theme-color-primary);
}

.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 6px 0;
}

.panel-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 100%;
  color: var(--theme-color-text-placeholder);
  font-size: 13px;
}

/* ─── Tree items ───────────────────────────────────── */

.tree-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  font-size: 13px;
  color: var(--theme-color-text);
  cursor: pointer;
}

.tree-item:hover {
  background-color: var(--theme-color-bg-hover);
}

.tree-icon {
  width: 16px;
  height: 16px;
  color: var(--theme-color-warning);
  flex-shrink: 0;
}

/* ─── Search ───────────────────────────────────────── */

.search-result {
  padding: 6px 12px;
  font-size: 13px;
  color: var(--theme-color-text);
  cursor: pointer;
}

.search-result:hover {
  background-color: var(--theme-color-bg-hover);
}

/* ─── Git ──────────────────────────────────────────── */

.git-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  font-size: 13px;
  color: var(--theme-color-text);
}

/* ─── Notifications ────────────────────────────────── */

.notif-item {
  display: flex;
  gap: 10px;
  padding: 8px 12px;
  align-items: flex-start;
}

.notif-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: var(--theme-color-primary);
  flex-shrink: 0;
  margin-top: 4px;
}

.notif-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--theme-color-text);
}

.notif-desc {
  font-size: 12px;
  color: var(--theme-color-text-secondary);
  margin-top: 2px;
}

/* ─── Outline ──────────────────────────────────────── */

.outline-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  font-size: 13px;
  color: var(--theme-color-text);
  cursor: pointer;
}

.outline-item:hover {
  background-color: var(--theme-color-bg-hover);
}

/* ─── Center content ───────────────────────────────── */

.center-content {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 24px;
  padding: 32px;
  box-sizing: border-box;
  background-color: var(--theme-color-bg-base);
}

.center-content__header {
  text-align: center;
}

.center-content__title {
  display: block;
  font-size: 22px;
  font-weight: 600;
  color: var(--theme-color-text-base);
}

.center-content__subtitle {
  display: block;
  font-size: 13px;
  color: var(--theme-color-text-secondary);
  margin-top: 4px;
}

.center-content__info {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  max-width: 560px;
}

.info-card {
  padding: 16px;
  border-radius: var(--theme-radius-md);
  background-color: var(--theme-color-bg-surface);
  border: 1px solid var(--theme-color-border-light);
}

.info-card__icon {
  font-size: 24px;
  margin-bottom: 8px;
}

.info-card__title {
  font-size: 14px;
  font-weight: 600;
  color: var(--theme-color-text-base);
  margin-bottom: 4px;
}

.info-card__desc {
  font-size: 12px;
  color: var(--theme-color-text-secondary);
  line-height: 1.5;
}

.center-content__status {
  display: flex;
  gap: 8px;
  font-size: 12px;
  color: var(--theme-color-text-secondary);
}

.center-content__status strong {
  color: var(--theme-color-primary);
}
</style>


