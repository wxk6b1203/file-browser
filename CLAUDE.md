# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Wails v2 desktop application with Go backend + Vue 3 frontend. The app is a file browser supporting multiple storage backends (S3, Alibaba OSS, SFTP).

## Commands

### Development
```bash
wails dev          # Live-reload dev mode (Go + Vite HMR)
```

### Production Build
```bash
wails build        # Production build → build/bin/
```

### Frontend Only (from `frontend/` directory)
```bash
npm install        # Install dependencies
npm run dev        # Vite dev server
npm run build      # Type-check + production build
npm run test:unit  # Vitest unit tests
```

### Go Tests
```bash
go test ./...      # Run all Go tests (integration tests skip without credentials)
```

## Architecture

### Backend (Go)
- **`app.go`**, **`main.go`**: Wails app entry point. Public methods on `App` struct are auto-bound to frontend.
- **`folder/`**: Core abstraction layer. `Manager` interface defines unified file operations (List, Stat, Copy, Move, Delete, Mkdir).
- **`folder/<backend>/`**: Concrete drivers: `s3`, `alibaba-oss` (registered as `"oss"`), `sftp`.
- **`config/`**: App-level config structs (`AppOptions`, `FolderOptions`).
- **`logging/`**: Zap-based logging with rotating file support. Call `logging.InitLogging()` once.

### Frontend (Vue 3 + TypeScript)
- Vite + Vue 3 Composition API (`<script setup lang="ts">`)
- Element Plus (auto-imported), Tailwind CSS v4
- vue-i18n with locale JSONs in `src/locales/{en,zh}.json` (default: `zh`)
- Pinia for state, vue-router for routing
- Custom components: `SplitPane` and `Tabs` in `src/components/`

### Wails Bindings
After adding/changing a method in `app.go`, regenerate bindings:
```bash
wails generate module
```
Generated TS stubs land in `frontend/wailsjs/go/main/`.

## Key Patterns

### Adding a New Storage Driver

1. Create `folder/<name>/` with `options.go` (typed config struct + `Validate()`) and `driver.go`
2. Embed `folder.BaseDriver` — only override methods the backend supports
3. Register in `init()`:
   ```go
   func init() {
       folder.RegisterDriver[Options]("driver-name", "Human-readable description.", New)
   }
   ```
4. Constructor: `func New(ctx context.Context, opt *folder.DriverOptions, cfg *Options) (folder.Manager, error)`

### Registry & Multi-Instance
`folder/registry.go` manages driver factories and named instances. Use `CreateInstance` / `GetInstance` / `DeleteInstance`.

### Error Handling
Use sentinel errors in `folder/errors.go` (`ErrNotFound`, `ErrAlreadyExist`, `ErrUnsupported`, etc.). Wrap with `%w` for `errors.Is()` compatibility.

## Conventions

- Go module: `github.com/wxk6b1203/file-util-manager`
- Struct tags use both `json` and `yaml`
- Frontend components use barrel exports via `index.ts`
- Component READMEs are in Chinese (zh-CN)
- No env var fallback for cloud credentials — always explicit config

## Testing

Driver tests are integration tests requiring credentials via environment variables:
- S3: `S3_ACCESS_KEY_ID`, `S3_ACCESS_KEY_SECRET`, `S3_REGION`, `S3_BUCKET`
- OSS: `OSS_ACCESS_KEY_ID`, `OSS_ACCESS_KEY_SECRET`, `OSS_REGION`, `OSS_BUCKET`

Tests skip automatically when credentials are unset.
