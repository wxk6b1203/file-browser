本文针对大文件的传输场景，介绍如何使用Go SDK V2新增的Copier模块进行文件拷贝。

## **注意事项**

-   本文示例代码以华东1（杭州）的地域ID`cn-hangzhou`为例，默认使用外网Endpoint，如果您希望通过与OSS同地域的其他阿里云产品访问OSS，请使用内网Endpoint。关于OSS支持的Region与Endpoint的对应关系，请参见[OSS地域和访问域名](https://help.aliyun.com/zh/oss/user-guide/regions-and-endpoints#concept-zt4-cvy-5db)。

-   本文以从环境变量读取访问凭证为例。如何配置访问凭证，请参见[配置访问凭证](https://help.aliyun.com/zh/oss/developer-reference/configure-access-credentials-by-using-oss-sdk-for-go-v2)。

-   要进行拷贝文件，您必须拥有源文件的读权限及目标Bucket的读写权限。

-   不支持跨地域拷贝。例如不能将华东1（杭州）地域存储空间中的文件拷贝到华北1（青岛）地域。

-   拷贝文件时，您需要确保源Bucket和目标Bucket均未设置合规保留策略，否则报错The object you specified is immutable.。


## **方法定义**

#### **拷贝管理器介绍**

当需要将对象从存储空间复制到另外一个存储空间，或者修改对象的属性时，您可以通过拷贝接口或者分片拷贝接口来完成这个操作。这两个接口有其适用的场景，例如：

-   拷贝接口（CopyObject）只适合拷贝 5GiB 以下的对象；

-   分片拷贝接口（UploadPartCopy）支持拷贝大于5GiB 的对象，但不支持元数据指令（x-oss-metadata-directive）和标签指令（x-oss-tagging-directive）参数，拷贝时需要主动设置需要复制的元数据和标签。


**Go SDK V2新增拷贝管理器**Copier提供了通用的拷贝接口，隐藏了接口的差异和实现细节，可根据拷贝的请求参数自动选择合适的接口复制对象。Copier的常用方法定义如下：

```
type Copier struct {
  ...
}

// 用于创建新的拷贝管理器
func (c *Client) NewCopier(optFns ...func(*CopierOptions)) *Copier

// // 用于拷贝文件
func (c *Copier) Copy(ctx context.Context, request *CopyObjectRequest, optFns ...func(*CopierOptions)) (*CopyResult, error)
```

#### **请求参数列表**

| 参数名 | 类型  | 说明  |
| --- | --- | --- |
| ctx | context.Context | 请求的上下文，可以用来设置请求的总时限 |
| request | \\*CopyObjectRequest | 设置具体接口的请求参数，具体请参见[CopyObjectRequest](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#CopyObjectRequest) |
| optFns | ...func(\\*CopierOptions) | （可选）配置选项，具体请参见[CopierOptions](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#CopierOptions) |

其中，CopyObjectRequest的常用参数列举如下：

| 参数名 | 类型  | 说明  |
| --- | --- | --- |
| Bucket | \\*string | 指定目标存储空间名称 |
| Key | \\*string | 指定目标对象名称 |
| SourceBucket | \\*string | 指定源存储空间名称 |
| SourceKey | \\*string | 指定源对象名称 |
| ForbidOverwrite | \\*string | 指定CopyObject操作时是否覆盖同名目标Object |
| Tagging | \\*string | 指定Object的对象标签，可同时设置多个标签，例如TagA=A&TagB=B。 |
| TaggingDirective | \\*string | 指定如何设置目标Object的对象标签。取值如下： - Copy（默认值）：复制源Object的对象标签到目标 Object。 - Replace：忽略源Object的对象标签，直接采用请求中指定的对象标签。 |

CopierOptions选项的常用参数列举如下：

| 参数名 | 类型  | 说明  |
| --- | --- | --- |
| PartSize | int64 | 指定分片大小，默认值为 64MiB |
| ParallelNum | int | 指定上传任务的并发数，默认值为 3。针对的是单次调用的并发限制，而不是全局的并发限制 |
| MultipartCopyThreshold | int64 | 使用分片拷贝的阈值，默认值为 200MiB |
| LeavePartsOnError | bool | 当拷贝失败时，是否保留已拷贝的分片，默认不保留 |
| DisableShallowCopy | bool | 不使用浅拷贝行为，默认使用 |

#### **返回值列表**

| 返回值名 | 类型  | 说明  |
| --- | --- | --- |
| result | \\*CopyResult | 接口返回值，当 err 为nil 时有效，具体请参见[CopyResult](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#CopyResult) |
| err | error | 请求的状态，当请求失败时，err 不为 nil |

## **示例代码**

您可以使用以下代码将对象从源存储空间拷贝到目标存储空间并修改对象的属性。

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
	region         string // 存储区域
	srcBucketName  string // 源存储空间名称
	srcObjectName  string // 源对象名称
	destBucketName string // 目标存储空间名称
	destObjectName string // 目标对象名称
)

// init函数用于初始化命令行参数
func init() {
	flag.StringVar(&region, "region", "", "The region in which the bucket is located.")
	flag.StringVar(&srcBucketName, "src-bucket", "", "The name of the source bucket.")
	flag.StringVar(&srcObjectName, "src-object", "", "The name of the source object.")
	flag.StringVar(&destBucketName, "dest-bucket", "", "The name of the destination bucket.")
	flag.StringVar(&destObjectName, "dest-object", "", "The name of the destination object.")
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 检查源存储空间名称是否为空
	if len(srcBucketName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, bucket name required")
	}

	// 检查存储区域是否为空
	if len(region) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, region required")
	}

	// 如果目标存储空间名称未指定，则使用源存储空间名称
	if len(destBucketName) == 0 {
		destBucketName = srcBucketName
	}

	// 检查源对象名称是否为空
	if len(srcObjectName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, src object name required")
	}

	// 检查目标对象名称是否为空
	if len(destObjectName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, destination object name required")
	}

	// 配置OSS客户端
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建文件拷贝器
	c := client.NewCopier()

	// 构建拷贝对象的请求
	request := &oss.CopyObjectRequest{
		Bucket:            oss.Ptr(destBucketName),  // 目标存储空间名称
		Key:               oss.Ptr(destObjectName),  // 目标对象名称
		SourceKey:         oss.Ptr(srcObjectName),   // 源对象名称
		SourceBucket:      oss.Ptr(srcBucketName),   // 源存储空间名称
		StorageClass:      oss.StorageClassStandard, // 指定存储类型为标准类型
		MetadataDirective: oss.Ptr("Replace"),       // 不拷贝源对象元数据
		TaggingDirective:  oss.Ptr("Replace"),       // 不拷贝源对象标签
	}

	// 执行拷贝对象的操作
	result, err := c.Copy(context.TODO(), request)
	if err != nil {
		log.Fatalf("failed to copy object %v", err) // 如果拷贝失败，记录错误并退出
	}

	// 打印拷贝对象的结果
	log.Printf("copy object result:%#v\n", result)
}
```

## **相关文档**

-   关于拷贝管理器的API接口，请参见[Copy](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Copier.Copy)。

-   关于拷贝管理器的更多信息，请参见[开发者指南](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2/blob/master/DEVGUIDE-CN.md#%E6%8B%B7%E8%B4%9D%E7%AE%A1%E7%90%86%E5%99%A8copier)。