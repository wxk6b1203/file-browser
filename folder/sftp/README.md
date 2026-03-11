# SFTP File Folder Driver

基于 SSH File Transfer Protocol (SFTP) 的远程文件系统驱动，通过 SSH 连接提供安全的远程文件管理能力。

## 实现的接口

| Interface              | Status |
|------------------------|--------|
| `folder.Manager`       | ✅      |
| `folder.Reader`        | ✅      |
| `folder.Writer`        | ✅      |
| `folder.HealthChecker` | ✅      |
| `folder.Closer`        | ✅      |

## 配置项 (`sftp.Options`)

| Field        | Type   | Required | Description                                          |
|--------------|--------|----------|------------------------------------------------------|
| `address`    | string | ✅        | SFTP 服务器地址 (hostname 或 IP)                      |
| `port`       | int    |          | SSH 端口，默认 `22`                                   |
| `username`   | string | ✅        | SSH 用户名                                           |
| `password`   | string | ①        | 密码认证                                              |
| `privateKey` | string | ①        | PEM 格式的私钥内容                                    |
| `passphrase` | string |          | 私钥的加密口令（仅当 privateKey 加密时需要）           |
| `rootPath`   | string |          | 远程根路径，所有操作相对于此路径 (如 `/home/user/data`)|

> ① `password` 和 `privateKey` 至少需要提供一个。两者同时存在时，先尝试私钥认证，再回退到密码认证。

## 特性

- **原子重命名 / 移动**: SFTP 的 `Rename` 操作在同一文件系统上是原子性的
- **递归操作**: 支持递归列举、递归删除、递归复制目录
- **权限保留**: 复制文件时自动保留源文件权限
- **自动创建父目录**: 写入文件时自动确保父目录存在
- **连接健康检查**: `Ping` 通过获取工作目录来验证连接是否存活

## 使用示例

```go
import _ "github.com/wxk6b1203/file-util-manager/folder/sftp"

mgr, err := folder.CreateInstance(ctx, "SFTP", "my-server", &folder.DriverOptions{
    Name: "my-server",
    Config: map[string]any{
        "address":  "192.168.1.100",
        "port":     22,
        "username": "deploy",
        "password": "secret",
        "rootPath": "/data/files",
    },
})
```
