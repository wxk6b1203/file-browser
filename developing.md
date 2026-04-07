# Developing

## Goal

构建一个基于 Wails 2.11 的多驱动文件管理系统，支持 `S3`、`OSS`、`SFTP`、`Local` 等连接类型，整体交互参考 VS Code。

## Design Understanding

### Product Model

我对这个产品的理解不是“单页文件选择器”，而是“面向多连接、多工作区、多任务的桌面文件工作台”。

- 一个连接是一个可持久化的远程入口。
- 左侧 Explorer 负责管理和打开这些入口。
- 中央区域不是单一页面，而是持续存在的工作区。
- 工作区中的每个 tab 都代表一次上下文：空白页、连接文件面板、设置页、新建连接页。
- `SplitPane + Tabs` 不是装饰，而是核心交互容器，后续所有主要内容都应该在这个容器里承载。

### Interaction Model

交互上要尽量遵守 VS Code 风格，但不是简单复制视觉，而是复制其“工作流节奏”：

- 左侧负责导航和资源入口。
- 中央负责具体工作内容。
- 右侧负责次级信息，如任务和通知。
- 设置、新建连接、文件面板都进入中央工作区，而不是弹窗优先。
- 用户可以保留多个上下文，并在不同 pane 中并行处理。

### Domain Boundary Understanding

这个项目的关键不是把所有逻辑直接压进驱动层，而是分清三层边界：

1. 驱动层
   - 只负责单个存储后端的文件能力。
   - 不负责连接配置仓库。
   - 不负责跨驱动拖放。
   - 不负责前端工作区语义。

2. 应用层
   - 负责连接定义、连接状态、文件操作编排、搜索聚合、通知事件、任务生命周期。
   - 把 `folder/` 提供的能力组合成产品可用的行为。

3. 前端工作区层
   - 负责面板、tab、树、菜单、任务与通知的交互组织。
   - 不直接感知驱动细节，只消费应用层暴露的能力和状态。

### Incremental Delivery Strategy

这个产品不能一口气把所有目标做完，合理的推进方式应该是：

1. 先打通“配置 -> 连接 -> 工作区壳层 -> 新建连接 -> 打开连接”的最小闭环。
2. 再把“文件列表/目录树/基础操作”补进连接面板。
3. 然后再做跨驱动拖放、异步任务、搜索聚合。
4. 最后补细节体验，如通知体系、图标映射、全局设置、快捷键完善。

这意味着当前阶段的前端连接面板可以先是“连接上下文面板”而不是最终完整文件浏览器，只要数据链路、工作区模型和组件边界是正确的，后续就能平滑替换。

## Current Progress

- 已完成项目初步探索，确认后端现状、前端现状、可复用基础组件与当前缺口。
- 已确认 `folder/` 目录下已有驱动抽象、驱动注册表、传输管理器与四种驱动实现，可作为后续连接层和文件操作层的底座。
- 已确认前端已有 `SplitPane`、`Tabs`、`Skeleton`、`Blank`、主题系统、快捷键系统，但当前页面仍是示例数据，不是面向文件管理系统的真实信息架构。
- 已完成第一批后端基础设施改造：
  - `config/` 已支持主配置与连接配置的 YAML/JSON 读写。
  - 已支持默认配置路径解析与 `--config/-c` 启动参数。
  - `main.go` 已改为通过配置初始化日志与应用运行时。
  - 已新增 `bootstrap/` 负责启动装配。
  - 已新增 `config` 单元测试并通过。
- 已完成第二批后端连接底座改造：
  - 已新增 `connection/` 模块，提供连接定义仓库与连接服务。
  - 已支持列出驱动、列出连接、保存连接、删除连接、打开连接、关闭连接、列出连接状态。
  - 已补充内置驱动注册导入，运行时可直接识别 `Local/S3/OSS/SFTP`。
  - 已重新生成 Wails JS 绑定。
- 已完成第三批前端工作区壳层改造：
  - 已新增正式的 `AppShell` 入口并替换根路由。
  - 已接入连接 store 和工作区 store。
  - 已实现左侧 Explorer、中央 Welcome/Settings/New Connection/Connection Overview tabs 的最小闭环。
  - 已接入 `CtrlOrCmd+Shift+N` 与 `CtrlOrCmd+.` 快捷键到工作区行为。
- 已完成第四批文件浏览最小链路：
  - 后端已新增 `fileops/`，支持按连接列目录。
  - `App` 已暴露 `ListConnectionDirectory()`。
  - 中央连接 tab 已从占位概览页演进为真实文件面板。
  - 文件面板已支持目录进入、根路径返回、列表/图标视图切换与刷新。
- 已完成第五批基础文件操作：
  - 后端已支持新建目录、重命名、删除。
  - 文件面板已接入工具栏“新建目录”和条目右键菜单。
  - 真实接口已通过 Wails 绑定暴露给前端。
- 已完成第六批 Explorer 与工作区联动增强：
  - Explorer 已支持按连接懒加载目录树，仅展示目录节点。
  - 目录树节点的“展开”和“导航”已拆分，避免点击目录时误触发展开逻辑。
  - 工作区已新增按连接维度维护的路径状态，Explorer 可直接驱动中央文件面板跳转到目标目录。
  - 文件面板内部导航会反向同步当前路径状态，为后续搜索定位、任务跳转、拖放目标定位提供统一入口。
- 已完成第七批异步搜索最小链路：
  - 后端已新增 `search/` 应用层模块，基于连接服务和递归列举能力做跨连接并发搜索。
  - 搜索已支持通过 Wails 事件渐进推送 `started/result/error/completed` 四类事件。
  - 前端左侧搜索面板已替换占位视图，可发起搜索、取消搜索、查看渐进结果与错误。
  - 搜索结果已可直接反向打开对应连接，并将中央文件面板定位到结果所在目录。
- 已完成第八批任务面板与下载任务最小链路：
  - 后端已新增 `transfer/` 应用层模块，基于 `folder.TransferManager` 管理传输任务。
  - 当前已打通“远程文件下载到本地临时目录”的真实任务链路。
  - 前端右侧任务面板已替换占位视图，可查看任务进度、速度、状态、错误，并支持取消和清理。
  - 文件面板右键菜单已接入“下载到临时目录”，任务创建后可在任务面板查看。
- 已完成第九批外部拖入上传最小链路：
  - 后端传输服务已支持“本地绝对路径 -> 当前连接目录”的上传任务提交。
  - 正式 `AppShell` 已挂载 OS 文件拖放监听，不再只在示例页内生效。
  - 拖入外部文件到中央文件面板时，前端会根据目标面板的激活 tab 解析连接上下文，并为每个本地路径创建上传任务。
  - 上传任务已接入现有右侧任务面板，与下载任务共用一套状态展示与控制逻辑。
- 已完成第十批通知面板最小链路：
  - 前端已新增 `notifications` store 和右侧通知面板。
  - 搜索错误、连接打开/关闭错误、传输失败/取消、面板级上传失败已统一写入通知流。
  - 右侧通知面板已替换占位视图，可查看来源、级别、时间，并支持单条移除和清空。
- 已完成第十一批全局设置页真实配置闭环：
  - `App` 已新增应用配置读取/保存 Wails 绑定。
  - 设置页已从占位卡片改为真实表单，可编辑语言、主题、日志、搜索和传输配置。
  - 设置保存后会即时同步当前运行时的日志、搜索默认参数和传输临时目录。
  - `AppShell` 启动时会主动拉取后端配置，并把主题与语言应用到前端运行时。

## Execution Log

### 2026-04-01 Compressed Summary

- 完成项目结构探索，明确 Wails + Go 驱动层 + Vue 工作区的整体边界。
- 完成配置、启动、连接服务、驱动注册、Wails 绑定和前端 AppShell 的第一批闭环。
- 打通文件浏览最小链路、基础文件操作、Explorer 目录树、异步搜索、任务面板、外部拖入上传、通知面板和设置页真实配置闭环。
- 关键验证覆盖：`go test ./...`、前端 `type-check` / `build` / `build-only`、必要时 `wails generate module`。

## Current Architecture Notes

### Backend

- `main.go`
  - 已改为通过 `bootstrap.Initialize()` 解析启动参数、加载配置、初始化日志。
  - 已补充内置驱动空导入，运行时驱动注册完整。
  - 当前仍只绑定 `App` 和 `render.Manager`，后续会继续扩展业务服务绑定。
- `app.go`
  - 已持有运行时配置对象与连接服务。
  - 已暴露连接相关 Wails 方法。
  - 已接入基础文件浏览服务。
  - 已接入异步搜索服务。
  - 已接入传输任务服务。
  - 已接入应用主配置的读取/保存入口。
  - 保存配置后会即时同步日志、搜索和传输服务的运行参数。
  - 当前仍未接入独立通知服务等更高层聚合模块。
- `config/config.go`
  - 已支持主配置与连接配置结构。
  - 已支持基于 `viper` 的配置读取。
  - 已支持 YAML/JSON 读写与默认路径解析。
  - 已接入设置页读写链路。
  - 当前仍缺少更细粒度的应用层配置仓库/变更历史。
- `logging/logging.go`
  - 已支持多输出目标、stdout/stderr、文件轮转。
  - 适合保留，后续由配置模块驱动初始化参数。
- `folder/`
  - `driver.go` 提供统一 `Manager` 接口。
  - `registry.go` 提供驱动注册与实例生命周期管理。
  - `transfer_manager.go` 提供异步上传/下载任务管理。
  - `s3`、`alibaba-oss`、`sftp`、`local` 已有较完整驱动实现。
- `connection/`
  - 已具备连接配置仓库与连接实例生命周期管理。
  - 当前还缺少探活刷新、凭证测试、连接编辑冲突处理等增强能力。
- `render/`
  - 当前只处理前端渲染信号/拖放信号的骨架。
  - 还没有真正承接文件面板、面板拖放、跨连接传输的业务事件。
- `shortcut/`
  - 已实现前后端事件桥接，可直接扩展为全局快捷键入口。
- `fileops/`
  - 已支持按连接列目录。
  - 已支持新建目录、重命名、删除。
  - 后续继续扩展上传、下载与跨驱动拖放。
- `search/`
  - 已支持搜索请求管理、请求取消、跨连接并发搜索和事件回推。
  - 已支持运行时更新默认并发数和结果上限。
  - 当前搜索实现基于递归列举，后续可按驱动能力补原生搜索优化。
- `transfer/`
  - 已支持下载任务提交、任务列表、取消、移除和清理已完成任务。
  - 已支持本地路径上传到当前连接目录。
  - 已支持运行时切换传输临时目录。
  - 当前已打通“下载到本地临时目录 + 外部文件拖入上传”，后续继续扩展目录传输和跨连接编排。

### Frontend

- `frontend/src/views/Skeleton`
  - 布局能力已复用到正式 `AppShell`。
  - 搜索已替换为真实左侧面板；任务面板和通知面板已替换为真实右侧面板。
- `frontend/src/composables/useFileDrop.ts`
  - 已在保留 Go 侧拖放信号的同时，向前端派发面板级 OS 文件 drop 事件，供应用层消费上传逻辑。
- `frontend/src/views/AppShell`
  - 已成为正式应用入口。
  - 已接入中央工作区 Tabs 和左侧 Explorer。
- `frontend/src/views/Blank/FrontPage.vue`
  - 已改为可复用的工作区欢迎页入口。
- `frontend/src/components/Workspace/ConnectionOverviewTab.vue`
  - 已演进为真实文件面板组件。
  - 已接入新建目录、重命名、删除等基础操作。
- `frontend/src/components/SplitPane`
  - 已具备多面板分割能力。
- `frontend/src/components/Tabs`
  - 已具备多标签和分割树能力，可作为中央工作区的核心容器。
- `frontend/src/stores/connections.ts`
  - 已接入后端连接接口。
- `frontend/src/stores/workspace.ts`
  - 已接管中央工作区 tab 生命周期。
- `frontend/src/stores/search.ts`
  - 已接管搜索请求状态、结果流和错误流。
- `frontend/src/stores/tasks.ts`
  - 已接管传输任务列表轮询、取消、移除和清理。
- `frontend/src/stores/notifications.ts`
  - 已接管应用级通知列表、移除与清空。
- `frontend/src/composables/useShortcut.ts`
  - 已按当前产品目标重组基础快捷键。
  - 后续仍需补充文件操作相关快捷键。
- `frontend/src/stores/theme.ts`
  - 已有主题切换能力，后续所有新组件都应基于现有 CSS 变量体系。
- `frontend/src/stores/settings.ts`
  - 已负责应用配置的拉取、保存和前端运行时主题/语言同步。
- `frontend/src/router/index.ts`
  - 当前仍保留 Hello/About 等示例路由，需要替换成实际应用入口。

## Target Module Split

### Backend target modules

1. 配置模块
   - 负责启动参数解析。
   - 负责主配置读写。
   - 负责远程连接配置读写。
   - 负责根据文件后缀识别 YAML/JSON，默认 YAML。
   - 负责输出标准化后的路径与运行配置给其他模块。

2. 日志模块
   - 保持 `logging/` 为底层实现。
   - 新增由配置驱动的初始化入口。

3. 连接模块
   - 负责连接定义的 CRUD。
   - 负责从连接定义创建/销毁 `folder.Manager` 实例。
   - 负责连接探活、连接状态缓存、连接能力暴露。

4. 文件操作模块
   - 负责目录列举、文件信息、重命名、删除、新建目录等回调。
   - 负责上传/下载/跨面板拖放的调度。
   - 负责统一封装 `folder.TransferManager`。
   - 负责跨驱动复制时的“下载到临时目录 -> 上传到目标”的编排。

5. 搜索模块
   - 负责接受前端搜索请求。
   - 负责并发搜索已连接驱动。
   - 负责异步推送搜索结果和搜索状态。

6. 应用模块
   - 聚合配置、日志、连接、文件操作、搜索、快捷键、通知。
   - 暴露 Wails 绑定方法。
   - 维护全局应用状态与生命周期。

### Frontend target modules

1. App Shell
   - 以 `SkeletonLayout` 为底座。
   - 左侧：目录、搜索、设置入口。
   - 中央：Tab + SplitPane 工作区。
   - 右侧：任务面板、通知面板。

2. Explorer
   - 展示保存的连接。
   - 双击连接发起连接并加载目录。
   - 展开后展示目录树，仅显示目录。

3. Workspace Tabs
   - 工作区 tab 支持空白页、文件面板、设置页、新建连接页。
   - 新内容进入最后聚焦的面板。

4. File Panel
   - 支持图标视图/列表视图。
   - 支持右键菜单。
   - 支持外部拖入上传。
   - 支持面板间拖放触发跨驱动传输。

5. Search Panel
   - 发起后端搜索。
   - 渐进接收异步搜索结果。

6. Settings Page
   - 编辑全局配置。

7. Connection Form
   - 支持按驱动类型动态渲染参数表单。
   - 便于未来扩展新驱动。

## Proposed Config Structure

### Main config file

建议新增独立运行时配置结构，主配置只存放应用级参数，不把密钥直接散落在不同地方。

```yaml
app:
  locale: zh
  theme: system
  temp_dir: ./data/tmp

log:
  level: info
  outputs:
    - stdout
    - ./logs/app.log

paths:
  connections_file: ./config/connections.yaml
  state_file: ./config/state.yaml

search:
  max_concurrency: 4
  result_limit: 500

transfer:
  temp_dir: ./data/transfers
  overwrite_strategy: rename
```

### Connections config file

连接配置建议单独存放，顶层为连接列表，便于增删改查和未来扩展分组。

```yaml
connections:
  - id: local-default
    name: Local
    driver: Local
    enabled: true
    root: C:/Users/wxk6b
    config: {}

  - id: s3-prod
    name: Prod S3
    driver: S3
    enabled: true
    root: /
    config:
      region: ap-southeast-1
      bucket: my-bucket
      access_key_id: xxx
      access_key_secret: xxx
```

说明：

- 复用 `folder.DriverOptions` 的公共字段思想，但不要直接把前端表单与驱动底层结构硬耦合。
- 建议新增应用层 `Connection` 实体，再由连接服务转换成 `folder.DriverOptions`。
- `connections_file` 可以是 `.yaml` 或 `.json`，由后缀决定序列化方式。

## Proposed Phases

### Phase 1

目标：把后端运行骨架搭起来，替换示例入口。

- 建立配置读取/保存模块。
- 接入启动参数与默认配置路径。
- 配置驱动日志初始化。
- 建立应用级 `AppServices` 装配。
- 定义连接实体与连接仓库。

### Phase 2

目标：打通“连接配置 -> 连接实例 -> 基础文件浏览”主链路。

- 实现连接 CRUD。
- 实现连接打开/关闭。
- 实现目录树加载与文件面板基础数据接口。
- 前端重构 `Skeleton` 为真实应用壳。

### Phase 3

目标：打通文件操作与任务面板。

- 重命名、删除、新建目录。
- 上传、下载、跨连接拖放。
- 任务历史与状态面板。

### Phase 4

目标：打通搜索、设置、通知与细节交互。

- 并发搜索与异步结果回推。
- 全局设置页与快捷键。
- 通知中心。
- 空白页、新建连接页、图标视图/列表视图完善。

## Immediate Next Step

下一步进入第三个前端/应用层切片：

- 扩展文件面板的目录上传、目录下载与更完整拖放交互。
- 视需要把任务面板从轮询刷新升级为事件流。
- 视需要把通知流从前端聚合升级为应用层事件流。

## Current Iteration

### Transfer Event Stream Upgrade

当前任务面板已经具备真实上传/下载任务展示能力，但状态同步仍依赖前端轮询 `ListTransferTasks()`。

这一版准备把传输任务同步链路升级为：

1. `folder.TransferManager`
   - 增加任务观察者回调。
   - 在 `submit/running/progress/final/remove/removeAll` 生命周期节点发出任务事件。
   - 顺手补齐当前任务快照读写的并发安全，避免任务结构被后台 goroutine 更新时前台同时读取。

2. `transfer.Service` + `App`
   - 继续保留 `ListTransferTasks()` 作为首次同步和手动刷新兜底接口。
   - 新增 `transfer:event` Wails 运行时事件，用于推送任务增量更新。

3. `frontend/src/stores/tasks.ts`
   - 改为常驻订阅 `transfer:event`。
   - store 内部负责增量合并、移除、排序以及失败/取消通知。
   - 组件层不再负责开启和关闭轮询。

### Why This Way

- 任务状态本质是后端异步过程，事件流比前端定时拉取更符合模型。
- 保留 `ListTransferTasks()` 可以避免初始化丢事件时前端无数据可恢复。
- 观察者挂在 `TransferManager` 而不是文件面板组件，可以覆盖上传、下载、取消、清理等所有入口。

### Current Status

- 已完成现状确认：搜索已使用事件流，任务仍是轮询。
- 已确定本轮先做“事件驱动 + 并发安全”两件事，不在这一轮同时扩展目录级传输。
- 已完成 `TransferManager` 事件观察能力接入，并新增任务快照读写锁。
- 已完成 `transfer:event` 运行时事件发射，前端任务 store 已改为常驻订阅。
- 已移除任务面板组件级轮询生命周期，任务同步职责已下沉到 store。
- 已补充 `folder.TransferManager` 观察者单测，覆盖完成和移除事件。
- 本轮验证结果：
  - `go test ./folder ./transfer ./connection ./fileops ./search ./...` 通过
  - `npm run type-check` 通过
  - `npm run build-only` 通过

## Next Iteration

### Upload Completion Refresh

当前上传任务已经能进入任务面板，但中央文件面板仍需要用户手动点击刷新才能看到新文件。

这一轮准备补齐的链路是：

1. `tasks` store
   - 在接收到“上传完成”任务事件时，只处理一次。
   - 解析任务目标连接和目标目录。
   - 发出前端工作区级目录刷新事件。

2. `ConnectionOverviewTab`
   - 监听目录刷新事件。
   - 仅当事件命中当前连接且目标目录正好是当前打开目录时，触发重载。

### Why This Way

- 不把文件面板刷新逻辑反向耦合进后端传输模块。
- 不要求文件面板主动理解全部任务状态，只消费“该目录需要刷新”这一更稳定的前端语义。
- 为后续目录上传、跨连接拖放完成后的自动刷新保留同一条前端事件通道。

### Current Status

- 已新增前端工作区目录刷新事件通道，用于承接任务完成后的目录重载请求。
- 已在 `tasks` store 内接入“上传完成 -> 发目录刷新事件”的一次性处理逻辑。
- 已在 `ConnectionOverviewTab` 内接入目录刷新监听，命中当前连接和当前目录时自动重载。
- 本轮验证结果：
  - `npm run type-check` 通过
  - `npm run build-only` 通过

## Next Slice

### Local Directory Upload

当前从系统外部拖入文件已能排队上传，但拖入本地目录仍然报“不支持目录上传”。

这一轮准备补齐：

1. `transfer.Service`
   - 接受本地目录路径。
   - 递归创建远端目录结构。
   - 为目录中的每个文件提交独立上传任务。

2. `App` / Wails 绑定
   - 把上传接口返回值从单个任务 ID 扩展为任务 ID 列表。

3. `AppShell`
   - 适配批量任务返回值。
   - 目录拖入后立即请求当前目录刷新，使新建目录无需等文件任务完成才可见。

### Why This Way

- 继续复用现有 `TransferManager`，不引入新的“目录任务”模型。
- 目录上传在任务面板里仍表现为多个文件任务，符合当前数据结构。
- 目录创建本身不是传输任务，所以需要在前端补一次即时目录刷新。

### Current Status

- 已支持本地目录路径递归展开为“远端目录创建 + 多个文件上传任务”。
- 已把上传接口的返回值扩展为任务 ID 列表，并已同步重新生成 Wails 绑定。
- 已完成 `AppShell` 对批量上传返回值的适配，目录拖入后会立即触发当前目录刷新。
- 已新增 `transfer` 单测，覆盖文件上传计划和目录上传计划的展开逻辑。
- 本轮验证结果：
  - `wails generate module` 通过
  - `go test ./transfer ./folder ./connection ./fileops ./search ./...` 通过
  - `npm run type-check` 通过
  - `npm run build-only` 通过

## Next Slice

### Directory Download And Explorer Sync

当前目录上传已经可用，但目录下载到本地临时目录还未接入；同时左侧 Explorer 目录树不会跟随目录刷新事件同步变化。

这一轮准备补齐：

1. `transfer.Service`
   - 支持远端目录递归下载到本地临时目录。
   - 继续保持“一个文件对应一个任务”的任务模型。

2. `ConnectionOverviewTab`
   - 对目录和文件统一开放“下载到临时目录”入口。
   - 适配批量下载任务返回值。

3. `ExplorerPanel`
   - 监听现有目录刷新事件。
   - 在目标目录节点已加载的情况下增量重载该节点，而不是整棵树重置。

### Current Status

- 已支持远端目录递归下载到本地临时目录，并继续沿用“一个文件对应一个传输任务”的任务模型。
- 已把文件面板中的“下载到临时目录”扩展到目录条目，并适配批量任务返回值。
- 已新增下载计划测试，覆盖目录下载和空目录下载两类规划逻辑。
- 已把 Explorer 接到目录刷新事件流上，已加载目录节点会按事件增量重载。
- 本轮验证结果：
  - `wails generate module` 通过
  - `go test ./transfer ./folder ./connection ./fileops ./search ./...` 通过
  - `npm run type-check` 通过
  - `npm run build-only` 通过

## Next Slice

### Cross-Panel Transfer

当前外部拖入上传和目录下载都已打通，但“从一个面板拖到另一个面板”还没有真正编排成跨连接传输。

这一轮准备补齐：

1. 前端拖拽源
   - 文件面板中的文件/目录条目接入内部 panel drag payload。
   - 通过 `Tabs -> SplitPane` 透传 panel drop 事件到 `AppShell`。

2. 应用层传输编排
   - 新增“源连接条目 -> 目标连接目录”的传输入口。
   - 后端负责把“下载到本地临时目录 -> 下载完成后上传到目标连接”串起来。

3. 前端目标刷新
   - 跨连接拖放发起后，立即刷新目标目录。
   - 上传完成后继续复用现有任务事件流做后续刷新。

### Why This Way

- 不把两段任务依赖关系塞到前端 store。
- 继续复用现有传输任务模型，任务面板里自然显示“先下载、后上传”两阶段任务。
- 保持驱动层纯粹，跨连接编排仍留在应用层。

### Current Status

- 已把文件面板条目接成内部 panel drag 源，支持文件和目录条目跨面板拖拽。
- 已把 `Tabs -> SplitPane` 的 panel drop 事件透传到 `AppShell`，正式工作区可接收内部面板拖放。
- 已新增应用层跨连接传输入口，后端会把“下载到本地临时目录 -> 下载完成后上传到目标连接”串起来。
- 已新增跨连接目录传输规划测试，覆盖目标目录与文件映射规则。
- 本轮验证结果：
  - `wails generate module` 通过
  - `go test ./transfer ./folder ./connection ./fileops ./search ./...` 通过
  - `npm run type-check` 通过
  - `npm run build-only` 通过

## Next Slice

### Transfer Follow-up Error Visibility

当前跨连接传输已经会在下载完成后自动创建上传任务，但如果“后续上传创建”失败，这个失败还没有显式暴露给前端，只会静默结束。

这一轮准备补齐：

1. `transfer.Service`
   - 在 follow-up 上传创建失败时发出服务层错误事件。

2. `tasks` store
   - 消费新的 `transfer:event` 错误事件。
   - 写入通知面板，补齐用户可见性。

### Why This Way

- 不强行把“上传创建失败”伪装成普通传输任务。
- 保持任务模型稳定，把链路编排错误作为独立事件处理。

### Current Status

- 已为 `transfer:event` 增加服务层错误事件类型，用于承接 follow-up 上传创建失败。
- 已在 `transfer.Service` 内补齐后续上传创建失败的显式事件发射。
- 已在前端 `tasks` store 内消费该错误事件，并写入通知中心。
- 本轮验证结果：
  - `go test ./transfer ./folder ./connection ./fileops ./search ./...` 通过
  - `npm run type-check` 通过
  - `npm run build-only` 通过

## Recommended Backend Packages

建议在保留现有 `config/`、`folder/`、`logging/`、`shortcut/` 的基础上，新增应用层 package：

- `appstate/`
  - 定义应用级聚合结构，如 `Services`、运行时状态、事件总线封装。
- `bootstrap/`
  - 启动参数解析、配置加载、日志初始化、服务装配。
- `connection/`
  - 定义连接实体。
  - 管理连接配置仓库。
  - 管理连接实例生命周期。
- `fileops/`
  - 面向前端暴露文件浏览、重命名、删除、新建目录、上传、下载、跨连接拖放。
- `search/`
  - 承接跨连接并发搜索与结果回推。
- `notify/`
  - 统一通知模型，供右侧通知面板消费。
- `workspace/`
  - 管理前端工作区相关状态，如最近打开连接、活动面板、设置页 tab 等。

说明：

- `folder/` 保持驱动层，不把连接配置仓库和跨驱动业务编排塞进去。
- `connection/` 应该依赖 `folder/`，而不是反过来。
- `App` 只做 Wails 绑定入口，不直接承载全部业务逻辑。

## Recommended Backend Entities

### Main config entities

- `config.AppConfig`
- `config.LogConfig`
- `config.PathConfig`
- `config.SearchConfig`
- `config.TransferConfig`
- `config.UIConfig`

### Connection entities

建议不要直接把 `folder.DriverOptions` 作为配置文件实体暴露给前端，建议新增：

- `connection.Definition`
  - `id`
  - `name`
  - `driver`
  - `enabled`
  - `root`
  - `readonly`
  - `description`
  - `tags`
  - `config`
- `connection.State`
  - `connected`
  - `lastError`
  - `lastConnectedAt`
  - `capabilities`

### Runtime services

- `connection.Repository`
  - 负责连接配置读写。
- `connection.Manager`
  - 负责 `Open`、`Close`、`ListDefinitions`、`GetState`、`GetFolderManager`。
- `fileops.Service`
  - 负责文件操作与传输编排。
- `search.Service`
  - 负责搜索任务管理。
- `notify.Service`
  - 负责通知事件积累和查询。

## Recommended Frontend Structure

### Views

- `frontend/src/views/AppShell/`
  - 替代当前演示版 `SkeletonView.vue`，作为正式应用入口。
- `frontend/src/views/Blank/`
  - 保留并重构为空白工作区欢迎页。
- `frontend/src/views/ConnectionForm/`
  - 新建连接/编辑连接页面。
- `frontend/src/views/Settings/`
  - 全局设置页面。
- `frontend/src/views/FilePanel/`
  - 文件浏览主页面。

### Components

- `frontend/src/components/ExplorerTree/`
  - 左侧连接树和目录树。
- `frontend/src/components/SearchPanel/`
  - 搜索输入与结果列表。
- `frontend/src/components/TaskPanel/`
  - 上传下载任务列表。
- `frontend/src/components/NotificationPanel/`
  - 通知列表。
- `frontend/src/components/FileList/`
  - 列表视图。
- `frontend/src/components/FileGrid/`
  - 图标视图。
- `frontend/src/components/ConnectionBadge/`
  - 驱动类型、连接状态展示。

### Stores

- `frontend/src/stores/app.ts`
  - 应用级初始化状态、错误状态。
- `frontend/src/stores/connections.ts`
  - 连接定义、连接状态、连接树数据。
- `frontend/src/stores/workspace.ts`
  - 中央 tabs、活动 pane、打开的文件面板与设置页。
- `frontend/src/stores/tasks.ts`
  - 传输任务列表。
- `frontend/src/stores/notifications.ts`
  - 通知中心数据。
- `frontend/src/stores/settings.ts`
  - 前端缓存的设置数据。

### Composables

- `frontend/src/composables/useConnections.ts`
- `frontend/src/composables/useWorkspaceTabs.ts`
- `frontend/src/composables/useContextMenu.ts`
- `frontend/src/composables/useFileIcons.ts`
- `frontend/src/composables/useBackendEvents.ts`

## First Coding Tranche

第一批代码建议先做基础设施，不先碰复杂拖放。顺序如下：

1. 后端配置模块重构
   - 定义新配置结构。
   - 接入 `viper`。
   - 支持主配置和连接配置的读写。
   - 接入启动参数。

2. 后端应用装配
   - 引入 `bootstrap` 或等价模块。
   - `main.go` 改为“加载配置 -> 初始化日志 -> 创建服务 -> 绑定 Wails”。
   - `app.go` 改为聚合服务入口。

3. 连接管理模块
   - 保存连接定义。
   - 列出连接定义。
   - 打开/关闭连接实例。
   - 返回连接状态与能力。

4. 前端应用壳重构
   - 路由入口切到正式 `AppShell`。
   - 左侧目录面板改成连接树。
   - 中央保持 `Tabs + SplitPane`。
   - 空白页接入新建连接/设置入口。

5. 新建连接页与设置页
   - 先打通表单提交、读取、保存。
   - 暂不做搜索与复杂传输。

## Risks And Constraints

- 当前 `frontend/wailsjs/go/main/App.d.ts` 只有 `Greet()`，说明后续每新增 Wails 方法都要同步生成绑定。
- 当前路由和页面仍有示例代码，重构时要避免把演示视图和正式视图混在同一路径下。
- 驱动注册名当前是 `Local`、`S3`、`OSS`、`SFTP`，前后端必须统一使用这组值。
- 搜索接口目前不在驱动层，后续要么扩展可选 `Searcher` 接口，要么由应用层基于 `List(..., Recursive)` 做统一搜索。
- 跨驱动拖放会涉及临时目录、冲突策略、失败补偿和任务状态展示，不适合放在第一批实现。

## Next Slice

### File Type Icons

当前文件面板和搜索面板已经具备真实数据链路，但文件图标仍然只有“目录/文件”两种静态表达，这会直接影响可读性，也不利于后续扩展预览、右键动作和搜索结果识别。

这一轮准备补齐：

1. `frontend/src/composables/useFileIcons.ts`
   - 建立统一的文件图标解析层。
   - 以“目录类型 + 特殊文件名 + 扩展名映射”的顺序解析图标。
   - 设计为纯前端可扩展，不把图标语义耦合进后端接口。

2. `ConnectionOverviewTab`
   - 列表视图和网格视图统一改为消费图标解析层。
   - 保留目录开合图标差异，文件按类型映射。

3. `SearchPanel`
   - 搜索结果与文件面板共享同一套图标规则，避免同一文件在不同区域显示不一致。

### Why This Way

- 图标映射属于工作区表现层，不应该散落到多个页面组件里重复判断。
- 后续如果要细化更多文件类型、增加品牌化图标或按主题调整色彩，只需要收敛修改一个入口。
- 先建立“图标解析层”，比在页面里继续堆 `v-if` 更利于维护。

### Current Status

- 已新增 `useFileIcons.ts`，统一处理目录图标、特殊配置文件图标和常见扩展名图标映射。
- 已把文件面板和搜索面板改为复用统一图标解析逻辑。
- 已新增 `useFileIcons` 单测，覆盖目录、代码文件、特殊配置文件、媒体文件和默认回退逻辑。
- 本轮验证结果：
  - `npm run test:unit -- useFileIcons` 通过
  - `npm run type-check` 通过
  - `npm run build-only` 通过

## Next Slice

### Settings Runtime Config

当前设置页虽然已经存在于中央工作区，但之前只是主题切换和说明文案，和 `config/` 模块没有真正连通。这会导致“设置页已经出现，但实际运行配置仍只能改文件”的体验断层，也不符合 `original.md` 中“设置菜单传到后端保存”的最初要求。

这一轮准备补齐：

1. 后端配置绑定
   - `App` 暴露应用主配置读取和保存接口。
   - 保存后重新加载标准化配置，避免前端直接持有未归一化路径。

2. 运行时同步
   - 日志配置保存后立即重新初始化。
   - 搜索服务默认并发和结果上限支持热更新。
   - 传输服务临时目录支持热更新。

3. 前端设置页
   - 新增 `settings` store 统一负责配置拉取、保存和主题/语言同步。
   - 把设置页改造成真实表单，而不是继续保留占位说明卡。
   - 运行中容易导致“半生效”的路径类配置先保持只读展示。

### Why This Way

- 现有 `config/` 模块已经具备 YAML/JSON 读写能力，缺的是应用层闭环，不需要再重新设计一套配置格式。
- 路径类配置和连接仓库路径如果在运行时直接切换，会牵扯连接服务与仓库重建；这一轮先只开放可安全热更新的字段，边界更清晰。
- 主题和语言同步放在前端 `settings` store，而不是散落在页面组件里，更利于后续扩展启动恢复、最近设置和导入导出能力。

### Current Status

- 已新增 `GetAppConfig()` 与 `SaveAppConfig()` Wails 绑定。
- 已在保存配置后即时同步日志初始化、搜索默认参数和传输临时目录。
- 已新增 `frontend/src/stores/settings.ts`，负责拉取配置、保存配置并把主题/语言应用到前端运行时。
- `AppShell` 启动时已主动拉取后端配置，避免设置页未打开前前端仍停留在默认主题/语言。
- 设置页已改为真实表单，当前可编辑语言、主题、应用临时目录、日志级别与输出、搜索参数、传输临时目录和覆盖策略。
- 连接配置文件路径和状态文件路径当前改为只读展示，避免运行中切换配置仓库导致服务半生效。
- 本轮验证结果：
  - `wails generate module` 通过
  - `go test ./config ./search ./transfer` 通过
  - `npm run build-only` 通过

## Next Slice

### Remaining Workspace Cleanup

设置页闭环之后，剩余最明显的工作缺口主要集中在两类：

1. 壳层清理
   - `router` 里仍保留 `Skeleton/About` 等示例路由。
   - 一些空状态和示例文案仍带有“后续接入”的旧阶段描述。

2. 文件面板能力补齐
   - 文件双击打开仍是占位提示，尚未形成真正的文件预览/打开策略。
   - 文件面板上下文菜单仍可以继续扩展为“复制路径 / 刷新父目录 / 连接级操作”等更完整动作。

## Latest Audit

### 2026-04-01 Project Audit Compressed Summary

- 对照 `original.md`、当前 `developing.md` 和代码状态做了阶段审计。
- 结论：项目主链路已经从“壳层搭建”进入“正确性与交互收口”阶段。
- 当时的优先级被调整为：先修后端路径语义、恢复 `go test ./...` 基线，再清理示例路由和继续文件面板细节。
- 审计确认任务事件流、外部拖入上传等旧计划已有实现，后续文档需以实际代码状态为准。

## Current Iteration

### Stability Baseline: Cross-Platform Path Fix

这一轮开始正式执行上面的稳定性基线计划，并优先处理“必须兼容 Windows / Linux / macOS 文件与路径语义”的要求。

### Design Strategy

这次修复没有走“只修 macOS `/var` -> `/private/var`”的局部补丁，而是按跨平台原则收敛：

1. `folder/local`
   - 根路径在校验阶段统一 canonical 化。
   - 路径越界判断不再依赖字符串前缀，而改为基于 `filepath.Rel` 的父子关系判断。
   - 这样可以避免前缀误判，也更符合 Windows 盘符、Linux/macOS 目录边界和 symlink 解析后的实际语义。

2. `config`
   - 配置文件路径解析统一走“absolute + nearest existing parent canonicalization”。
   - 对于存在 symlink 的工作目录或配置目录，主配置路径、日志路径、连接配置路径和传输目录都回到同一套标准化口径。

3. 测试层
   - 配置测试断言改为按 canonical path 比较，而不是假设临时目录一定保留原始 symlink 形式。

### Completed Changes

- 已更新 `folder/local/options.go`
  - `RootPath` 在验证通过后会进一步 canonical 化，避免后续子路径校验时根路径和目标路径不在同一 filesystem view 下比较。

- 已更新 `folder/local/driver.go`
  - `isSubPath()` 已从字符串 `HasPrefix` 改为 `filepath.Rel` 判断。
  - 这一步同时提升了跨平台兼容性，避免 Windows 盘符与大小写、Unix 路径前缀碰撞带来的误判风险。

- 已更新 `config/config.go`
  - 新增配置路径 canonical 解析逻辑。
  - `ResolveAppConfigPath()`、连接配置加载/保存现在会输出统一口径的标准化路径。

- 已更新 `config/config_test.go`
  - 测试改为以 canonical 临时目录为断言基准。

### Verification

- `go test ./config ./folder/local ./fileops` 通过。
- `go test ./...` 通过。

### Result

这一轮完成后，之前阻塞主线的三个问题已经收敛：

- `Local` 驱动基础文件操作恢复正常。
- `fileops` 不再被本地路径误判拖垮。
- 配置路径在 symlink 场景下的行为与驱动层语义已经统一到同一方向。

### Next Recommended Step

稳定性基线第一阶段已经完成。随后已继续完成正式入口收口：

- 已清理 `frontend/src/router/index.ts` 中遗留的 `/blank`、`/skeleton`、`/about` 示例路由。
- 正式路由入口现在只保留 `AppShell`。
- `FrontPage` 组件仍保留给工作区欢迎页复用，但不再作为独立公开演示路由暴露。
- `cd frontend && npm run build-only` 通过，前端构建产物中不再额外拆出 `AboutView` / `SkeletonView` 的公开路由 chunk。

下一步建议继续：

1. 梳理仍保留在仓库中的示例页面与开发遗留视图，决定哪些继续复用、哪些彻底下沉或删除。
2. 进入文件面板“文件打开策略”，结束当前“双击文件仍是占位提示”的状态。
3. 为连接表单补“测试连接”和更明确的错误恢复反馈。

## Current Iteration

### File Open Strategy

在完成路径稳定性修复和正式路由收口后，这一轮继续处理文件面板最明显的交互缺口：双击文件不再只弹占位提示，而是进入真实打开流程。

### Strategy

当前阶段不直接在前端内嵌编辑器或预览器，而是采用一个更稳妥的统一策略：

1. 前端文件面板
   - 双击文件或右键“打开”都会调用统一后端入口。

2. 后端传输服务
   - 先把目标文件下载到传输临时目录。
   - 下载完成后触发 follow-up 动作。
   - follow-up 不再只支持“下载后上传”，现在同时支持“下载后用系统默认程序打开”。

3. 跨平台打开实现
   - macOS 使用 `open`
   - Windows 使用 `cmd /c start`
   - Linux 使用 `xdg-open`

这样做的原因是：

- 不需要当前阶段就引入文件预览器、文本编辑器和二进制查看器的复杂分流。
- 对 `S3 / OSS / SFTP / Local` 四类驱动都适用。
- 继续复用现有传输任务、事件流和通知体系，而不是再造一套下载后打开的异步管理逻辑。

### Completed Changes

- 已更新 `transfer/service.go`
  - 新增 `OpenFile()`。
  - follow-up 机制从单一上传扩展为“上传 / 打开”两种动作。
  - 新增跨平台本地文件打开逻辑。

- 已更新 `app.go`
  - 新增 `OpenConnectionFile()` Wails 绑定入口。

- 已重新生成 Wails 绑定
  - 前端生成代码已包含新的 `OpenConnectionFile()` 方法。

- 已更新 `frontend/src/components/Workspace/ConnectionOverviewTab.vue`
  - 双击文件现在会触发真实打开流程。
  - 文件右键菜单新增“打开”。

- 已更新中英文 locale
  - 去掉“文件打开后续阶段接入”的占位文案。
  - 新增“打开 / 正在准备并打开文件”的反馈文案。

- 已补充 `transfer/service_test.go`
  - 覆盖“下载完成后触发打开 follow-up”路径。

### Verification

- `wails generate module` 通过。
- `go test ./transfer ./...` 通过。
- `cd frontend && npm run build-only` 通过。

### Current Product State

到这一轮为止，文件面板的文件动作已经从“只能浏览目录和下载”推进到：

- 目录：双击进入。
- 文件：双击准备本地副本并交给系统默认程序打开。
- 右键：打开 / 下载 / 重命名 / 删除。

### Next Recommended Step

下一步应继续做两件事：

1. 梳理并清理仓库内仍残留的示例组件、示例 store 和开发演示视图。
2. 为连接表单补“测试连接”，让新建连接不必先保存再发现参数错误。

### Additional Cleanup

在文件打开策略落地后，已继续完成一轮仓库清理：

- 已删除未被正式入口引用的示例视图：
  - `AboutView`
  - `Hello/HomeView`
  - `SkeletonView`

- 已删除未被使用的示例组件与配套资源：
  - `HelloWorld.vue`
  - `TheWelcome.vue`
  - `WelcomeItem.vue`
  - 示例 icons 目录下的旧欢迎页图标组件
  - `stores/counter.ts`
  - `HelloWorld` 单测

- 已更新 `frontend/src/views/Skeleton/index.ts`
  - 去除对 `SkeletonView` 的无效导出，保留正式应用仍在使用的 `SkeletonLayout`、按钮列和类型导出。

### Additional Verification

- 在删除示例残留后再次执行 `cd frontend && npm run build-only`，通过。

### Updated Next Step

当前最值得继续推进的是：

1. 为连接表单补“测试连接”动作。
2. 再考虑是否需要把“本地临时副本打开”进一步演进为“文本文件预览 / 编辑器内打开”。

## Current Iteration

### Default Download Path + Save As + Connection Test

根据新的需求，这一轮继续补齐三块能力：

1. 默认下载路径
   - 用户可在设置页配置默认下载目录。
   - 如果该目录已配置且实际存在，双击文件时会优先下载到该目录，再交给系统默认程序打开。
   - 如果默认下载目录未配置或目录已失效，则自动回退到传输临时目录，不阻断文件打开流程。

2. 文件菜单下载动作
   - 文件/目录右键菜单现在包含：
     - 打开
     - 下载
     - 下载到临时目录
     - 文件额外支持“另存为”
     - 重命名
     - 删除

3. 连接表单测试连接
   - 新建/编辑连接页新增“测试连接”按钮。
   - 测试连接使用 one-shot manager，不会污染已保存连接实例，也不会要求先保存后才能验证参数。

### Design Notes

这次实现里，下载相关语义被明确拆开了：

- 双击打开
  - 面向“临时获得一个本地副本并立刻打开”。
  - 优先使用默认下载目录；不可用时回退到临时目录。

- 下载
  - 面向“把文件或目录下载到默认下载目录”。
  - 默认下载目录可用时写入该目录；否则退回临时目录，保证动作不会失效。

- 另存为
  - 面向“用户手动决定目标文件路径”。
  - 通过 Wails 的 `SaveFileDialog` 由 Go 侧调起系统保存对话框。

### Completed Changes

- 已扩展 `config.TransferSection`
  - 新增 `download_dir / downloadDir` 配置字段。

- 已更新 `transfer.Service`
  - 已支持默认下载目录。
  - 已开始使用已有的 `overwriteStrategy` 处理默认下载目录内的重名目标。
  - 已新增：
    - `Download()`
    - `DownloadToDirectory()`
    - `DownloadFileToPath()`
    - 默认下载目录与临时目录之间的回退逻辑

- 已更新 `App` Wails 绑定
  - 新增：
    - `DownloadConnectionFile()`
    - `SaveConnectionFileAs()`
    - `PickDownloadDirectory()`
    - `TestConnection()`

- 已更新设置页
  - 传输设置新增“默认下载目录”字段。
  - 已新增“选择目录”按钮，使用系统目录选择对话框。

- 已更新文件面板
  - 双击文件现在优先下载到默认下载目录再打开。
  - 右键菜单新增“下载”“下载到临时目录”“另存为”。

- 已更新连接服务
  - 新增 `connection.Service.Test()`。
  - 测试连接通过 `folder.NewManager()` 构造临时 manager，并在支持时执行 `Ping()`。

- 已更新连接表单
  - 新增“测试连接”按钮。
  - 当前表单可在不保存的情况下直接验证驱动参数是否可用。

- 已重新生成 Wails 绑定并同步前端调用。

### Verification

- `wails generate module` 通过。
- `go test ./...` 通过。
- `cd frontend && npm run build-only` 通过。

### Current Product State

到这一轮为止，文件面板和连接表单的主体验已经从“能跑最小链路”提升到“具备基础可用性”：

- 文件可以双击真实打开。
- 下载不再只依赖临时目录。
- 文件菜单已有“下载 / 下载到临时目录 / 另存为”。
- 连接参数可以先测试再保存。

### What Still Needs Development

结合当前代码状态，下一阶段最值得做的不是继续堆边角功能，而是下面三项：

1. 连接表单体验增强
   - 为测试连接结果增加更明确的能力展示和错误定位，而不是只用一句 toast。
   - 对各驱动配置项补更严格的前端校验和必填提示。

2. 文件面板细节增强
   - 增加“复制路径 / 刷新当前目录 / 在 Explorer 中定位”等高频上下文动作。
   - 明确目录下载完成后的反馈和默认下载目录命中/回退提示。

3. 任务与通知信息质量提升
   - 当前任务面板能看到传输状态，但还缺“打开来源目录 / 打开本地文件”这类任务后置动作。
   - 通知目前仍偏事件罗列，后续可以增加可跳转上下文和分组。

## Current Iteration

### Connection Form Validation + Test Result Panel

在补齐“测试连接”按钮之后，这一轮继续把连接表单从“能测”提升到“可读、可纠错”。

### Completed Changes

- 已更新 `frontend/src/components/Workspace/connectionSchemas.ts`
  - 为驱动字段补充前端 required 元信息。
  - 当前已按真实后端 `Validate()` 规则覆盖 `S3 / OSS / SFTP` 的主要必填项。

- 已更新 `frontend/src/components/Workspace/ConnectionFormTab.vue`
  - 新增前端基础校验：
    - 连接名称必填
    - 驱动类型必填
    - `S3 / OSS` 的 `region / bucket / accessKeyID / accessKeySecret` 必填
    - `SFTP` 的 `address / username` 必填
    - `SFTP` 额外要求“密码或私钥至少填写一个”
  - 新增页内测试结果面板，不再只依赖 toast。
  - 测试成功后会展示：
    - 驱动
    - 能力数量
    - 能力标签列表
  - 测试失败时会在页内保留错误信息，便于用户对照修改。
  - 表单字段变化后会自动清空已过时的测试结果和相关校验错误，避免旧结果误导用户。

- 已更新 locale
  - 新增连接表单校验提示、测试结果标题、能力标签和状态文案。

### Verification

- `go test ./...` 通过。
- `cd frontend && npm run build-only` 通过。

### Current Product State

到这一轮为止，连接表单已经具备以下可用性特征：

- 可以在保存前测试连接。
- 会先做一轮前端基础校验，再请求后端测试。
- 测试结果会保留在表单页内，而不只是闪过一条通知。

### Updated Next Step

结合当前完成度，下一步最值得做的是：

1. 文件面板高频动作补齐
   - 复制路径
   - 刷新当前目录
   - 在 Explorer 中定位

2. 任务完成后的后置动作
   - 从任务面板直接“打开本地文件 / 打开下载目录”

3. 通知质量提升
   - 让通知能直接跳回相关连接、目录或任务，而不是只做被动展示

## Current Iteration

### File Panel High-Frequency Actions + Task Post-Actions

这一轮继续把“能完成传输”推进到“传输完成后可继续处理结果”，重点不是再加一个新面板，而是把文件面板和任务面板补成更完整的闭环。

### Design Strategy

这次实现分成两部分：

1. 文件面板高频动作
   - 增加“复制路径”。
   - 把刷新从只有图标的按钮收口为明确的文本动作，降低可发现性问题。
   - 这里复制的是统一驱动层暴露的远程路径字符串，不把不同驱动额外分裂成多种 UI 语义。

2. 任务完成后的后置动作
   - 对已完成的下载任务增加：
     - 打开本地文件
     - 打开所在目录
   - 这些动作不放在前端直接猜平台命令，而是继续通过 Go 后端统一封装。
   - 原因是“打开文件 / 在系统文件管理器中定位”本质上是平台相关能力，应该由后端负责兼容 Windows / Linux / macOS。

### Cross-Platform Notes

为满足“兼容 Windows、Linux、macOS 的文件和路径”这个要求，这一轮没有把平台判断散落到前端，而是统一封装在传输服务内：

- 打开本地路径
  - macOS: `open`
  - Windows: `cmd /c start`
  - Linux: `xdg-open`

- 在系统文件管理器中定位
  - macOS:
    - 文件: `open -R`
    - 目录: `open`
  - Windows:
    - 文件: `explorer /select,`
    - 目录: `explorer`
  - Linux:
    - 退化为打开所在目录，因为 `xdg-open` 没有统一的“选中文件”协议

同时，这部分平台命令选择逻辑被提取为可测试函数，而不是只靠运行时人工验证。

### Completed Changes

- 已更新 `transfer.Service`
  - 新增：
    - `OpenLocalPath()`
    - `RevealLocalPath()`
  - 新增跨平台命令构造函数：
    - `commandForOpen()`
    - `commandForReveal()`
  - 打开/定位前会先校验本地路径存在，避免前端只得到模糊失败。

- 已更新 `app.go`
  - 新增 Wails 绑定：
    - `OpenLocalPath()`
    - `RevealLocalPath()`

- 已更新 `frontend/src/components/Workspace/ConnectionOverviewTab.vue`
  - 文件/目录右键菜单新增“复制路径”。
  - 刷新按钮现在显示明确文案，而不只是图标。
  - 复制路径使用 Wails runtime 剪贴板能力，不依赖浏览器沙箱实现。

- 已更新 `frontend/src/components/TaskPanel/TaskPanel.vue`
  - 已完成的下载任务现在显示：
    - 打开本地文件
    - 打开所在目录
    - 移除
  - 运行中任务仍只保留“取消”，避免动作过载。

- 已更新 locale
  - 补齐文件面板和任务面板新增动作文案。

- 已更新 `transfer/service_test.go`
  - 新增平台命令构造测试，覆盖 Windows / Linux / macOS 的打开与定位命令选择。

### Verification

- `wails generate module` 通过。
- `go test ./...` 通过。
- `cd frontend && npm run build-only` 通过。

### Current Product State

到这一轮为止，传输链路已经从“任务完成后用户自己去系统里找文件”推进到：

- 文件面板可以直接复制远程路径。
- 下载完成后可以从任务面板直接打开本地结果。
- 下载完成后可以直接打开所在目录继续处理。

### Updated Next Step

当前最值得继续推进的是：

1. 通知质量提升
   - 让通知能直接跳转到相关任务、连接或目录。

2. 文件面板状态反馈增强
   - 明确目录下载完成、默认下载目录命中、回退到临时目录等结果反馈。

3. 连接表单高级校验
   - 继续补齐更多驱动字段的格式校验，而不只是 required 检查。

## Current Iteration

### Actionable Notifications

这一轮继续处理通知面板的信息闭环问题：通知不再只是被动显示错误文本，而是尽可能把用户直接带回任务或相关目录。

### Design Strategy

这里没有把“跳转能力”硬编码在通知面板组件里，而是拆成三层：

1. `notifications store`
   - 通知项新增 `action` 元数据。
   - 动作当前先支持两类：
     - 打开任务面板
     - 打开相关连接目录

2. `shell store`
   - 左右侧栏激活状态从 `AppShell` 局部 `ref` 提升到共享 store。
   - 这样通知面板、任务面板和其他工作区组件都能安全切换到 `Explorer / Tasks / Notifications`，避免组件间隐式耦合。

3. 触发源补上下文
   - 任务流里的“失败 / 取消 / follow-up 失败”通知，携带“打开任务面板”动作。
   - `AppShell` 中“上传任务创建失败 / 跨连接传输创建失败”通知，携带“打开相关目录”动作。

### Completed Changes

- 已新增 `frontend/src/stores/shell.ts`
  - 统一管理：
    - 左侧 `explorer / search`
    - 右侧 `tasks / notifications`

- 已更新 `frontend/src/stores/notifications.ts`
  - `NotificationItem` 现在支持 `action` 元数据。

- 已更新 `frontend/src/components/NotificationPanel/NotificationPanel.vue`
  - 通知卡片现在会按动作类型显示按钮。
  - 点击后会：
    - 切到任务面板
    - 或打开对应连接目录并切到 Explorer

- 已更新 `frontend/src/stores/tasks.ts`
  - 传输失败、取消、follow-up 失败通知现在可直接跳到任务面板。

- 已更新 `frontend/src/views/AppShell/AppShellView.vue`
  - 上传失败、跨连接传输创建失败通知现在可直接跳到相关连接目录。
  - 左右侧栏状态已改为使用 `shell store` 管理。

- 已更新 locale
  - 新增通知动作按钮文案。
  - 通知空态说明同步更新为“含后续动作”。

### Verification

- `go test ./...` 通过。
- `cd frontend && npm run build-only` 通过。

### Current Product State

到这一轮为止，通知面板已经不再只是事件列表，而是具备基础导航能力：

- 传输失败/取消通知可以把用户带回任务面板。
- 创建上传或跨连接任务失败时，可以直接回到相关连接目录继续处理。

### Updated Next Step

当前最值得继续推进的是：

1. 文件面板状态反馈增强
   - 区分“默认下载目录命中”和“回退到临时目录”。
   - 为目录下载补更聚合的完成提示。

2. 连接表单高级校验
   - 补更多格式和取值校验，而不只是 required。

3. 任务面板精细化
   - 如果继续提升，可补“聚焦具体任务”而不只是切到任务面板。

## Current Iteration

### Layout Overflow + Search Cancel Bug Fix

这一轮优先处理两个影响基础使用体验的 bug：

1. 右侧任务/通知面板默认未展开时，中央工作区没有真正吃满剩余空间。
2. 搜索取消按钮已接线，但运行状态和事件收口不正确，表现为“无法停止”。

### Root Cause

#### 1. 侧栏留白

问题不在 `AppShell`，而在 `SplitPane` 的尺寸分配语义：

- 面板最小化后，`SplitPanePanel` 的视觉尺寸确实被压到了 `0px`。
- 但 `SplitPane` 内部的尺寸计算仍把该面板历史比例算进总宽度。
- 结果是：
  - 右侧面板视觉上收起了
  - 中央面板却没有拿回这部分宽度
  - 最终表现为右侧残留一块空白区域

#### 2. 搜索取消

问题不在按钮点击本身，而在前端 store 的状态管理：

- `cancel()` 只调用后端 `CancelSearch()`，但前端仍保持：
  - `running = true`
  - `requestId` 继续保留
- 同时，事件过滤逻辑只有在“当前存在 requestId”时才会做比对。
- 这会导致取消后的迟到事件仍可能继续落回当前 UI，搜索看起来就像没有真正停止。

### Completed Changes

- 已更新 `frontend/src/components/SplitPane/SplitPane.vue`
  - 尺寸计算现在会把已最小化面板视为 `0` 宽度。
  - 剩余可用空间会重新分配给未最小化面板。
  - 只有一个可见面板时，该面板会直接占满可用空间。

- 已更新全局根布局和骨架布局样式
  - `frontend/src/assets/main.css`
  - `frontend/src/views/Skeleton/SkeletonLayout.vue`
  - `frontend/src/components/SplitPane/SplitPane.vue`
  - `frontend/src/components/SplitPane/SplitPanePanel.vue`
  - 当前统一补上：
    - `min-width: 0`
    - `min-height: 0`
    - 根节点 `overflow: hidden`
  - 目标是确保窗口高度不超出视口，超长内容统一在内部滚动容器里滚动。

- 已更新 `frontend/src/stores/search.ts`
  - `cancel()` 现在会先在前端立即收口：
    - `running = false`
    - 清空当前 `requestId`
    - 保留 `summary.cancelled = true`
  - 被取消请求后续返回的事件现在会被彻底忽略，不再污染当前 UI。
  - 正常完成时也会清空 `requestId`，避免旧请求继续落入当前状态。

### Verification

- `go test ./...` 通过。
- `cd frontend && npm run build-only` 通过。

### Current Product State

到这一轮为止：

- 右侧面板默认折叠时，中央工作区会真正占满剩余区域。
- 根布局不再依赖页面整体向外撑高，滚动应集中在内部面板。
- 搜索取消会立即结束前端运行状态，并忽略被取消请求的迟到事件。

### Updated Next Step

当前如果继续修 bug，最值得继续看的是：

1. 在 `wails dev` 实际运行态再验证一次滚动边界
   - 重点看 Explorer、Search、Task、Notification、ConnectionOverview 这几个长列表面板。

2. 搜索后端的深层取消响应
   - 当前前端已能立即停止，但如果某些驱动的递归列举不及时响应 `ctx.Done()`，仍可继续优化驱动层或搜索服务层的中断粒度。

## Current Iteration

### Native-Style File List View

这一轮继续处理文件面板的使用感问题，目标不是再做一层“卡片列表”，而是把列表视图收口到接近 Finder / Windows Explorer 的原生明细列表。

### Design Strategy

这一轮的列表视图按以下原则重做：

1. 视觉结构
   - 从“一个个卡片框”改为“表头 + 分隔行”的明细列表。
   - 行之间只保留细分隔线和悬停高亮，弱化卡片感。

2. 列模型
   - 当前支持：
     - 名称
     - 修改时间
     - 大小
     - 类型
   - 其中“名称”固定展示，其余字段可勾选开关。

3. 排序模型
   - 点击表头即可排序。
   - 再次点击同一列切换升序/降序。
   - 目录始终保持在文件之前，符合常见文件管理器习惯。

4. 偏好持久化
   - 列显示和排序状态落到本地存储。
   - 切换连接或重开应用后仍保留当前偏好。

### Completed Changes

- 已重做 `frontend/src/components/Workspace/ConnectionOverviewTab.vue` 的列表视图
  - 新增原生明细列表表头。
  - 行样式改为细分隔线列表，不再使用卡片行。
  - 网格视图保留不变。

- 已新增列显示控制
  - 当前可勾选：
    - 修改时间
    - 大小
    - 类型
  - 名称列固定保留。

- 已新增列表排序
  - 支持按：
    - 名称
    - 修改时间
    - 大小
    - 类型
    排序。

- 已新增本地偏好持久化
  - 列显示状态和排序状态会写入本地存储。

- 已更新 locale
  - 补充字段面板、列名和文件类型文案。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-06 - Add Inline Delete Unit Tests

### Scope

- 为文件面板内联删除补充稳定的单元测试。
- 初始尝试挂载完整 `ConnectionOverviewTab`，但 Element Plus 按需 CSS 在当前 Vitest 环境中会通过 `element-plus/es/.../style/css` 拉入真实 CSS，导致测试环境在加载 SFC 前失败。
- 为避免把测试建立在脆弱的 CSS mock 上，本轮把内联删除的路径规划逻辑抽成纯 helper，并对 helper 与快捷键配置做纯单测。

### Completed Changes

- 新增 `inlineDelete.ts`：
  - `buildInlineDeletePaths(targetPaths, orderedPaths)`
  - `removeInlineDeletePath(paths, targetPath)`
- `ConnectionOverviewTab` 的内联删除状态复用该 helper：
  - 多选进入待确认状态时，按当前列表顺序排列待确认路径。
  - 单项取消/确认时，只移除该路径。
- 新增单测 `inlineDelete.spec.ts`，覆盖：
  - 多选待删除路径按当前列表顺序排列。
  - 目标路径不在当前有序列表时，回退到目标输入顺序。
  - 单项取消/确认只移除一个待确认路径。
  - `DEFAULT_SHORTCUTS` 不再注册 `delete -> Delete`，避免和 Wails/Edit Delete 冲突。

### Verification

- `cd frontend && npm run test:unit -- src/components/Workspace/__tests__/inlineDelete.spec.ts` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-06 - Fix Delete Shortcut Conflict And Per-Item Delete Confirmation

### Root Cause

- 文件面板仍处理 `Delete` 键，会和 Wails/macOS 菜单中的 `Edit -> Delete` 语义产生冲突。
- 多选进入内联删除确认后，所有待确认项共享同一个确认动作：
  - 点击任意一个条目的“删除”，会批量删除全部待确认项。
  - 这不符合“每个文件/目录独立确认”的交互预期。

### Completed Changes

- 文件面板不再拦截 `Delete` 键：
  - `Delete` 保留给 Wails/系统菜单语义。
  - 文件删除快捷键只使用用户指定的 `Backspace`。
- 全局 shortcut 默认表移除 `delete -> Delete`：
  - 避免全局快捷键层继续对 `Delete` 产生事件。
- 内联删除确认改为按条目独立执行：
  - 多选后可以让多个条目同时进入待确认状态。
  - 点击某一项的“删除”只删除该项。
  - 点击某一项的“取消”只取消该项的待确认状态。
  - 删除完成后本地移除该条目，其他待确认条目保持原状态。
- 删除忙碌状态从全局布尔值改为按路径记录：
  - 当前正在删除的条目按钮禁用。
  - 其他条目保持可见，避免误认为会一起删除。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

### Current Product State

到这一轮为止，文件列表视图已经从“偏卡片式文件块”收口到更接近系统文件管理器的使用方式：

- 列表更接近原生明细视图。
- 可按列排序。
- 可勾选显示字段。
- 偏好可持续保留。

### Updated Next Step

如果继续沿这个方向优化，当前最值得做的是：

1. 明细列表多选
   - 支持 Shift / Cmd(Ctrl) 选择、多项批量操作。

2. 列宽调整
   - 允许拖动列头调整宽度，更接近原生资源管理器。

3. 键盘导航
   - 上下选择、回车打开、Delete 删除、F2 重命名。

## 2026-04-01 to 2026-04-05 - Compressed Change Summary

### File Panel Interaction

- 文件面板从卡片式列表演进为更接近 Finder / Explorer 的明细列表和图标视图。
- 增加多选、Shift 区间选择、Ctrl/Cmd 切换选择、键盘导航、批量复制路径、批量下载、批量删除、列显示开关、排序和列宽拖动。
- 实现中央工作区内、中央工作区到 Explorer 树、跨连接和同连接的拖放；同连接目录移动走 `Move`，跨连接走 transfer。
- 修复拖放后源列表不收口、刷新回顶、底部留白、多选嵌套路径重复、拖放反馈不清晰等问题。
- 增加多选拖拽预览，改进文件面板工具栏、breadcrumbs、状态条和视图切换按钮。

### Explorer / Tabs / Workspace

- Explorer 支持拖放悬停自动展开、通知动作跳转、连接删除内联确认、删除连接后关闭相关 tab、新建连接保存后关闭表单 tab。
- Welcome 页改为普通可关闭 tab；修复最后一个 tab 偶发关不掉和 split 关闭一侧后残留空位的问题。
- 为 split collapse 补充回归测试，后续发现 `SplitPanePanel` 测试需要 i18n 挂载。
- 分割工作区时增加文件列表缓存，避免 tab 迁移/重挂载导致无意义重新加载。

### Connection Form / Config / Runtime State

- 修复新建和编辑连接表单状态串扰、`Local rootPath` 回显、`OSS/S3/SFTP` 非默认驱动配置回填被 watcher 重置的问题。
- 当前驱动编辑语义收口为：`Local/SFTP` 使用 `rootPath`，`S3/OSS` 使用 `prefix`，不再对这些驱动暴露通用逻辑 `root` 作为主编辑入口。
- `connections.yml/yaml` 读取从 Viper 切换为 `yaml.v3`；主配置仍在后续阶段单独处理。
- 编辑已连接连接后会重建运行时实例；前端 `connections.hydrate()` 改成 pending promise 模式，避免并发 hydrate 竞态。
- 保存连接后新增显式连接配置刷新事件，让 Explorer 和文件面板清理缓存并按新配置重载。

### Drag / Drop / WebView Compatibility

- 修复外部 OS 文件拖入中央工作区上传链路；文件 drop target 下沉到叶子 `TabGroup`，避免多分割时外层大遮罩和子面板遮罩叠加。
- 增加 drop/dragend 全局兜底清理，处理 WebKit / WebView2 事件顺序差异导致的遮罩残留。
- 后续又针对内部跨面板拖放补充 WebView2 兼容：自定义 MIME 不可见时回退到共享内存拖拽状态和标准 `text/plain` marker。

### Theme / Layout / Visual Polish

- 新增并接入 `2026 Dark` 主题，基于 VS Code 2026 Dark 色板映射到项目主题变量。
- 针对 Explorer、Tabs、文件列表选中态做组件级覆写。
- 增加 Explorer 和文件列表字号设置、默认值和恢复默认按钮；图标视图也接入文件列表字号。
- 统一滚动条样式的后续收口发生在 4/7，未归入本段压缩。

### Path / Empty Directory / Driver Semantics

- 统一前端 remote path 身份，修复多选、拖放、目录树刷新中 raw path 和 normalized path 混用的问题。
- 后端目录 listing 统一返回空切片 `[]`，避免空目录经 Wails 传成 `null` 后导致前端渲染异常。
- 修复目录树空/非空目录切换导致的递归渲染和初始化问题，并拆掉 workspace 组件静态循环依赖。

### Modification Time Preservation

- 上传和下载增加文件 mtime 保留：`Local/SFTP` 尽量写真实 mtime，`S3/OSS` 用 metadata 记录原始 mtime 并在下载时恢复。
- 目录 mtime 通过 transfer finalizer 延迟到所有子任务完成后回填。
- 修复 finalizer 任务先完成后登记的竞态，以及对象存储无显式目录 marker 时虚拟目录未纳入下载计划的问题。
- 对没有显式目录 marker 的对象存储虚拟目录，只能用子文件最新 mtime 作为近似值，这是对象存储语义边界。

### Validation Snapshot

- 该阶段多次执行并恢复通过：`go test ./...`、`go test ./folder/...`、`go test ./transfer/...`、前端 `type-check`、`build-only`、相关 Vitest 单测、`git diff --check`。
- 运行副作用文件 `logs/app.log` 多次出现变更，原则上不作为功能变更提交。

## 2026-04-06 - Fix OSS/S3 Directory Move Semantics

### Root Cause

- OSS 和 S3 驱动之前把 `Move()` / `Rename()` 实现成单对象 `CopyObject + DeleteObject`。
- 这对普通文件成立，但对象存储目录是 prefix/listing 语义：
  - 拖动目录时，源路径实际代表 `dir/` 下的一组 objects。
  - 单对象 copy 不会递归复制子对象。
  - `Delete()` 也只在入参已经带 `/` 时才递归删除 prefix，而上层 `fileops.MoveEntry()` 会把路径清洗成不带尾部 `/`。
- 因此 OSS 目录拖动会失败或不完整；S3 也存在同类隐患。

### Completed Changes

- `OSS` / `S3` 的 `Copy()`：
  - 先 `Stat()` 判断源路径类型。
  - 文件走原有单对象 copy。
  - 目录走 prefix 级递归 copy。
- `OSS` / `S3` 的 `Move()`：
  - 文件走单对象 copy + delete。
  - 目录走 prefix copy，然后 prefix delete。
  - 增加目标 prefix 位于源 prefix 内部的防御性校验，避免直接调用驱动时把目录移动到自身子目录。
- `OSS` / `S3` 的 `Rename()`：
  - 统一走 `Move()`，让目录重命名也获得 prefix 语义。
- `OSS` / `S3` 的 `Delete()`：
  - 不再只依赖尾部 `/` 判断目录。
  - 对不带 `/` 的路径也会先 `Stat()`，确认是目录时执行 prefix delete。

### Tests

- `folder/alibaba-oss/driver_test.go`
  - 新增 `MoveDirectory` 集成测试：目录内有子文件，移动目录后源子文件应消失，目标子文件应存在。
- `folder/s3/driver_test.go`
  - 新增同等 `MoveDirectory` 集成测试。
- 无云端凭据时，这两组集成测试继续按既有规则自动跳过。

### Verification

- `go test ./folder/... ./fileops` 通过。
- `go test ./...` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-06 - Fix File List Header Resize Interaction

### Root Cause

- 中央工作区列表视图的表头列宽拖动依赖 `mousemove/mouseup` 更新宽度，但没有把 resize 状态和文件面板的拖放 / 排序交互隔离。
- 在 WebView 里拖动 header/resizer 时，容易出现：
  - resize 后触发 header click 排序。
  - 触发表格区域的 drag/drop 状态。
  - 鼠标离开窗口或窗口失焦后 resize 状态没有稳定收口。

### Completed Changes

- 新增 `resizingColumn` 状态作为显式交互锁。
- 表头 click 改为 `onHeaderClick()`：
  - resize 期间不触发排序。
- 表头按钮增加 `@dragstart.prevent`：
  - 避免浏览器 / WebView 把 header 拖动识别成原生 drag。
- 列表根区域的 `dragover/dragleave/drop` 在 resize 期间直接屏蔽并清理 drop 指示。
- 列宽 resize 的全局 `mousemove/mouseup` 改为 capture 阶段监听：
  - 避免被内部元素事件时序干扰。
- 增加收口兜底：
  - `Escape`
  - `window blur`
  - `document visibilitychange`
  - 组件卸载 / transient state reset

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-06 - Refine File List Adjacent Column Resize

### Root Cause

- 列表视图表头拖动之前仍是“只修改当前列宽度”：
  - 拖动某列右侧 resizer，只会改该列宽。
  - `name` 列使用 `minmax(width, 1fr)`，会让剩余空间吸收变化，表现不像原生文件列表的相邻分割线。
- 用户期望的行为是：
  - 拖中间分割线向左，左列缩小、右列扩大。
  - 拖中间分割线向右，左列扩大、右列缩小。
  - 只影响分割线两侧相邻两列，不影响其他列。

### Completed Changes

- `ColumnResizeState` 从单列状态改为成对状态：
  - `leftKey`
  - `rightKey`
  - `startLeftWidth`
  - `startRightWidth`
- resize 计算改为 clamp 后的相邻列联动：
  - `left = startLeft + delta`
  - `right = startRight - delta`
  - 总宽度保持不变。
  - 两列都遵守各自最小宽度。
- 所有列表列宽都改为确定的 `px` track：
  - 不再让 `name` 列用 `1fr` 吸收剩余空间。
- 最后一列不再显示 resize hit area，因为右侧没有相邻列。
- 表头取消可见 divider：
  - 移除 resizer hover 背景。
  - 移除 header bottom border。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-06 - Add Mtime Preservation Regression Tests

### Review Findings

- 目录 mtime 保留仍有两个测试缺口：
  - 缺少完整 `processFollowUp(followUpUpload)` 路径的回归测试。
  - 缺少“显式目录 marker mtime 优先于虚拟目录推导 mtime”的回归测试。
- 新增显式 marker 优先级测试后，暴露了一个真实缺陷：
  - `buildLocalDownloadPlan()` 只记录目录时间值，没有记录来源。
  - 当显式目录 marker 的 mtime 早于子文件 mtime 时，后续虚拟目录推导会用子文件时间覆盖 marker 时间。

### Completed Changes

- 为目录 mtime 聚合新增来源标记：
  - `directoryModTimeSource.modTime`
  - `directoryModTimeSource.explicit`
- `mergeDirectoryModTime()` 现在遵循以下规则：
  - 显式目录 marker 有有效 mtime 时，优先级高于虚拟目录推导。
  - 显式目录没有 mtime 时，仍允许后续用子文件时间作为虚拟目录近似值。
  - 虚拟目录之间仍使用子文件最新 mtime 作为近似值。
- 新增完整 follow-up upload 路径测试：
  - 通过真实 `Local` 连接执行 `processFollowUp(followUpUpload)`。
  - 验证 follow-up upload 完成后，deferred directory finalizer 会回填目标目录 mtime。
- 新增显式目录 marker 优先级测试：
  - 同一目录同时存在显式 marker mtime 和更晚的子文件 mtime。
  - 验证最终目录计划保留 marker mtime，不被子文件推导覆盖。

### Verification

- `go test ./transfer/...` 通过。
- `go test ./...` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-06 - Replace Wails Default Logo

### Design Direction

- 根据项目定位，Logo 使用“文件夹 + 远程连接节点”的组合：
  - 文件夹代表文件管理器主语义。
  - 顶部连接节点代表 `Local / SFTP / S3 / OSS` 等多接入模式。
  - 色彩使用当前项目接近 VS Code 2026 Dark 主题的青蓝主色，避免继续沿用 Wails 默认视觉。
- 采用可维护 SVG 作为源文件，再生成 Wails 构建和前端需要的 PNG/ICO 资源。

### Completed Changes

- 新增矢量源：
  - `assets/logo.svg`
- 从同一个源文件生成并替换：
  - `assets/logo.png`
  - `build/appicon.png`
  - `build/windows/icon.ico`
  - `frontend/public/favicon.ico`
- 生成过程中发现 ImageMagick 对部分 SVG 滤镜/渐变输出不稳定，已把图标主体改成基础实体色矢量形状，确保 PNG/ICO 转换结果稳定。

### Verification

- 图标资源格式和尺寸已检查：
  - `assets/logo.png` 为 `1024x1024` PNG。
  - `build/appicon.png` 为 `1024x1024` PNG。
  - `build/windows/icon.ico` 包含 `256/128/64/32/24/16` 多尺寸图标。
  - `frontend/public/favicon.ico` 为 `32x32` favicon。
- `git diff --check` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-06 - SFTP Private Key Path Support

### Requirement

- SFTP 私钥认证需要支持两种输入：
  - `privateKey`: 私钥文本本身。
  - `privateKeyPath`: 本地私钥文件路径。
- 两者存在其一即可连接；如果给出路径，驱动读取文件内容后作为私钥认证材料。

### Completed Changes

- 后端 `folder/sftp.Options` 新增 `privateKeyPath` 字段，并把认证校验改为 `password / privateKey / privateKeyPath` 三选一。
- SFTP 驱动连接时新增私钥材料解析链路：
  - 优先使用 `privateKey` 文本。
  - 如果没有文本，则读取 `privateKeyPath` 指向的文件。
  - 为兼容旧误用场景，如果 `privateKey` 不像私钥文本且没有显式 `privateKeyPath`，会尝试把 `privateKey` 当路径读取。
  - 路径支持 `~` 展开，并会拒绝空文件和目录路径。
- 前端 SFTP 表单新增 `Private Key Path` 输入项，原私钥输入项改名为 `Private Key Text`。
- 前端校验改为密码、私钥路径、私钥文本三选一。
- 更新 SFTP README 和集成测试环境变量说明，新增 `SFTP_PRIVATE_KEY_PATH`。
- 新增 `folder/sftp/options_test.go`，覆盖：
  - `privateKeyPath` 通过配置校验。
  - 从私钥路径读取文件内容。
  - 兼容旧配置中把路径填入 `privateKey` 的场景。
  - 同时存在文本和路径时优先使用文本。
  - `privateKey` 不是私钥文本且存在显式路径时使用显式路径。

### Verification

- `go test ./folder/sftp` 通过。
- `go test ./...` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。
- `git diff --check` 通过。

## 2026-04-06 - Fix SFTP Absolute RootPath Boundary Check

### Root Cause

- SFTP `fullPath()` 之前使用 `path.Join("/", RootPath, relPath)` 生成远端路径，但边界检查里的 root 使用 `"/" + RootPath`。
- 当 `RootPath` 是 `/home/ping` 时：
  - 实际路径会生成 `/home/ping`。
  - 边界 root 会错误生成 `//home/ping`。
  - 因此合法根路径会被误判为越界，导致 `List("")` 等操作失败。
- 当 `RootPath` 是 `/` 时，`Validate()` 会把它归一成空 root，所以没有触发这个问题。

### Completed Changes

- 新增 `normalizeRemoteRoot()`，把 SFTP remote root 统一规整为单一绝对路径：
  - `""` 和 `/` -> `/`
  - `/home/ping` 和 `home/ping` -> `/home/ping`
- `fullPath()` 改为先归一 root，再拼接相对路径和做目录边界检查。
- 新增测试覆盖：
  - `RootPath=/home/ping` 下根目录、子路径、调用方传绝对路径、越界路径。
  - `RootPath=/` 下根目录和子路径。

### Verification

- `go test ./folder/sftp` 通过。
- `go test ./...` 通过。
- 使用真实 SFTP 连接验证 `rootPath=/home/ping` 且用户 `ping` 时 `List("")` 通过，返回 `31` 个条目。

## 2026-04-06 - Normalize S3 and OSS Prefix Roots

### Finding

- S3/OSS 也存在和 SFTP root 类似的路径语义风险。
- 旧逻辑只去掉 `prefix` 尾部 `/`，没有去掉前导 `/`：
  - `/home/ping` 会变成 `/home/ping/`。
  - `/` 会变成 `/`。
- 对象存储 key 通常按 `home/ping/` 这类无前导 slash 的 prefix 组织，因此前端如果把 prefix 填成类似本地路径的 `/home/ping`，驱动会查找错误前缀。

### Completed Changes

- S3 和 OSS 都新增 `normalizePrefix()`：
  - `""` 和 `/` -> `""`
  - `/home/ping` 和 `home/ping` -> `home/ping/`
  - 多余首尾 slash 会被收口。
- `DriverOptions.Root` 与驱动自身 `prefix` 合并时也统一走 `normalizePrefix()`，避免生成带前导 slash 的对象 key 前缀。
- 新增 S3/OSS 单测覆盖：
  - prefix 归一化。
  - `/root` + `/child/` 合并为 `root/child/`。
  - `fullKey("")` 和 `fullKey("/file.txt")` 不产生前导 slash。

### Verification

- `go test ./folder/s3 ./folder/alibaba-oss ./folder/sftp` 通过。
- `go test ./...` 通过。

## 2026-04-06 - Rewrite Top-Level README Documents

### Goal

- Replace the default Wails template README with project-specific documentation.
- Keep `README.md` as the default English entry point.
- Add `README_cn.md` for Chinese readers.

### Completed Changes

- Rewrote `README.md` to cover:
  - Project positioning and current storage backends.
  - Main feature set.
  - Backend and frontend architecture.
  - Storage driver semantics and object-store directory boundaries.
  - Configuration files and example connection definitions.
  - Development and test commands.
  - Known boundaries, including SFTP host key verification and object-store prefix semantics.
- Added `README_cn.md` with equivalent Chinese documentation.

### Verification

- `git diff --check` 通过。

## 2026-04-06 - Add Commercial Authorization Notice To AGPL License

### Goal

- Keep the project under AGPL-3.0-only.
- Add a commercial authorization path for modified non-open-source use.
- Treat accepted bugfix or functional enhancement contributions, excluding comment-only/non-functional edits, as qualifying for contributor commercial authorization.

### Design Note

- Did not rewrite the GNU AGPL v3 text itself because the FSF license text requires verbatim copying and GPL/AGPL terms do not allow arbitrary further restrictions on the AGPL grant.
- Added a project-specific licensing notice before the verbatim AGPL text.
- The notice phrases the commercial license requirement as an alternative path for users who do not make corresponding modified source code available under AGPL-3.0-only to entitled recipients or network users.
- Added an inbound contribution license grant so accepted external contributions can be included in future commercial-license distributions.

### Completed Changes

- Prepended `LICENSE` with:
  - Project copyright notice.
  - AGPL-3.0-only statement.
  - Dual licensing and one-time commercial authorization notice.
  - Contributor commercial authorization definition.
  - Contribution relicensing grant for accepted contributions.
- Updated `README.md` and `README_cn.md` license sections to point to `LICENSE` and summarize the commercial authorization rule.

### Verification

- `git diff --check` 通过。

## 2026-04-06 - Add Tag Release Build GitHub Action

### Goal

- Build release artifacts automatically when a Git tag is pushed.
- Cover Windows, macOS, and Linux for amd64 and arm64.

### Completed Changes

- Added `.github/workflows/release-build.yml`.
- The workflow builds these targets with Wails:
  - `linux/amd64` on `ubuntu-22.04`
  - `linux/arm64` on `ubuntu-22.04-arm`
  - `windows/amd64` on `windows-2025`
  - `windows/arm64` on `windows-11-arm`
  - `darwin/amd64` on `macos-15-intel`
  - `darwin/arm64` on `macos-15`
- Each build is packaged as a `.tar.gz` artifact and uploaded to the Actions run.
- A follow-up release job downloads all artifacts and uploads them to the GitHub Release for the pushed tag.

### Notes

- Linux builds install Wails GTK/WebKitGTK dependencies before building.
- The workflow uses native runner architecture labels instead of relying on Wails/CGO cross-compilation.
- ARM64 runner availability depends on GitHub-hosted runner support for the repository/plan.

### Verification

- `ruby -e 'require "yaml"; YAML.load_file(".github/workflows/release-build.yml")'` 通过。
- `git diff --check` 通过。

## 2026-04-06 - Pin GitHub Actions Node.js To 24.14.0 LTS

### Finding

- Node.js `24.14.0` is a current LTS release in the `Krypton` line.
- The frontend `package.json` engine range is `^20.19.0 || >=22.12.0`, so Node.js `24.14.0` satisfies the project constraint.

### Completed Changes

- Updated `.github/workflows/release-build.yml` from `NODE_VERSION: '22'` to `NODE_VERSION: '24.14.0'`.

### Verification

- `ruby -e 'require "yaml"; YAML.load_file(".github/workflows/release-build.yml")'` 通过。
- `git diff --check` 通过。

## 2026-04-06 - Fix GitHub Release Job Repository Context

### Root Cause

- The `release` job downloaded artifacts but did not check out the repository.
- `gh release view/create/upload` can invoke git to infer repository context, so the job failed with `fatal: not a git repository`.

### Completed Changes

- Added `actions/checkout@v4` to the `release` job.
- Added `GH_REPO: ${{ github.repository }}`.
- Passed `--repo "$GH_REPO"` to `gh release view`, `gh release create`, and `gh release upload`.

### Verification

- `ruby -e 'require "yaml"; YAML.load_file(".github/workflows/release-build.yml")'` 通过。
- `git diff --check` 通过。

## 2026-04-06 - Move Production Unix Default Config Path To User Config Directory

### Root Cause

- The app resolved the default app config path from the current working directory.
- In a production macOS `.app` launch, the working directory can be unstable and may not be the project/source directory.
- This caused production builds to look for `config.yaml` in the wrong place.

### Completed Changes

- Updated default app config path resolution:
  - Explicit `--config` / `-c` still wins.
  - If the working directory looks like the project source root, the app still uses working-directory config files. This keeps `wails dev` convenient.
  - Otherwise, on non-Windows platforms, the default path is `~/.config/file-browser/config.yaml`.
  - Windows keeps the existing working-directory fallback for now.
- Because relative app config paths are normalized against the resolved config file directory, default `connections.yaml`, `state.yaml`, logs, temp paths, and transfer paths now also resolve under `~/.config/file-browser/` in production Unix launches.
- `LoadAppConfig()` now creates the resolved config file parent directory if it does not exist.
- Added config tests for source-root fallback, Unix user config fallback, directory creation, and existing `config.yml` discovery under the user config directory.
- Updated `README.md` and `README_cn.md` configuration sections.

### Verification

- `go test ./config` 通过。
- `go test ./...` 通过。
- `git diff --check` 通过。

## 2026-04-06 - Change Default Transfer Temp Directory

### Goal

- Move the default transfer temporary directory out of the project/current working directory.
- Use OS-specific temp locations by default.

### Completed Changes

- Replaced the old relative default `tmp/transfers` with `config.DefaultTransferTempDirPath()`.
- New defaults:
  - Unix-like systems: `/tmp/file-browser/transfers`
  - Windows: `%USERPROFILE%\AppData\Local\Temp\file-browser\transfers`
- Updated app config normalization so an empty `transfer.temp_dir` uses the OS-specific default path.
- Updated `transfer.Service` fallback paths so empty runtime temp dir also uses the same OS-specific default.
- Updated frontend placeholder text and README configuration notes.
- Extended config tests to assert the default transfer temp directory.

### Verification

- `go test ./config ./transfer` 通过。
- `go test ./...` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。
- `git diff --check` 通过。

## 2026-04-06 - Correct Default Transfer Temp Directory Suffix

### Completed Changes

- Corrected the default transfer temp directory suffix from singular `transfer` to plural `transfers`.
- Final defaults:
  - Unix-like systems: `/tmp/file-browser/transfers`
  - Windows: `%USERPROFILE%\AppData\Local\Temp\file-browser\transfers`
- Updated README and frontend setting placeholder text accordingly.

### Verification

- `go test ./config ./transfer` 通过。
- `cd frontend && npm run type-check` 通过。
- `git diff --check` 通过。

## 2026-04-07 - Add Windows EXE Suffix To Release Build Output

### Goal

- Keep release artifact archive names unchanged.
- Ensure Windows binaries generated by `wails build` have the `.exe` suffix inside the archive.

### Completed Changes

- Added `binary_ext` to the release build matrix.
- Set `binary_ext: '.exe'` only for Windows targets.
- Updated the Wails output name to `${APP_NAME}-${matrix.suffix}${matrix.binary_ext}`.
- `.tar.gz` artifact names remain unchanged.

### Verification

- `ruby -e 'require "yaml"; YAML.load_file(".github/workflows/release-build.yml")'` 通过。
- `git diff --check` 通过。

## 2026-04-07 - File Panel Back/Forward Navigation History

### Goal

- Add back/forward navigation for the central file browser panel.
- Keep the feature scoped to file browser tabs only; settings, forms, welcome tabs, and other panels should not react.
- Keep history independent per file browser panel and avoid adding visible toolbar buttons.

### Design

- Implemented a local per-component navigation history stack in `ConnectionOverviewTab`.
- Added a pure helper `navigationHistory.ts` so stack behavior is testable without mounting the full Element Plus/Wails component.
- History records successful directory navigations only:
  - User path navigation uses `push`.
  - Initial load, refresh, transfer refresh, directory refresh, and connection config reset use `reset` so invalidated navigation state is cleared.
  - Back/forward history navigation itself uses `none` so the selected history index can move without creating a duplicate entry.
- Shortcut handling stays local to the file browser root via capture-phase keydown handling instead of using global `DEFAULT_SHORTCUTS`; this prevents non-file tabs from responding.
- Supported shortcuts:
  - macOS: `Cmd + [` and `Cmd + ]`.
  - Windows/Linux and cross-platform fallback: `Alt + Left` and `Alt + Right`.
- Editable elements are ignored so path history does not trigger while typing in inline create inputs or other form controls.

### Completed Changes

- Added `frontend/src/components/Workspace/navigationHistory.ts`.
- Added `frontend/src/components/Workspace/__tests__/navigationHistory.spec.ts`.
- Updated `ConnectionOverviewTab.vue` to record navigation history on successful loads and perform local back/forward navigation from the history stack.
- Relaxed the file viewport keydown early return so empty directories can still handle navigation-related keyboard behavior such as Backspace parent navigation.

### Verification

- `cd frontend && npm run test:unit -- src/components/Workspace/__tests__/navigationHistory.spec.ts` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。
- `git diff --check` 通过。

### Follow-up Adjustment

- Expanded shortcut handling from viewport-only bubbling to a window-level listener guarded by file-browser scope checks.
- The active file panel now responds when focus is on its tab group/header, not only when a file row/tile or viewport has focus.
- Scope guards prevent inactive cached panels and non-file tabs from handling the shortcuts:
  - Events inside the panel root are accepted.
  - Events inside the central shell are accepted only when the panel's tab is the active tab in the active tab group.
  - Editable targets remain ignored.

### Welcome Page Shortcut Hint

- Added file-panel back/forward shortcut descriptions to the welcome page shortcut list.
- Shortcut display is platform-aware:
  - macOS: `Cmd + [` / `Cmd + ]`
  - Windows/Linux: `Alt + Left` / `Alt + Right`

## 2026-04-07 - Add Japanese Locale

### Goal

- Add Japanese UI translations and expose Japanese in the settings locale selector.

### Completed Changes

- Added `frontend/src/locales/ja.json` with the same message structure as `zh` and `en`.
- Added the `ja` option to the settings locale selector.
- Updated the settings store supported locale whitelist to include `ja`.
- Added a locale parity unit test to prevent future `zh/en/ja` key drift.

### Verification

- `cd frontend && npm run test:unit -- src/__tests__/localeParity.spec.ts` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。
- `git diff --check` 通过。

## 2026-04-07 - Refresh Explorer Parent After Inline Directory Creation

### Goal

- After creating a directory from the central file panel, update the corresponding Explorer tree parent without globally refreshing the tree.

### Design

- Reused the existing `workspace:connection-directory-refresh` event because Explorer already handles it as a loaded-node local refresh.
- Extended the event payload with:
  - `source: 'mutation'` for direct file operations such as inline directory creation.
  - optional `origin` so the source file panel can ignore its own event after it has already reloaded.
- The Explorer tree keeps its existing behavior: refresh only the target parent node when that node is already loaded in `childrenByKey`; otherwise no eager load occurs.

### Completed Changes

- `ConnectionOverviewTab` now emits a directory refresh event after successful inline directory creation and local reload.
- `ConnectionOverviewTab` ignores refresh events from its own origin to avoid duplicate reloads.
- `useConnectionDirectoryRefresh` now supports mutation-originated refreshes and optional origin metadata.

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run test:unit -- src/__tests__/localeParity.spec.ts src/components/Workspace/__tests__/navigationHistory.spec.ts` 通过。
- `cd frontend && npm run build-only` 通过。
- `git diff --check` 通过。

## 2026-04-07 - Add Master Branch Auto Build Workflow

### Goal

- Add an automated GitHub Actions build for every push to `master`.
- Reuse the existing release build matrix and packaging shape, but do not create or update GitHub Releases.

### Completed Changes

- Added `.github/workflows/master-build.yml`.
- The workflow builds the same six targets as `Release Build`:
  - Linux amd64 / arm64
  - Windows amd64 / arm64
  - macOS amd64 / arm64
- Windows binaries keep the `.exe` suffix inside the tarball.
- Artifacts are uploaded as GitHub Actions artifacts with a 14-day retention period.
- Concurrency cancels in-progress master builds when a newer master push arrives.

### Verification

- `ruby -e 'require "yaml"; YAML.load_file(".github/workflows/master-build.yml")'` 通过。
- `git diff --check` 通过。

## 2026-04-07 - Fix WebView2 Internal Cross-Panel Drop Detection

### Problem

- On Unix/WebKit, dragging entries between split file panels triggered cross-connection transfer correctly.
- On Windows/WebView2, releasing after the same drag produced no upload action and no console error.

### Root Cause

- Internal file-panel drag detection required `dataTransfer.types` to contain the custom MIME marker `application/x-splitpane-drag`.
- WebView2 can omit or hide that custom type during `dragenter/drop`.
- The application already stores the active internal drag payload in module-level state for security-restricted drag phases, but the drop resolver and panel drop layer still rejected the event before reading that state.

### Completed Changes

- Added `isInternalDragEvent()` to `splitPaneDragState`.
- `connectionEntryDrop.resolveConnectionEntryDragPayload()` now accepts either:
  - the custom `DataTransfer` marker, or
  - an active internal drag payload in shared state.
- `usePanelFileDrop` now uses the same internal-drag detector for split-panel drop routing.
- External OS file drags still win when `dataTransfer.types` contains `Files`, so stale internal drag state cannot hijack Finder / Explorer uploads.
- Added a WebView2-style regression test where `dataTransfer.types` does not include the custom marker, but the active internal drag payload is still resolved.

### Verification

- `cd frontend && npm run test:unit -- src/composables/__tests__/connectionEntryDrop.spec.ts` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-07 - Global and File Panel Search Shortcuts

### Goal

- `Ctrl/Cmd + Shift + F` opens the global search panel.
- `Ctrl/Cmd + F` opens a local search box inside the active file browser tab only.
- Local search filters the current loaded list without calling the backend and without scrolling with the file list.

### Completed Changes

- Updated the global shortcut preset so `search` uses `CmdOrCtrl+Shift+F` instead of `CmdOrCtrl+F`.
- Added the global search shortcut handler in `AppShellView`; it switches the left panel to the existing search view.
- Added file-panel local search in `ConnectionOverviewTab`:
  - fixed overlay at the top-right of the list/grid body
  - close button
  - `Ctrl/Cmd + F` toggles the local search box: open when closed, close and clear filtering when open
  - Enter commits the search query
  - Escape closes local search
  - match case option
  - whole-text option
  - regex option with inline invalid-regex feedback
- Local search filters the sorted in-memory entries and the filtered result is also used for keyboard navigation, select-all, active item lookup, and selected item counting.
- Search state is cleared on directory load/reload so stale filters do not survive destructive state changes.
- Updated the welcome page shortcut list and reduced the shortcut row/key font sizes so the additional search shortcuts fit.
- Added `zh` / `en` / `ja` locale strings for the new shortcut descriptions and local-search UI.

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run test:unit -- src/__tests__/localeParity.spec.ts src/composables/__tests__/connectionEntryDrop.spec.ts` 通过。
- `cd frontend && npm run build-only` 通过。
- `git diff --check` 通过。

## 2026-04-07 - Search Result Navigation to File Panel

### Goal

- Clicking a global search result should navigate the central workspace to the result connection and the containing directory.
- The matching file or directory should be selected and scrolled into view without forcing a jump past the current viewport bounds.
- Search result rows should occupy a stable full width in the search panel.

### Completed Changes

- Added a one-shot `revealPath` state to the workspace store for each connection.
- Search results now set `revealPath` before opening the target connection tab.
- `ConnectionOverviewTab` consumes the reveal target after loading or restoring the directory cache:
  - selects the target entry
  - sets it as the active path
  - calls `scrollIntoView({ block: 'nearest', inline: 'nearest' })`
- Directory results open the directory itself; file results open the parent directory and reveal the file.
- Search result rows now use full-width button rows instead of grid gap sizing, so each result has a fixed row width within the panel.

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run test:unit -- src/__tests__/localeParity.spec.ts` 通过。
- `cd frontend && npm run build-only` 通过。
- `git diff --check` 通过。

## 2026-04-07 - Full Unit Test Sweep

### Result

- `go test ./...` 通过。
- 首次 `cd frontend && npm run test:unit` 失败在 `SplitPane.spec.ts`，原因是 `SplitPanePanel` 当前使用 `useI18n()`，但测试挂载没有安装 i18n 插件。
- 已在 `frontend/src/components/SplitPane/__tests__/SplitPane.spec.ts` 为测试挂载补充最小 i18n plugin。
- 修复后 `cd frontend && npm run test:unit` 通过。
- `cd frontend && npm run type-check` 通过。
- `git diff --check` 通过。

### Note

- `frontend/tmp-tabs-debug.spec.ts` 当前仍会在全量 Vitest 中打印调试输出；它不导致测试失败，但会污染测试日志。

## 2026-04-07 - Normalize Scrollbar Appearance Across WebViews

### Goal

- WebKit and WebView2 render native scrollbars differently.
- Normalize the app scrollbar appearance so Explorer, file lists, task panels, notification panels, tabs, and Element Plus internal scroll containers use a thin theme-colored style.

### Completed Changes

- Added global scrollbar variables:
  - `--ui-scrollbar-size: 8px`
  - `--ui-scrollbar-thumb-size: 5px`
- Added global native scrollbar rules in `frontend/src/assets/main.css`:
  - Firefox `scrollbar-width: thin`
  - WebKit/WebView2 `::-webkit-scrollbar` and thumb styling
  - transparent track/corner
  - theme-driven thumb and hover colors
- Added Element Plus scrollbar bridge for `.el-scrollbar__bar` / `.el-scrollbar__thumb`.
- Updated `TabBar` to use the same scrollbar variables instead of its previous local `2px` WebKit override.

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-07 - Harden Cross-Panel File Drag Drop for WebView2

### Problem

- Cross-panel file entry drag/drop worked on Unix WebKit, but Windows WebView2 could release the drag without triggering the cross-connection upload path.
- The previous MIME fallback was not sufficient because the actual drop target in WebView2 can be the leaf `TabGroup`, and WebView2 can also clear or reorder drag lifecycle events differently from WebKit.

### Root Cause

- The leaf `TabGroup` installed file-drop handling for OS file drops, but `enablePanelDrag` was hard-coded to `false`.
- On WebView2, a drop that lands on the leaf `TabGroup` can be consumed before the parent `SplitPanePanel` receives the internal panel-drop event.
- File item `dragend` and window-level drag cleanup were also clearing `activeInternalDrag` immediately; if WebView2 delivers `dragend` before `drop`, the payload is gone before the target can consume it.

### Completed Changes

- `TabGroup` now receives `enablePanelDrag`, injects its owning `SplitPanePanel` identity, and can emit `panelDrop` directly from the leaf group.
- `TabNodeRenderer` forwards `panel-drop` from leaf `TabGroup` nodes, not only from nested `SplitPane` nodes.
- `usePanelFileDrop` no longer calls `preventDefault()` for internal drags it is not allowed to handle, so parent handlers can still receive them.
- Internal drop handling now calls `stopPropagation()` only after the current target has accepted and emitted the drop event, avoiding duplicate parent handling.
- Internal drag `DataTransfer` now writes both:
  - custom MIME marker: `application/x-splitpane-drag`
  - standard `text/plain` marker for stricter WebView2 drag/drop activation
- Internal drag state cleanup is delayed briefly and identity-guarded so a WebView2 `dragend` cannot clear payload state before the target `drop` consumes it.

### Compatibility Notes

- WebKit still uses the custom MIME path when it is available.
- WebView2 can fall back to shared in-memory drag state and the `text/plain` marker.
- External OS file drops remain excluded by `Files` detection and continue through the Wails file-drop path.

### Verification

- `cd frontend && npm run test:unit -- src/composables/__tests__/connectionEntryDrop.spec.ts` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。
- `git diff --check` 通过。

## 2026-04-07 - Finalize File Panel Selection and Drag Overlay Cleanup

### Goal

- Prevent WebKit/WebView2 native text selection from reappearing when double-clicking file names.
- Ensure cross-panel internal drag overlays are always cleared even when WebView2 delivers drop events to a child component that stops bubbling.

### Frontend Review Notes

- Vite / Vue configuration does not need platform-specific branching for this issue.
- The right place to fix this is the shared browser-event compatibility layer:
  - file-panel selection suppression in `ConnectionOverviewTab`
  - split/tab drag overlay lifecycle in `usePanelFileDrop`
- WebKit and WebView2 differ in drag/drop ordering and propagation, so overlay cleanup must not depend on a single bubbling `drop` path.

### Completed Changes

- `ConnectionOverviewTab` now blocks `selectstart` inside the file browser except editable inputs.
- The second click in a double-click sequence calls `preventDefault()` and clears native selection before WebView/WebKit can select the file name text.
- File-browser CSS now explicitly applies `user-select: none` and `-webkit-user-select: none` to non-input descendants.
- `ConnectionOverviewTab` listens for `dragend` / `drop` in capture phase, so its local drop indicators are cleared even if a child drop handler stops propagation.
- File-panel global `dragend` / `drop` cleanup is scoped to the active file tab, the event target panel, or the panel that owns the current drag session, so inactive KeepAlive tabs are not reset accidentally.
- `usePanelFileDrop` now installs capture-phase window-level `drop` / `dragend` reset hooks, plus `blur` / `visibilitychange` fallbacks, so TabGroup/SplitPanePanel masks cannot remain stuck after a child handles the drop.

### Verification

- `cd frontend && npm run test:unit -- src/composables/__tests__/connectionEntryDrop.spec.ts` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。
