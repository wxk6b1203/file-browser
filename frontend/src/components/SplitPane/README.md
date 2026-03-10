# SplitPane 分割面板

可拖拽调整大小的分割面板组件，支持水平 / 垂直方向、多面板嵌套、最小 / 最大尺寸约束、延迟更新模式以及丰富的样式自定义。

## 组件总览

| 组件 | 说明 |
|---|---|
| `SplitPane` | 父容器，提供布局方向和拖拽上下文 |
| `SplitPanePanel` | 子面板，承载内容并接受尺寸 / 样式配置 |
| `SplitPaneDivider` | 分隔条（面板间自动渲染，通常无需手动使用） |

## 快速开始

```vue
<script setup lang="ts">
import { SplitPane, SplitPanePanel } from '@/components/SplitPane'
</script>

<template>
  <SplitPane layout="horizontal" style="height: 400px">
    <SplitPanePanel size="30%">
      <p>左侧面板</p>
    </SplitPanePanel>
    <SplitPanePanel>
      <p>右侧面板（自动填充剩余空间）</p>
    </SplitPanePanel>
  </SplitPane>
</template>
```

> **注意**：`SplitPane` 容器本身使用 `width: 100%; height: 100%`，需要确保其父元素有明确的宽高，否则面板不会正确渲染。

---

## API

### `<SplitPane>` Props

| Prop | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `layout` | `'horizontal' \| 'vertical'` | `'horizontal'` | 分割方向：`horizontal` 为左右分割，`vertical` 为上下分割 |
| `lazy` | `boolean` | `false` | 开启后面板尺寸仅在拖拽结束时才更新（适用于面板内容渲染开销较大的场景） |
| `gap` | `number` | `0` | 面板之间的间隔距离（像素）。间隔空间不参与面板大小分配 |
| `indicatorWidth` | `string` | `'2px'` | 分隔条指示器的粗细（短边），CSS 值，如 `'3px'`、`'0.2em'` |
| `indicatorHeight` | `string` | `'24px'` | 分隔条指示器的长度（长边），CSS 值，如 `'40px'`、`'50%'` |

### `<SplitPane>` Events

| 事件名 | 回调参数 | 说明 |
|---|---|---|
| `resizeStart` | `(index: number, sizes: number[])` | 拖拽开始，`index` 为分隔条索引，`sizes` 为所有面板当前像素尺寸 |
| `resize` | `(index: number, sizes: number[])` | 拖拽中持续触发（`lazy` 模式下不触发） |
| `resizeEnd` | `(index: number, sizes: number[])` | 拖拽结束，`sizes` 为最终像素尺寸 |
| `resetCenter` | `(index: number, sizes: number[])` | 双击分隔条指示器后触发，两侧面板被重置为等宽/等高 |

### `<SplitPanePanel>` Props

| Prop | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `size` | `number \| string` | — | 初始大小。支持百分比 `'30%'`、像素 `'200px'`、纯数字（按 px 处理） |
| `minSize` | `number \| string` | — | 最小大小，格式同 `size` |
| `maxSize` | `number \| string` | — | 最大大小，格式同 `size` |
| `resizable` | `boolean` | `true` | 是否允许通过拖拽调整大小 |
| `borderRadius` | `string` | — | 面板圆角，例如 `'8px'`、`'0 8px 8px 0'`。设置后分隔条会自动避让圆角区域，只保留中间部分 |
| `backgroundColor` | `string` | — | 面板背景色，例如 `'#f5f5f5'`、`'rgba(0,0,0,0.05)'` |
| `customStyle` | `CSSProperties` | — | 自定义行内样式对象，会与内部样式合并 |
| `customClass` | `string` | — | 自定义 CSS 类名 |

### `<SplitPanePanel>` Events

| 事件名 | 回调参数 | 说明 |
|---|---|---|
| `update:size` | `(value: number)` | 面板大小因拖拽改变时触发，可配合 `v-model:size` 使用 |

---

## 使用示例

### 1. 基础水平分割

```vue
<SplitPane layout="horizontal" style="height: 300px">
  <SplitPanePanel size="30%" :min-size="100" background-color="#f0f9ff">
    侧栏
  </SplitPanePanel>
  <SplitPanePanel>
    主内容
  </SplitPanePanel>
</SplitPane>
```

### 2. 垂直分割

```vue
<SplitPane layout="vertical" style="height: 500px">
  <SplitPanePanel size="200px" min-size="80px">
    顶部工具栏
  </SplitPanePanel>
  <SplitPanePanel>
    编辑区
  </SplitPanePanel>
  <SplitPanePanel size="150px" min-size="60px">
    底部终端
  </SplitPanePanel>
</SplitPane>
```

### 3. 三栏布局 + 圆角 + 背景色

```vue
<SplitPane layout="horizontal" style="height: 100vh">
  <SplitPanePanel
    size="20%"
    min-size="160px"
    max-size="400px"
    border-radius="8px 0 0 8px"
    background-color="#fafafa"
  >
    导航
  </SplitPanePanel>
  <SplitPanePanel size="50%">
    主内容
  </SplitPanePanel>
  <SplitPanePanel
    size="30%"
    border-radius="0 8px 8px 0"
    background-color="#fafafa"
  >
    详情
  </SplitPanePanel>
</SplitPane>
```

### 4. 嵌套分割（水平 + 垂直）

```vue
<SplitPane layout="horizontal" style="height: 100vh">
  <SplitPanePanel size="250px" min-size="180px">
    文件树
  </SplitPanePanel>
  <SplitPanePanel>
    <!-- 嵌套的垂直分割 -->
    <SplitPane layout="vertical">
      <SplitPanePanel size="70%">
        代码编辑器
      </SplitPanePanel>
      <SplitPanePanel size="30%" min-size="100px">
        终端
      </SplitPanePanel>
    </SplitPane>
  </SplitPanePanel>
</SplitPane>
```

### 5. 禁止某个面板调整大小

将 `resizable` 设为 `false` 后，该面板与相邻面板之间的分隔条将不可拖拽（分隔条两侧面板都为 `resizable` 时才可拖拽）。

```vue
<SplitPane layout="horizontal" style="height: 300px">
  <SplitPanePanel size="200px" :resizable="false">
    固定宽度侧栏
  </SplitPanePanel>
  <SplitPanePanel>
    可伸缩区域
  </SplitPanePanel>
</SplitPane>
```

### 6. Lazy 模式

启用 `lazy` 后，面板在拖拽过程中不会实时更新大小，仅在松手后一次性应用。适合内容渲染开销大的场景。

```vue
<SplitPane layout="horizontal" :lazy="true" style="height: 300px">
  <SplitPanePanel size="50%">
    <HeavyComponent />
  </SplitPanePanel>
  <SplitPanePanel size="50%">
    <AnotherHeavyComponent />
  </SplitPanePanel>
</SplitPane>
```

### 7. 监听事件

```vue
<script setup lang="ts">
function onResizeStart(index: number, sizes: number[]) {
  console.log('拖拽开始', index, sizes)
}
function onResize(index: number, sizes: number[]) {
  console.log('拖拽中', index, sizes)
}
function onResizeEnd(index: number, sizes: number[]) {
  console.log('拖拽结束', index, sizes)
}
</script>

<template>
  <SplitPane
    layout="horizontal"
    style="height: 300px"
    @resize-start="onResizeStart"
    @resize="onResize"
    @resize-end="onResizeEnd"
  >
    <SplitPanePanel size="40%">A</SplitPanePanel>
    <SplitPanePanel size="60%">B</SplitPanePanel>
  </SplitPane>
</template>
```

### 8. 双击分隔条回到中间位置

双击分隔条中间的指示器（indicator），该分隔条两侧的面板会自动等分，同时遵守 `minSize` / `maxSize` 约束。此行为内置，无需额外配置。

```vue
<script setup lang="ts">
function onResetCenter(index: number, sizes: number[]) {
  console.log('已重置为居中', index, sizes)
}
</script>

<template>
  <SplitPane
    layout="horizontal"
    style="height: 300px"
    @reset-center="onResetCenter"
  >
    <SplitPanePanel size="30%" min-size="100px">
      拖拽后双击分隔条指示器 → 两侧面板等宽
    </SplitPanePanel>
    <SplitPanePanel>
      右侧面板
    </SplitPanePanel>
  </SplitPane>
</template>
```

### 9. 面板间隔（gap）

设置 `gap` 在面板之间添加固定间隔。间隔空间由分隔条占据，不参与面板大小的分配计算。

```vue
<SplitPane layout="horizontal" :gap="12" style="height: 300px">
  <SplitPanePanel size="30%" border-radius="8px" background-color="#f5f5f5">
    侧栏
  </SplitPanePanel>
  <SplitPanePanel border-radius="8px" background-color="#f5f5f5">
    主内容
  </SplitPanePanel>
  <SplitPanePanel size="25%" border-radius="8px" background-color="#f5f5f5">
    详情
  </SplitPanePanel>
</SplitPane>
```

> 配合 `borderRadius` 和 `backgroundColor` 使用可以实现卡片式布局效果。

### 10. 使用 `v-model:size` 双向绑定

```vue
<script setup lang="ts">
import { ref } from 'vue'
const sidebarWidth = ref(300)
</script>

<template>
  <SplitPane layout="horizontal" style="height: 300px">
    <SplitPanePanel v-model:size="sidebarWidth" min-size="150px">
      侧栏宽度：{{ sidebarWidth }}px
    </SplitPanePanel>
    <SplitPanePanel>
      主内容
    </SplitPanePanel>
  </SplitPane>
</template>
```

### 11. 自定义样式（customStyle / customClass）

```vue
<SplitPane layout="horizontal" style="height: 300px">
  <SplitPanePanel
    size="50%"
    :custom-style="{ padding: '16px', border: '1px solid #eee' }"
    custom-class="my-panel"
  >
    自定义面板
  </SplitPanePanel>
  <SplitPanePanel size="50%">
    普通面板
  </SplitPanePanel>
</SplitPane>
```

---

## 尺寸计算规则

1. **百分比 `'30%'`**：占父容器主轴方向尺寸的 30%。
2. **像素 `'200px'` 或数字 `200`**：固定像素值，内部会换算为百分比参与分配。
3. **未指定**：所有未设置 `size` 的面板平分剩余空间。
4. **溢出处理**：当所有面板的 `size` 总和超过 100% 时，组件会等比缩放至 100%。

## 文件结构

```
SplitPane/
├── index.ts                 # 统一导出
├── types.ts                 # 类型定义
├── SplitPane.vue            # 父容器组件
├── SplitPanePanel.vue       # 子面板组件
├── SplitPaneDivider.vue     # 分隔条组件
└── composables/
    ├── useContainer.ts      # 容器尺寸监听（基于 @vueuse/core）
    ├── useSize.ts           # 尺寸解析与百分比/像素计算
    └── useResize.ts         # 拖拽逻辑与约束限制
```

## 类型导出

```ts
import type {
  SplitLayout,
  SplitPanePanelProps,
  SplitPaneContext,
  PanelState,
} from '@/components/SplitPane'
```

## 依赖

- **Vue 3.5+**
- **@vueuse/core** — 用于 `useElementSize` 监听容器尺寸变化

