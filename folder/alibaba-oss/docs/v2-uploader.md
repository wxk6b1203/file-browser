本文针对文件的传输场景，介绍如何使用Go SDK V2新增的上传管理器Uploader模块进行文件上传。

## **注意事项**

-   本文示例代码以华东1（杭州）的地域ID`cn-hangzhou`为例，默认使用外网Endpoint，如果您希望通过与OSS同地域的其他阿里云产品访问OSS，请使用内网Endpoint。关于OSS支持的Region与Endpoint的对应关系，请参见[OSS地域和访问域名](https://help.aliyun.com/zh/oss/user-guide/regions-and-endpoints#concept-zt4-cvy-5db)。

-   本文以从环境变量读取访问凭证为例。如何配置访问凭证，请参见[配置访问凭证](https://help.aliyun.com/zh/oss/developer-reference/configure-access-credentials-by-using-oss-sdk-for-go-v2)。

-   要进行大文件上传，您必须有`oss:PutObject`权限。具体操作，请参见[为RAM用户授予自定义的权限策略](https://help.aliyun.com/zh/oss/user-guide/common-examples-of-ram-policies#section-ucu-jv0-zip)。


## **方法定义**

#### 上传管理器功能简介

Go SDK V2新增上传管理器Uploader提供了通用的上传接口，隐藏了底层接口的实现细节，提供便捷的文件上传能力。

-   上传管理器Uploader底层利用分片上传接口，把大文件或者流分成多个较小的分片并发上传，提升上传的性能。

-   上传管理器Uploader同时提供了**断点续传**的能力，即在上传过程中，记录已完成的分片状态，如果出现网络中断、程序异常退出等问题导致文件上传失败，甚至重试多次仍无法完成上传，再次上传时，可以通过断点记录文件恢复上传。


上传管理器Uploader的常用方法如下：

```
type Uploader struct {
  ...
}

// 用于创建新的上传管理器
func (c *Client) NewUploader(optFns ...func(*UploaderOptions)) *Uploader 

// 用于上传文件流
func (u *Uploader) UploadFrom(ctx context.Context, request *PutObjectRequest, body io.Reader, optFns ...func(*UploaderOptions)) (*UploadResult, error)

// 用于上传本地文件
func (u *Uploader) UploadFile(ctx context.Context, request *PutObjectRequest, filePath string, optFns ...func(*UploaderOptions)) (*UploadResult, error)
```

#### **请求参数列表**

| 参数名 | 类型  | 说明  |
| --- | --- | --- |
| ctx | context.Context | 请求的上下文 |
| request | \\*PutObjectRequest | 上传对象的请求参数，和 PutObject 接口的请求参数一致，具体请参见[PutObjectRequest](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#PutObjectRequest) |
| body | io.Reader | 需要上传的流 - 当 body 只支持io.Reader类型，必须先把数据缓冲在内存中，然后才能上传该部分 - 当 body 同时支持 io.Reader, io.Seeker 和 io.ReaderAt 类型时，不需要把数据缓存在内存里 |
| filePath | string | 本地文件路径 |
| optFns | ...func(\\*UploaderOptions) | （可选）配置选项 |

其中，UploaderOptions常用参数说明列举如下：

| 参数名 | 类型  | 说明  |
| --- | --- | --- |
| PartSize | int64 | 指定分片大小，默认值为 6MiB |
| ParallelNum | int | 指定上传任务的并发数，默认值为 3。针对的是单次调用的并发限制，而不是全局的并发限制 |
| LeavePartsOnError | bool | 当上传失败时，是否保留已上传的分片，默认不保留 |
| EnableCheckpoint | bool | 是否开启断点上传功能，默认不开启 **说明** EnableCheckpoint参数目前仅对UploadFile 接口有效，UploadFrom接口暂不支持 |
| CheckpointDir | string | 指定记录文件的保存路径，例如 /local/dir/, 当EnableCheckpoint 为 true时有效 |

当使用NewUploader实例化实例时，您可以指定多个配置选项来自定义对象的上传行为。也可以在每次调用上传接口时，指定多个配置选项来自定义每次上传对象的行为。

-   设置Uploader的配置参数

    ```
    u := client.NewUploader(func(uo *oss.UploaderOptions) {
      uo.PartSize = 10 * 1024 * 1024
    })
    ```

-   设置每次上传请求的配置参数

    ```
    request := &oss.PutObjectRequest{Bucket: oss.Ptr("bucket"), Key: oss.Ptr("key")}
    result, err := u.UploadFile(context.TODO(), request, "/local/dir/example", func(uo *oss.UploaderOptions) {
      uo.PartSize = 10 * 1024 * 1024
    })
    ```


## **示例代码**

您可以通过以下代码使用上传管理器上传本地文件到存储空间。

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
	flag.StringVar(&objectName, "object", "", "The name of the source object.")
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
		log.Fatalf("invalid parameters, source object name required")
	}

	// 加载默认配置并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建上传管理器
	u := client.NewUploader()

	// 定义本地文件路径，需要替换为您的实际本地文件路径和文件名称
	localFile := "/Users/yourLocalPath/yourFileName"

	// 执行上传文件的操作
	result, err := u.UploadFile(context.TODO(),
		&oss.PutObjectRequest{
			Bucket: oss.Ptr(bucketName),
			Key:    oss.Ptr(objectName)},
		localFile)
	if err != nil {
		log.Fatalf("failed to upload file %v", err)
	}

	// 打印上传文件的结果
	log.Printf("upload file result:%#v\n", result)
}
```

## **常见使用场景**

### **使用上传管理器启动断点续传功能**

您可以使用以下代码设置上传管理器Uploader的配置参数，启动断点续传功能。

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
	flag.StringVar(&objectName, "object", "", "The name of the source object.")
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
		log.Fatalf("invalid parameters, source object name required")
	}

	// 加载默认配置并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建上传器，并启用断点续传功能
	u := client.NewUploader(func(uo *oss.UploaderOptions) {
		uo.CheckpointDir = "/Users/yourLocalPath/checkpoint/" // 指定断点记录文件的保存路径
		uo.EnableCheckpoint = true        // 开启断点续传
	})

	// 定义本地文件路径，需要替换为您的实际本地文件路径和文件名称
	localFile := "/Users/yourLocalPath/yourFileName"

	// 执行上传文件的操作
	result, err := u.UploadFile(context.TODO(),
		&oss.PutObjectRequest{
			Bucket: oss.Ptr(bucketName),
			Key:    oss.Ptr(objectName)},
		localFile)
	if err != nil {
		log.Fatalf("failed to upload file %v", err)
	}

	// 打印上传文件的结果
	log.Printf("upload file result:%#v\n", result)
}
```

### **使用上传管理器上传本地文件流**

您可以通过以下代码使用上传管理器上传本地文件流。

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

	// 加载默认配置并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建上传器
	u := client.NewUploader()

	// 将“/Users/yourLocalPath/yourFileName”替换为实际的文件路径和文件名称，打开本地文件并创建io.Reader实例
	file, err := os.Open("/Users/yourLocalPath/yourFileName")
	if err != nil {
		log.Fatalf("failed to open local file %v", err)
	}
	defer file.Close()

	var r io.Reader = file

	// 执行上传文件流的操作
	result, err := u.UploadFrom(context.TODO(),
		&oss.PutObjectRequest{
			Bucket: oss.Ptr(bucketName),
			Key:    oss.Ptr(objectName),
		},
		r, // 上传文件流
	)

	if err != nil {
		log.Fatalf("failed to upload file stream %v", err)
	}

	// 打印上传文件流的结果
	log.Printf("upload file stream, etag: %v\n", oss.ToString(result.ETag))
}
```

### **使用上传管理器设置分片大小和并发数**

您可以使用以下代码设置上传管理器Uploader的配置参数，设置分片大小和并发数。

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
	flag.StringVar(&objectName, "object", "", "The name of the source object.")
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
		log.Fatalf("invalid parameters, source object name required")
	}

	// 加载默认配置并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建上传器
	u := client.NewUploader(func(uo *oss.UploaderOptions) {
		uo.PartSize = 1024 * 1024 * 5     // 设置分片大小为5MB
		uo.ParallelNum = 5                // 设置并行上传数为5
	})

	// 定义本地文件路径，需要替换为您的实际本地文件路径和文件名
	localFile := "/Users/yourLocalPath/yourFileName"

	// 执行上传文件的操作
	result, err := u.UploadFile(context.TODO(),
		&oss.PutObjectRequest{
			Bucket: oss.Ptr(bucketName),
			Key:    oss.Ptr(objectName)},
		localFile)
	if err != nil {
		log.Fatalf("failed to upload file %v", err)
	}

	// 打印上传文件的结果
	log.Printf("upload file result:%#v\n", result)
}
```

### **使用上传管理器设置上传回调**

如果您希望在文件上传后通知应用服务器，可参考以下代码示例。

```
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
	flag.StringVar(&objectName, "object", "", "The name of the source object.")
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
		log.Fatalf("invalid parameters, source object name required")
	}

	// 加载默认配置并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建上传器，并启用断点续传功能
	u := client.NewUploader(func(uo *oss.UploaderOptions) {
		uo.PartSize = 1024 * 1024 * 5 // 设置分片大小为5MB
		uo.ParallelNum = 5            // 设置并行上传数为5
	})

	// 定义本地文件路径，需要替换为您的实际本地文件路径
	localFile := "/Users/yourLocalPath/yourFileName"

	// 定义回调参数
	callbackMap := map[string]string{
		"callbackUrl":      "https://example.com:23450/callback",                                                        // 设置回调服务器的URL，例如https://example.com:23450。
		"callbackBody":     "bucket=${bucket}&object=${object}&size=${size}&my_var_1=${x:my_var1}&my_var_2=${x:my_var2}", // 设置回调请求体。
		"callbackBodyType": "application/x-www-form-urlencoded",                                                          //设置回调请求体类型。
	}

	// 将回调参数转换为JSON并进行Base64编码，以便将其作为回调参数传递
	callbackStr, err := json.Marshal(callbackMap)
	if err != nil {
		log.Fatalf("failed to marshal callback map: %v", err)
	}
	callbackBase64 := base64.StdEncoding.EncodeToString(callbackStr)

	// 定义自定义回调参数
	callbackVarMap := map[string]string{}
	callbackVarMap["x:my_var1"] = "thi is var 1"
	callbackVarMap["x:my_var2"] = "thi is var 2"

	// 将回调参数转换为JSON并进行Base64编码，以便将其作为回调参数传递
	callbackVarStr, err := json.Marshal(callbackVarMap)
	if err != nil {
		log.Fatalf("failed to marshal callback var: %v", err)
	}
	callbackVarBase64 := base64.StdEncoding.EncodeToString(callbackVarStr)

	// 执行上传大文件的操作
	result, err := u.UploadFile(context.TODO(),
		&oss.PutObjectRequest{
			Bucket:      oss.Ptr(bucketName),
			Key:         oss.Ptr(objectName),
			Callback:    oss.Ptr(callbackBase64),    // 填写回调参数
			CallbackVar: oss.Ptr(callbackVarBase64), // 填写自定义回调参数
		},
		localFile)
	if err != nil {
		log.Fatalf("failed to upload file %v", err)
	}

	// 打印上传文件的结果
	log.Printf("upload file result:%#v\n", result)
}
```

### **使用上传管理器显示进度条**

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
	flag.StringVar(&objectName, "object", "", "The name of the source object.")
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
		log.Fatalf("invalid parameters, source object name required")
	}

	// 加载默认配置并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建上传管理器
	u := client.NewUploader(func(uo *oss.UploaderOptions) {
		uo.PartSize = 1024 * 1024 * 5 // 设置分片大小为5MB
		uo.ParallelNum = 3            // 设置并行上传数为3
	})

	// 替换为您的实际本地文件路径
	localFile := "/Users/yourLocalPath/yourFileName"

	// 执行上传大文件的操作
	result, err := u.UploadFile(context.TODO(),
		&oss.PutObjectRequest{
			Bucket: oss.Ptr(bucketName),
			Key:    oss.Ptr(objectName),
			ProgressFn: func(increment, transferred, total int64) {
				fmt.Printf("increment:%v, transferred:%v, total:%v\n", increment, transferred, total)
			}, // 进度回调函数，用于显示上传进度
		},
		localFile)
	if err != nil {
		log.Fatalf("failed to upload file %v", err)
	}

	// 打印上传文件的结果
	log.Printf("upload file result:%#v\n", result)
}
```

## **相关文档**

-   关于上传管理器的更多信息，请参见[开发者指南](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2/blob/master/DEVGUIDE-CN.md#%E4%B8%8A%E4%BC%A0%E7%AE%A1%E7%90%86%E5%99%A8uploader)。

-   关于大文件上传的API接口，请参见[UploadFrom](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Uploader.UploadFrom)和[UploadFile](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Uploader.UploadFile)。