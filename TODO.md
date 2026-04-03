# TODO

## Preserve File Modification Time During Upload / Download

### Problem

- 当前上传和下载链路不会保留源文件的修改时间。
- 具体表现：
  - 上传本地文件到远端后，目标端修改时间不会保留为源文件时间。
  - 从远端下载到本地后，本地文件时间不会恢复为远端文件时间。
  - 跨连接传输本质上走下载 + 上传，同样不会保留修改时间。

### Current Findings

- `folder.TransferRequest` 目前没有携带修改时间字段。
- `folder.WriteOptions` 目前也没有修改时间字段。
- `fallbackUpload()` 只读取本地文件大小做进度统计，不传递 `ModTime()`。
- `fallbackDownload()` 只创建本地文件并写入内容，没有在完成后调用 `os.Chtimes()`。
- `Local` / `SFTP` / `S3` / `OSS` 的驱动写入实现都没有统一处理“保留源修改时间”。
- `S3` / `OSS` 不能直接改对象系统自己的 `LastModified`，只能通过对象 metadata 保存原始时间。

### Constraints

- `Local` / `SFTP`：
  - 可以做成真正保留远端/本地文件修改时间。
- `S3` / `OSS`：
  - 无法覆盖对象系统 `LastModified`。
  - 只能把原始修改时间写入 metadata，并在下载时恢复到本地文件时间。

### Proposed Plan

1. 扩展传输模型
- 在 `folder/transfer.go` 的 `TransferRequest` 增加：
  - `PreserveModTime bool`
  - `SourceModTime *time.Time`
- 在 `folder/driver.go` 的 `WriteOptions` 增加：
  - `ModTime *time.Time`

2. 上传时传递本地 mtime
- `folder/transfer_manager.go` 的 `fallbackUpload()`：
  - 在 `os.Stat(req.LocalPath)` 后读取 `info.ModTime()`
  - 写入 `TransferRequest` / `WriteOptions`
- `folder/s3/transfer.go`、`folder/alibaba-oss/transfer.go`：
  - 在专用上传链路里同样读取本地 mtime

3. 下载后恢复本地 mtime
- `folder/transfer_manager.go` 的 `fallbackDownload()`：
  - 下载完成后调用 `os.Chtimes(req.LocalPath, now, sourceModTime)`
- `sourceModTime` 来源：
  - `Local` / `SFTP`：直接使用 `Stat().LastModified`
  - `S3` / `OSS`：优先读取 metadata 中保存的原始 mtime，没有则退回对象自己的 `LastModified`

4. 驱动侧支持
- `folder/local/driver.go`
  - 上传写完后，对目标文件执行 `os.Chtimes()`
- `folder/sftp/driver.go`
  - 上传写完后，通过 SFTP 客户端设置远端 mtime
- `folder/s3/driver.go`
  - 上传时把原始 mtime 写入 metadata，例如 `fileutil-mtime`
  - 下载/读取时优先取 metadata
- `folder/alibaba-oss/driver.go`
  - 同 S3，使用 metadata 保存原始 mtime

5. 跨连接传输
- `transfer/service.go`
  - 需要把源端 mtime 串到“下载到临时文件 -> 再上传”的链路里
  - 否则跨连接仍会丢失修改时间

### Recommended Delivery Order

1. 先做最小可用版
- 下载恢复本地 mtime
- `Local` / `SFTP` 上传保留远端 mtime
- `S3` / `OSS` 上传写 metadata，下载恢复

2. 再补测试
- `Local -> Local`
- `Local -> SFTP`
- `Local -> S3`
- `Local -> OSS`
- `S3/OSS -> Local`
- 跨连接目录传输

3. 再决定是否做 UI 一致性增强
- 目前即使做完最小可用版，`S3` / `OSS` 列表显示的仍可能是对象系统 `LastModified`
- 如果需要 UI 也展示“原始修改时间”，则需要额外设计 `List/Stat` 优先读取 metadata 的策略

### Notes

- 推荐先实现“最小可用版”，不要一开始就把 `S3/OSS` 的列表显示语义改复杂。
- 先把实际传输的时间保留下来，再考虑 UI 展示是否要和原始时间完全一致。
