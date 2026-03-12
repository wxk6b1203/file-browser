本文介绍如何通过简单下载方法将存储空间（Bucket）中的文件（Object）下载到本地，此方法操作简便，适合快速将云端存储的文件下载到本地。

## **注意事项**

-   本文示例代码以华东1（杭州）的地域ID`cn-hangzhou`为例，默认使用外网Endpoint，如果您希望通过与OSS同地域的其他阿里云产品访问OSS，请使用内网Endpoint。关于OSS支持的Region与Endpoint的对应关系，请参见[OSS地域和访问域名](https://help.aliyun.com/zh/oss/user-guide/regions-and-endpoints#concept-zt4-cvy-5db)。

-   本文以从环境变量读取访问凭证为例。如何配置访问凭证，请参见[配置访问凭证](https://help.aliyun.com/zh/oss/developer-reference/configure-access-credentials-by-using-oss-sdk-for-go-v2)。


## **权限说明**

阿里云账号默认拥有全部权限。阿里云账号下的RAM用户或RAM角色默认没有任何权限，需要阿里云账号或账号管理员通过[RAM Policy](https://help.aliyun.com/zh/oss/ram-policy-overview/)或[Bucket Policy](https://help.aliyun.com/zh/oss/user-guide/oss-bucket-policy/)授予操作权限。

| **API** | **Action** | **说明** |
| --- | --- | --- |
| GetObject | `oss:GetObject` | 下载Object。 |
| `oss:GetObjectVersion` | 下载Object时，如果通过versionId指定了Object的版本，则需要授予此操作的权限。 |
| `kms:Decrypt` | 下载Object时，如果Object的元数据包含X-Oss-Server-Side-Encryption: KMS，则需要此操作的权限。 |

## **方法定义**

```
func (c *Client) GetObject(ctx context.Context, request *GetObjectRequest, optFns ...func(*Options)) (*GetObjectResult, error)
```

#### **请求参数列表**

| 参数名 | 类型  | 说明  |
| --- | --- | --- |
| ctx | context.Context | 请求的上下文，可以用来设置请求的总时限 |
| request | \\*GetObjectRequest | 设置具体接口的请求参数，具体请参见[GetObjectRequest](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#GetObjectRequest) |
| optFns | ...func(\\*Options) | （可选）接口级的配置参数, 具体请参见[Options](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Options) |

#### **返回值列表**

| 返回值名 | 类型  | 说明  |
| --- | --- | --- |
| result | \\*GetObjectResult | 接口返回值，当 err 为nil 时有效，具体请参见[GetObjectResult](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#GetObjectResult) |
| err | error | 请求的状态，当请求失败时，err 不为 nil |

## **示例代码**

您可以使用以下代码将存储空间中的文件下载到本地。

```
package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// 定义全局变量
var (
	region     string // 存储区域
	bucketName string // 存储空间名称
	objectName string // 对象名称
)

// init函数用于初始化命令行参数
func init() {
	flag.StringVar(&region, "region", "", "The region in which the bucket is located.")
	flag.StringVar(&bucketName, "bucket", "", "The name of the bucket.")
	flag.StringVar(&objectName, "object", "", "The name of the object.")
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 检查bucket名称是否为空
	if len(bucketName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, bucket name required")
	}

	// 检查region是否为空
	if len(region) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, region required")
	}

	// 检查object名称是否为空
	if len(objectName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, object name required")
	}

	// 定义输出文件路径
	outputFile := "downloaded.file" // 替换为你希望保存的文件路径

	// 加载默认配置并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建获取对象的请求
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
	}

	// 执行获取对象的操作并处理结果
	result, err := client.GetObject(context.TODO(), request)
	if err != nil {
		log.Fatalf("failed to get object %v", err)
	}
	defer result.Body.Close() // 确保在函数结束时关闭响应体

	// 一次性读取整个文件内容
	data, err := io.ReadAll(result.Body)
	if err != nil {
		log.Fatalf("failed to read object %v", err)
	}

	// 将内容写入到文件
	err = os.WriteFile(outputFile, data, 0644)
	if err != nil {
		log.Fatalf("failed to write to output file %v", err)
	}

	log.Printf("file downloaded successfully to %s", outputFile)
}
```

## **常见使用场景**

### **根据限定条件下载**

当从Bucket中下载单个文件（Object）时，您可以指定基于文件最后修改时间或ETag（文件内容标识符）的条件限制。只有当这些条件得到满足时才会执行下载操作；如果不满足，则会返回错误并且不会触发下载。利用限定条件下载不仅可以减少不必要的网络传输和资源消耗，还能提高下载效率。

OSS支持的限定条件如下：

**说明**

-   If-Modified-Since和If-Unmodified-Since可以同时存在。If-Match和If-None-Match也可以同时存在。

-   您可以通过ossClient.getObjectMeta方法获取ETag。


| **参数** | **描述** |
| --- | --- |
| IfModifiedSince | 如果指定的时间早于实际修改时间，则正常传输文件，否则返回错误（304 Not modified）。 |
| IfUnmodifiedSince | 如果指定的时间等于或者晚于文件实际修改时间，则正常传输文件，否则返回错误（412 Precondition failed）。 |
| IfMatch | 如果指定的ETag和OSS文件的ETag匹配，则正常传输文件，否则返回错误（412 Precondition failed）。 |
| IfNoneMatch | 如果指定的ETag和OSS文件的ETag不匹配，则正常传输文件，否则返回错误（304 Not modified）。 |

以下示例代码展示了如何使用限定条件下载。

```
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// 定义全局变量
var (
	region     string // 存储区域
	bucketName string // 存储空间名称
	objectName string // 对象名称
)

// init函数用于初始化命令行参数
func init() {
	flag.StringVar(&region, "region", "", "The region in which the bucket is located.")
	flag.StringVar(&bucketName, "bucket", "", "The name of the bucket.")
	flag.StringVar(&objectName, "object", "", "The name of the object.")
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 检查bucket名称是否为空
	if len(bucketName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, bucket name required")
	}

	// 检查region是否为空
	if len(region) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, region required")
	}

	// 检查object名称是否为空
	if len(objectName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, object name required")
	}

	// 加载默认配置并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 指定本地文件路径
	localFile := "download.file"

	// 假设Object最后修改时间为2024年10月21日18:43:02，则填写的UTC早于该时间时，将满足IfModifiedSince的限定条件，并触发下载行为。
	date := time.Date(2024, time.October, 21, 18, 43, 2, 0, time.UTC)

	// 假设ETag为e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855，则填写的ETag与Object的ETag值相等时，将满足IfMatch的限定条件，并触发下载行为。
	etag := "\"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855\""

	// 创建下载对象到本地文件的请求
	getRequest := &oss.GetObjectRequest{
		Bucket:          oss.Ptr(bucketName),                   // 存储空间名称
		Key:             oss.Ptr(objectName),                   // 对象名称
		IfModifiedSince: oss.Ptr(date.Format(http.TimeFormat)), // 指定IfModifiedSince参数
		IfMatch:         oss.Ptr(etag),                         // 指定IfMatch参数
	}

	// 执行下载对象到本地文件的操作并处理结果
	result, err := client.GetObjectToFile(context.TODO(), getRequest, localFile)
	if err != nil {
		log.Fatalf("failed to get object to file %v", err)
	}

	log.Printf("get object to file result:%#v\n", result)
}
```

### **打印下载文件的进度条**

当您在下载文件时，可以使用进度条实时了解下载进度，避免因为等待时间过长而感到不安或怀疑任务是否卡住。

以下示例代码展示了如何使用进度条查看下载文件的进度。

```
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// 定义全局变量
var (
	region     string // 存储区域
	bucketName string // 存储空间名称
	objectName string // 对象名称
)

// init函数用于初始化命令行参数
func init() {
	flag.StringVar(&region, "region", "", "The region in which the bucket is located.")
	flag.StringVar(&bucketName, "bucket", "", "The name of the bucket.")
	flag.StringVar(&objectName, "object", "", "The name of the object.")
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 检查bucket名称是否为空
	if len(bucketName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, bucket name required")
	}

	// 检查region是否为空
	if len(region) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, region required")
	}

	// 检查object名称是否为空
	if len(objectName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, object name required")
	}

	// 加载默认配置并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 指定本地文件路径
	localFile := "download.file"

	// 创建下载对象到本地文件的请求
	getRequest := &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
		ProgressFn: func(increment, transferred, total int64) {
			fmt.Printf("increment:%v, transferred:%v, total:%v\n", increment, transferred, total)
		}, //进度回调函数,显示下载进度
	}

	// 执行下载对象到本地文件的操作并处理结果
	result, err := client.GetObjectToFile(context.TODO(), getRequest, localFile)
	if err != nil {
		log.Fatalf("failed to get object to file %v", err)
	}

	log.Printf("get object to file result:%#v\n", result)
}
```

### **批量下载文件到本地**

```
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// 定义全局变量
var (
	region     string // 存储区域
	bucketName string // 存储空间名称
	prefix     string // 对象前缀（文件夹路径）
	localDir   string // 本地下载目录
	maxWorkers int    // 最大并发数
	maxKeys    int    // 每次列举的最大对象数
)

// DownloadTask 下载任务结构
type DownloadTask struct {
	ObjectKey string
	LocalPath string
	Size      int64
}

// DownloadResult 下载结果结构
type DownloadResult struct {
	ObjectKey string
	Success   bool
	Error     error
	Size      int64
}

// init函数用于初始化命令行参数
func init() {
	flag.StringVar(&region, "region", "", "The region in which the bucket is located.")
	flag.StringVar(&bucketName, "bucket", "", "The name of the bucket.")
	flag.StringVar(&prefix, "prefix", "", "The prefix (folder path) to download.")
	flag.StringVar(&localDir, "local-dir", "./downloads", "Local directory to save downloaded files.")
	flag.IntVar(&maxWorkers, "workers", 5, "Maximum number of concurrent downloads.")
	flag.IntVar(&maxKeys, "max-keys", 1000, "Maximum number of objects to list at once.")
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 检查必要参数
	if len(bucketName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, bucket name required")
	}

	if len(region) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, region required")
	}

	// 确保前缀以/结尾（如果不是空字符串）
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	// 创建本地下载目录
	if err := os.MkdirAll(localDir, 0755); err != nil {
		log.Fatalf("failed to create local directory: %v", err)
	}

	// 加载默认配置并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	fmt.Printf("开始批量下载，存储空间: %s, 前缀: %s, 本地目录: %s\n", bucketName, prefix, localDir)

	// 列举所有需要下载的对象
	tasks, err := listObjects(client, bucketName, prefix)
	if err != nil {
		log.Fatalf("failed to list objects: %v", err)
	}

	if len(tasks) == 0 {
		fmt.Println("没有找到需要下载的文件")
		return
	}

	fmt.Printf("找到 %d 个文件需要下载\n", len(tasks))

	// 执行批量下载
	results := batchDownload(client, tasks, maxWorkers)

	// 统计下载结果
	var successCount, failCount int
	var totalSize int64
	for _, result := range results {
		if result.Success {
			successCount++
			totalSize += result.Size
		} else {
			failCount++
			fmt.Printf("下载失败: %s, 错误: %v\n", result.ObjectKey, result.Error)
		}
	}

	fmt.Printf("\n下载完成! 成功: %d, 失败: %d, 总大小: %s\n",
		successCount, failCount, formatBytes(totalSize))
}

// listObjects 列举存储空间中指定前缀的所有对象
func listObjects(client *oss.Client, bucketName, prefix string) ([]DownloadTask, error) {
	var tasks []DownloadTask
	var continuationToken *string

	for {
		// 创建列举对象请求
		request := &oss.ListObjectsV2Request{
			Bucket:            oss.Ptr(bucketName),
			Prefix:            oss.Ptr(prefix),
			MaxKeys:           int32(maxKeys),
			ContinuationToken: continuationToken,
		}

		// 执行列举操作
		result, err := client.ListObjectsV2(context.TODO(), request)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		// 处理列举结果
		for _, obj := range result.Contents {
			// 跳过文件夹对象（以/结尾且大小为0）
			if strings.HasSuffix(*obj.Key, "/") && obj.Size == 0 {
				continue
			}

			// 计算本地文件路径
			relativePath := strings.TrimPrefix(*obj.Key, prefix)
			localPath := filepath.Join(localDir, relativePath)

			tasks = append(tasks, DownloadTask{
				ObjectKey: *obj.Key,
				LocalPath: localPath,
				Size:      obj.Size,
			})
		}

		// 检查是否还有更多对象
		if result.NextContinuationToken == nil {
			break
		}
		continuationToken = result.NextContinuationToken
	}

	return tasks, nil
}

// batchDownload 执行批量下载
func batchDownload(client *oss.Client, tasks []DownloadTask, maxWorkers int) []DownloadResult {
	taskChan := make(chan DownloadTask, len(tasks))
	resultChan := make(chan DownloadResult, len(tasks))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go downloadWorker(client, bucketName, taskChan, resultChan, &wg)
	}

	// 发送下载任务
	go func() {
		for _, task := range tasks {
			taskChan <- task
		}
		close(taskChan)
	}()

	// 等待所有工作协程完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果并显示进度
	var results []DownloadResult
	completed := 0
	total := len(tasks)

	for result := range resultChan {
		results = append(results, result)
		completed++

		if result.Success {
			fmt.Printf("✓ [%d/%d] %s (%s)\n",
				completed, total, result.ObjectKey, formatBytes(result.Size))
		} else {
			fmt.Printf("✗ [%d/%d] %s - 错误: %v\n",
				completed, total, result.ObjectKey, result.Error)
		}
	}

	return results
}

// downloadWorker 下载工作协程
func downloadWorker(client *oss.Client, bucketName string, taskChan <-chan DownloadTask,
	resultChan chan<- DownloadResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range taskChan {
		result := DownloadResult{
			ObjectKey: task.ObjectKey,
			Size:      task.Size,
		}

		// 创建本地文件目录
		if err := os.MkdirAll(filepath.Dir(task.LocalPath), 0755); err != nil {
			result.Error = fmt.Errorf("failed to create directory: %w", err)
			resultChan <- result
			continue
		}

		// 检查文件是否已存在且大小一致
		if fileInfo, err := os.Stat(task.LocalPath); err == nil {
			if fileInfo.Size() == task.Size {
				result.Success = true
				resultChan <- result
				continue
			}
		}

		// 创建下载请求
		getRequest := &oss.GetObjectRequest{
			Bucket: oss.Ptr(bucketName),
			Key:    oss.Ptr(task.ObjectKey),
		}

		// 执行下载
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		_, err := client.GetObjectToFile(ctx, getRequest, task.LocalPath)
		cancel()

		if err != nil {
			result.Error = err
		} else {
			result.Success = true
		}

		resultChan <- result
	}
}

// formatBytes 格式化字节数为可读格式
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
```

#### **使用方法**

```
# 编译测试代码为可执行文件
go build -o oss-batch-download main.go

# 下载指定文件夹
./oss-batch-download -region oss-cn-hangzhou -bucket my-bucket -prefix images/2024/

# 自定义本地目录和并发数
./oss-batch-download -region oss-cn-hangzhou -bucket my-bucket -prefix documents/ -local-dir ./downloads -workers 10

# 下载整个存储空间
./oss-batch-download -region oss-cn-hangzhou -bucket my-bucket -prefix ""
```

#### **输出示例**

程序运行时会显示详细的下载进度：

```
开始批量下载，存储空间: my-bucket, 前缀: images/2024/, 本地目录: ./downloads
找到 150 个文件需要下载
✓ [1/150] images/2024/photo1.jpg (2.3 MB)
✓ [2/150] images/2024/photo2.png (1.8 MB)
...
下载完成! 成功: 148, 失败: 2, 总大小: 1.2 GB
```

## 相关文档

-   关于下载到本地文件的完整示例代码，请参见[GitHub示例](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2/blob/master/sample/get_object_to_file.go)。

-   关于下载到本地文件的API接口说明，请参见[GetObjectToFile](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Client.GetObjectToFile)和[GetObject](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Client.GetObject)。