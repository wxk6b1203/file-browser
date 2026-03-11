# Local File System Driver

本地文件系统驱动，直接操作本机的文件和目录。

## 实现的接口

| Interface              | Status |
|------------------------|--------|
| `folder.Manager`       | ✅      |
| `folder.Reader`        | ✅      |
| `folder.Writer`        | ✅      |
| `folder.HealthChecker` | ✅      |

## 配置项 (`local.Options`)

| Field      | Type   | Required | Description                                          |
|------------|--------|----------|------------------------------------------------------|
| `rootPath` | string |          | 本地根路径，所有操作相对于此路径 (如 `C:\data\files` 或 `/home/user/data`)。为空时表示系统根目录（谨慎使用） |

## 特性

- **原子重命名 / 移动**: `os.Rename` 在同一文件系统上是原子性的
- **递归操作**: 支持递归列举、递归删除、递归复制目录
- **权限保留**: 复制文件时自动保留源文件权限
- **自动创建父目录**: 写入文件时自动确保父目录存在
- **连接健康检查**: `Ping` 通过检查根路径是否存在来验证驱动是否可用
- **路径穿越保护**: 当配置了 `rootPath` 时，所有路径操作都会检查是否越界
- **符号链接检测**: `Stat` 可以识别并报告符号链接类型

## 使用示例

```go
import _ "github.com/wxk6b1203/file-util-manager/folder/local"

mgr, err := folder.CreateInstance(ctx, "Local", "my-local", &folder.DriverOptions{
    Name: "my-local",
    Config: map[string]any{
        "rootPath": "/data/files",
    },
})
```

## 注意事项

- `rootPath` 为空时驱动可以访问整个文件系统，建议始终配置 `rootPath` 限制访问范围
- 跨文件系统的 `Move` 操作可能会失败，此时可以使用 `Copy` + `Delete` 代替
- Windows 和 Unix 路径分隔符均受支持，内部会自动转换

