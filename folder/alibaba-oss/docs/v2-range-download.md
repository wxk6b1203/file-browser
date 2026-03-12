本文介绍如何使用范围下载方法，帮助您高效地获取文件中的特定部分数据。

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
| request | \\*GetObjectRequest | 设置具体接口的请求参数，例如设置Range指定下载范围，RangeBehavior指定标准行为范围下载，具体请参见[GetObjectRequest](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#GetObjectRequest) |
| optFns | ...func(\\*Options) | （可选）接口级的配置参数, 具体请参见[Options](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Options) |

#### **返回值列表**

| 返回值名 | 类型  | 说明  |
| --- | --- | --- |
| result | \\*GetObjectResult | 接口返回值，当 err 为nil 时有效，具体请参见[GetObjectResult](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#GetObjectResult) |
| err | error | 请求的状态，当请求失败时，err 不为 nil |

**重要**

-   假设现有大小为1000 Bytes的Object，则指定的正常下载范围应为0~999。**如果指定范围不在有效区间，会导致Range不生效，响应返回值为200，并传送整个Object的内容。**请求不合法的示例及返回说明如下：

    -   若指定了Range: bytes=500-2000，**此时范围末端取值不在有效区间**，返回整个文件的内容，且HTTP Code为200。

    -   若指定了Range: bytes=1000-2000，**此时范围首端取值不在有效区间**，返回整个文件的内容，且HTTP Code为200。

-   您可以在请求中增加请求头**x-oss-range-behavior:standard**指定标准行为范围下载，改变指定范围不在有效区间时OSS的下载行为。假设现有大小为1000 Bytes的Object：

    -   若指定了Range: bytes=500-2000，**此时范围末端取值不在有效区间**，返回500~999字节范围内容，且HTTP Code为206。

    -   若指定了Range: bytes=1000-2000，**此时范围首端取值不在有效区间**，返回HTTP Code为416，错误码为InvalidRange。


## **示例代码**

以下代码展示了在请求中增加请求头RangeBehavior:standard来指定标准下载行为，下载指定范围内的文件数据。

```
package main

import (
	"context"
	"flag"
	"io"
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

	// 创建获取对象的请求
	request := &oss.GetObjectRequest{
		Bucket:        oss.Ptr(bucketName),     // 存储空间名称
		Key:           oss.Ptr(objectName),     // 对象名称
		Range:         oss.Ptr("bytes=15-35"), // 指定下载范围
		RangeBehavior: oss.Ptr("standard"),     // 指定标准行为范围下载
	}

	// 执行获取对象的操作并处理结果
	result, err := client.GetObject(context.TODO(), request)
	if err != nil {
		log.Fatalf("failed to get object %v", err)
	}
	defer result.Body.Close() // 确保在函数结束时关闭响应体

	log.Printf("get object result:%#v\n", result)

	// 读取对象的内容
	data, _ := io.ReadAll(result.Body)
	log.Printf("body:%s\n", data)
}
```

## 相关文档

-   关于范围下载的完整示例代码，请参见[GitHub示例](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2/blob/master/sample/get_object.go)。

-   关于范围下载的API接口说明，请参见[GetObject](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Client.GetObject)。