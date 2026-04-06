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

### 2026-04-01

- 完成项目结构探索与模块拆分。
- 建立 `developing.md`，记录目标架构、阶段计划、模块建议与风险。
- 完成配置模块第一版实现：
  - `AppConfig`
  - `ConnectionsConfig`
  - 默认值与路径标准化
  - YAML/JSON 加载与保存
- 完成启动链路第一版实现：
  - `bootstrap.Initialize()`
  - `bootstrap.ParseStartupOptions()`
  - `main.go` 使用配置驱动日志初始化
- 验证结果：
  - `go test ./config ./bootstrap` 通过
  - `go test ./...` 通过
- 完成连接模块第一版实现：
  - `connection.FileRepository`
  - `connection.Service`
  - 连接状态缓存
  - 连接打开/关闭与驱动实例生命周期管理
  - `App` 新增连接相关 Wails 绑定方法
- 完成运行时驱动注册补齐：
  - `main.go` 已通过空导入注册 `Local`、`S3`、`OSS`、`SFTP`
- 完成前后端桥接更新：
  - 重新生成 `frontend/wailsjs/go/main/App.*`
  - 重新生成 `frontend/wailsjs/go/models.ts`
- 最新验证结果：
  - `go test ./connection ./config ./bootstrap` 通过
  - `go test ./...` 通过
- 完成前端工作区第一版实现：
  - `frontend/src/views/AppShell/AppShellView.vue`
  - `frontend/src/stores/connections.ts`
  - `frontend/src/stores/workspace.ts`
  - `frontend/src/components/ExplorerTree/ExplorerPanel.vue`
  - `frontend/src/components/Workspace/*`
- 完成前端路由与快捷键接线：
  - 根路由切换到 `AppShell`
  - `CtrlOrCmd+Shift+N` 打开新建连接页
  - `CtrlOrCmd+.` 打开设置页
- 当前前端完成度说明：
  - Explorer 已接真实连接数据
  - 新建/编辑连接已接真实保存接口
  - 双击连接可建立连接并在中央打开真实文件面板
  - 文件面板当前已具备目录浏览、新建目录、重命名、删除
  - 上传下载、跨面板拖放、文件预览仍未接入
- 最新前端验证结果：
  - `npm run type-check` 通过
  - `npm run build` 通过
- 完成文件浏览最小链路：
  - `fileops.Service.ListDirectory()`
  - `App.ListConnectionDirectory()`
  - 重新生成 Wails 绑定
  - 中央连接 tab 已接入真实目录列表
- 完成基础文件操作：
  - `fileops.Service.CreateDirectory()`
  - `fileops.Service.RenameEntry()`
  - `fileops.Service.DeleteEntry()`
  - `App` 新增对应 Wails 绑定
  - 文件面板已接入“新建目录 / 重命名 / 删除”
- 最新综合验证结果：
  - `go test ./fileops ./connection ./...` 通过
  - `npm run type-check` 通过
  - `npm run build` 通过
- 完成 Explorer 目录树增强：
  - 左侧连接节点已支持懒加载目录树。
  - 目录树当前仅渲染目录，不渲染文件。
  - 连接节点继续保持“展开目录树”和“打开中央文件面板”分离。
- 完成 Explorer 与文件面板联动增强：
  - 工作区新增连接级路径状态。
  - 点击左侧目录节点会在对应连接 tab 中直接定位到该目录。
  - 文件面板内部切换目录后会回写路径状态，保证左侧导航和中央上下文使用同一套路径语义。
- 完成异步搜索最小链路：
  - 新增 `search.Service`，支持多连接并发搜索和请求取消。
  - `App` 新增 `StartSearch()` 与 `CancelSearch()` Wails 绑定。
  - 搜索结果通过 `search:event` 渐进推送到前端，而不是等待整体完成后一次性返回。
  - 左侧搜索面板已接真实搜索输入、连接范围筛选、结果列表和错误展示。
  - 点击搜索结果会打开对应连接 tab，并定位到结果所在目录。
- 完成任务面板与下载任务最小链路：
  - 新增 `transfer.Service`，封装 `folder.TransferManager` 并向前端暴露任务查询与控制接口。
  - `App` 新增下载任务、任务列表、取消任务、移除任务、清理已完成任务等 Wails 绑定。
  - 文件面板已支持对文件发起“下载到临时目录”任务。
  - 右侧任务面板已可展示传输方向、远程/本地路径、进度、速度、错误和操作按钮。
- 完成外部拖入上传最小链路：
  - `transfer.Service` 新增本地路径上传到当前连接目录的任务提交接口。
  - `App` 新增 `UploadConnectionLocalPath()` Wails 绑定。
  - 正式应用壳已启用 `useFileDrop()`，外部文件拖入中央连接 tab 会自动转成上传任务。
  - 当前上传链路先支持文件，不支持目录；目录上传将作为下一步扩展。
- 完成通知面板最小链路：
  - 新增 `notifications` store 与 `NotificationPanel`。
  - 右侧通知页签已替换为真实面板。
  - 搜索错误、连接错误、传输失败/取消、上传失败已统一汇聚到通知面板，而不再只依赖即时弹窗。
- 最新综合验证结果：
  - `npm run type-check` 通过
  - `npm run build-only` 通过
- 完成全局设置页真实配置闭环：
  - `App` 新增 `GetAppConfig()` 与 `SaveAppConfig()` Wails 绑定。
  - `config/` 已补充配置克隆能力，便于运行时安全读写。
  - `search.Service` 已支持更新默认并发和结果上限。
  - `transfer.Service` 已支持运行时切换临时目录。
  - 设置页已接入真实配置加载/保存，不再只是主题占位页。
  - `frontend/src/stores/settings.ts` 已负责配置拉取、保存和前端主题/语言同步。
- 本轮验证结果：
  - `wails generate module` 通过
  - `go test ./config ./search ./transfer` 通过
  - `npm run build-only` 通过
- 本轮额外验证说明：
  - `go test ./...` 在当前 sandbox 用户下仍会卡在既有的 `fileops` 本地驱动测试，报 `Access is denied`，看起来是本地目录权限环境问题，不是这一轮设置页改动引入的回归。
- 下一步：
  - 为文件面板继续补目录上传、目录下载和更完整的上下文菜单
  - 将任务面板从轮询刷新演进为事件驱动更新
  - 将通知流逐步从前端聚合演进为后端/应用层事件流

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

### 2026-04-01 Project Audit

这一轮不是继续按旧计划顺推功能，而是重新对照 `original.md`、`developing.md` 和实际代码状态做一次全项目审计，判断“现在最值得做的事”。

### Audit Conclusion

当前项目已经完成了原始诉求里的大部分主链路：

- 配置加载、日志初始化、连接仓库、连接打开/关闭已经打通。
- `Local / S3 / OSS / SFTP` 四类连接已经接入统一驱动模型。
- 左侧 Explorer、中央 `Tabs + SplitPane` 工作区、搜索、任务、通知、设置已经接入正式 `AppShell`。
- 文件面板已经支持目录浏览、刷新、视图切换、新建目录、重命名、删除。
- 搜索已经支持跨连接并发和渐进事件推送。
- 传输已经支持下载到本地临时目录、外部拖入上传、跨连接传输。
- 任务面板已经是事件驱动，不再只是旧文档里提到的轮询方案。
- 本地目录拖入上传已经支持目录递归，不只是文件。

这意味着项目当前不再处于“功能壳层未建好”的阶段，而是进入“核心能力已基本成形，但需要先修正确性与收口文档/壳层”的阶段。

### Main Gaps Found

这次审计确认的真正缺口，不是再继续堆一个新面板，而是以下三类问题：

1. 核心正确性回归
   - `go test ./...` 当前失败，且失败点集中在 `config`、`folder/local`、`fileops`。
   - `Local` 驱动把 `RootPath` 保留为 `filepath.Abs()` 结果，但路径校验时又对目标路径做了 `EvalSymlinks()`，在 macOS 的 `/var` -> `/private/var` 场景下会把合法路径误判为越界。
   - 这个问题直接导致 `Local` 驱动的 `Mkdir / Write / List / Exist / Rename / Delete / Copy / Move` 测试大面积失败，也连带打穿 `fileops` 的本地链路。
   - 由于 `Local` 本身就是产品四类目标驱动之一，这不是测试洁癖，而是当前最优先的真实功能阻塞。

2. 配置路径标准化语义不一致
   - `config.LoadAppConfig()` 当前返回的是 `filepath.Abs()` 结果，测试环境下会得到 `/var/...`。
   - 但运行期和驱动层的部分逻辑已经开始接触 symlink 解析后的真实路径 `/private/var/...`。
   - 结果是配置层和驱动层对“同一个路径”的标准化口径不一致，文档里虽然写了“路径标准化”，但代码语义还没有彻底统一。

3. 产品壳层仍残留示例工程痕迹
   - `frontend/src/router/index.ts` 仍保留 `/blank`、`/skeleton`、`/about` 等示例路由。
   - 前端生产构建仍输出 `AboutView`、`SkeletonView`、`FrontPage` 等额外 chunk，说明正式应用入口和演示页面尚未彻底收口。
   - 这不是 P0 功能阻塞，但会持续干扰后续信息架构和交付边界。

### What Is Already Outdated In This Document

本文件前面几处“Next Slice / Immediate Next Step / Current Iteration”已经有明显过时内容，这次审计后应按以下理解为准：

- “把任务面板从轮询升级为事件流”这件事，代码里已经完成。
- “外部拖入上传目前只支持文件，不支持目录”这条已经过时，当前上传计划构建逻辑已经支持目录递归。
- “下一步优先做传输事件流升级”已经不再是最优优先级，真正的第一优先级应该切到本地驱动和路径语义修复。

### Best Next Step

接下来最优先的一步，不是继续加文件预览或更多菜单，而是进入一个新的收敛阶段：

### Next Phase: Stability Baseline

目标：先把本地驱动和路径标准化语义修正到可靠状态，恢复可持续迭代的验证基线。

执行顺序建议如下：

1. 修复 `Local` 驱动路径标准化
   - 在 `folder/local` 内统一根路径和目标路径的 canonical 语义。
   - 明确是否以 `EvalSymlinks(root)` 作为比较基准。
   - 同时处理“目标不存在时校验父目录”和“根路径本身为 symlink”的情况。

2. 修复 `config` 路径断言口径
   - 统一 `LoadAppConfig()` 返回路径与测试断言的标准化方式。
   - 避免配置层返回值和驱动层比较基准不一致。

3. 恢复自动化验证基线
   - 目标是让 `go test ./...` 在本机环境下重新通过。
   - 在此基础上再继续做后续功能切片，否则后面每次传输、搜索、跨连接改动都会缺少可信回归保护。

4. 清理演示路由与示例页面残留
   - 精简 `router` 到正式应用入口。
   - 把示例页面明确下沉到开发专用入口，或直接移除不再需要的公开路由。

### Why This Is The Best Next Step

- 原始目标里 `Local` 是正式支持的驱动之一，当前它的基础路径行为不稳定，说明“产品最小可信闭环”还没完全站稳。
- 传输、搜索、跨连接这些更上层能力都建立在 `Manager` 路径语义正确之上，本地驱动不稳会持续制造假阳性功能完成感。
- 示例路由残留虽然不如驱动回归严重，但它会持续让前端架构边界发散，应该放在稳定性修复之后立即收口。

### After Stability Baseline

当上面的基线恢复后，再进入下一批功能，优先级建议是：

1. 文件面板“文件打开策略”
   - 明确双击文件后的行为：预览、下载到临时目录后打开，还是先做只读信息页。

2. Explorer / Connection 的连接质量增强
   - 新建连接页增加“测试连接”动作。
   - Explorer 增加更明确的连接状态反馈与错误恢复入口。

3. 通知与任务事件的应用层收口
   - 目前通知仍以前端聚合为主。
   - 后续可考虑把连接错误、搜索错误、传输错误进一步统一为应用层事件流。

### Verification Snapshot

本轮审计的实际验证结果如下：

- `cd frontend && npm run build-only` 通过。
- `go test ./...` 在脱离 sandbox 后可执行，但当前仍失败：
  - `config`: 默认配置路径断言与 macOS canonical path 存在差异。
  - `folder/local`: 多个基础操作被错误判定为 `invalid path`。
  - `fileops`: 受 `Local` 驱动回归影响失败。

### Execution Status

- 已完成全项目结构复查。
- 已完成原始需求与当前实现的差异分析。
- 已识别出文档中若干已经过时的下一步描述。
- 已重新判定下一阶段应以“稳定性基线修复”替代继续堆叠新功能。

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

## 2026-04-01 - File Browser Multi-Select And Keyboard Navigation

### Design Strategy

这一轮继续把文件面板往 Finder / Explorer 风格靠拢，但只补高频交互，不引入过重的表格系统：

1. 选择模型
   - 支持单选、Ctrl/Cmd 多选、Shift 区间选择。
   - 选择状态以当前排序后的路径序列为基准，避免排序后区间选择错乱。

2. 键盘模型
   - 支持方向键移动焦点。
   - 支持 Shift + 方向键扩展选择。
   - 支持 Enter 打开、Delete 删除、F2 重命名、Ctrl/Cmd + A 全选、Escape 清空选择。

3. 状态收敛
   - 刷新、切目录、排序后，会把选中项限制在当前目录可见数据集中。
   - 不保留已经不存在的路径，避免脏状态残留。

4. 视觉约束
   - 继续保持原生明细列表感，不把多选态做成厚重卡片。
   - 只增加轻量选中背景、活动焦点描边和容器焦点态。

### Completed Changes

- 已为 `frontend/src/components/Workspace/ConnectionOverviewTab.vue` 增加多选状态管理
  - 新增 `selectedPaths`、`selectionAnchorPath`、`activePath`。
  - 支持单选、切换选择、区间选择。

- 已为列表视图和网格视图接入统一选择交互
  - 点击即单选。
  - Ctrl/Cmd 点击切换选择。
  - Shift 点击做区间选择。
  - 右键前先收口为当前项，保证上下文菜单作用对象明确。

- 已增加批量操作入口
  - 选中多项后可直接批量下载。
  - 选中多项后可直接批量删除。
  - 可一键清空当前选择。

- 已增加键盘导航
  - 方向键移动活动项。
  - Shift + 方向键扩展选择范围。
  - Enter 打开当前项。
  - Delete 删除当前选择。
  - F2 重命名当前活动项。
  - Ctrl/Cmd + A 全选。
  - Escape 清空选择。

- 已增加选择状态收敛逻辑
  - 排序或刷新后自动过滤失效路径。
  - 当前活动项或锚点失效时自动回收。

- 已更新 locale
  - 补充“下载所选”“删除所选”“清除选择”“已选择 N 项”等文案。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

### Current Product State

到这一轮为止，文件面板已经具备更接近桌面文件管理器的基础交互闭环：

- 原生明细列表风格。
- 可勾选字段。
- 可按列排序。
- 支持多选。
- 支持键盘导航。
- 支持批量下载和批量删除。

### Updated Next Step

下一步最值得继续做的是：

1. 列宽拖动
   - 让明细视图进一步接近系统文件管理器。

2. 多选批量能力补齐
   - 批量重命名暂时不做，但可补批量复制路径、批量另存为限制提示等。

3. 更完整的键盘体验
   - Home/End、空格预选、Backspace 返回上级目录。

## 2026-04-01 - File Browser Resizable Columns

### Design Strategy

这一轮聚焦明细列表的列宽拖动，目标是让列表继续向系统文件管理器靠近，但不引入重量级表格组件：

1. 列宽模型
   - 每一列都有默认宽度和最小宽度。
   - 名称列保持弹性列，但其最小宽度可被拖动调整。
   - 其余列使用固定像素宽度，更接近 Finder / Explorer 的明细视图。

2. 交互模型
   - 在表头右侧提供拖动手柄。
   - 拖动时锁定为横向调整，避免误触发排序。
   - 拖动结束后立即持久化到本地存储。

3. 状态持久化
   - 列宽偏好独立存储，不与排序和字段显示状态混在一起。
   - 切换连接、切换目录、重开应用后继续沿用上次列宽。

4. 实现约束
   - 不替换现有原生明细列表结构。
   - 不引入新的第三方表格依赖。
   - 保持当前滚动容器和多选逻辑不受影响。

### Completed Changes

- 已为 `frontend/src/components/Workspace/ConnectionOverviewTab.vue` 增加列宽状态
  - 新增默认宽度、最小宽度和本地存储键。
  - 列模板宽度由静态字符串改为动态计算。

- 已为明细列表表头增加拖动手柄
  - 每个可见列头右侧都可拖动调宽。
  - 拖动时不会误触发排序点击。

- 已增加列宽拖动逻辑
  - 通过 `mousemove` / `mouseup` 实现拖拽生命周期。
  - 拖动期间锁定全局光标和文本选择，避免交互抖动。

- 已增加列宽持久化
  - 列宽会写入本地存储。
  - 重新进入应用后会恢复上次设置。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

### Current Product State

到这一轮为止，文件面板明细视图已经具备：

- 原生列表风格。
- 字段显示开关。
- 列排序。
- 多选与键盘导航。
- 列宽拖动。
- 列宽偏好持久化。

### Updated Next Step

下一步继续补原生体验时，优先级建议如下：

1. Home / End / Backspace 键盘补齐
   - 让键盘导航更接近桌面文件管理器。

2. 批量动作继续补齐
   - 比如批量复制路径、目录批量下载反馈聚合。

3. 选择框架继续强化
   - 视图切换时保留活动项滚动可见，减少跳变感。

## 2026-04-01 - File Browser Keyboard Completion

### Design Strategy

这一轮继续补文件面板的键盘能力，但只补最接近日常桌面文件管理器习惯的几个动作：

1. 边界跳转
   - `Home` 跳到当前列表第一项。
   - `End` 跳到当前列表最后一项。
   - 与 `Shift` 组合时继续扩展当前选择范围。

2. 层级返回
   - `Backspace` 返回上级目录。
   - 根目录下不触发无意义跳转。

3. 焦点可见性
   - 活动项变化后自动滚动到可见区域。
   - 视图切换后继续尝试把当前活动项滚动进视口，减少跳变感。

### Completed Changes

- 已为 `frontend/src/components/Workspace/ConnectionOverviewTab.vue` 增加边界键盘导航
  - `Home` 跳到第一项。
  - `End` 跳到最后一项。
  - `Shift + Home/End` 扩展区间选择。

- 已增加层级返回快捷键
  - `Backspace` 返回当前目录的父级路径。
  - 根目录下不会继续向上跳转。

- 已增加活动项滚动可见逻辑
  - 当前活动项变化时自动滚动到视口内。
  - 列表/网格视图切换时，会尝试保持活动项可见。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

### Current Product State

到这一轮为止，文件面板键盘体验已经包含：

- 方向键移动。
- `Shift` 区间扩展。
- `Ctrl/Cmd + A` 全选。
- `Enter` 打开。
- `Delete` 删除。
- `F2` 重命名。
- `Home` / `End` 边界跳转。
- `Backspace` 返回上级目录。

### Updated Next Step

下一步如果继续做交互增强，建议优先：

1. 批量复制路径
   - 和已有多选模型完全兼容，补齐高频批量动作。

2. 下载反馈聚合
   - 多选或目录下载时，把结果反馈从“任务数量”提升为更易懂的提示。

3. 更完整的 Explorer/Finder 细节
   - 比如空格预选、右键不打断已有多选、恢复最后活动项等。

### Incremental Update - Batch Copy Paths

- 文件面板工具栏已增加“复制所选路径”。
- 多选后会把当前选中项路径按换行拼接写入剪贴板。
- 这一动作复用现有多选状态，不额外引入新的选择模型。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-01 - Drag To Directory Targets

### Design Strategy

这一轮补的是“文件/目录拖到目录内”的真实文件管理器交互，目标覆盖三类入口：

1. 文件面板内部目录目标
   - 列表视图目录行可接收 drop。
   - 图标视图目录块可接收 drop。
   - 当前目录根区域可接收 drop。

2. 中央工作区到 Explorer 目标
   - 可从中央工作区把文件/目录拖到连接节点根目录。
   - 也可拖到目录树节点。

3. 同连接 / 跨连接语义分流
   - 同连接：按文件管理器直觉走 `Move`。
   - 跨连接：继续走现有 transfer 链路。
   - 不允许把目录拖进自己或自己的子目录。

### Completed Changes

- 后端已新增同连接移动入口
  - 新增 `MoveConnectionEntry()` Wails 绑定。
  - `fileops.Service` 新增 `MoveEntry()`。
  - 同连接拖放不再走“下载再上传”，而是直接调用后端 `Move`。

- 后端已补边界校验
  - 目录拖进自身或子目录会被拒绝。
  - 同父目录 no-op 会被前端和执行层共同跳过。

- 已新增共享拖放执行层
  - 新增 `frontend/src/composables/connectionEntryDrop.ts`。
  - 统一负责：
    - 拖拽 payload 构造
    - drop 可行性判断
    - 同连接 move
    - 跨连接 transfer
    - refresh 事件发射

- 文件面板已接入目录级 drop
  - 列表视图目录行可作为 drop target。
  - 图标视图目录块可作为 drop target。
  - 当前目录根区域可作为 drop target。
  - 支持单项和多项拖拽。

- Explorer 树已接入 drop
  - 连接节点代表根目录，可直接接收拖放。
  - 目录节点可接收拖放到该目录下。
  - 节点有独立高亮反馈。

- 中央工作区 panel drop 已统一收口
  - `AppShellView` 的 panel-level drop 改为复用共享拖放执行层。
  - 同连接和跨连接不再各写一套分支。

- 已补前端提示文案
  - 新增拖放移动成功、拖放传输排队、拖放准备完成等提示。

### Verification

- `go test ./...` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

### Current Product State

到这一轮为止，项目已经支持：

- 在列表视图把文件/目录拖到另一个目录。
- 在图标视图把文件/目录拖到另一个目录。
- 在中央工作区把文件/目录拖到目标连接根目录。
- 在中央工作区把文件/目录拖到 Explorer 目录树节点。
- 同连接 move。
- 跨连接 transfer。

### Updated Next Step

如果继续收口这块交互，下一步最值得做的是：

1. 右键与多选拖放的最终语义
   - 是否保留已有多选，还是右键前收口为单项，需要统一规则。

2. 拖放后的目录自动展开
   - 在 Explorer 树悬停目录一段时间后自动展开，更接近桌面文件管理器。

3. 拖放进度反馈
   - 对目录级跨连接拖放增加更明确的聚合提示，而不只是任务数量。

### Incremental Update - Drag Consistency And Busy State

- 已为连接条目拖放新增生命周期事件。
- 文件面板现在会在受影响目录进入拖放执行期时显示遮罩，并临时禁止操作。
- 对同连接 move，源目录中的被拖动项会先做临时隐藏，finish 后再强制刷新，避免文件拖走后仍残留在中央工作区。
- `go test ./...` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

### Incremental Update - Source List Match Strategy

- 中央工作区源列表是否受拖放影响，已从“按源目录路径判断”改为“按当前列表里是否实际展示了这些 source path 判断”。
- 这样从中央工作区拖到文件树时，只要当前列表中确实存在被拖走的文件/目录，就会立即进入隐藏和 finish 后刷新流程。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

### Incremental Update - Source View Dir And Direct Removal

- 拖放 payload 现在明确携带 `sourceViewDir`，表示拖拽发起时中央工作区所在目录。
- 源列表受影响判断改为直接匹配 `sourceConnectionId + sourceViewDir`。
- 同连接 move 成功 finish 后，会先从当前 `items` 中直接移除对应 source path，再执行 reload 校准。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

### Incremental Update - Direct Removal On Drag End

- 新增最近一次成功拖放结果的消费接口。
- 中央工作区文件面板在 `dragend` 时会直接消费成功的同连接 move 结果。
- 命中后会立即从当前 `items` 删除 source item，再执行 reload 校准。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

### Incremental Update - Normalized Path Matching

- 中央工作区文件列表对拖放源路径的隐藏、删除、选中匹配，已统一改为规范化路径比较。
- 现在会先把路径规整成统一的 remote path 形式，再进行命中判断，避免因斜杠或前缀格式差异导致删不掉。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

### Incremental Update - Relaxed Source Panel Match

- 源工作区文件列表对同连接 move 的收口，不再要求当前目录先精确匹配后才执行。
- 现在只要命中同一个连接，就先尝试移除 source paths；目标目录刷新仍按当前目录匹配。
- 这样源列表收口不会再被目录比较条件卡住。
- cd frontend && npm run type-check 通过。
- cd frontend && npm run build-only 通过。

### Incremental Update - Preserve Source Scroll Position

- 同连接拖放成功后，源工作区文件面板不再主动 `reload()`。
- 现在源面板只做本地删项和隐藏状态回收，目标目录刷新仍然保留。
- 这样长列表不会因为源面板刷新而回到顶部。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

### Incremental Update - Clamp Scroll After Local Removal

- 源工作区文件面板在本地删项后，新增滚动位置钳位。
- 当内容高度变短且原本正停在底部时，会把 `scrollTop` 自动压回新的最大可滚动位置，避免底部出现空白。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-02 - Explorer Hover Expand During Drag

### Design Strategy

这一轮补的是目录树在拖放过程中的悬停自动展开，目标是让深层目录拖放不再要求用户先手动逐级展开：

1. 触发条件
   - 仅在连接条目拖拽命中可投放节点时触发。
   - 只对当前未展开、未加载中的节点生效。

2. 交互策略
   - 悬停一小段时间后再自动展开，避免掠过时误展开。
   - 离开节点或完成 drop 时立即取消定时器。

3. 架构分层
   - 节点组件只负责悬停计时和事件上抛。
   - Explorer 面板负责真正加载子节点并展开，避免递归节点自己管理异步加载。

### Completed Changes

- Explorer 树节点已增加拖放悬停展开定时器。
- 节点离开或 drop 时会清理悬停展开定时器。
- Explorer 面板已接入 hover-expand 事件，并在需要时自动加载子目录后展开。
- 节点已增加轻量悬停展开反馈样式。

### Verification

- cd frontend && npm run type-check 通过。
- cd frontend && npm run build-only 通过。

### Incremental Update - Remove In-Panel Move Reload

- 中央工作区内部把文件拖到目录时，已移除 `executeDrop()` 中对 move 成功后的主动 `reload()`。
- 现在这一场景只保留本地删项、忙碌态回收和滚动钳位，不再整页刷新回顶。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-02 - Drag Feedback Aggregation

### Design Strategy

这一轮不改拖放行为，只优化成功反馈，让提示信息更接近日常文件管理器：

1. 聚合口径
   - 不再只报“创建了多少任务”。
   - 改为优先报“处理了多少项，到哪里”，必要时附带任务数和跳过数。

2. 统一出口
   - 中央工作区内部拖放。
   - 中央工作区拖到 Explorer 树。
   - panel 级跨组拖放。
   统一复用同一套反馈构建逻辑。

3. 目标标签
   - 反馈优先展示实际目标目录名称。
   - 根目录场景回退到连接名称。

### Completed Changes

- 已为连接条目拖放结果增加聚合统计字段。
- 已新增共享拖放反馈 helper，统一生成 success message key 和参数。
- 已统一三个拖放入口的 success toast。
- 已补中英文 locale：
  - 多项移动摘要
  - 多项传输摘要
  - 目录/根目录目标标签
  - skipped 项反馈

### Verification

- cd frontend && npm run type-check 通过。
- cd frontend && npm run build-only 通过。

## 2026-04-02 - Multi-Select Drag Edge Cases

### Design Strategy

这一轮继续收口多选拖放里最容易导致异常的两类边界：

1. 嵌套选择去重
   - 如果已经选中了父目录，再同时选中它的子目录或子文件，拖放时只保留父目录。
   - 避免同一批里重复 move / 重复 transfer。

2. 部分无效项跳过
   - 多选中若只有个别项的目标非法，不再一票否决整批。
   - 改为只跳过冲突项，其余有效项继续执行。

### Completed Changes

- 已为连接条目拖放新增嵌套选择压缩逻辑。
- 已让 `buildConnectionEntryDragPayload()` 在构造 payload 时自动去掉被父目录覆盖的子项。
- 已让拖放计划阶段在同连接场景下对“目录拖进自己子目录”改为逐项跳过，而不是整批失败。
- 已让跨连接拖放同样复用嵌套选择压缩逻辑，避免目录和其子项重复传输。
- 已新增前端单测覆盖：
  - 父目录 + 子项去重
  - 同连接多选里部分非法项跳过
  - 跨连接多选时嵌套项压缩

### Verification

- `cd frontend && npm run test:unit -- connectionEntryDrop.spec.ts` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-02 - Add 2026 Dark Theme

### Design Strategy

这一轮只做主题系统适配，不扩展编辑器语法高亮能力。

1. 适配范围
   - 当前项目主题系统只支持全局 UI token（`--theme-*`）。
   - 因此不直接导入 VS Code 的完整 theme JSON，也不处理 `tokenColors`。
   - 只抽取其中对桌面文件管理器界面真正有意义的颜色层级：背景、文本、边框、强调色、语义色、滚动条、阴影。

2. 映射策略
   - `editor.background` / `sideBar.background` / `panel.background` 作为主背景分层。
   - `focusBorder` / `button.background` / `list.focusOutline` 汇总为主强调色。
   - `foreground` / `descriptionForeground` / `disabledForeground` 形成文本层级。
   - `gitDecoration.*` / `errorForeground` 等补足 success / warning / danger 语义色。

3. 实现方式
   - 新增独立主题文件 `frontend/src/assets/themes/2026-dark.css`。
   - 在 `frontend/src/assets/theme.css` 导入。
   - 在 `frontend/src/stores/theme.ts` 注册为内置深色主题，设置页自动可选，无需额外 UI 代码。

### Completed Changes

- 已新增 `2026 Dark` 内置主题。
- 已新增 `frontend/src/assets/themes/2026-dark.css`，按现有主题变量面映射 VS Code 2026 Dark 色板。
- 已在主题入口中导入该主题。
- 已在 `BUILTIN_THEMES` 中注册该主题，设置页会自动展示。
- 已明确文档约束：当前只是 UI 主题适配，不是 VS Code JSON 的完整导入器。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-02 - 2026 Dark Component Polish

### Design Strategy

这一轮继续细调 `2026 Dark`，但不再扩展主题变量面，而是做组件级覆写。

原因很直接：

1. `2026 Dark` 的质感不只来自主色板，还来自具体控件层级
   - Tabs 的活动页签是深底 + 细蓝线，不是单纯主色文字。
   - 文件列表的 hover / selected / active 是三层不同灰度和强调线。
   - Explorer 的树节点和徽标也需要更低对比、更接近 VS Code 的表面色。

2. 这些差异已经超出当前 `--theme-*` 通用变量面
   - 如果继续往全局变量里塞，会把本来简单的主题系统拖成半个设计 token 平台。
   - 对当前项目不划算。

3. 因此策略改为
   - 基础色板仍走 `--theme-*`。
   - `2026-dark.css` 内对 Explorer、Tabs、文件列表追加少量主题专属覆写。
   - 只影响 `2026-dark`，不污染其他主题。

### Completed Changes

- 已细调 Explorer：
  - panel 背景层次更贴近 VS Code 2026 Dark。
  - 树节点 hover 改为更轻的白色蒙层。
  - driver pill 和菜单 hover 改为更原生的低对比灰。

- 已细调 Tabs：
  - tab bar 背景和分隔线改为更接近 `titleBar` / `tabsBackground` 组合。
  - 活动 tab 改为深底 + 青蓝下边线。
  - inactive / hover / close hover 层级已收紧。

- 已细调文件浏览器：
  - 列表壳层、表头、行分隔线改为更接近 Finder / VS Code 深色文件面板的低对比层级。
  - 列表和图标视图的 `selected` 改为更接近 `list.inactiveSelectionBackground`。
  - `active + selected` 改为青蓝焦点态，而不是继续只依赖通用主色混合。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Adjustable Explorer and File List Font Sizes

### Design Strategy

这一轮处理两个目标：

1. 先把目录树默认字体下调一档
   - 原来 Explorer 树节点主要跟随全局 `body` 字号，看起来偏大。
   - 现在将目录树默认字号明确设为 `13px`。

2. 再把目录树和中央工作区列表视图字号做成正式设置项
   - 不做前端临时本地状态。
   - 直接进入主配置 `ui` 区段，和 `locale/theme` 一样走保存、加载、热应用链路。

3. 运行时应用方式
   - 设置保存后，前端通过 CSS 变量统一下发：
     - `--ui-explorer-font-size`
     - `--ui-file-list-font-size`
   - Explorer 树节点和文件列表视图只消费变量，不再把字号散落在组件里。

4. 边界控制
   - 两个字号都限制在 `11px ~ 18px`。
   - 默认值都为 `13px`。
   - 次级文字和表头字号带下限，避免调小后不可读。

### Completed Changes

- 已为 `config.UISection` 新增：
  - `explorerFontSize`
  - `fileListFontSize`
- 已在后端配置默认值中加入两个字号，默认都为 `13`。
- 已在配置归一化阶段增加字号边界钳位，范围 `11 ~ 18`。
- 已在设置页增加两个可调输入项：
  - 目录树字体大小
  - 文件列表字体大小
- 已在设置保存/加载时实时把字号写入根节点 CSS 变量。
- 已让 Explorer 树节点字号、状态字和 loading 文案都消费 `--ui-explorer-font-size`。
- 已让中央工作区列表视图表头、行高、正文和图标都消费 `--ui-file-list-font-size`。
- 已新增配置测试覆盖：
  - 缺省配置默认值
  - 超出范围时的钳位行为

### Verification

- `wails generate module` 通过。
- `go test ./config ./...` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Font Size Defaults and Explorer Tree Drop Refresh Fix

### Design Strategy

这一轮同时处理三个相关体验问题：

1. 字号设置补完整
   - 上一轮已经把目录树和列表视图字号做成配置项。
   - 这一轮继续补“恢复默认”，避免用户调过之后只能手动改回默认值。

2. 图标视图接入同一套字号设置
   - `fileListFontSize` 不应该只作用于列表视图正文。
   - 同一连接面板在列表/图标视图之间切换时，字号体验应保持一致。
   - 因此图标视图的文件名、辅助信息和图标也纳入同一套字号变量。

3. 目录树拖放后的无意义刷新
   - 目录树本身只展示目录，不展示文件。
   - 所以“把文件从中央工作区拖到目录树”时，目录树通常不需要刷新。
   - 之前同连接 move / 跨连接 transfer 无论拖的是文件还是目录，都会发目录刷新事件，导致 Explorer 树重新加载并可能回顶。
   - 这一轮改为：只有目录结构真的变化时才刷新树；单纯文件拖放不再触发树刷新。

4. 真需要刷新时保住滚动位置
   - 目录拖放仍然需要局部刷新 Explorer 树。
   - 对这类刷新增加滚动位置保留，避免长目录树回到顶部。

### Completed Changes

- 已为设置页两个字号项增加“恢复默认”按钮，默认回到 `13px`。
- 已让图标视图的文件名、辅助信息和图标尺寸复用 `--ui-file-list-font-size`。
- 已收紧连接条目拖放刷新策略：
  - 文件 move / transfer 不再触发 Explorer 树目录刷新。
  - 只有目录 move / transfer 才会刷新对应树节点。
- 已去掉拖放后一些不必要的父目录树刷新，减少无效重渲染。
- 已为 Explorer 树局部刷新增加滚动位置恢复，避免长树回顶。
- 已同步更新前端单测断言，覆盖新增的 `isDirectory` 拖放计划字段。

### Verification

- `cd frontend && npm run test:unit -- connectionEntryDrop.spec.ts` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Fix Connection Form State Bleed Across Tabs

### Design Strategy

这个问题的根因不在连接表单字段绑定本身，而在 Tabs 渲染层：

1. `ConnectionFormTab` 是同一个 Vue 组件类型
   - 新建连接和编辑连接都复用它。
   - 但它们属于不同 tab，应该拥有各自独立的实例状态。

2. 之前 `TabGroup` 里用动态 `<component :is="...">` 渲染活动 tab
   - 没有按 `tab.id` 提供 key。
   - 当两个 tab 恰好都是 `ConnectionFormTab` 时，Vue 会原地复用组件实例，只更新 props。
   - 结果就是前一个 tab 里的响应式 `form` 状态被后一个 tab 继承，出现“编辑 Local 却看到 S3 表单”的串状态问题。

3. 正确修法应该放在 Tabs 层
   - 不是在 `ConnectionFormTab` 内再堆更多 reset/watch 补丁。
   - 而是让每个 tab 都按 `tab.id` 拥有自己的组件实例。
   - 同时使用 `KeepAlive` 保留各 tab 的未保存状态，避免切换 tab 时丢表单。

### Completed Changes

- 已在 `frontend/src/components/Tabs/TabGroup.vue` 中为动态活动组件增加 `:key="activeTab.id"`。
- 已用 `KeepAlive` 包裹活动 tab 组件。
- 现在相同组件类型但不同 tab 的内容不会再共用同一个实例状态。
- 新建连接页和编辑连接页会各自保留自己的表单状态，不再互相污染。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Normalize Root and RootPath in Connection Form

### Design Strategy

这一轮修的是连接表单里一个语义不一致问题：

1. 当前表单同时存在两套“根路径”来源
   - 通用字段 `root`
   - Local / SFTP 驱动配置字段 `config.rootPath`

2. 后端驱动层会把 `DriverOptions.Root` 和 `config.rootPath` 合并使用
   - 所以即使只填了通用 `root`，连接也可能正常工作。
   - 但编辑页如果只回显 `config.rootPath`，就会表现成“RootPath 没有回显”。

3. 正确处理方式不是只修显示
   - 而是让 Local / SFTP 这类有 `rootPath` 语义的驱动，在表单层把 `root` 和 `config.rootPath` 统一。
   - 这样新建、编辑、切换驱动和保存时都保持一致。

### Completed Changes

- 已新增驱动级根路径字段判断：当前对 `Local` 和 `SFTP` 识别 `rootPath` 配置项。
- 已在编辑态 `applyDefinition()` 时统一 `root` 和 `config.rootPath`：
  - 任意一边有值，都会同步回表单。
- 已在提交 `buildDefinition()` 时统一写回：
  - `root`
  - `config.rootPath`
- 已在切换驱动时同步重建后的 root 语义，避免字段切换后再次出现一边有值另一边为空。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Stop Exposing Logical Root for Current Drivers

### Design Strategy

这一轮继续收口连接表单里的“逻辑根路径”语义。

1. 对当前项目已支持的驱动来说，通用 `root` 基本没有独立价值
   - `Local` / `SFTP` 已有自己的 `rootPath`
   - `S3` / `OSS` 已有自己的 `prefix`

2. 如果在表单里继续同时暴露通用 `root`
   - 会出现两个路径字段表达同一件事。
   - 用户很难判断应该填哪一个。
   - 编辑时也容易出现“一边有值、一边空白”的错觉。

3. 因此当前策略改为
   - 对已有驱动，表单只展示驱动自己的路径参数。
   - 通用 `root` 不再作为主编辑入口展示。
   - 旧连接如果历史上把值写在通用 `root`，编辑时会自动迁移回显到对应驱动参数中。

### Completed Changes

- 已让连接表单对以下驱动只展示驱动自己的路径字段：
  - `Local` / `SFTP` → `rootPath`
  - `S3` / `OSS` → `prefix`
- 已隐藏这些驱动下的通用 `Root` 输入框。
- 已把旧数据里的 `root` 作为兼容迁移来源：
  - 编辑时会自动回填到对应驱动参数。
- 已在保存时对这批驱动统一写回驱动参数，并将通用 `root` 清空。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Switch Connections Config Loading Away from Viper

### Design Strategy

这次只改连接配置读取链路，不动应用主配置：

1. 现状确认
   - `LoadConnectionsConfig()` 之前确实走的是 `viper` + `Unmarshal`。
   - `SaveConnectionsConfig()` 写 YAML 时其实已经是 `yaml.v3`。
   - 也就是说，连接配置此前是“写用 yaml.v3，读用 viper”的不一致状态。

2. 改动目标
   - `connections.yaml` / `connections.yml` 读取改为直接使用 `yaml.v3`。
   - `connections.json` 读取改为标准库 `encoding/json`。
   - 不再让连接配置读取依赖 `viper`。

3. 范围控制
   - 这一轮不改 `LoadAppConfig()`。
   - `AppConfig` 仍保持现状，避免把两个配置链路混在同一轮修改里。

### Completed Changes

- 已把 `LoadConnectionsConfig()` 从 `viper` 读取改为直接读取文件内容。
- 已新增 `decodeConnectionsConfigFile()`：
  - `.yaml/.yml` → `yaml.v3`
  - `.json` → `encoding/json`
- 已保留原有 `ConnectionsConfig.Normalize()` 流程，确保读取后的字段仍会归一化。
- 已新增 YAML 读取测试，覆盖 `connections.yaml` 的基本解析行为。

### Verification

- `go test ./config ./...` 通过。

## 2026-04-03 - Stabilize Explorer Connection Delete Menu

### Design Strategy

这次修的是 Explorer 连接节点右侧操作菜单里“删除连接”不稳定的问题。

1. 现象判断
   - 连接节点操作菜单此前基于 `el-dropdown`。
   - 实际表现是菜单样式和弹出行为都不稳定，删除动作有时无法正常弹出确认链路。

2. 根因判断
   - 删除逻辑本身在 `ExplorerPanel` 里，并没有明显错误。
   - 真正不稳的是 `ExplorerTreeNode` 里的菜单触发层。
   - 对这种窄行树节点来说，继续依赖 `el-dropdown` 的命令分发不够稳，尤其在节点点击、菜单触发和弹层之间更容易出现交互异常。

3. 修复策略
   - 把连接节点菜单从 `el-dropdown` 改成更直接的 `el-popover`。
   - 菜单项改为显式按钮点击，不再依赖 dropdown command。
   - 给触发按钮和菜单按钮都补 `type="button"`，并显式阻断点击冒泡。
   - 删除确认仍复用原有 `ExplorerPanel` 的 `ElMessageBox.confirm` 逻辑，不改业务语义。

### Completed Changes

- 已将 Explorer 连接节点菜单从 `el-dropdown` 改为 `el-popover`。
- 已把菜单内容改为显式按钮列表：
  - 打开
  - 编辑
  - 关闭
  - 删除
- 已在菜单组件内部新增 `menuVisible` 状态，点击动作后主动收起菜单。
- 已给节点上的交互按钮补充显式 `type="button"`，避免默认按钮行为干扰。
- 已新增本地菜单样式，确保删除项的危险态和整体排版稳定。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Move Explorer Delete Confirmation Into Inline Drawer

### Design Strategy

这次不是继续推翻 Explorer 节点右侧三点菜单，而是只替换“删除确认”这一段交互。

1. 现象重新确认
   - 原来的三点按钮以及 dropdown / popover 本身是能正常工作的。
   - 真正不稳定的是删除后的确认弹框。
   - 这更像是 WebView 环境下弹层确认组件的兼容问题，而不是菜单入口本身的问题。

2. 修复策略
   - 恢复 Explorer 连接节点原来的 dropdown 菜单入口。
   - 删除菜单项不再走 `ElMessageBox.confirm`。
   - 改为在 Explorer 面板内部打开一个右侧内嵌抽屉，专门承接连接操作和删除确认。
   - 这样保留原有菜单交互习惯，同时彻底绕开不稳定的确认弹框。

3. 设计取舍
   - `open` / `edit` / `close` 仍保留为菜单中的快捷动作。
   - `delete` 进入抽屉后再进行二次确认。
   - 抽屉是 Explorer 内部布局的一部分，不依赖额外 popup 或 teleport。

### Completed Changes

- 已恢复 Explorer 连接节点右侧三点菜单为 `el-dropdown`。
- 已移除 `ExplorerPanel` 中对 `ElMessageBox.confirm` 的依赖。
- 已新增 Explorer 右侧内嵌操作抽屉，展示：
  - 当前连接名称
  - 驱动类型
  - 连接状态
  - 打开 / 编辑 / 关闭 操作
  - 删除区域
- 已将删除改为抽屉内二次确认按钮，不再依赖弹框。
- 已补充相关中英文文案。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Refine Explorer Delete Interaction Into Inline Row Actions

### Design Strategy

上一轮把删除确认从弹框改成了 Explorer 内部抽屉，但实际交互还是偏重。

1. 新问题
   - 抽屉会额外占用 Explorer 右侧空间。
   - 删除态里仍然混入了编辑、关闭等其他动作。
   - 这和“只做删除确认”的目标不一致。

2. 最终收口
   - 保留原有三点 dropdown。
   - 点击 `删除` 后，不再打开抽屉。
   - 改为在当前连接节点右侧直接切换成两个小按钮：
     - `确认删除`
     - `取消删除`
   - 同时隐藏该行原本的：
     - 三点菜单
     - 已连接 / 已保存 状态

3. 这样做的原因
   - 删除确认就地完成，不改变 Explorer 总体布局。
   - 视觉聚焦在当前连接行，不会让操作范围扩散。
   - 继续规避 WebView 下不稳定的确认弹框。

### Completed Changes

- 已移除 Explorer 内嵌抽屉式删除确认方案。
- 已将连接删除改为节点行内确认模式。
- 已在 `ExplorerTreeNode` 中增加删除待确认状态：
  - 待确认时显示 `确认删除` / `取消删除`
  - 同时隐藏状态标签和三点菜单
- 已在 `ExplorerPanel` 中收口删除状态管理：
  - `pendingDeleteConnectionId`
  - `deleteBusyConnectionId`
- 已保留 dropdown 菜单作为删除入口，不改用户已有操作习惯。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Close Connection Tabs After Connection Deletion

### Design Strategy

删除连接后，Explorer 树里的节点虽然会消失，但如果对应的工作区 tab 还留着，UI 就会进入失效状态。

1. 需要同步关闭的内容
   - 文件浏览 tab：`connection:{id}`
   - 连接编辑 tab：`connection-form:{id}`

2. 正确做法
   - 不把这类清理逻辑散落在 Explorer 组件里。
   - 统一收口到 `workspace store`，由它负责删除连接后对应 tab 和路径状态的清理。

3. 额外收口
   - 同时清除该连接的 `connectionTabState` 路径缓存。
   - 如果删完后工作区没有任何 tab，则恢复 welcome tab。

### Completed Changes

- 已在 `workspace store` 新增 `closeConnectionTabs(connectionId)`。
- 已统一关闭以下 tab：
  - `connection:{id}`
  - `connection-form:{id}`
- 已在删除连接成功后调用该方法。
- 已在清理 tab 的同时移除该连接的路径状态缓存。
- 已保证删完后若工作区为空，会自动恢复 welcome tab。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Close New Connection Form Tab After Save

### Design Strategy

新建连接保存成功后，如果 `New Connection` 表单 tab 还留在工作区，会让流程显得没有收口。

1. 当前问题
   - 保存成功后已经会打开对应连接的文件浏览 tab。
   - 但 `connection-form:new` 仍然保留在工作区里。
   - 这会造成一个已经完成使命的表单页残留在界面上。

2. 修复策略
   - 在 `workspace store` 增加通用 `closeTabById(tabId)`。
   - 新建连接保存成功后：
     - 先打开对应连接浏览页
     - 再关闭 `connection-form:new`

3. 范围控制
   - 这一轮只修“新建连接保存后关闭新建表单 tab”。
   - 不强行改变“编辑连接保存后是否关闭编辑 tab”的语义。

### Completed Changes

- 已在 `workspace store` 新增 `closeTabById(tabId)`。
- 已在 `ConnectionFormTab` 的创建模式保存成功后关闭 `connection-form:new`。
- 已保留原有“保存后打开连接浏览页”的行为。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Prevent New Connection Form From Reusing Saved Connection Identity

### Design Strategy

这次问题不是“加号按钮路由错了”，而是新建连接表单的实例状态在 `KeepAlive` 下被复用了。

1. 现象
   - 新建连接保存成功后，再点击加号打开“新建连接”。
   - 表单会继续带着上一次保存出来的连接身份。
   - 继续保存时会落到“更新原连接”，而不是创建新连接。

2. 根因
   - `create` 模式保存成功后，表单把 `form.id` 写成了已保存连接的真实 id。
   - 后面即使关闭了 `connection-form:new` tab，`KeepAlive` 仍可能复用这份表单实例。
   - 下一次重新打开时，如果实例没先清空，就会把旧 id 一并带回来。

3. 修复策略
   - `create` 模式保存成功后，不再把 `form.id` 写成已保存连接 id。
   - 在关闭 `connection-form:new` 之前，先把表单重置回空白新建状态。
   - `edit` 模式仍保留写回 `form.id` 的行为。

### Completed Changes

- 已将 `ConnectionFormTab.submit()` 调整为：
  - `edit` 模式才写回 `form.id`
  - `create` 模式不会保留已保存连接 id
- 已在 `create` 模式保存成功后，关闭 tab 前先 `applyDefinition(null)` 重置表单。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Add Notification Indicator to Sidebar Button

### Design Strategy

这次只做一个轻量提示：当通知栏里存在事件时，在右侧通知按钮上显示一个小标记点。

1. 需求边界
   - 不显示数字，只显示“有通知”的状态。
   - 不改通知面板数据结构。
   - 颜色使用主题主色调，而不是固定红点。

2. 实现策略
   - 在侧边栏按钮配置 `SidebarButton` 上增加可选 `indicator` 字段。
   - 通知按钮根据 `notifications.unreadCount > 0` 动态决定是否显示标记。
   - 标记点直接作为按钮内部元素渲染，不额外引入复杂徽章依赖。

3. 好处
   - 改动范围小，只触及侧边栏按钮层。
   - 后续如果要升级成数字徽章，也可以基于同一入口继续扩展。

### Completed Changes

- 已在 `SidebarButton` 类型中新增 `indicator?: boolean`。
- 已在 `SidebarButtonColumn` 中为按钮增加主色小圆点标记渲染。
- 已将右侧通知按钮改为响应式配置：
  - `notifications.unreadCount > 0` 时显示标记
  - 否则隐藏

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Avoid Reloading File Lists After Tab Split Remounts

### Design Strategy

用户反馈分割工作区后，连接文件面板会重新加载列表，例如把 B 拖到右侧分割后，A 和 B 都会重新刷新。

1. 现象判断
   - 这不是有意做的“分割后强制刷新”。
   - 根因是当前 Tabs 分割实现会重建对应的 `TabGroup`。
   - 对连接文件面板来说，这会触发组件重新挂载，而 `ConnectionOverviewTab` 在 `onMounted()` 时原本会无条件 `load()`。

2. 设计取舍
   - 不直接改 Tabs 分割结构，因为那会牵涉更大的组件树和缓存语义。
   - 先做更稳的方案：把连接文件面板的浏览状态缓存到 `workspace store`。
   - 这样即使 split 后组件重新挂载，也可以优先复用已有列表快照，而不是重新请求目录。

3. 缓存内容
   - 当前路径
   - 当前目录项列表
   - 仍保留原有的显式刷新、目录切换、目录变更事件刷新能力

### Completed Changes

- 已在 `workspace store` 的 `connectionTabState` 中增加文件列表快照缓存。
- 已新增：
  - `setConnectionBrowserState(connectionId, path, items)`
  - `getConnectionBrowserState(connectionId)`
- 已让 `ConnectionOverviewTab.load()` 在成功加载后写入缓存。
- 已让 `ConnectionOverviewTab` 在本地删项时同步更新缓存。
- 已让 `ConnectionOverviewTab` 挂载时优先读取缓存：
  - 若缓存路径与目标路径一致，则直接复用缓存
  - 否则才真正请求后端目录

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Fix Last Tab Sometimes Not Closing

### Design Strategy

这次修的是“最后一个 tab 有时点关闭却关不掉”的问题。

1. 根因
   - `useTabTree.removeTab()` 在移除某个 group 的最后一个 tab 时，会尝试把该 group 从树里替换掉。
   - 如果这个 group 同时又是整棵 tab tree 的最后一个节点，`replaceNode()` 会返回 `null`。
   - 旧逻辑只在 `result` 存在时才写回 `tree.value`。
   - 结果就是最后一个 tab 被删掉后，树状态没有真正更新，表现成“点击关闭无效”。

2. 修复策略
   - 当 `replaceNode()` 返回 `null` 时，不再忽略。
   - 直接把树落成一个空的 `tabs` group：
     - `tabs: []`
     - `activeId: ''`
   - 后续再由上层 workspace 的 welcome 补位逻辑接管。

3. 好处
   - 这是 tab 树层面的根修，不依赖某个具体页面。
   - 之后无论关闭的是连接页、表单页还是设置页，最后一个 tab 的行为都会一致。

### Completed Changes

- 已修正 `useTabTree.removeTab()`：
  - 当删除最后一个 tab 导致树为空时，仍会写回一个空 `tabs` group`
- 已保留原有上层 welcome 补位逻辑，不额外改变 UX。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Make Welcome Page a Normal Closable Workspace Tab

### Design Strategy

这次调整的是 welcome 页的角色定位。

1. 之前的问题
   - welcome 页虽然显示在中央工作区里，但并不是普通 tab。
   - 它被当成“保底占位页”：
     - 默认不可关闭
     - 其他 tab 全关后会自动补回来
     - 打开其他 tab 时还会被特殊移除
   - 这和文件列表页、设置页的正常 tab 语义不一致。

2. 调整目标
   - 初始启动时仍然打开 welcome 页。
   - 但 welcome 页本身要变成普通 central workspace tab：
     - 可关闭
     - 关闭后不自动补回
     - 打开其他 tab 时不再被特殊替换

3. 设计取舍
   - 保留 `workspace.ensureWelcomeTab()` 作为“初始打开 welcome”的入口。
   - 去掉 welcome 作为空布局保底页的特殊行为。
   - 空工作区允许存在，不强制用 welcome 占位。

### Completed Changes

- 已将 welcome tab 改为 `closable: true`。
- 已移除 `setLayout()` 中“空布局自动补 welcome”的逻辑。
- 已移除打开其他 tab 时自动删除 welcome 的占位逻辑。
- 已移除 `closeTabById()` / `closeConnectionTabs()` 中“空布局自动补 welcome”的逻辑。
- 初始启动仍然会通过 `workspace.ensureWelcomeTab()` 打开一次 welcome 页。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Replace File Panel Context Menus With Mouse-Following Custom Menus

### Design Strategy

这轮处理三个相关问题：

1. WebView 原生右键菜单没有被屏蔽
2. 文件面板列表视图的右键菜单没有跟随鼠标
3. 图标视图卡片尺寸和排版不够统一

#### 1. 原生右键菜单

- 在当前 WebView 环境里，浏览器原生 `contextmenu` 仍会冒出来。
- 这会和应用自己的右键交互冲突。
- 因此改为在前端入口统一拦截默认 `contextmenu` 事件。

#### 2. 文件面板右键菜单

- 原来列表和图标视图都依赖 Element Plus 的 `el-dropdown trigger=contextmenu`。
- 这种做法在当前场景里存在两个问题：
  - 菜单定位更偏向绑定元素，而不是鼠标点位
  - 很难做到“真正跟随鼠标”的体验
- 因此文件面板右键菜单统一改为自绘固定定位菜单：
  - 右键时记录鼠标坐标
  - 通过 `Teleport` 渲染到 `body`
  - 打开后按视口边界做钳位
  - 点击外部、滚动、缩放、按 `Escape` 都会关闭

#### 3. 图标视图统一性

- 原来的 tile 更像自由高度卡片，内容左对齐。
- 现在统一改成固定卡片尺寸，图标、文件名、元信息整体居中。
- 这样视觉上更像标准文件管理器的图标视图。

### Completed Changes

- 已在 `frontend/src/main.ts` 中统一屏蔽原生 `contextmenu`。
- 已移除文件面板里基于 `el-dropdown trigger=contextmenu` 的右键菜单实现。
- 已为列表视图和图标视图接入统一的自定义右键菜单：
  - 打开
  - 复制路径
  - 下载
  - 下载到临时目录
  - 另存为（仅文件）
  - 重命名
  - 删除
- 已让右键菜单按鼠标位置显示，并在边界处自动调整位置。
- 已让图标视图 tile 统一尺寸并改为内容居中排版。
- 已让图标视图中的文件名支持两行截断，避免卡片高度被文本撑乱。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Fix Stuck Interaction After Opening Directory From File Panel

### Background

- 在文件面板支持自定义右键菜单和条目拖拽后，目录项仍然保持 `draggable=true`。
- 这会让“双击进入目录”和“原生 HTML5 拖拽开始”之间存在竞争关系。
- 一旦浏览器先触发了条目拖拽，而目录加载又立即替换了当前 DOM，旧节点上的 `dragend` 就可能来不及把临时状态清掉。
- 结果表现为：进入目录后界面像被卡住，后续 tab、连接、断连等操作都可能失效或异常。

### Decision

- 先不继续赌浏览器对 `dblclick` / `dragstart` / `dragend` 的具体时序。
- 目录打开前直接强制清理这块文件面板的所有瞬时交互状态：
  - 自定义右键菜单
  - 当前拖放高亮
  - 当前连接条目拖拽 session
  - SplitPane 内部拖拽共享状态
- 同时补一层全局 `dragend` / `drop` 收口，避免源节点在切目录或卸载时漏掉清理。

### Completed Changes

- 已将文件面板的目录双击改为显式 `@dblclick.stop.prevent`，减少与默认拖拽/冒泡的竞争。
- 已新增 `resetTransientInteractionState()`，统一清理：
  - 右键菜单
  - 拖放指示状态
  - 当前 drag session
  - `clearActiveInternalDrag()`
- 已在 `load()` 和 `openItem()` 开始前执行上述清理，确保目录切换前先收口旧状态。
- 已补充全局 `dragend` / `drop` 监听，在浏览器完成原生拖拽时再次兜底清理。
- 已在组件卸载时改为统一走同一套收口逻辑，而不是只关右键菜单。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

### Follow-up

- 在进一步复盘后，判断仅靠“切目录前清理状态”仍不足以彻底解决问题。
- 用户补充“多点几次就卡住，目录树也会卡”，这更符合浏览器原生拖拽被误触发后的表现，而不是单个 Vue 弹层状态没关掉。
- 因此又补了一层更前置的防护：
  - 记录目录项的 `mousedown.detail`
  - 当检测到同一项正处于双击窗口时，直接取消该项的 `dragstart`
- 这样目录双击进入和条目拖拽不再共用同一条浏览器原生起手路径。

### Additional Changes

- 已为列表视图和图标视图条目增加 `mousedown` 记录。
- 已新增 `suppressedDrag` 临时状态，用于标记双击窗口内不允许启动拖拽的条目。
- 已在 `onItemDragStart()` 中对双击窗口内的同一条目执行 `event.preventDefault()` 并立刻收口临时状态。
- 已在拖拽结束和统一收口逻辑里清理 `suppressedDrag`。

### Root Cause Refinement

- 用户补充后，问题特征进一步收窄为：
  - 进入非空目录基本正常
  - 进入空目录后更容易把整个工作区交互搞乱
- 这说明问题不只是条目拖拽本身，更关键的是“空目录会切换到另一套 DOM 分支”。
- 原实现里：
  - 非空目录走列表/网格视口
  - 空目录直接卸载视口，换成单独的 `file-browser__empty`
- 这种在点击序列中整块替换交互容器的方式，在 WebView 里容易让焦点、拖拽、全局监听状态发生错位。

### Structural Fix

- 已去掉“空目录独立页面分支”。
- 现在无论目录是否为空，都保持列表视图 / 图标视图的同一套视口容器常驻。
- 空目录仅在对应视口内部渲染占位内容，不再卸载：
  - `browserViewportRef`
  - 列表/网格容器
  - 对应的键盘、拖放和焦点承载节点
- 这样进入空目录时，不会再在点击过程中替换整块工作区交互 DOM。

## 2026-04-03 - Fix Explorer Tree Freeze When Switching Between Empty And Non-Empty Directories

### Root Cause

- 进一步排查目录树后，发现 `ExplorerTreeNode.vue` 里递归子节点模板存在明确错误：
  - 同一个 `<ExplorerTreeNode>` 同时使用了 `v-for` 和 `v-else`
- 这在 Vue 模板里属于高风险写法。
- 当节点 children 在“空目录”和“非空目录”之间来回切换时，递归分支的 patch 行为会变得不稳定。
- 用户复现路径“目录树里非空目录 -> 空目录 -> 非空目录，循环几次后卡死”与这个模板问题高度一致。

### Completed Changes

- 已将递归子节点渲染改为明确的两层结构：
  - `v-else` 放在 `<template>` 上
  - `v-for` 只保留在内部的 `<ExplorerTreeNode>` 上
- 这样目录树在空/非空子节点之间切换时，Vue 的条件分支和递归列表渲染职责明确分离，不再混用。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Break Workspace Component Cycles Causing Uninitialized Variable Errors

### Root Cause

- 用户提供的报错已经足够明确：
  - `Unhandled Promise Rejection: ReferenceError: Cannot access uninitialized variable.`
  - `Unhandled Promise Rejection: TypeError: null is not an object (evaluating '$setup.items.length')`
- 第二个报错只是连锁反应，说明 `ConnectionOverviewTab` 的 `setup` 在更早阶段已经失败。
- 进一步检查后，`workspace store` 存在组件循环依赖：
  - `workspace.ts` 静态导入 `ConnectionOverviewTab.vue` / `ConnectionFormTab.vue` / `WelcomeTab.vue`
  - 这些组件又反向导入 `useWorkspaceStore()`
- 这种 ESM 循环在 Vue/Wails WebView 里很容易表现成随机的 “Cannot access uninitialized variable”。
- 目录树频繁切换目录、重开 tab、重挂载组件时，会更容易触发这类问题。

### Completed Changes

- 已将 `workspace store` 中的工作区组件注册改为 `defineAsyncComponent()` 懒加载：
  - `WelcomeTab`
  - `ConnectionFormTab`
  - `SettingsTab`
  - `ConnectionOverviewTab`
- 这样 `workspace.ts` 不再在模块初始化阶段直接静态拉起这些组件，从而拆掉 store 与工作区组件之间的强循环依赖。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Normalize Empty Directory Listing To Empty Slices Across All Drivers

### Decision

- 是，后端返回空列表比返回 `null` 更合适。
- 对当前项目来说，目录 listing 的语义是“集合”，空目录应该稳定表示为 `[]`。
- 如果返回 `null`：
  - 前端所有 `length` / `filter` / `map` 都要额外防御
  - Wails/JS 桥接时更容易把空目录和异常状态混淆
  - 目录树与文件面板都会被空目录边界条件放大成运行时错误

### Completed Changes

- 已在 `fileops.Service.ListDirectory()` 中统一保证：空结果返回空切片，不返回 `nil`。
- 已在 `App.ListConnectionDirectory()` 中再补一层保护，确保对前端暴露的方法也不返回 `nil`。
- 已统一调整所有驱动的 `List()` 初始化行为，确保空目录从驱动层开始就构造空切片：
  - `folder/local/driver.go`
  - `folder/sftp/driver.go`
  - `folder/s3/driver.go`
  - `folder/alibaba-oss/driver.go`
- 已覆盖 `Local` 的回归测试，验证空目录 listing 不会返回 `nil`。

### Verification

- `go test ./fileops ./folder/local` 通过。
- `cd frontend && npm run type-check` 通过。

## 2026-04-03 - Revert Overeager Global Drop Cleanup That Broke File Drag-Drop

### Root Cause

- 文件/目录拖动到其他目录失效，根因大概率来自前一轮为修空目录卡死而补的“全局拖拽收口”。
- 当时在 `ConnectionOverviewTab` 里加了：
  - `window.addEventListener('dragend', onWindowDragComplete, true)`
  - `window.addEventListener('drop', onWindowDragComplete, true)`
- 因为使用了捕获阶段，`window` 会在目标目录自己的 `drop` 处理器之前先执行。
- 而 `onWindowDragComplete()` 又会调用 `resetTransientInteractionState()`，其中包含 `clearActiveInternalDrag()`。
- 结果就是：真正的目标目录还没来得及读取拖拽 payload，内部拖拽状态就已经被提前清空，导致拖放失效。

### Completed Changes

- 已将这两个全局监听改回普通冒泡阶段，不再在捕获阶段抢先执行。
- 已将全局拖拽收口改成 `setTimeout(..., 0)` 延后执行，确保目标目录的本地 `drop` 逻辑先完成。
- 这次等于把“过早清空内部拖拽状态”的改动实质性回退掉了。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Unify Remote Path Identity Across Workspace Selection and Explorer Tree

### Root Cause

- 这次不是单点 bug，而是“路径身份模型”在前端被混用了两套：
  - 中央工作区的选中态、活动项已经逐步切到了规范化 path；
  - 但列表排序后的路径序列、区间选择、目录树节点 key、拖放刷新命中条件里，仍混着原始 `item.path`。
- 结果会直接表现成两类问题：
  - `Shift` / `Ctrl(Cmd)` 多选不稳定，因为选择状态和当前列表顺序不是同一套 key；
  - 同连接目录拖放后，中央工作区能删项，但目录树刷新命不中已加载节点，导致树里残留旧目录。

### Design Decision

- 把“远程路径身份”提升为显式共享规则，不再让工作区、目录树、拖放模块各自清理一次路径。
- 前端统一使用规范化后的 remote path 作为：
  - 选中 key
  - 活动项 key
  - 拖放命中 key
  - 目录树节点 key
  - 目录刷新 key
- 原始 `item.path` 继续保留给展示和后端调用，但不再作为前端状态主键。

### Completed Changes

- 新增共享路径工具：
  - `frontend/src/composables/remotePath.ts`
  - 统一提供 `normalizeRemotePath()`、`joinRemotePath()`、`parentRemotePath()`、`baseRemotePath()`、`isRemotePathWithin()`
- `connectionEntryDrop.ts` 已全部改用共享路径工具，不再保留本地 `cleanRemotePath()` 私有实现。
- 中央工作区 `ConnectionOverviewTab.vue` 已统一到规范化 path：
  - `orderedPaths()`
  - `findItemByPath()`
  - `reconcileSelection()`
  - 目录级 drop target
  - DOM `data-entry-path`
  - 列表/图标项 key
- 目录树 `ExplorerPanel.vue` 已统一到规范化 path：
  - `nodeKey()`
  - `buildDirectoryChildren()`
  - `buildRefreshNode()`
  - `onDirectoryRefresh()`
- 同连接目录移动完成后，目录树现在会先做本地收口：
  - 源父节点已加载时，先立即移除被移动目录
  - 被移动目录的旧 subtree cache 会一起清理
  - 目标节点如果已加载，再做局部 `loadChildren()` 校准
- 拖放生命周期事件补充了 `sourceDirectoryPaths`，用于目录树精确知道“哪些目录被移动了”。

### Backend Audit

- 已同步复核后端 `fileops` 路径语义：
  - `cleanPath()`
  - `joinChildPath()`
  - `MoveEntry()`
  - `ListDirectory()`
- 结论是后端当前的路径规范化语义与前端这次统一后的 remote path 规则一致：
  - 空根目录统一为 `""`
  - 反斜杠统一转正斜杠
  - 多余前导 `/` 不作为路径身份的一部分
- 这轮后端不需要再额外改逻辑，问题主要在前端状态主键不一致。

### Verification

- `go test ./fileops` 通过。
- `cd frontend && npm run test:unit -- connectionEntryDrop.spec.ts` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Improve Multi-Selection Drag Preview In Workspace File Browser

### Problem

- 中央工作区当前拖拽源虽然已经支持多选，但浏览器默认 drag preview 只会抓取一个源元素。
- 结果就是：
  - 多选拖拽时用户视觉上仍像是在拖一个文件
  - 列表视图和图标视图都缺少“这是多项拖拽”的即时反馈

### Decision

- 不额外设计一套独立图标系统，直接复用当前文件面板的真实 DOM 作为 `dataTransfer.setDragImage()` 的来源。
- 这样有两个好处：
  - 拖拽预览和当前列表/图标视图风格天然一致
  - 不需要手写第二套文件/目录图标映射

### Completed Changes

- 在 `ConnectionOverviewTab.vue` 中新增自定义 drag preview 生成逻辑：
  - 多选时最多克隆前 3 个已选项作为叠片预览
  - 列表视图显示多行叠片
  - 图标视图显示多卡片叠片
- 右上角增加数量徽标，直接显示当前拖拽项数量。
- 拖拽结束、取消拖拽、切目录或组件卸载时，都会主动清理临时 drag preview DOM，避免残留。
- 查询源 DOM 节点时补了 `CSS.escape()` 降级，兼容 WebView 环境。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Reconnect Live Connections After Save And Serialize Frontend Hydration

### Problem

- 连接编辑页保存后，仓库里的配置虽然更新了，但如果该连接已经处于已连接状态，运行中的后端实例仍会继续复用旧配置。
- 同时，前端 `connections` store 的 `hydrate()` 在并发调用时只是看到 `loading=true` 就直接返回，后续调用方不会等待第一次加载完成。
- 这两者叠加后，最容易表现成：
  - 编辑连接后仍在跑旧的 `Local rootPath` / `SFTP` 地址 / 云端配置
  - 编辑页或其他依赖 `definitionMap/stateMap` 的组件偶发拿到未完成的 hydration 状态

### Root Cause

- 后端 `connection.Service.Save()` 原来只负责校验并写入仓库，不处理已连接实例的生命周期。
- `connection.Service.Open()` 又会在 `states[id].Connected == true` 时直接返回，不重新建实例。
- 前端 `connections` store 的 `hydrate()` 没有像 `settings` store 那样维护 `pendingHydration` promise。

### Completed Changes

- 后端：
  - `connection.Service.Save()` 现在会在保存成功后清理该连接旧实例。
  - 如果保存前该连接处于已连接状态，则会立即按新配置重新 `Open()`，确保运行时实例与最新配置一致。
  - 同时会同步刷新 `states` 中的 `Name` / `Driver` / `Capabilities` / `Connected` 状态，避免 `Open()` 被旧的 `Connected=true` 短路。
- 前端：
  - `frontend/src/stores/connections.ts` 增加 `pendingHydration`，并发 `hydrate()` 现在会复用同一个 promise，而不是直接返回。
  - `saveConnection()` 现在保存后会同时刷新：
    - `definitions`
    - `states`
  - 这样编辑成功后，Explorer 和工作区状态会立即反映最新连接状态。

### Regression Test

- 新增后端测试：已连接 `Local` 连接修改 `rootPath` 后，会自动重建实例并切到新目录。
  - `connection/connection_test.go`

### Verification

- `go test ./connection ./fileops` 通过。
- `go test ./...` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Refresh Explorer And Workspace Explicitly After Connection Reconfiguration

### Root Cause

- 修完后端“保存即重连”后，运行时实例已经变成了新配置，但前端仍缺少“连接已重配置”的显式刷新链路。
- 原来的保存流程只是：
  - `saveConnection()`
  - `refreshConnections()`
  - `workspace.openConnection(saved.id, saved.name)`
- 这会导致两个缓存继续留在旧状态：
  - Explorer 的 `childrenByKey`
  - Workspace 的 `connectionTabState.items`
- 因为保存连接既不是目录刷新事件，也不是拖放事件，所以原有监听链路根本不会命中。

### Completed Changes

- 新增专门的前端事件：
  - `frontend/src/composables/useConnectionConfigRefresh.ts`
  - 事件名：`workspace:connection-config-refresh`
- 连接表单保存成功后：
  - 先清理该连接的 workspace 浏览缓存
  - 发出“连接已重配置”事件
  - 再把工作区切回该连接根目录 tab
- `ConnectionOverviewTab.vue` 现在会监听该事件：
  - 命中当前连接时清空旧缓存
  - 强制回到连接根并重新 `load('')`
- `ExplorerPanel.vue` 现在会监听该事件：
  - 清掉该连接原有目录树缓存
  - 如果该连接节点原本已展开，则按新配置重新加载根节点

### Effect

- 保存连接后，中央工作区不再“偶发继续显示旧目录”。
- 保存连接后，目录树不再继续挂着旧的根目录子项。
- 这次不再依赖“再次打开同一个 tab”碰运气，而是把“连接语义已变化”作为显式刷新事件处理。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Tighten File Browser Toolbar Controls And Fix View Toggle Alignment

### Problem

- 文件面板顶部工具栏按钮过大、间距偏松，连续操作看起来像几块分散的按钮，不像文件管理器里的紧凑工具条。
- 列表/图标视图切换按钮没有真正做内容居中，图标在按钮内的视觉位置偏怪。

### Completed Changes

- 将顶部主操作改成连续的小号按钮组：
  - `新建目录`
  - `刷新`
  - `字段`
- 将批量操作区也改成连续的小号按钮组：
  - `下载所选`
  - `复制所选路径`
  - `删除所选`
  - `清空选择`
- 调整工具栏整体布局：
  - 缩小间距
  - 统一按钮高度和内边距
  - 移动端下保持左对齐
- 修正视图切换按钮：
  - 改成固定尺寸的方形按钮
  - `display: inline-flex`
  - `align-items/justify-content: center`
  - 图标字号与按钮尺寸重新匹配

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-03 - Compact Breadcrumbs And Status Bar Into A Secondary Toolbar

### Problem

- 顶部主按钮已经收紧后，下面的 breadcrumbs 和状态条仍然偏松散：
  - 路径导航像一串独立文字按钮
  - 状态信息像简单文本拼接
- 整体不够像文件管理器里的次级工具条。

### Completed Changes

- 将路径导航包装成紧凑的胶囊式 breadcrumbs 容器：
  - 根节点单独加粗
  - 分隔符和按钮尺寸统一缩小
  - hover 改成轻量主色叠层
- 将连接状态、项目数量、选中数量改成右侧状态胶囊：
  - 在线状态保留小圆点
  - 选中数量使用强调样式
- 新增 `file-browser__subbar`，把路径条和状态条组成同一层次的次级工具条。
- 移动端下自动改为上下堆叠，避免过窄时挤坏布局。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-04 - Fix OSS Connection Form Replay During Edit Hydration

### Root Cause

- 连接编辑页初始化时，表单默认驱动是 `Local`。
- 编辑 `OSS` 连接时，`applyDefinition()` 会把 `form.driver` 从 `Local` 改成 `OSS`。
- 但当前代码在真正 `applyDefinition(def)` 之前，就先把 `hydratedEditModel` 置为 `true`。
- 这会导致 `watch(() => form.driver)` 把这次“编辑态回填驱动”误判成“用户主动切换驱动”，进而调用 `onDriverChanged()`，把刚回填的 OSS 配置重置成默认空值。

### Completed Changes

- 调整编辑页初始化顺序：
  - 先 `applyDefinition(def)`
  - 后 `hydratedEditModel.value = true`
- 这样驱动 watch 只会在真正完成编辑态回填之后才接管后续用户操作，不会破坏已有配置的回显。

### Effect

- `OSS` 编辑页现在能正确回显：
  - `region`
  - `bucket`
  - `accessKeyID`
  - `accessKeySecret`
  - `securityToken`
  - `endpoint`
  - `prefix`
  - `forcePathStyle`
  - `useCName`
  - `disableSSL`
- 同类问题对 `S3` / `SFTP` 的非默认驱动编辑回填也一并更稳。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-04 - Fix Non-Local Driver Config Replay Being Reset By Async Driver Watch

### Investigation

- 重新审查了驱动相关前后端链路：
  - 后端 `connections.yaml -> config.LoadConnectionsConfig() -> connection.definitionFromConfig()` 读取 `OSS` 配置是完整的。
  - 仓库中的 `connections.yaml` 里 `OSS` 字段名与前端 schema、后端 `folder/alibaba-oss.Options` 完全一致：
    - `accessKeyID`
    - `accessKeySecret`
    - `securityToken`
    - `endpoint`
    - `prefix`
    - `forcePathStyle`
    - `useCName`
    - `disableSSL`
- 这说明问题不在后端持久化、解码或 Wails 绑定，而在前端表单回填时序。

### Root Cause

- `Local` 能回显、`OSS` 不能回显，关键差异是：
  - 表单默认驱动就是 `Local`
  - 编辑 `Local` 时，`form.driver` 没有变化
  - 编辑 `OSS/S3/SFTP` 时，`form.driver` 会从 `Local` 改成目标驱动
- `ConnectionFormTab` 里存在 `watch(() => form.driver)`。
- `applyDefinition(def)` 虽然先回填了 `config`，但这个 `watch` 默认是异步调度的：
  - `form.driver = 'OSS'` 会先排队一个 watch job
  - `applyDefinition()` 继续执行，把 `form.config` 写成正确的 OSS 配置
  - 当前 tick 结束前，编辑态初始化又把 `hydratedEditModel` 设成了 `true`
  - 随后异步 watch 执行，把这次“编辑态回填”误判成“用户切换驱动”
  - `onDriverChanged()` 把 `form.config` 重置成默认空配置
- 这就是为什么：
  - `Local` 正常
  - `OSS` 不正常
  - 而且同类问题理论上也会影响 `S3`、`SFTP`

### Completed Changes

- 新增 `suppressDriverWatch`，在 `applyDefinition()` 回填期间临时屏蔽 `watch(form.driver)`。
- 等表单完整落地并进入下一次 DOM 更新后，再恢复这条 watcher。
- 这次修复针对的是“编辑态回填期间的驱动切换误判”，不是 OSS 特判，因此对所有非默认驱动都生效。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-04 - Restore OS File Drop Uploads Into The Central Workspace

### Root Cause

- 外部文件/目录拖入中央工作区上传，依赖的是前端的 panel-level OS file drop 协调链路：
  - `usePanelFileDrop()` 在面板 DOM `drop` 时记录 panel 身份
  - `useFileDrop()` 再通过 Wails `OnFileDrop()` 拿到真实本地路径并触发上传
- 这条链路要求对应的 `SplitPane` 显式开启 `enableFileDrop`。
- 但当前代码里有两处没有把这个能力真正贯通：
  - `AppShell -> Tabs` 没有开启 `enable-file-drop`
  - `Tabs -> TabNodeRenderer -> SplitPane` 也没有继续向内传递 `enableFileDrop`
- 结果是：
  - 中央工作区未分割时，中心 panel 不一定会登记成 OS drop target
  - 中央工作区已分割时，内部 split panel 也不会登记成 OS drop target
  - 最终 `OnFileDrop()` 虽然可能收到路径，但拿不到对应 panel/tab 上下文，上传不会发起。

### Completed Changes

- 为 `Tabs` 新增并贯通 `enableFileDrop` prop。
- `AppShell` 中央工作区的 `Tabs` 现在显式开启 `enable-file-drop`。
- `Tabs -> TabNodeRenderer -> SplitPane` 全链路继续向内传递 `enableFileDrop`，确保分割后的子 panel 也能接收外部文件拖放。
- `SkeletonLayout` 外层主分栏 `SplitPane` 也开启了 `enable-file-drop`：
  - 未分割时由中心 panel 兜底接收
  - 已分割时由更内层 panel 优先命中。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-04 - Clear Split Panel OS Drop Overlays Globally After Drop Completion

### Problem

- 中央工作区分割成多个 panel 后，外部文件拖放上传已经恢复可用。
- 但存在一个收口问题：
  - 从右侧 panel 横向移动到左侧 panel 并释放后
  - 左侧 panel 的 drop overlay 可能残留，不会自动消失。

### Root Cause

- `usePanelFileDrop()` 目前每个 panel 都独立维护自己的：
  - `isDragOver`
  - `dragCounter`
- 这套状态主要依赖局部的 `dragenter / dragleave / drop` 来收口。
- 分割后 panel 数量增多，指针在多个 panel 之间穿梭时，局部事件次序在 WebView 下并不稳定：
  - 某个 panel 可能收到多次 `dragenter`
  - 但未必能在 drop 结束时收到对等的 `dragleave`
- 结果就是某个 panel 的 overlay 状态残留。

### Completed Changes

- 新增全局事件 `workspace:panel-os-file-drop-reset`。
- 在 `useFileDrop()` 中：
  - DOM `drop` 发生时立即广播一次 reset
  - Wails `OnFileDrop()` 收到真实路径时再广播一次 reset
- 在 `usePanelFileDrop()` 中：
  - 所有 panel 统一监听这个 reset 事件
  - 一旦收到，就直接清空本地 `isDragOver` 和 `dragCounter`
- 这样不再依赖每个 panel 自己能否完整收到配对的 `dragleave`，drop 完成后所有残留 overlay 都会统一收掉。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-04 - Move OS File Drop Targets Down To Leaf Tab Groups To Avoid Overlay Stacking

### Problem

- 外部文件拖入中央工作区时：
  - 未分割，上传正常
  - 分割后，也能分别拖入左右子面板
- 但如果中央工作区存在多分隔布局，会出现遮罩层叠加：
  - 外层中间大遮罩
  - 内层各个子面板自己的遮罩
- 视觉和交互都不对，因为真正应该接收外部文件拖放的是“当前叶子工作区”，不是外层所有容器一起响应。

### Root Cause

- 之前为了恢复外部文件拖入上传，给两层容器都打开了 `enableFileDrop`：
  - `SkeletonLayout` 外层主分栏 `SplitPane`
  - `Tabs` 内部的 `SplitPane`
- 这样在多分隔场景下，同一个 OS drag 会同时命中：
  - 外层 center panel
  - 内层 split panel
- 于是就出现“中间大遮罩 + 子面板遮罩”叠加。
- 这说明 OS 文件拖放目标设计层级有误：
  - 外部文件拖放应该只落到最末端的 leaf workspace
  - 不该让更外层的 layout 容器也参与同类遮罩。

### Completed Changes

- 将 OS 文件拖放目标下沉到 `Tabs` 的叶子 `TabGroup`：
  - `TabGroup` 现在直接接入 `usePanelFileDrop()`，作为真正的外部文件 drop target。
- `TabsContext` 新增 `enableFileDrop`，由 `Tabs` 统一向下提供。
- `TabNodeRenderer` 改为：
  - 叶子 `TabGroup` 继续接收 `enableFileDrop`
  - 中间层 `SplitPane` 不再启用自己的 OS file drop overlay
- `SkeletonLayout` 外层主分栏 `SplitPane` 关闭 `enable-file-drop`。
- `usePanelFileDrop()` 新增 `resolveGroupHost()`：
  - 既兼容以前的 `SplitPanePanel -> 内部 tab group`
  - 也兼容现在直接把 `TabGroup` 自己作为 drop target。

### Result

- 未分割时：
  - 当前 tab 内容区会显示一层 drop overlay
- 多分割时：
  - 只有鼠标所在的叶子工作区会显示 overlay
- 不再出现外层大遮罩覆盖内层子面板遮罩的重叠情况。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-04 - Add Regression Tests For Split Collapse After Closing One Side

### Goal

- 为“左右分割时关闭左边最后一个 tab，左侧空余位置偶发仍残留”的问题补充可执行回归测试。
- 先固定问题，再做修复，避免继续靠人工极端场景复现。

### Added Tests

- `frontend/src/components/Tabs/__tests__/useTabTree.spec.ts`
  - 新增纯树变换测试：
    - 关闭 split 一侧最后一个 tab 后
    - 布局树应从 `split` 塌缩成剩余的单个 `tabs` 节点
- `frontend/src/components/Tabs/__tests__/Tabs.spec.ts`
  - 新增 UI 层测试：
    - 挂载真实 `Tabs`
    - 创建左右 split
    - 关闭左侧最后一个 tab
    - 断言 `SplitPane` DOM 与 `.split-pane-panel` 不应残留

### Result

- 纯树变换测试通过，说明 `useTabTree.removeTab()` 在数据结构层面已经能把 split 塌成单 child。
- 新增的 UI 回归测试稳定失败：
  - `SplitPane` 在关闭左侧最后一个 tab 后仍然存在
  - 这证明问题确实不在树模型本身，而在 UI 层 `Tabs -> TabNodeRenderer -> SplitPane` 的卸载 / 收口链路。

### Verification

- `cd frontend && npm run test:unit -- src/components/Tabs/__tests__/useTabTree.spec.ts src/components/Tabs/__tests__/Tabs.spec.ts`
  - `useTabTree.spec.ts` 通过
  - `Tabs.spec.ts` 新增回归用例失败（这是当前已确认的待修复问题）

## 2026-04-04 - Fix Split Pane Residual Space After Closing One Side's Last Tab

### Root Cause

- 之前已经补了回归测试，确认问题不在树模型层，而在 UI 层：
  - `useTabTree.spec.ts` 证明纯树变换下，`split -> single tabs` 可以正确塌缩
  - `Tabs.spec.ts` 则证明真实组件运行时，关闭左侧最后一个 tab 后，`SplitPane` 仍会残留
- 继续沿着失败用例排查后，定位到 `useTabTree.removeTab()`：
  - 它之前采用的是“先原地修改当前 group 的 proxy，再根据 `node.tabs.length` 决定是否调用 `replaceNode()`”的写法
  - 在纯对象测试里这条逻辑没问题
  - 但在组件真实运行时，这个分支会退化成：
    - 左侧 group 被改成空 tabs
    - 整体 tree 仍保留原来的 `split`
    - 最终形成一个空的左 group + 仍然存在的 `SplitPane`
- 调试输出已经验证了这一点：
  - `update:modelValue` 收到的是带空左 group 的 `split`
  - 而不是应该塌成的单个右侧 `tabs`

### Completed Changes

- 将 `useTabTree.removeTab()` 改成纯不可变更新：
  - 先基于 `rawGroup` 计算 `nextTabs`
  - 如果删完为空，直接 `replaceNode(tree.value, rawGroup.id, null)`
  - 如果删完还有 tab，则构造 `nextGroup`，再整体替换回树里
- 不再依赖“先原地改 proxy，再看长度”的分支判断。
- 这样数据结构层和真实组件运行时都走同一条确定性的替换路径。

### Verification

- `cd frontend && npm run test:unit -- src/components/Tabs/__tests__/useTabTree.spec.ts src/components/Tabs/__tests__/Tabs.spec.ts` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-04 - Preserve Modification Time On Upload And Download

### Goal

- 按既定最小方案，为上传、下载和跨连接传输补上“保留文件修改时间”。
- 优先保证文件级语义正确：
  - `Local` / `SFTP` 直接写远端 mtime。
  - `S3` / `OSS` 通过对象 metadata 保存原始 mtime，下载时恢复到本地。
- 目录本身的 mtime 这轮暂不处理，避免把范围扩展到目录创建和回填链路。

### Completed Changes

- 在 `folder` 层补了统一 mtime 工具：
  - 新增 `fileutil-mtime` metadata key。
  - 新增 metadata 合并、解析和本地 `os.Chtimes()` 辅助。
- 扩展传输和写入参数：
  - `folder.TransferRequest` 新增 `PreserveModTime`、`SourceModTime`。
  - `folder.WriteOptions` 新增 `ModTime`。
- `TransferManager` fallback 路径补齐 mtime：
  - 上传时从本地文件读取 `ModTime()`，传给 `WriteOptions.ModTime`。
  - 下载时从 `Stat()` 的 `LastModified/Metadata` 解析源时间，写完后 `os.Chtimes()` 恢复本地时间。
- 各驱动补齐写入与下载语义：
  - `Local` 写完文件后调用 `os.Chtimes()`。
  - `SFTP` 写完文件后调用 `client.Chtimes()`。
  - `S3` / `OSS` 上传时把原始 mtime 写入对象 metadata，下载时优先从 metadata 恢复到本地文件；没有 metadata 时回退到对象自己的 `LastModified`。
- `transfer.Service` 统一收口：
  - 所有上传/下载/跨连接 follow-up 任务默认都带 `PreserveModTime: true`。
  - 这样跨连接“先下载到临时文件，再上传”也会自然保留原始文件时间。

### Notes

- 这轮没有改前端展示逻辑。
- 因此 `S3` / `OSS` 在文件列表中显示的仍然是对象系统自己的 `LastModified`，不是 metadata 里的原始文件时间。
- 但实际下载到本地的文件时间，以及再次上传到其他连接后的文件时间，已经会被保留。

### Verification

- `go test ./folder/...` 通过。
- `go test ./transfer/...` 通过。
- `go test ./...` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-06 - Add Backspace Inline Delete Shortcut

### Root Cause

- 文件面板删除已经改为内联确认，但键盘 `Backspace` 仍优先执行“返回上级目录”。
- 对文件管理器语义来说，当存在已选中文件/目录时，`Backspace` 应优先作用于选中项，而不是目录导航。
- 右键菜单中的删除项没有提示对应快捷键，用户无法发现该键盘入口。

### Completed Changes

- `Backspace` 快捷键优先级调整：
  - 当前有选中项时，进入内联删除待确认状态。
  - 选中一个或多个文件/目录都走同一套 `deleteSelected()` / `beginInlineDelete()` 逻辑。
  - 当前没有选中项时，保留原来的返回上级目录行为。
- 右键菜单 `删除` 项右侧增加 `⌫` 快捷键标记。
- 右键菜单项布局增加左右分布，快捷键标记固定在最右侧。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-06 - Inline File Delete Confirmation

### Root Cause

- 文件删除仍依赖 `ElMessageBox.confirm`。
- 当前 WebView 下 Element 弹框链路不稳定，删除确认应收口到文件面板内部状态。
- 内联确认按钮需要放在文件名右侧，但长文件名不能覆盖按钮；按钮必须在布局层级上固定可见。

### Completed Changes

- 删除触发入口改为内联确认：
  - 右键菜单 `删除` 只让当前项进入确认态。
  - 顶部 `删除所选` 让当前选中项进入批量确认态。
  - 键盘 `Delete` 也复用同一套入口。
- 列表视图中，确认按钮显示在文件名右侧：
  - 文件名使用 flex ellipsis。
  - `删除 / 取消` 按钮使用 `flex: 0 0 auto` 和更高层级，避免被长文件名覆盖。
- 图标视图中，确认按钮显示在文件名下方，仍跟随对应文件卡片。
- 文件项根节点从 `button` 改为 `div role="button"`：
  - 避免在按钮内嵌套按钮造成无效 HTML。
  - 仍保留现有点击、双击、右键和拖放事件。
- 删除确认完成后按原逻辑执行 `DeleteConnectionEntry()`，再刷新当前目录。
- 新增 `workspace.fileBrowser.cancelDelete` 中英文文案。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

## 2026-04-06 - Show Files In Explorer Tree And Inline New Folder

### Root Cause

- Explorer 目录树之前在构建子节点时直接过滤掉非目录项，只展示目录。
- Explorer 节点图标与中央工作区各自维护，文件类型图标没有统一匹配规则。
- 中央工作区单击只会选中，不支持再次单击取消选中；双击打开和单击选择需要明确分离。
- 新建目录依赖 `ElMessageBox.prompt`，在当前 WebView 下弹框链路不稳定。
- 右键命中文本时，WebView 仍可能产生原生文本选区。

### Completed Changes

- Explorer 树节点类型扩展为：
  - `connection`
  - `directory`
  - `file`
- Explorer 加载目录时不再过滤文件，文件节点作为叶子节点展示：
  - 文件节点不显示可用展开箭头。
  - 文件节点不接收拖放目标。
  - 文件节点不触发悬停展开。
  - 双击文件节点调用 `OpenConnectionFile()`。
- Explorer 图标改用中央工作区同一套 `resolveFileIcon()` 匹配逻辑。
- 中央工作区单击选择改为延迟提交：
  - 单击未选中项：选中。
  - 单击已选中项：取消选中。
  - 双击时取消挂起的单击选择，再执行打开。
  - `Shift` 和 `Ctrl/Cmd` 多选逻辑保持即时执行。
- 新建目录改为内联首项：
  - 点击“新建目录”后滚动到顶部。
  - 列表视图首行显示目录图标和输入框。
  - 图标视图首个卡片显示目录图标和输入框。
  - 输入框 blur 或 Enter 且文本非空时调用 `CreateConnectionDirectory()`。
  - Escape 或空文本 blur 时取消。
- 文件面板根节点和右键菜单增加 `user-select: none`，输入框显式恢复 `user-select: text`。

### Verification

- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

### Added Tests

- `folder/transfer_manager_test.go`
  - 新增 fallback upload 保留 mtime 测试。
  - 新增 fallback download 恢复本地 mtime 测试。
- `folder/local/driver_test.go`
  - 新增 `Local` 直写保留 mtime 测试。

## 2026-04-04 - Extend Modification Time Preservation To Directories

### Root Cause

- 文件 mtime 之前已经打通，但目录不能沿用同样策略。
- 对 `Local` / `SFTP` 来说，目录如果在创建后立刻 `Chtimes()`，后续子文件上传或下载仍会再次改写父目录时间。
- 因此目录 mtime 必须延后到“整批文件任务全部结束后”再统一回填。

### Completed Changes

- 新增可选能力 `folder.DirectoryModTimeSetter`：
  - `Local` / `SFTP` 直接对真实目录调用 `Chtimes()`。
  - `S3` / `OSS` 通过目录 marker object 的 metadata 写入原始目录 mtime。
- `S3` / `OSS` 目录 `Stat()` 现在会优先探测 `dir/` marker：
  - 如果存在 marker，就返回 metadata 里的原始目录 mtime。
  - 仍兼容没有 marker 的虚拟目录回退逻辑。
- `S3` / `OSS` 递归 `List()` 现在会把目录 marker 也作为目录项返回，便于传输计划识别目录层级。
- `transfer.Service` 的目录传输计划已扩展：
  - 上传计划会记录本地目录 mtime。
  - 下载计划会记录远端目录 mtime。
  - 跨连接计划会把源目录 mtime 带到目标目录。
- 新增目录 finalize 链路：
  - 目录下载：等所有文件下载结束后，再回填本地目录 mtime。
  - 本地目录上传：等所有文件上传结束后，再回填远端目录 mtime。
  - 跨连接目录传输：等 follow-up upload 全部结束后，再回填目标目录 mtime。
- 空目录是单独处理的：
  - 没有文件任务时，会在建目录后立即回填目录 mtime。

### Notes

- 对 `S3` / `OSS`，只有存在显式目录 marker 的目录才能携带原始目录 mtime。
- 如果对象存储里目录只是由文件 key 隐式推导出来、没有 marker object，那么仍然只能当作虚拟目录处理，这类目录没有可恢复的原始 mtime。
- 这与当前对象存储语义一致，也是这轮实现的边界。

### Verification

- `go test ./transfer/...` 通过。
- `go test ./folder/...` 通过。
- `go test ./...` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

### Updated Tests

- `transfer/service_test.go`
  - 调整目录上传 / 下载 / 跨连接计划测试，覆盖目录计划结构和 mtime 载体。

## 2026-04-04 - Fix Directory Mtime Finalizer Races And Virtual Object Directories

### Root Cause

- 目录 mtime finalize 之前存在两个竞态窗口：
  - 任务可能在 `taskFinalizers` 建立映射前就已经完成，导致 finalize 永远收不到终态事件。
  - 跨连接 follow-up upload 在 `attachDirectoryFinalizeTask()` 之前就可能已经结束，导致 deferred finalize 永远不 settle。
- 另外，`buildLocalDownloadPlan()` 之前只收显式目录项，没把对象存储里仅由文件 key 隐式推导出来的虚拟子目录纳入计划。

### Completed Changes

- 在注册和 attach 目录 finalizer 后，立即对已完成任务做一次 reconciliation：
  - 如果任务已经处于 `Completed/Failed/Cancelled`，直接补跑 `processDirectoryFinalizers()`。
  - 这样不会再丢失“先完成、后登记”的终态。
- `buildLocalDownloadPlan()` 现在会从文件路径反推出虚拟子目录：
  - 例如只有 `photos/2026/trip.png` 时，也会把 `photos/2026` 纳入目录计划。
  - 如果没有显式目录 marker，就用子文件的最新 mtime 作为该虚拟目录的近似时间来源。
- 这套虚拟目录推导会自然传递到：
  - 下载目录回填
  - 本地上传目录回填
  - 跨连接目录回填

### Verification

- `go test ./transfer/...` 通过。
- `go test ./...` 通过。
- `cd frontend && npm run type-check` 通过。
- `cd frontend && npm run build-only` 通过。

### Added Tests

- `transfer/service_test.go`
  - 新增“任务已完成后再注册 finalizer”回归。
  - 新增“任务已完成后再 attach deferred finalizer”回归。
  - 新增“从文件 key 推导虚拟目录并携带 mtime”回归。

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
