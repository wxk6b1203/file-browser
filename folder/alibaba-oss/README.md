# Alibaba Cloud OSS Driver

Full-featured driver for [Alibaba Cloud Object Storage Service (OSS)](https://www.alibabacloud.com/product/oss).

## Implemented interfaces

| Interface              | Status |
|------------------------|--------|
| `folder.Manager`       | ✅      |
| `folder.Reader`        | ✅      |
| `folder.Writer`        | ✅      |
| `folder.HealthChecker` | ✅      |
| `folder.Presigner`     | ✅      |
| `folder.Closer`        | ✅      |

## Options (`oss.Options`)

| Field             | Type   | Required | Description                                                                 |
|-------------------|--------|:--------:|-----------------------------------------------------------------------------|
| `region`          | string | ✅        | OSS region ID, e.g. `cn-hangzhou`, `ap-southeast-1`                         |
| `bucket`          | string | ✅        | OSS bucket name                                                             |
| `accessKeyID`     | string | ✅        | Alibaba Cloud AccessKey ID                                                  |
| `accessKeySecret` | string | ✅        | Alibaba Cloud AccessKey Secret                                              |
| `securityToken`   | string |          | STS temporary security token (for RAM role / temporary credentials)         |
| `endpoint`        | string |          | Custom endpoint URL, e.g. `https://oss-cn-hangzhou.aliyuncs.com`           |
| `prefix`          | string |          | Key prefix prepended to all paths, acting as a virtual sub-directory        |
| `forcePathStyle`  | bool   |          | Use path-style addressing `https://endpoint/bucket/key` instead of virtual-hosted |
| `useCName`        | bool   |          | Set to `true` when the endpoint is a CName bound to the bucket              |
| `disableSSL`      | bool   |          | Use HTTP instead of HTTPS                                                   |

> **Note**: Credentials are passed explicitly via `Options` — no environment variable / ECS RAM role fallback.
> If you need RAM role credentials, pass the token obtained from STS via `securityToken`.

## Usage

### Register and create a named instance

```go
import (
    "context"
    "github.com/wxk6b1203/file-util-manager/folder"
    _ "github.com/wxk6b1203/file-util-manager/folder/alibaba-oss" // side-effect: registers "oss" driver
)

mgr, err := folder.CreateInstance(ctx, "oss", "my-oss", &folder.DriverOptions{
    Name: "my-oss",
    Config: map[string]any{
        "region":          "cn-hangzhou",
        "bucket":          "my-bucket",
        "accessKeyID":     "LTAI...",
        "accessKeySecret": "secret",
    },
})
```

### Construct directly (without the registry)

```go
import (
    "context"
    alioss "github.com/wxk6b1203/file-util-manager/folder/alibaba-oss"
    "github.com/wxk6b1203/file-util-manager/folder"
)

mgr, err := alioss.New(ctx, &folder.DriverOptions{Name: "my-oss"}, &alioss.Options{
    Region:          "cn-hangzhou",
    Bucket:          "my-bucket",
    AccessKeyID:     "LTAI...",
    AccessKeySecret: "secret",
})
```

### Custom endpoint (e.g. internal VPC endpoint)

```go
&alioss.Options{
    Region:          "cn-hangzhou",
    Bucket:          "my-bucket",
    AccessKeyID:     "LTAI...",
    AccessKeySecret: "secret",
    Endpoint:        "https://oss-cn-hangzhou-internal.aliyuncs.com",
}
```

### YAML configuration

```yaml
driver: oss
name: my-oss
config:
  region: cn-hangzhou
  bucket: my-bucket
  accessKeyID: LTAI...
  accessKeySecret: secret
  prefix: uploads/          # optional: scope all operations to this prefix
  endpoint: ""              # optional: override endpoint
  forcePathStyle: false
  useCName: false
  disableSSL: false
```

## Operations

| Operation                 | Method                     | Notes                                                    |
|---------------------------|----------------------------|----------------------------------------------------------|
| List directory            | `List(ctx, dir, opt)`      | Supports delimiter-based (non-recursive) and recursive listing; pagination handled internally |
| Get file metadata         | `Stat(ctx, path)`          | Falls back to virtual-directory probe when key is not found |
| Check existence           | `Exist(ctx, path)`         | Implemented via `Stat`                                   |
| Download file             | `Read(ctx, path)`          | Returns `io.ReadCloser`; caller must close               |
| Upload file               | `Write(ctx, path, body, opt)` | Supports `ContentType` and custom metadata            |
| Delete file or directory  | `Delete(ctx, path)`        | Paths ending with `/` trigger recursive batch deletion (1 000 objects/batch) |
| Copy object               | `Copy(ctx, PathOp)`        | Server-side copy via OSS `CopyObject`                    |
| Move object               | `Move(ctx, PathOp)`        | Copy + Delete (non-atomic)                               |
| Rename object             | `Rename(ctx, path, name)`  | Copy to new name in same directory + Delete source       |
| Create virtual directory  | `Mkdir(ctx, dir)`          | Uploads a zero-byte key with a trailing `/`              |
| Health check              | `Ping(ctx)`                | Issues `HeadObject(".ping")`; 404 is treated as reachable |
| Pre-signed download URL   | `PresignRead(ctx, path, opt)`  | Default expiry: 15 min                               |
| Pre-signed upload URL     | `PresignWrite(ctx, path, opt)` | Default expiry: 15 min                               |
| Close / release resources | `Close()`                  | Sets internal client to `nil`                            |

## Running the integration tests

All parameters are supplied via environment variables. Required variables must be set for the tests to run; missing required variables cause the test to be **skipped** (not failed).

| Variable              | Required | Description                                    |
|-----------------------|:--------:|------------------------------------------------|
| `OSS_ACCESS_KEY_ID`   | ✅        | AccessKey ID                                   |
| `OSS_ACCESS_KEY_SECRET` | ✅      | AccessKey Secret                               |
| `OSS_REGION`          | ✅        | Region ID, e.g. `cn-hangzhou`                  |
| `OSS_BUCKET`          | ✅        | Bucket name                                    |
| `OSS_ENDPOINT`        |          | Custom endpoint URL                            |
| `OSS_PREFIX`          |          | Key prefix for all test objects                |
| `OSS_SECURITY_TOKEN`  |          | STS temporary token                            |
| `OSS_FORCE_PATH_STYLE`|          | `"true"` to enable path-style addressing       |
| `OSS_USE_CNAME`       |          | `"true"` if endpoint is a CName               |
| `OSS_DISABLE_SSL`     |          | `"true"` to use HTTP                           |

```bash
export OSS_ACCESS_KEY_ID=LTAI...
export OSS_ACCESS_KEY_SECRET=secret
export OSS_REGION=cn-hangzhou
export OSS_BUCKET=my-test-bucket

go test ./folder/alibaba-oss/... -v -run TestRunOSSDriver
```

The test suite covers the full object lifecycle:
`Ping → Capabilities → Mkdir → Write → Stat → Exist → ExistNonExistent → List → Read → Copy → Move → Rename → PresignRead → PresignWrite → Delete`

Each run writes objects under a timestamped path (`test-runs/TestRunOSSDriver-YYYYMMDD-HHMMSS/`) to avoid collisions between concurrent runs.

## Known limitations

- **Move / Rename are non-atomic**: both operations are implemented as Copy + Delete. A crash between the two steps may leave both source and destination objects alive.
- **No multipart upload**: `Write` uses a single `PutObject` call. For objects larger than 5 GB, use the OSS SDK's multipart upload API directly.
- **No bucket-level operations**: creating or deleting buckets is outside the scope of this driver.
- **Close is advisory**: after `Close()` is called the internal client is set to `nil`. Calling any method on a closed driver will panic. Callers are responsible for not using a driver after closing it.
