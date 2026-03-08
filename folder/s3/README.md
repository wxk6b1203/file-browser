# AWS S3 File Folder Driver

Full-featured driver for Amazon S3 and S3-compatible backends (MinIO, Cloudflare R2, etc.).

## Implemented interfaces

| Interface              | Status |
|------------------------|--------|
| `folder.Manager`       | ✅      |
| `folder.Reader`        | ✅      |
| `folder.Writer`        | ✅      |
| `folder.HealthChecker` | ✅      |
| `folder.Closer`        | ✅      |

## Options (`s3.Options`)

| Field             | Type   | Required | Description                                              |
|-------------------|--------|----------|----------------------------------------------------------|
| `region`          | string | ✅        | AWS region (e.g. `us-east-1`)                            |
| `bucket`          | string | ✅        | S3 bucket name                                           |
| `accessKeyID`     | string | ✅        | AWS access key ID                                        |
| `accessKeySecret` | string | ✅        | AWS secret access key                                    |
| `sessionToken`    | string |          | STS session token (for temporary credentials)            |
| `endpoint`        | string |          | Custom endpoint URL (for MinIO, R2, etc.)                |
| `forcePathStyle`  | bool   |          | Use path-style addressing (required by some S3-compat)   |
| `disableSSL`      | bool   |          | Use HTTP instead of HTTPS                                |
| `prefix`          | string |          | Key prefix prepended to all paths (virtual sub-directory)|

> **Note**: Credentials are passed explicitly via Options — no environment variable / AWS profile fallback.

## Usage

```go
import _ "github.com/wxk6b1203/file-util-manager/folder/s3"

mgr, err := folder.CreateInstance(ctx, "s3", "my-bucket", &folder.DriverOptions{
    Name: "my-bucket",
    Config: map[string]any{
        "region":          "us-east-1",
        "bucket":          "my-bucket",
        "accessKeyID":     "AKIA...",
        "accessKeySecret": "secret",
    },
})
```
