追加上传是指在已上传的追加类型文件（Appendable Object）末尾直接追加内容。本文介绍如何使用OSS Go SDK进行追加上传。

## **注意事项**

-   本文示例代码以华东1（杭州）的地域ID`cn-hangzhou`为例，默认使用外网Endpoint，如果您希望通过与OSS同地域的其他阿里云产品访问OSS，请使用内网Endpoint。关于OSS支持的Region与Endpoint的对应关系，请参见[OSS地域和访问域名](https://help.aliyun.com/zh/oss/user-guide/regions-and-endpoints#concept-zt4-cvy-5db)。

-   本文以从环境变量读取访问凭证为例。如何配置访问凭证，请参见[配置访问凭证](https://help.aliyun.com/zh/oss/developer-reference/configure-access-credentials-by-using-oss-sdk-for-go-v2)。

-   当文件不存在时，调用追加上传接口会创建一个追加类型文件。

-   当文件已存在时：

    -   如果文件为追加类型文件，且设置的追加位置和文件当前长度相等，则直接在该文件末尾追加内容。

    -   如果文件为追加类型文件，但是设置的追加位置和文件当前长度不相等，则抛出PositionNotEqualToLength异常。

    -   如果文件为非追加类型文件，例如通过简单上传的文件类型为Normal的文件，则抛出ObjectNotAppendable异常。


## **权限说明**

阿里云账号默认拥有全部权限。阿里云账号下的RAM用户或RAM角色默认没有任何权限，需要阿里云账号或账号管理员通过[RAM Policy](https://help.aliyun.com/zh/oss/ram-policy-overview/)或[Bucket Policy](https://help.aliyun.com/zh/oss/user-guide/oss-bucket-policy/)授予操作权限。

| **API** | **Action** | **说明** |
| --- | --- | --- |
| AppendObject | `oss:PutObject` | 以追加写的方式上传文件（Object）。 |
| `oss:PutObjectTagging` | 以追加写的方式上传文件（Object）时，如果通过x-oss-tagging指定Object的标签，则需要此操作的权限。 |

## **方法定义**

针对文件追加上传的场景，Go SDK V2新增了AppendFile接口以模仿文件的读写行为，用于操作存储空间里的对象，以下列举了AppendFile与AppendObject接口的具体说明：

| 接口名 | 说明  |
| --- | --- |
| Client.AppendObject | 追加上传, 最终文件最大支持5GiB 支持CRC64数据校验（默认启用） 支持进度条 请求body类型为io.Reader, 当支持io.Seeker类型时，具备失败重传（该接口为非幂等接口，重传时可能出现失败） |
| Client.AppendFile | 与Client.AppendObject接口能力一致 优化了重传时失败后容错处理 包含AppendOnlyFile接口 AppendOnlyFile.Write AppendOnlyFile.WriteFrom |

### **高级版追加上传API：AppendFile**

调用AppendFile接口以追加写的方式上传数据。如果对象不存在，则创建追加类型的对象。如果对象存在，并且不为追加类型的对象，则返回错误。

AppendFile接口定义如下。

```
type AppendOnlyFile struct {
...
}

func (c *Client) AppendFile(ctx context.Context, bucket string, key string, optFns ...func(*AppendOptions)) (*AppendOnlyFile, error)
```

#### **请求参数列表**

| 参数名 | 类型  | 说明  |
| --- | --- | --- |
| ctx | context.Context | 请求的上下文 |
| bucket | string | 设置存储空间名 |
| key | string | 设置对象名 |
| optFns | ...func(\\*AppendOptions) | （可选）追加文件时的配置选项 |

其中，AppendOptions的参数说明列举如下：

| 参数  | 类型  | 说明  |
| --- | --- | --- |
| RequestPayer | \\*string | 启用了请求者付费模式时，需要设置为'requester' |
| CreateParameter | \\*AppendObjectRequest | 用于首次上传时，设置对象的元信息，包括ContentType，Metadata，权限，存储类型等，具体请参见[AppendObjectRequest](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#AppendObjectRequest) |

#### **返回值列表**

| 返回值名 | 类型  | 说明  |
| --- | --- | --- |
| file | \\*AppendOnlyFile | 追加文件的实例，当 err 为nil 时有效，具体请参见[AppendOnlyFile](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#AppendOnlyFile) |
| err | error | 打开追加文件时的状态，当失败时，err 不为 nil |

其中，AppendOnlyFile接口说明如下：

| 接口名 | 说明  |
| --- | --- |
| Close() error | 关闭文件句柄，释放资源 |
| Write(b \\[\\]byte) (int, error) | 将b中的数据写入到数据流中，返回写入的字节数和遇到的错误 |
| WriteFrom(r io.Reader) (int64, error) | 将r中的数据写入到数据流中，返回写入的字节数和遇到的错误 |
| Stat() (os.FileInfo, error) | 获取对象的信息，包括对象大小、最后修改时间以及元信息 |

### **基础版追加上传API：AppendObject**

```
func (c *Client) AppendObject(ctx context.Context, request *AppendObjectRequest, optFns ...func(*Options)) (*AppendObjectResult, error)
```

#### **请求参数列表**

| 参数名 | 类型  | 说明  |
| --- | --- | --- |
| ctx | context.Context | 请求的上下文，可以用来设置请求的总时限 |
| request | \\*AppendObjectRequest | 设置具体接口的请求参数，具体请参见[AppendObjectRequest](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#AppendObjectRequest) |
| optFns | ...func(\\*Options) | （可选）接口级的配置参数, 具体请参见[Options](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Options) |

#### **返回值列表**

| 返回值名 | 类型  | 说明  |
| --- | --- | --- |
| result | \\*AppendObjectResult | 接口返回值，当 err 为nil 时有效，具体请参见[AppendObjectResult](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#AppendObjectResult) |
| err | error | 请求的状态，当请求失败时，err 不为 nil |

## **示例代码**

### **使用AppendFile追加上传**

```
package main

import (
	"context"
	"flag"
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

	// 检查存储空间名称是否为空
	if len(bucketName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, bucket name required")
	}

	// 检查存储区域是否为空
	if len(region) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, region required")
	}

	// 检查对象名称是否为空
	if len(objectName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, object name required")
	}

	// 配置OSS客户端
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建一个追加文件的实例
	f, err := client.AppendFile(context.TODO(),
		bucketName,
		objectName,
		func(ao *oss.AppendOptions) {
			ao.CreateParameter = &oss.AppendObjectRequest{
				Acl: oss.ObjectACLPrivate, // 设置对象的访问权限为私有
				Metadata: map[string]string{
					"user": "jack", // 设置对象的元数据
				},
				Tagging: oss.Ptr("key=value"), // 设置对象的标签
			}
		})

	if err != nil {
		log.Fatalf("failed to append file %v", err)
	}
	defer f.Close() // 确保文件流在函数结束时关闭

	// 打开本地文件 example1.txt
	lf, err := os.Open("/local/dir/example1.txt")
	if err != nil {
		log.Fatalf("failed to open local file %v", err)
	}

	// 将 example1.txt 的内容追加到 OSS 对象
	_, err = f.WriteFrom(lf)
	if err != nil {
		log.Fatalf("failed to append file %v", err)
	}
	lf.Close() // 关闭本地文件

	// 打开本地文件 example2.txt
	lf, err = os.Open("/local/dir/example2.txt")
	if err != nil {
		log.Fatalf("failed to open local file %v", err)
	}

	// 将 example2.txt 的内容追加到 OSS 对象
	_, err = f.WriteFrom(lf)
	if err != nil {
		log.Fatalf("failed to append file %v", err)
	}
	lf.Close() // 关闭本地文件

	// 读取本地文件 example3.txt 的内容
	lb, err := os.ReadFile("/local/dir/example3.txt")
	if err != nil {
		log.Fatalf("failed to read local file %v", err)
	}

	// 将 example3.txt 的内容追加到 OSS 对象
	_, err = f.Write(lb)
	if err != nil {
		log.Fatalf("failed to append file %v", err)
	}

	// 打印成功追加文件的日志信息
	log.Printf("append file successfully")
}
```

### **使用AppendObject追加上传**

```
package main

import (
	"context"
	"flag"
	"log"
	"strings"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// 定义全局变量
var (
	region     string
	bucketName string
	objectName string
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

	// 定义追加文件的初始位置
	var (
		position = int64(0)
	)

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

	// 加载默认配置并设置凭证提供者和region
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 定义要追加的内容
	content := "hi append object"

	// 创建AppendObject请求
	request := &oss.AppendObjectRequest{
		Bucket:   oss.Ptr(bucketName),
		Key:      oss.Ptr(objectName),
		Position: oss.Ptr(position),
		Body:     strings.NewReader(content),
	}

	// 执行AppendObject请求并处理结果
	// 第一次追加上传的位置是0，返回值中包含下一次追加的位置
	result, err := client.AppendObject(context.TODO(), request)
	if err != nil {
		log.Fatalf("failed to append object %v", err)
	}

	// 创建第二次AppendObject请求
	request = &oss.AppendObjectRequest{
		Bucket:   oss.Ptr(bucketName),
		Key:      oss.Ptr(objectName),
		Position: oss.Ptr(result.NextPosition), //从第一次AppendObject返回值中获取NextPosition
		Body:     strings.NewReader("hi append object"),
	}

	// 执行第二次AppendObject请求并处理结果
	result, err = client.AppendObject(context.TODO(), request)
	if err != nil {
		log.Fatalf("failed to append object %v", err)
	}

	log.Printf("append object result:%#v\n", result)
}
```

## **常见使用场景**

### **追加上传并显示进度条**

```
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// 定义全局变量
var (
	region     string
	bucketName string
	objectName string
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

	// 定义追加文件的初始位置
	var (
		position = int64(0)
	)

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

	// 加载默认配置并设置凭证提供者和region
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 定义要追加的内容
	content := "hi append object"

	// 创建AppendObject请求
	request := &oss.AppendObjectRequest{
		Bucket:   oss.Ptr(bucketName),
		Key:      oss.Ptr(objectName),
		Position: oss.Ptr(position),
		Body:     strings.NewReader(content),
		ProgressFn: func(increment, transferred, total int64) {
			fmt.Printf("increment:%v, transferred:%v, total:%v\n", increment, transferred, total)
		}, // 进度回调函数，用于显示上传进度
	}

	// 执行第一次AppendObject请求并处理结果
	// 第一次追加上传的位置是0，返回值中包含下一次追加的位置
	result, err := client.AppendObject(context.TODO(), request)
	if err != nil {
		log.Fatalf("failed to append object %v", err)
	}

	// 创建第二次AppendObject请求
	request = &oss.AppendObjectRequest{
		Bucket:   oss.Ptr(bucketName),
		Key:      oss.Ptr(objectName),
		Position: oss.Ptr(result.NextPosition), //从第一次AppendObject返回值中获取NextPosition
		Body:     strings.NewReader("hi append object"),
		ProgressFn: func(increment, transferred, total int64) {
			fmt.Printf("increment:%v, transferred:%v, total:%v\n", increment, transferred, total)
		}, // 进度回调函数，用于显示上传进度
	}

	// 执行第二次AppendObject请求并处理结果
	result, err = client.AppendObject(context.TODO(), request)
	if err != nil {
		log.Fatalf("failed to append object %v", err)
	}

	// 打印追加对象的versionId
	log.Printf("append object result:%#v\n", *result.VersionId)
}
```

## 相关文档

-   关于追加上传的完整示例代码，请参见[GitHub示例](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2/blob/master/sample/append_file.go)和[开发者指南](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2/blob/master/DEVGUIDE-CN.md#%E7%B1%BB%E6%96%87%E4%BB%B6file-like)。

-   关于追加上传的高级API接口，请参见[AppendFile](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Client.AppendFile)。

-   关于追加上传的基础API接口，请参见[AppendObject](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Client.AppendObject)。