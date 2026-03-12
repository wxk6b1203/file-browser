本文针对文件的传输场景，介绍如何使用Go SDK V2新增的下载管理器Downloader模块进行文件下载。

## **注意事项**

-   本文示例代码以华东1（杭州）的地域ID`cn-hangzhou`为例，默认使用外网Endpoint，如果您希望通过与OSS同地域的其他阿里云产品访问OSS，请使用内网Endpoint。关于OSS支持的Region与Endpoint的对应关系，请参见[OSS地域和访问域名](https://help.aliyun.com/zh/oss/user-guide/regions-and-endpoints#concept-zt4-cvy-5db)。

-   本文以从环境变量读取访问凭证为例。如何配置访问凭证，请参见[配置访问凭证](https://help.aliyun.com/zh/oss/developer-reference/configure-access-credentials-by-using-oss-sdk-for-go-v2)。

-   要进行文件下载，您必须有`oss:GetObject`权限。具体操作，请参见[为RAM用户授予自定义的权限策略](https://help.aliyun.com/zh/oss/user-guide/common-examples-of-ram-policies#section-ucu-jv0-zip)。


## **方法定义**

#### **下载管理器功能简介**

Go SDK V2新增下载管理器Downloader提供了通用的下载接口，隐藏了底层接口的实现细节，提供便捷的文件下载能力。

-   下载管理器Downloader底层利用范围下载，把文件自动分成多个较小的分片进行并发下载，提升下载的性能。

-   下载管理器Downloader同时提供了**断点续传**的能力，即在下载过程中，记录已完成的分片状态，如果出现网络中断、程序异常退出等问题导致文件下载失败，甚至重试多次仍无法完成下载，再次下载时，可以通过断点记录文件恢复下载。


下载管理器Downloader的常用方法如下：

```
type Downloader struct {
  ...
}

// 用于创建新的下载管理器
func (c *Client) NewDownloader(optFns ...func(*DownloaderOptions)) *Downloader

// 用于下载文件
func (d *Downloader) DownloadFile(ctx context.Context, request *GetObjectRequest, filePath string, optFns ...func(*DownloaderOptions)) (result *DownloadResult, err error)
```

#### **请求参数列表**

| 参数名 | 类型  | 说明  |
| --- | --- | --- |
| ctx | context.Context | 请求的上下文，可以用来设置请求的总时限 |
| request | \\*GetObjectRequest | 设置具体接口的请求参数，具体请参见[GetObjectRequest](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#GetObjectRequest) |
| filePath | string | 本地文件路径 |
| optFns | ...func(\\*DownloaderOptions) | （可选）配置选项 |

其中，DownloaderOption常用参数说明列举如下：

| 参数名 | 类型  | 说明  |
| --- | --- | --- |
| PartSize | int64 | 指定分片大小，默认值为 6MiB |
| ParallelNum | int | 指定下载任务的并发数，默认值为 3。针对的是单次调用的并发限制，而不是全局的并发限制 |
| EnableCheckpoint | bool | 是否记录断点下载信息，默认不记录 |
| CheckpointDir | string | 指定记录文件的保存路径，例如 /local/dir/, 当EnableCheckpoint 为 true时有效 |
| VerifyData | bool | 恢复下载时，是否要校验已下载数据的CRC64值，默认不校验, 当EnableCheckpoint 为 true时有效 |
| UseTempFile | bool | 下载文件时，是否使用临时文件，默认使用。先下载到 临时文件上，当成功后，再重命名为目标文件 |

当使用NewDownloader实例化实例时，您可以指定多个配置选项来自定义对象的下载行为。也可以在每次调用下载接口时，指定多个配置选项来自定义每次下载对象的行为。

-   设置Downloader的配置参数：

    ```
    d := client.NewDownloader(func(do *oss.DownloaderOptions) {
      do.PartSize = 10 * 1024 * 1024
    })
    ```

-   设置每次下载请求的配置参数：

    ```
    request := &oss.GetObjectRequest{Bucket: oss.Ptr("bucket"), Key: oss.Ptr("key")}
    d.DownloadFile(context.TODO(), request, "/local/dir/example", func(do *oss.DownloaderOptions) {
      do.PartSize = 10 * 1024 * 1024
    })
    ```


## **示例代码**

您可以使用以下代码将存储空间中的文件下载到本地。

```
package main

import (
	"context"
	"flag"
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
	flag.StringVar(&objectName, "src-object", "", "The name of the source object.")
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

	// 检查源对象名称是否为空
	if len(objectName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, src object name required")
	}

	// 配置OSS客户端
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建下载器
	d := client.NewDownloader()

	// 构建获取对象的请求
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
	}

	// 定义本地文件路径
	localFile := "local-file"

	// 执行下载文件的请求
	result, err := d.DownloadFile(context.TODO(), request, localFile)
	if err != nil {
		log.Fatalf("failed to download file %v", err)
	}

	// 打印下载成功的信息
	log.Printf("download file %s to local-file successfully, size: %d", objectName, result.Written)
}
```

## **常见使用场景**

### **使用下载管理器设置分片大小和并发数**

您可以使用以下代码配置下载管理器的DownloaderOptions参数，设置分片大小和并发数。

```
package main

import (
	"context"
	"flag"
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
	flag.StringVar(&objectName, "src-object", "", "The name of the source object.")
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

	// 检查源对象名称是否为空
	if len(objectName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, src object name required")
	}

	// 配置OSS客户端
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建下载管理器
	d := client.NewDownloader()

	// 构建获取对象的请求
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
	}

	// 定义本地文件路径
	localFile := "local-file"

	// 设置下载器的配置参数
	downloaderOptions := func(do *oss.DownloaderOptions) {
		do.PartSize = 20 * 1024 * 1024 // 指定分片大小为20MiB
		do.ParallelNum = 6            // 指定下载任务的并发数为6
	}

	// 执行下载文件的请求
	result, err := d.DownloadFile(context.TODO(), request, localFile, downloaderOptions)
	if err != nil {
		log.Fatalf("failed to download file %v", err)
	}

	// 打印下载成功的信息
	log.Printf("download file %s to local-file successfully, size: %d", objectName, result.Written)
}
```

### **使用下载管理器启动断点续传功能**

您可以使用以下代码配置下载管理器的DownloaderOptions参数，启动断点续传功能。

```
package main

import (
	"context"
	"flag"
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
	flag.StringVar(&objectName, "src-object", "", "The name of the source object.")
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

	// 检查源对象名称是否为空
	if len(objectName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, src object name required")
	}

	// 配置OSS客户端
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建下载器
	d := client.NewDownloader()

	// 构建获取对象的请求
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
	}

	// 定义本地文件路径
	localFile := "local-file"

	// 设置下载器选项
	downloaderOptions := func(do *oss.DownloaderOptions) {
		do.EnableCheckpoint = true        // 启用记录断点下载信息
		do.CheckpointDir = "./checkpoint" // 指定断点下载信息存储的目录
		do.UseTempFile = true             // 下载文件时使用临时文件
	}

	// 执行下载文件的请求
	result, err := d.DownloadFile(context.TODO(), request, localFile, downloaderOptions)
	if err != nil {
		log.Fatalf("failed to download file %v", err)
	}

	// 打印下载成功的信息
	log.Printf("download file %s to local-file successfully, size: %d", objectName, result.Written)
}
```

## **相关文档**

-   关于下载管理器的更多信息，请参见[开发者指南](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2/blob/master/DEVGUIDE-CN.md#%E4%B8%8B%E8%BD%BD%E7%AE%A1%E7%90%86%E5%99%A8downloader)。

-   关于文件下载的API接口，请参见[DownloadFile](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Downloader.DownloadFile)。