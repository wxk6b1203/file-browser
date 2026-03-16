<template>
  <div class="home-view">
    <!-- Toolbar -->
    <div class="home-view__toolbar">
      <el-button size="small" @click="addTab">+ 添加标签</el-button>
      <el-button size="small" @click="resetLayout">重置布局</el-button>
      <span class="home-view__hint">
        拖拽标签头排序 · 拖到内容区上/下/左/右分屏 · 关闭标签自动合并
      </span>

      <!-- 主题切换 -->
      <div class="home-view__toolbar-spacer" />
      <el-dropdown trigger="click" @command="onThemeCommand">
        <el-button size="small">
          {{ currentTheme.dark ? '🌙' : '☀️' }} {{ currentTheme.label }}
          <el-icon class="el-icon--right"><i-ep-arrow-down /></el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item
              v-for="t in themes"
              :key="t.id"
              :command="t.id"
              :class="{ 'is-active': t.id === resolvedTheme }"
            >
              {{ t.dark ? '🌙' : '☀️' }} {{ t.label }}
            </el-dropdown-item>
            <el-dropdown-item divided :command="SYSTEM_THEME" :class="{ 'is-active': mode === SYSTEM_THEME }">
              🖥️ 跟随系统
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <!-- Tabs -->
    <div style="flex: 1; overflow: hidden;">
      <Tabs
        v-model="layout"
        :overlay-opacity="0.18"
        :min-split-width="120"
        :min-split-height="100"
        @tab-drag-start="onDragStart"
        @tab-drag-end="onDragEnd"
        @tab-reorder="onReorder"
        @tab-split="onSplit"
      />
    </div>

    <!-- Event log -->
    <div class="home-view__log">
      <div v-for="(log, i) in logs" :key="i" class="home-view__log-item">{{ log }}</div>
      <div v-if="logs.length === 0" class="home-view__log-empty">事件日志将显示在这里...</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, defineComponent, h, markRaw } from 'vue'
import { Tabs, genId, type TabNode, type TabItem } from '@/components/Tabs'
import { useTheme } from '@/composables/useTheme'
import { useShortcutMap } from '@/composables/useShortcut'

const { mode, themes, resolvedTheme, currentTheme, isDark, setTheme, SYSTEM_THEME } = useTheme()

function onThemeCommand(command: string) {
  setTheme(command)
}

// ─── Sample content components ──────────────────────────────

const COLORS = ['#e8f5e9', '#e3f2fd', '#fff3e0', '#fce4ec', '#f3e5f5', '#e0f7fa', '#fff9c4', '#efebe9']

const DemoPanel = markRaw(defineComponent({
  name: 'DemoPanel',
  props: {
    title: { type: String, default: 'Panel' },
    color: { type: String, default: '#f5f5f5' },
  },
  setup(props) {
    return () =>
      h(
        'div',
        {
          style: {
            width: '100%',
            height: '100%',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            backgroundColor: props.color,
            fontSize: '18px',
            fontWeight: '500',
            color: 'var(--theme-color-text)',
            userSelect: 'none',
            gap: '8px',
          },
        },
        [
          h('div', { style: { fontSize: '32px' } }, '📄'),
          h('div', {}, props.title),
          h('div', { style: { fontSize: '12px', color: 'var(--theme-color-text-secondary)' } }, '拖拽标签头试试'),
        ],
      )
  },
}))

// ─── State ──────────────────────────────────────────────────

let counter = 0

function makeTab(): TabItem {
  counter++
  const color = COLORS[counter % COLORS.length]!
  return {
    id: genId('tab'),
    label: `标签 ${counter}`,
    closable: true,
    component: DemoPanel,
    props: { title: `面板 ${counter}`, color },
  }
}

function makeInitialLayout(): TabNode {
  counter = 0
  return {
    type: 'tabs',
    id: 'root-group',
    activeId: '',
    tabs: [makeTab(), makeTab(), makeTab(), makeTab(), makeTab()],
  } satisfies TabNode
}

const initialLayout = makeInitialLayout()
// Set the first tab as active
if (initialLayout.type === 'tabs') {
  initialLayout.activeId = initialLayout.tabs[0]?.id ?? ''
}

const layout = ref<TabNode>(initialLayout)

function addTab() {
  const tab = makeTab()
  // Find the first group node and add the tab there
  const node = findFirstGroup(layout.value)
  if (node) {
    node.tabs.push(tab)
    node.activeId = tab.id
  }
}

function findFirstGroup(node: TabNode): (TabNode & { type: 'tabs' }) | null {
  if (node.type === 'tabs') return node
  for (const child of node.children) {
    const found = findFirstGroup(child)
    if (found) return found
  }
  return null
}

function resetLayout() {
  layout.value = makeInitialLayout()
  if (layout.value.type === 'tabs') {
    layout.value.activeId = layout.value.tabs[0]?.id ?? ''
  }
  logs.value = []
  addLog('🔄 布局已重置')
}

// ─── Event log ──────────────────────────────────────────────

const logs = ref<string[]>([])

function addLog(msg: string) {
  const time = new Date().toLocaleTimeString()
  logs.value.unshift(`[${time}] ${msg}`)
  if (logs.value.length > 50) logs.value.pop()
}

// ─── Test shortcut callbacks ────────────────────────────────

useShortcutMap({
  'new-folder':  () => addLog('⌨️ 快捷键触发: Ctrl+Shift+N (新建文件夹)'),
  'open-remote': () => addLog('⌨️ 快捷键触发: Ctrl+Shift+O (打开远程)'),
})

function onDragStart(tab: TabItem, groupId: string) {
  addLog(`🟢 拖拽开始: "${tab.label}" (组: ${groupId})`)
}

function onDragEnd(tab: TabItem, groupId: string) {
  addLog(`🔴 拖拽结束: "${tab.label}" (组: ${groupId})`)
}

function onReorder(groupId: string, oldIndex: number, newIndex: number) {
  addLog(`↔️ 排序: 组 ${groupId}, ${oldIndex} → ${newIndex}`)
}

function onSplit(tabId: string, zone: string) {
  addLog(`✂️ 分屏: 标签 ${tabId}, 方向: ${zone}`)
}
</script>

<style scoped>
.home-view {
  width: 100%;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: var(--theme-color-bg-base);
}

.home-view__toolbar {
  padding: 8px 16px;
  display: flex;
  gap: 12px;
  align-items: center;
  border-bottom: 1px solid var(--theme-color-border);
  flex-shrink: 0;
}

.home-view__toolbar-spacer {
  flex: 1;
}

.home-view__hint {
  font-size: 12px;
  color: var(--theme-color-text-secondary);
}

.home-view__log {
  height: 120px;
  overflow-y: auto;
  border-top: 1px solid var(--theme-color-border);
  padding: 8px 16px;
  font-size: 12px;
  font-family: monospace;
  background: var(--theme-color-bg-surface);
  flex-shrink: 0;
  transition: background-color 0.2s ease, border-color 0.2s ease;
}

.home-view__log-item {
  color: var(--theme-color-text);
}

.home-view__log-empty {
  color: var(--theme-color-text-placeholder);
}
</style>
