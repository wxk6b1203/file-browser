<template>
  <!-- OS file drag-and-drop mask overlay -->
  <Teleport to="body">
    <div v-if="isDragging" class="file-drag-mask">
      <div class="file-drag-mask__label">松开以导入文件</div>
    </div>
  </Teleport>

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

    <!-- ─── SplitPane Minimize Demo ─────────────────────── -->
    <div class="home-view__split-demo">
      <div class="home-view__split-toolbar">
        <strong>SplitPane 最小化演示</strong>
        <el-button size="small" @click="leftMin = !leftMin">
          {{ leftMin ? '↔ 展开左侧栏' : '← 最小化左侧栏' }}
        </el-button>
        <el-button size="small" @click="rightMin = !rightMin">
          {{ rightMin ? '↔ 展开右侧栏' : '→ 最小化右侧栏' }}
        </el-button>
        <el-button size="small" @click="splitRef?.togglePanel(1)">
          ref 切换中间面板
        </el-button>
        <el-button size="small" @click="bottomMin = !bottomMin">
          {{ bottomMin ? '↕ 展开底栏' : '↓ 最小化底栏' }}
        </el-button>
        <span style="font-size:12px;color:var(--theme-color-text-secondary)">
          提示：最小化后拖动窗口大小再还原 → 自动居中
        </span>
      </div>

      <!-- 外层水平分割：左侧栏 | 中间（含垂直分割） | 右侧栏 -->
      <SplitPane
        ref="splitRef"
        layout="horizontal"
        :gap="6"
        class="home-view__split-container"
        @panel-minimized="(i: number) => addLog(`📌 水平面板 ${i} 最小化`)"
        @panel-restored="(i: number) => addLog(`📌 水平面板 ${i} 已还原`)"
      >
        <SplitPanePanel
          v-model:minimized="leftMin"
          size="20%"
          min-size="80px"
          border-radius="8px"
          background-color="var(--theme-color-bg-surface)"
        >
          <div class="demo-panel">
            <div style="font-size:24px">📂</div>
            <div>左侧栏</div>
            <div class="demo-panel__sub">horizontal → 向左折叠</div>
          </div>
        </SplitPanePanel>

        <SplitPanePanel border-radius="8px">
          <!-- 内层垂直分割：主区域 | 底栏 -->
          <SplitPane
            layout="vertical"
            :gap="6"
            @panel-minimized="(i: number) => addLog(`📌 垂直面板 ${i} 最小化`)"
            @panel-restored="(i: number) => addLog(`📌 垂直面板 ${i} 已还原`)"
          >
            <SplitPanePanel
              border-radius="8px"
              background-color="var(--theme-color-bg-surface)"
            >
              <div class="demo-panel">
                <div style="font-size:24px">📝</div>
                <div>主内容区</div>
              </div>
            </SplitPanePanel>

            <SplitPanePanel
              v-model:minimized="bottomMin"
              size="30%"
              min-size="50px"
              border-radius="8px"
              background-color="var(--theme-color-bg-surface)"
            >
              <div class="demo-panel">
                <div style="font-size:24px">💻</div>
                <div>底栏</div>
                <div class="demo-panel__sub">vertical → 向下折叠</div>
              </div>
            </SplitPanePanel>
          </SplitPane>
        </SplitPanePanel>

        <SplitPanePanel
          v-model:minimized="rightMin"
          size="20%"
          min-size="80px"
          border-radius="8px"
          background-color="var(--theme-color-bg-surface)"
        >
          <div class="demo-panel">
            <div style="font-size:24px">🔍</div>
            <div>右侧栏</div>
            <div class="demo-panel__sub">horizontal → 向右折叠</div>
          </div>
        </SplitPanePanel>
      </SplitPane>
    </div>

    <!-- ─── SplitPane maxPanelSize Demo ─────────────────────── -->
    <div class="home-view__split-demo">
      <div class="home-view__split-toolbar">
        <strong>maxPanelSize 最大宽度限制演示</strong>
        <span style="font-size:12px;color:var(--theme-color-text-secondary)">
          max-panel-size="60%" — 拖拽时任何面板都不超过容器的 60%；双击分割线居中同样受限
        </span>
      </div>
      <SplitPane
        layout="horizontal"
        :gap="6"
        max-panel-size="60%"
        class="home-view__split-container"
        @resize-end="(_i: number, sizes: number[]) => addLog(`📏 maxPanel resize: [${sizes.map(s => Math.round(s)).join(', ')}]`)"
        @reset-center="(_i: number, sizes: number[]) => addLog(`📏 maxPanel dblclick: [${sizes.map(s => Math.round(s)).join(', ')}]`)"
      >
        <SplitPanePanel
          border-radius="8px"
          background-color="var(--theme-color-bg-surface)"
        >
          <div class="demo-panel">
            <div style="font-size:24px">🅰️</div>
            <div>面板 A</div>
            <div class="demo-panel__sub">max 60%</div>
          </div>
        </SplitPanePanel>
        <SplitPanePanel
          border-radius="8px"
          background-color="var(--theme-color-bg-surface)"
        >
          <div class="demo-panel">
            <div style="font-size:24px">🅱️</div>
            <div>面板 B</div>
            <div class="demo-panel__sub">max 60%</div>
          </div>
        </SplitPanePanel>
        <SplitPanePanel
          border-radius="8px"
          background-color="var(--theme-color-bg-surface)"
        >
          <div class="demo-panel">
            <div style="font-size:24px">©️</div>
            <div>面板 C</div>
            <div class="demo-panel__sub">max 60%</div>
          </div>
        </SplitPanePanel>
      </SplitPane>
    </div>

    <!-- Tabs -->
    <div style="flex: 1; overflow: hidden;">
      <Tabs
        ref="tabsRef"
        v-model="tabLayout"
        :overlay-opacity="0.18"
        :min-split-width="120"
        :min-split-height="100"
        @tab-drag-start="onDragStart"
        @tab-drag-end="onDragEnd"
        @tab-reorder="onReorder"
        @tab-split="onSplit"
        @tab-activate="onTabActivate"
        @update:model-value="onTabLayoutChange"
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
import { ref, nextTick, defineComponent, h, markRaw } from 'vue'
import { Tabs, genId, type TabNode, type TabItem } from '@/components/Tabs'
import { SplitPane, SplitPanePanel } from '@/components/SplitPane'
import { useTheme } from '@/composables/useTheme'
import { useShortcutMap } from '@/composables/useShortcut'
import { useFileDrop } from '@/composables/useFileDrop'
import { OnLayoutChange } from '../../../wailsjs/go/render/Manager'

const { mode, themes, resolvedTheme, currentTheme, isDark, setTheme, SYSTEM_THEME } = useTheme()
const { isDragging } = useFileDrop()

function onThemeCommand(command: string) {
  setTheme(command)
}

// ─── SplitPane minimize demo ────────────────────────────────

const tabsRef = ref<InstanceType<typeof Tabs>>()
const splitRef = ref<InstanceType<typeof SplitPane>>()
const leftMin = ref(false)
const rightMin = ref(false)
const bottomMin = ref(false)

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

const tabLayout = ref<TabNode>(initialLayout)

function addTab() {
  const tab = makeTab()
  // Find the first group node and add the tab there
  const node = findFirstGroup(tabLayout.value)
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
  tabLayout.value = makeInitialLayout()
  if (tabLayout.value.type === 'tabs') {
    tabLayout.value.activeId = tabLayout.value.tabs[0]?.id ?? ''
  }
  leftMin.value = false
  rightMin.value = false
  bottomMin.value = false
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
  'rename': () => addLog('⌨️ 快捷键触发: F2 (重命名)'),
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
  // 树变更后 DOM 在下一 tick 才刷新，必须等 nextTick 再查
  nextTick(() => {
    if (!tabsRef.value) return
    const rects = tabsRef.value.getAllNodeRects()
    console.log(`📐 分屏后所有节点 Rect (${rects.size} 个):`, rects)
    // 将 Map<string, NodeRect> 转为普通对象后发给 Go 后端
    if (import.meta.env.VITE_APP_ENV !== 'internal') {
      OnLayoutChange(Object.fromEntries(rects))
    }
  })
}

function onTabActivate(tab: TabItem, groupId: string) {
  addLog(`🔘 选中: "${tab.label}" (组: ${groupId})`)
  nextTick(() => {
    if (!tabsRef.value) return
    const rect = tabsRef.value.getNodeRect(groupId)
    if (rect) {
      console.log(
        `📐 [${tab.label}] 所在组 ${groupId} 的 Rect:`,
        `x=${Math.round(rect.x)}, y=${Math.round(rect.y)}, ` +
          `w=${Math.round(rect.width)}, h=${Math.round(rect.height)}`,
      )
    }
  })
}

function onTabLayoutChange(_newTree: TabNode) {
  // tree changed (drag / split / reorder) — no-op, just for v-model sync
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

/* ── SplitPane demo ───────────────────────────────── */

.home-view__split-demo {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  border-bottom: 1px solid var(--theme-color-border);
}

.home-view__split-toolbar {
  padding: 6px 16px;
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.home-view__split-container {
  height: 240px;
  padding: 4px;
}

.demo-panel {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  user-select: none;
  gap: 4px;
  font-size: 14px;
  color: var(--theme-color-text);
}

.demo-panel__sub {
  font-size: 11px;
  color: var(--theme-color-text-secondary);
}

/* ── Log / Tabs ───────────────────────────────────── */

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

/* ── OS file drag-and-drop mask ─────────────────────── */

.file-drag-mask {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: color-mix(in srgb, var(--theme-color-primary, #409eff) 15%, transparent);
  border: 3px dashed var(--theme-color-primary, #409eff);
  pointer-events: none;
}

.file-drag-mask__label {
  padding: 12px 24px;
  background: var(--theme-color-bg-overlay, rgba(255 255 255 / 0.9));
  border-radius: 10px;
  font-size: 16px;
  font-weight: 600;
  color: var(--theme-color-primary, #409eff);
  box-shadow: 0 4px 16px rgb(0 0 0 / 0.12);
}
</style>
