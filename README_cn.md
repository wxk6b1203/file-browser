# File Browser

[English README](README.md)

File Browser 是一个基于 Wails v2 的桌面文件管理器，后端使用 Go，前端使用 Vue 3。项目目标是把本地目录、SFTP 服务器、S3/OSS 对象存储等不同接入方式统一到同一个文件工作区里，并采用接近 VS Code 的交互模型：连接目录树、中央 Tab、分割面板、任务面板、通知面板、列表/图标视图和拖拽操作。

当前支持的存储后端：

- 本地文件系统
- SFTP
- Amazon S3 或兼容 S3 的对象存储
- Alibaba Cloud OSS

## 功能概览

- 支持 Local、SFTP、S3、OSS 多连接管理。
- VS Code 风格骨架：左侧 Explorer、中央可分割 Tab、右侧任务/通知面板。
- 文件面板支持列表视图和图标视图。
- 列表视图支持字段显示、排序、列宽拖动、多选、键盘导航和内联创建/删除。
- Explorer 目录树支持展示连接、目录和文件。
- 支持打开、重命名、删除、新建目录、复制、移动、上传、下载、另存为等文件操作。
- 支持同连接内拖拽移动，以及跨连接拖拽传输。
- 支持从操作系统外部拖入文件或目录并上传到当前工作区。
- 支持传输任务进度、取消、任务清理、默认下载路径、下载后打开和定位本地路径。
- 上传/下载/跨连接传输会在后端能力允许时保留文件和目录修改时间。
- 支持搜索、通知、连接测试和 UI 偏好设置。
- 支持多主题，包括接近 VS Code 的 `2026 Dark` 主题。

## 项目结构

```text
.
├── app.go                 # 暴露给前端的 Wails App 方法
├── config/                # 应用配置和连接配置读写
├── connection/            # 连接配置仓库和运行时连接生命周期
├── fileops/               # 面向 UI 的文件操作服务
├── folder/                # 存储驱动抽象和具体驱动
│   ├── local/             # 本地文件系统驱动
│   ├── sftp/              # SFTP 驱动
│   ├── s3/                # S3 驱动
│   └── alibaba-oss/       # Alibaba Cloud OSS 驱动
├── transfer/              # 高层传输编排
├── search/                # 面向连接的搜索服务
├── shortcut/              # 前后端快捷键事件桥接
├── logging/               # Zap 日志初始化
├── frontend/              # Vue 3 + Vite + Element Plus 前端
├── original.md            # 原始产品需求
├── developing.md          # 开发过程、设计和决策记录
└── TODO.md                # 后续任务记录
```

## 架构说明

### 后端

后端使用 Go 编写，通过 Wails 把 `app.go` 中的公开方法绑定给前端调用。

- `folder.Manager` 定义统一文件操作接口：`List`、`Stat`、`Exist`、`Copy`、`Move`、`Rename`、`Delete`、`Mkdir`。
- 可选驱动接口包括 `Reader`、`Writer`、`Transferer`、`Presigner`、`HealthChecker`、`Closer`。
- `folder/transfer_manager.go` 负责异步上传/下载任务、进度、速度、取消和 fallback 流式传输。
- `connection/` 负责保存连接定义，并管理已连接的运行时实例。
- `fileops/` 封装 UI 触发的文件操作，避免前端直接处理驱动差异。
- `transfer/` 负责目录传输计划、跨连接传输、follow-up upload、目录 finalizer 和修改时间恢复等高层流程。

如果新增或修改 `App` 的公开方法，需要重新生成 Wails 绑定：

```bash
wails generate module
```

生成文件会写入 `frontend/wailsjs/`。

### 前端

前端使用 Vue 3、TypeScript、Vite、vue-router、Pinia、Element Plus 和 vue-i18n。

重要目录：

- `frontend/src/views/AppShell/` - 主桌面外壳。
- `frontend/src/components/ExplorerTree/` - 连接和目录树。
- `frontend/src/components/Workspace/` - 文件面板、连接表单、设置页和欢迎页。
- `frontend/src/components/Tabs/` - 可拖拽分割的 Tab 工作区。
- `frontend/src/components/SplitPane/` - 分割面板基础组件。
- `frontend/src/components/TaskPanel/` - 传输任务面板。
- `frontend/src/components/NotificationPanel/` - 通知面板。
- `frontend/src/stores/` - workspace、connections、tasks、settings、theme、search、notifications 等 Pinia store。
- `frontend/src/assets/themes/` - 内置主题。
- `frontend/src/locales/` - 中英文文案。

## 存储驱动

| 驱动 | 用途 | 说明 |
|---|---|---|
| `Local` | 本地文件系统 | `rootPath` 会把操作范围限定到某个本地目录。 |
| `SFTP` | SSH File Transfer Protocol | 支持密码、私钥文本、私钥文件路径三种认证输入。 |
| `S3` | Amazon S3 或兼容对象存储 | 使用对象 key prefix 模拟目录。 |
| `OSS` | Alibaba Cloud OSS | 使用对象 key prefix 模拟目录。 |

对象存储没有真正的目录移动原语。S3/OSS 的目录移动和复制是通过列举 prefix、逐个复制对象、再删除源 prefix 来实现的。

S3/OSS 的对象 `LastModified` 由服务端控制，不能像本地文件一样直接覆盖。项目会在可行时把原始修改时间保存到对象 metadata，并在下载或跨连接传输回文件系统类后端时恢复。没有显式目录 marker 的对象存储虚拟目录，只能使用子文件时间做近似推导。

## 配置

默认情况下，应用会在工作目录下查找：

- `config.yaml` - 应用配置。
- `connections.yaml` - 已保存连接配置。
- `state.yaml` - UI/运行时状态。
- `logs/app.log` - 默认日志文件。

应用配置也支持 `config.yml` 和 `config.json`。连接配置支持 YAML 和 JSON。

最小 `connections.yaml` 示例：

```yaml
connections:
  - id: local-documents
    name: Documents
    driver: Local
    enabled: true
    config:
      rootPath: /Users/example/Documents

  - id: sftp-home
    name: SFTP Home
    driver: SFTP
    enabled: true
    config:
      address: 10.0.0.10
      port: 22
      username: example
      privateKeyPath: ~/.ssh/id_rsa
      rootPath: /home/example

  - id: s3-projects
    name: S3 Projects
    driver: S3
    enabled: true
    config:
      region: us-east-1
      bucket: example-bucket
      accessKeyId: YOUR_ACCESS_KEY_ID
      accessKeySecret: YOUR_ACCESS_KEY_SECRET
      prefix: projects/

  - id: oss-assets
    name: OSS Assets
    driver: OSS
    enabled: true
    config:
      region: oss-cn-hangzhou
      bucket: example-bucket
      accessKeyId: YOUR_ACCESS_KEY_ID
      accessKeySecret: YOUR_ACCESS_KEY_SECRET
      endpoint: oss-cn-hangzhou.aliyuncs.com
      prefix: assets/
```

不要提交真实密钥。云厂商凭据是显式连接配置项，正常应用连接不使用环境变量兜底。

## 开发

环境要求：

- Go，版本需满足 `go.mod`。
- Node.js `^20.19.0` 或 `>=22.12.0`。
- Wails v2 CLI。

启动桌面开发模式：

```bash
wails dev
```

构建生产包：

```bash
wails build
```

前端命令需要在 `frontend/` 下执行：

```bash
npm install
npm run dev
npm run build
npm run test:unit
npm run type-check
npm run build-only
```

## 测试

运行全部 Go 测试：

```bash
go test ./...
```

前端检查：

```bash
cd frontend
npm run test:unit
npm run type-check
npm run build-only
```

驱动集成测试没有凭据时会自动跳过。

S3 集成测试环境变量：

```text
S3_ACCESS_KEY_ID
S3_ACCESS_KEY_SECRET
S3_REGION
S3_BUCKET
```

OSS 集成测试环境变量：

```text
OSS_ACCESS_KEY_ID
OSS_ACCESS_KEY_SECRET
OSS_REGION
OSS_BUCKET
```

SFTP 集成测试环境变量：

```text
SFTP_ADDRESS
SFTP_USERNAME
SFTP_PASSWORD 或 SFTP_PRIVATE_KEY 或 SFTP_PRIVATE_KEY_PATH
SFTP_PASSPHRASE
SFTP_PORT
SFTP_ROOT_PATH
SFTP_DIAL_TIMEOUT_SEC
```

## 已知边界

- SFTP 当前使用 `ssh.InsecureIgnoreHostKey()`，代码里已有支持 `known_hosts` 的 TODO。在未完成主机密钥校验前，不应把它视为适合不可信网络的加固实现。
- S3/OSS 目录操作基于 prefix，通常比本地/SFTP 目录操作更重。
- S3 单对象复制存在云厂商限制；超大对象的目录移动后续可能需要 multipart copy。
- 项目仍在持续开发中，`developing.md` 记录了设计决策、执行情况和后续思路。

## 许可证

本项目采用 AGPL-3.0-only，并提供单独的商业授权路径。详见 [LICENSE](LICENSE)。

简要说明：

- 你可以在 AGPL-3.0-only 下使用、修改、分发本项目，或通过网络提供本项目服务。
- 如果你修改了源代码，并且没有按 AGPL-3.0-only 向有权获得源码的接收方或网络用户提供对应修改源码，则必须先从版权持有人处取得单独的一次性商业授权。
- 被项目维护者接受的 bugfix 或功能拓展贡献，排除仅注释、纯文档或其他无功能效果的改动，可按 `LICENSE` 中的定义视为贡献者商业授权的合格贡献。

授权条款以 `LICENSE` 文件为准。
