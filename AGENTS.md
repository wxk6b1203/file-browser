# AGENTS.md

## Architecture

Wails v2 desktop app: Go backend + Vue 3 frontend. The Go side exposes methods to the frontend via Wails bindings (`app.go` → `frontend/wailsjs/go/main/`). The frontend is embedded at build time via `//go:embed all:frontend/dist`.

### Backend (Go)

- **`folder/`** — Core abstraction layer. `Manager` interface (`driver.go`) defines unified file operations (List, Stat, Copy, Move, Delete, Mkdir). Optional interfaces: `Reader`, `Writer`, `Presigner`, `HealthChecker`, `Closer`.
- **`folder/<backend>/`** — Concrete drivers: `s3`, `alibaba-oss` (registered as `"oss"`), `sftp`. Each lives in its own sub-package.
- **`config/`** — App-level config structs (`AppOptions`, `FolderOptions`).
- **`logging/`** — Zap-based logging with rotating file support via `timberjack`. Call `logging.InitLogging()` once. Uses `zap.S()` / `zap.L()` globally.

### Frontend (Vue 3 + TypeScript)

- Vite + Vue 3 Composition API (`<script setup lang="ts">`)
- Element Plus (auto-imported via `unplugin-vue-components`), Tailwind CSS v4
- vue-i18n with locale JSONs in `src/locales/{en,zh}.json` (default locale: `zh`)
- Pinia for state, vue-router for routing
- Custom components: `SplitPane` and `Tabs` (drag-to-split tab system) in `src/components/`

## Key Patterns

### Adding a new storage driver

Every driver follows the same pattern (see `folder/sftp/driver.go` as the simplest example):

1. Create `folder/<name>/` with `options.go` (typed config struct + `Validate()`) and `driver.go`
2. Embed `folder.BaseDriver` — only override methods the backend supports
3. Register in `init()` using the generic helper:
   ```go
   func init() {
       folder.RegisterDriver[Options]("driver-name", "Human-readable description of the driver.", New)
   }
   ```
4. Constructor signature: `func New(ctx context.Context, opt *folder.DriverOptions, cfg *Options) (folder.Manager, error)`
5. Backend-specific config goes in `DriverOptions.Config` map and is decoded via `folder.DecodeConfig` (JSON round-trip into the typed struct)

### Registry & multi-instance

`folder/registry.go` manages driver factories and named instances. Use `CreateInstance` / `GetInstance` / `DeleteInstance` for lifecycle. Driver types are global singletons; instances are per-name within a driver type.

### Error sentinels

Use the sentinel errors in `folder/errors.go` (`ErrNotFound`, `ErrAlreadyExist`, `ErrUnsupported`, `ErrReadOnly`, `ErrInvalidPath`). Wrap with `%w` for `errors.Is()` compatibility. Helper: `folder.IsNotFound(err)`.

### Wails bindings

Public methods on `App` struct in `app.go` are auto-bound. After adding/changing a method, regenerate bindings with `wails generate module`. Generated TS stubs land in `frontend/wailsjs/go/main/`.

## Build & Dev Commands

```bash
wails dev          # live-reload dev mode (Go + Vite HMR)
wails build        # production build → build/bin/
```

Frontend only (from `frontend/`):
```bash
npm install        # install deps
npm run dev        # Vite dev server
npm run build      # type-check + production build
npm run test:unit  # vitest
```

## Testing

Driver tests (`folder/s3/driver_test.go`, `folder/alibaba-oss/driver_test.go`) are **integration tests** that require real credentials via environment variables and skip automatically when unset:
- S3: `S3_ACCESS_KEY_ID`, `S3_ACCESS_KEY_SECRET`, `S3_REGION`, `S3_BUCKET`
- OSS: `OSS_ACCESS_KEY_ID`, `OSS_ACCESS_KEY_SECRET`, `OSS_REGION`, `OSS_BUCKET`

Test helpers: `envOrSkip(t, key)` skips gracefully; `testDir(t)` generates collision-free paths.

Run all Go tests: `go test ./...` (integration tests skip without credentials).

## Conventions

- Go module path: `github.com/wxk6b1203/file-util-manager`
- All struct tags use both `json` and `yaml` (dual serialization support)
- Frontend components use barrel exports via `index.ts` files
- Component READMEs are written in Chinese (zh-CN)
- No environment variable fallback for cloud credentials — always explicit config

