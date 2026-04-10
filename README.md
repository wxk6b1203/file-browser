# File Browser

[中文说明](README_cn.md)

<p align="center">
  <img src="assets/logo.svg" alt="File Browser logo" width="128" height="128">
</p>

File Browser is a Wails v2 desktop file manager for local and remote storage. It combines a Go backend with a Vue 3 frontend and is designed around a VS Code style workspace: connection explorer, split tabs, task panel, notifications, list/icon file views, and drag-and-drop file operations.

The project currently targets these storage backends:

- Local file system
- WebDAV
- SFTP
- Amazon S3 or S3-compatible object storage
- Alibaba Cloud OSS

The long-term goal is a unified file management system where local paths, WebDAV endpoints, SSH servers, and object-storage prefixes can be opened and operated through the same UI model.

## Highlights

- Multi-connection workspace for Local, WebDAV, SFTP, S3, and OSS.
- VS Code style shell with an explorer, central tab groups, draggable split panes, right-side task and notification panels.
- File panel with list view and icon view.
- Sortable and configurable file list columns, column resizing, multi-select, keyboard navigation, and inline create/delete flows.
- Explorer tree that can show connections, directories, and files.
- File operations including open, rename, delete, mkdir, copy, move, upload, download, and save as.
- Same-connection drag move and cross-connection drag transfer.
- External OS file and folder drag-in upload to a workspace panel.
- Transfer task tracking with progress, cancellation, download follow-up actions, default download directory, and open/reveal local path actions.
- File and directory modification-time preservation during transfer where the backend supports it.
- Search, notifications, connection testing, and configurable UI preferences.
- Theme system with built-in themes, including a VS Code inspired `2026 Dark` theme.

## Project Layout

```text
.
├── app.go                 # Wails App methods exposed to the frontend
├── config/                # App config and connections config loading/saving
├── connection/            # Saved connection repository and runtime manager lifecycle
├── fileops/               # High-level file operations used by the UI
├── folder/                # Storage driver abstraction and concrete drivers
│   ├── local/             # Local file system driver
│   ├── webdav/            # WebDAV driver
│   ├── sftp/              # SFTP driver
│   ├── s3/                # S3 driver
│   └── alibaba-oss/       # Alibaba Cloud OSS driver
├── transfer/              # Transfer orchestration above folder.TransferManager
├── search/                # Connection-aware search service
├── shortcut/              # Frontend-to-Go shortcut dispatch bridge
├── logging/               # Zap based logging setup
├── frontend/              # Vue 3 + Vite + Element Plus frontend
├── original.md            # Original product requirements
├── developing.md          # Development notes and design records
└── TODO.md                # Tracked follow-up items
```

## Architecture

### Backend

The backend is written in Go and exposed to the UI through Wails bindings.

- `folder.Manager` defines the unified file operation interface: `List`, `Stat`, `Exist`, `Copy`, `Move`, `Rename`, `Delete`, and `Mkdir`.
- Optional driver interfaces include `Reader`, `Writer`, `Transferer`, `Presigner`, `HealthChecker`, and `Closer`.
- `folder/transfer_manager.go` runs asynchronous upload and download tasks with progress, speed tracking, cancellation, and fallback streaming paths.
- `connection/` stores connection definitions and manages connected runtime instances.
- `fileops/` provides UI-facing operations and keeps backend-specific details out of the frontend.
- `transfer/` handles higher-level workflows such as folder transfer planning, cross-connection transfer, follow-up uploads, directory finalizers, and modification-time restoration.

Public methods on `App` in `app.go` are Wails-bound. After adding or changing an exported method, regenerate bindings:

```bash
wails generate module
```

Generated files are written to `frontend/wailsjs/`.

### Frontend

The frontend uses Vue 3, TypeScript, Vite, vue-router, Pinia, Element Plus, vue-i18n, and custom layout components.

Important areas:

- `frontend/src/views/AppShell/` - main desktop shell.
- `frontend/src/components/ExplorerTree/` - connection and directory explorer.
- `frontend/src/components/Workspace/` - file panels, connection forms, settings, and welcome tab.
- `frontend/src/components/Tabs/` - draggable tab and split-pane workspace system.
- `frontend/src/components/SplitPane/` - resizable pane layout primitives.
- `frontend/src/components/TaskPanel/` - transfer task UI.
- `frontend/src/components/NotificationPanel/` - notification UI.
- `frontend/src/stores/` - Pinia stores for workspace, connections, tasks, settings, theme, search, and notifications.
- `frontend/src/assets/themes/` - built-in themes.
- `frontend/src/locales/` - Chinese and English locale JSON files.

## Storage Backends

| Driver | Purpose | Notes |
|---|---|---|
| `Local` | Local file system access | Root path is mapped to a scoped local directory. |
| `WebDAV` | HTTP(S)-based WebDAV storage | Supports username/password auto auth, bearer token, scoped root path, request timeout, and optional insecure TLS. |
| `SFTP` | SSH File Transfer Protocol | Supports password, private key text, and private key path. |
| `S3` | Amazon S3 or compatible storage | Uses object prefixes as virtual directories. |
| `OSS` | Alibaba Cloud OSS | Uses object prefixes as virtual directories. |

Object stores do not have a native directory move primitive. Directory move/copy is implemented by listing a prefix, copying matching objects to the destination prefix, and deleting the source prefix after copy.

For S3 and OSS, object `LastModified` is controlled by the storage service and cannot be overwritten directly. The app stores original file modification time in object metadata where possible and uses it when downloading or transferring files back to a filesystem backend. Virtual directories without explicit marker objects use best-effort inferred directory times.

## Configuration

When an explicit config path is passed with `--config` / `-c`, that file is always used.

Without an explicit config path:

- When the app is launched from the project source root, the default files live in the working directory. This keeps `wails dev` convenient.
- In production-like Unix launches, including macOS app bundles and Linux binaries launched outside the source root, the default app config is `~/.config/file-browser/config.yaml`.

Default files:

- `config.yaml` - app settings.
- `connections.yaml` - saved connection definitions.
- `state.yaml` - persisted UI/runtime state.
- `logs/app.log` - default log file when file logging is enabled.

Default transfer temporary directory:

- Unix-like systems: `/tmp/file-browser/transfers`
- Windows: `%USERPROFILE%\AppData\Local\Temp\file-browser\transfers`

The app config can also be loaded from `config.yml` or `config.json`. Connection config supports YAML and JSON.

A minimal `connections.yaml` example:

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

  - id: webdav-docs
    name: WebDAV Documents
    driver: WebDAV
    enabled: true
    config:
      endpoint: https://dav.example.com/remote.php/dav/files/example
      username: example
      password: YOUR_PASSWORD
      rootPath: Documents

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

Do not commit real credentials. Cloud credentials are explicit connection config values; this project does not use environment-variable fallback for normal application connections.

## Development

Prerequisites:

- Go, compatible with the module in `go.mod`.
- Node.js `^20.19.0` or `>=22.12.0` for the frontend package.
- Wails v2 CLI.

Run the desktop app in development mode:

```bash
wails dev
```

Build a production package:

```bash
wails build
```

Frontend-only commands from `frontend/`:

```bash
npm install
npm run dev
npm run build
npm run test:unit
npm run type-check
npm run build-only
```

## Testing

Run all Go tests:

```bash
go test ./...
```

Frontend checks:

```bash
cd frontend
npm run test:unit
npm run type-check
npm run build-only
```

Driver integration tests skip automatically when credentials are not configured.

WebDAV driver tests use an in-memory test server and do not require external credentials.

S3 integration test variables:

```text
S3_ACCESS_KEY_ID
S3_ACCESS_KEY_SECRET
S3_REGION
S3_BUCKET
```

OSS integration test variables:

```text
OSS_ACCESS_KEY_ID
OSS_ACCESS_KEY_SECRET
OSS_REGION
OSS_BUCKET
```

SFTP integration test variables:

```text
SFTP_ADDRESS
SFTP_USERNAME
SFTP_PASSWORD or SFTP_PRIVATE_KEY or SFTP_PRIVATE_KEY_PATH
SFTP_PASSPHRASE
SFTP_PORT
SFTP_ROOT_PATH
SFTP_DIAL_TIMEOUT_SEC
```

## Known Boundaries

- SFTP currently uses `ssh.InsecureIgnoreHostKey()` with a TODO to support `known_hosts` verification. Do not treat it as hardened for untrusted networks yet.
- WebDAV can operate without credentials for anonymous endpoints, or with username/password and bearer token authentication depending on the server.
- S3 and OSS directory operations are prefix-based and can be more expensive than local/SFTP directory operations.
- S3 single-object copy has provider limits; very large object directory moves may require multipart copy support in the future.
- The project is still under active development. `developing.md` records design decisions, implementation notes, and follow-up work.

## License

This project is licensed under AGPL-3.0-only with a separate commercial authorization path. See [LICENSE](LICENSE).

In short:

- You may use, modify, distribute, and provide network access to this project under AGPL-3.0-only.
- If you modify the source code and do not make the corresponding modified source code available under AGPL-3.0-only to the recipients or network users who are entitled to receive it, you must obtain a separate one-time commercial authorization from the copyright holder.
- Accepted bugfix or functional enhancement contributions, excluding comment-only or other non-functional edits, are treated as qualifying contributions for the contributor commercial authorization described in `LICENSE`.

The `LICENSE` file is the controlling text for licensing terms.
