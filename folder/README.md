# Folder Driver

Unified file-system abstraction for local and remote backends.

## Core abstractions

- `Manager`: common file operations (`List`, `Stat`, `Copy`, `Move`, `Delete`, `Mkdir`)
- `BaseDriver`: embed in concrete drivers to inherit default (`ErrUnsupported`) implementations; override only what the backend supports
- `FileInfo`: normalized file metadata (`Name`, `Size`, `LastModified`, `Owner`, ...)
- `Capabilities`: capability flags for backend-specific behavior
- `Reader` / `Writer`: optional streaming interfaces
- `HealthChecker` / `Closer`: optional connection lifecycle hooks (`Ping`, `Close`)

## Common driver options

`DriverOptions` provides shared fields for all backends:

- `ID`: stable identifier
- `Name`: instance name
- `Description`: human-readable description
- `Driver`: driver type (`oss`, `s3`, `sftp`, ...)
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
    folder.RegisterDriver[mypackage.Options]("my-driver", mypackage.New)
}

func New(ctx context.Context, opt *folder.DriverOptions, cfg *Options) (folder.Manager, error) {
    return &Driver{
        BaseDriver: folder.NewBaseDriver(opt),
        cfg:        cfg,
    }, nil
}
```

5. Create instances:
   - `folder.NewManager(ctx, "my-driver", options)`
   - `folder.CreateInstance(ctx, "my-driver", "instance-a", options)`
