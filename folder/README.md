# Folder Driver

Unified file-system abstraction for local and remote backends.

## Core abstractions

- `Manager`: common file operations (`List`, `Stat`, `Copy`, `Move`, `Delete`, `Mkdir`)
- `BaseDriver`: embed in concrete drivers to inherit default (`ErrUnsupported`) implementations; override only what the backend supports
- `FileInfo`: normalized file metadata (`Name`, `Size`, `LastModified`, `Owner`, ...)
- `Capabilities`: capability flags for backend-specific behavior
- `Reader` / `Writer`: optional streaming interfaces
- `Transferer`: optional interface for optimized (multipart) upload/download with progress tracking
- `TransferManager`: coordinates async transfers across all drivers with progress, cancellation, and speed tracking
- `HealthChecker` / `Closer`: optional connection lifecycle hooks (`Ping`, `Close`)

## Common driver options

`DriverOptions` provides shared fields for all backends:

- `ID`: stable identifier
- `Name`: instance name
- `Description`: human-readable description
- `Driver`: driver type (`oss`, `s3`, `sftp`, `local`, ...)
- `Root`: logical root path/prefix
- `Enabled`: enable/disable flag
- `ReadOnly`: forbid mutating operations in driver implementation
- `TimeoutSec`: default operation timeout (optional)
- `Tags`: tags for grouping/filtering
- `Metadata`: custom key-value metadata
- `Config`: backend-specific raw options

## Multi-instance support

- `CreateInstance(ctx, driver, instance, options)`: create and register an instance
- `GetInstance(driver, instance)`: get an existing instance
- `ListInstances(driver)`: list all instance names for a driver type
- `DeleteInstance(driver, instance)`: remove and close one instance

## Built-in driver names

- `local` — Local file system
- `oss` — Alibaba Cloud OSS
- `s3` — Amazon S3
- `sftp` — SFTP

## Add a custom backend

1. Embed `folder.BaseDriver` in your struct
2. Override only the `Manager` methods your backend supports
3. (Optional) implement `folder.Reader`, `folder.Writer`, `folder.HealthChecker`, `folder.Closer`
4. Register in package `init()` with the generic helper:

```go
func init() {
    folder.RegisterDriver[mypackage.Options]("my-driver", "A brief description of what this driver does.", mypackage.New)
}
```

5. Create instances:
   - `folder.NewManager(ctx, "my-driver", options)`
   - `folder.CreateInstance(ctx, "my-driver", "instance-a", options)`

## Async transfer (upload / download) with progress

The `TransferManager` manages async file transfers with progress tracking, speed computation, and cancellation support.

### Quick start

```go
tm := folder.NewTransferManager()

// Submit an upload task — returns immediately with a task ID.
taskID, err := tm.Submit(mgr, "S3", "my-bucket", folder.TransferUpload, &folder.TransferRequest{
    RemotePath: "data/report.csv",
    LocalPath:  "/tmp/report.csv",
    PartSize:   10 * 1024 * 1024, // 10 MiB parts (0 = backend default)
    Concurrency: 5,               // parallel parts (0 = backend default)
})

// Query progress at any time.
task := tm.Progress(taskID)
fmt.Printf("%.1f%% — %d B/s\n",
    float64(task.BytesTransferred)/float64(task.TotalBytes)*100,
    task.BytesPerSecond)

// Cancel a running task.
tm.Cancel(taskID)

// List all tasks.
tasks := tm.List()

// Clean up finished tasks.
tm.RemoveAll()
```

### How driver dispatch works

| Driver implements `Transferer`? | Upload | Download |
|---|---|---|
| **Yes** (S3, OSS) | Multipart upload via SDK manager | S3: parallel range download; OSS: single-stream + progress |
| **No** (SFTP, Local) | Single-stream via `Writer.Write` + progress reader | Single-stream via `Reader.Read` + progress writer |

### Adding `Transferer` to a custom driver

Implement two methods on your driver struct:

```go
func (d *Driver) Upload(ctx context.Context, req *folder.TransferRequest, fn folder.ProgressFunc) error { ... }
func (d *Driver) Download(ctx context.Context, req *folder.TransferRequest, fn folder.ProgressFunc) error { ... }
```

Set `CanTransfer = true` in `Capabilities()` and add a compile-time check:

```go
var _ folder.Transferer = (*Driver)(nil)
```

