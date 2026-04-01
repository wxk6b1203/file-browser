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
