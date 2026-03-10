# Tabs 标签页 / 可拖拽分屏标签组件

支持 **拖拽排序**、**拖拽分屏**、**递归嵌套** 的标签页组件。底层分屏能力由同级 `SplitPane` 组件提供。

整个标签页布局以一棵 **树** 来描述——节点要么是「标签组 `TabGroupNode`」，要么是「分屏 `TabSplitNode`」。通过双向绑定 `v-model`，外部可以持久化 / 还原任意复杂的分屏布局。

---

## 组件总览

| 组件 | 说明 |
|---|---|
| `Tabs` | 根组件，接收树形数据，提供上下文与全局拖拽管理 |
| `TabGroup` | 一组标签（TabBar + 内容区），对应树中的 `TabGroupNode` |
| `TabBar` | 标签栏，包含多个 `TabHeader`，支持拖拽排序动画 |
| `TabHeader` | 单个标签头，可拖拽、可关闭 |
| `TabDropOverlay` | 拖拽悬停时的半透明分屏提示覆盖层 |

---

## 快速开始

```vue
<script setup lang="ts">
import { ref } from 'vue'
import { Tabs, type TabNode } from '@/components/Tabs'
import PanelA from './PanelA.vue'
import PanelB from './PanelB.vue'
import PanelC from './PanelC.vue'

const layout = ref<TabNode>({
  type: 'tabs',
  id: 'root',
  activeId: 'a',
  tabs: [
    { id: 'a', label: '面板 A', component: PanelA },
    { id: 'b', label: '面板 B', component: PanelB },
    { id: 'c', label: '面板 C', component: PanelC, closable: true },
  ],
})
</script>

<template>
  <div style="width: 100%; height: 600px">
    <Tabs v-model="layout" />
  </div>
</template>
```

> **注意**：`Tabs` 根组件使用 `width: 100%; height: 100%`，需要确保其父元素有明确的宽高。

---

## 数据结构

### `TabItem` — 单个标签定义

```ts
interface TabItem {
  id: string                        // 唯一标识
  label: string                     // 显示文本
  closable?: boolean                // 是否可关闭（默认 true）
  component?: Component | string    // 内容组件
  props?: Record<string, any>       // 传给内容组件的 props
}
```

### `TabGroupNode` — 标签组节点（叶子）

```ts
interface TabGroupNode {
  type: 'tabs'
  id: string
  tabs: TabItem[]
  activeId: string    // 当前激活的标签 id
}
```

### `TabSplitNode` — 分屏节点

```ts
interface TabSplitNode {
  type: 'split'
  id: string
  layout: 'horizontal' | 'vertical'    // 分屏方向
  children: TabNode[]                   // 子节点（递归）
  sizes?: (string | number)[]           // 初始面板大小，如 ['50%', '50%']
}
```

### `TabNode` — 联合类型（递归树）

```ts
type TabNode = TabGroupNode | TabSplitNode
```

#### 示例：预设分屏布局

```ts
const layout = ref<TabNode>({
  type: 'split',
  id: 'root-split',
  layout: 'horizontal',
  sizes: ['30%', '70%'],
  children: [
    {
      type: 'tabs',
      id: 'left',
      activeId: 'explorer',
      tabs: [
        { id: 'explorer', label: '资源管理器', component: Explorer },
      ],
    },
    {
      type: 'split',
      id: 'right-split',
      layout: 'vertical',
      sizes: ['60%', '40%'],
      children: [
        {
          type: 'tabs',
          id: 'editors',
          activeId: 'file1',
          tabs: [
            { id: 'file1', label: 'main.ts', component: Editor, props: { file: 'main.ts' } },
            { id: 'file2', label: 'App.vue', component: Editor, props: { file: 'App.vue' } },
          ],
        },
        {
          type: 'tabs',
          id: 'terminal',
          activeId: 'term1',
          tabs: [
            { id: 'term1', label: '终端', component: Terminal },
          ],
        },
      ],
    },
  ],
})
```

---

## API

### `<Tabs>` Props

| Prop | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `modelValue` / `v-model` | `TabNode` | — (必填) | 树形布局数据 |
| `barBackground` | `string` | `undefined` | 标签栏背景色 |
| `tabBackground` | `string` | `undefined` | 内容区背景色 |
| `overlayOpacity` | `number` | `0.15` | 拖拽分屏预览覆盖层的透明度 (0-1) |
| `minSplitWidth` | `number` | `100` | 允许垂直（左右）分屏的最小宽度（px） |
| `minSplitHeight` | `number` | `80` | 允许水平（上下）分屏的最小高度（px） |

### `<Tabs>` Events

| 事件名 | 回调参数 | 说明 |
|---|---|---|
| `update:modelValue` | `(value: TabNode)` | 布局数据变更（配合 `v-model`） |
| `tabDragStart` | `(tab: TabItem, groupId: string)` | 标签头开始被拖拽 |
| `tabDragEnd` | `(tab: TabItem, groupId: string)` | 标签头拖拽结束 |
| `tabReorder` | `(groupId: string, oldIndex: number, newIndex: number)` | 标签在同一组内完成排序 |
| `tabSplit` | `(tabId: string, zone: 'top'\|'bottom'\|'left'\|'right')` | 标签被拖拽到分屏区域并触发分屏 |

---

## 拖拽行为说明

### 在 TabBar 内拖拽（排序）

- 按住标签头即可拖拽，一个「幽灵副本」会附着鼠标跟随移动，不超出视口边界。
- 拖拽过程中，同组内其他标签会以 CSS `transform` 动画实时让位。
- 当拖拽位置越过相邻标签的中心线时，视为位置交换。
- 释放鼠标后完成排序。

### 在内容区拖拽（分屏）

| 指针所在区域 | 行为 |
|---|---|
| 上 1/3 | 水平分屏，被拖标签占据 **上半部分** |
| 下 1/3 | 水平分屏，被拖标签占据 **下半部分** |
| 左 1/4 | 垂直分屏，被拖标签占据 **左半部分** |
| 右 1/4 | 垂直分屏，被拖标签占据 **右半部分** |
| 中间 | 移动标签到该组（不分屏） |

- 当内容区尺寸不满足 `minSplitWidth` / `minSplitHeight` 约束时，分屏退化为移动。
- 分屏可以递归进行——分出来的子面板仍然是完整的 `Tabs`，可以继续拖拽分屏。
- 当一个标签组的所有标签被拖走或关闭后，该组会自动从树中移除，父级分屏节点也会自动折叠。

---

## 样式定制

```vue
<Tabs
  v-model="layout"
  bar-background="#f5f5f5"
  tab-background="#fff"
  :overlay-opacity="0.2"
/>
```

---

## 文件结构

```
Tabs/
├── index.ts                      # 导出入口
├── types.ts                      # 类型定义 & 注入 key
├── Tabs.vue                      # 根组件（递归渲染树、全局拖拽管理、幽灵副本）
├── TabGroup.vue                  # 标签组（TabBar + 内容区 + 覆盖层）
├── TabBar.vue                    # 标签栏（排序动画）
├── TabHeader.vue                 # 单个标签头（拖拽入口）
├── TabDropOverlay.vue            # 分屏预览覆盖层
├── README.md                     # 本文件
└── composables/
    ├── useTabTree.ts             # 树操作（增删移拆合）
    ├── useDrag.ts                # 拖拽状态管理
    └── useDropZone.ts            # 指针位置 → 分屏区域计算
```

---

## 导出

```ts
// 组件
export { Tabs, TabGroup, TabBar, TabHeader, TabDropOverlay }

// 类型
export type {
  TabItem,
  TabGroupNode,
  TabSplitNode,
  TabNode,
  DropZone,
  DragState,
  TabsContext,
}

// 工具
export { tabsContextKey, genId }
```

