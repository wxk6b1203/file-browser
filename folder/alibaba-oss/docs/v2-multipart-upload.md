OSS提供的分片上传（Multipart Upload）功能，将要上传的较大文件（Object）分成多个分片（Part）来分别上传，上传完成后再调用CompleteMultipartUpload接口将这些Part组合成一个Object。

## **注意事项**

-   本文示例代码以华东1（杭州）的地域ID`cn-hangzhou`为例，默认使用外网Endpoint，如果您希望通过与OSS同地域的其他阿里云产品访问OSS，请使用内网Endpoint。关于OSS支持的Region与Endpoint的对应关系，请参见[地域和Endpoint](https://help.aliyun.com/zh/oss/user-guide/regions-and-endpoints#concept-zt4-cvy-5db)。

-   本文以从环境变量读取访问凭证为例。如何配置访问凭证，请参见[配置访问凭证](https://help.aliyun.com/zh/oss/developer-reference/configure-access-credentials-by-using-oss-sdk-for-go-v2)。

-   要分片上传，您必须有`oss:PutObject`权限。具体操作，请参见[为RAM用户授予自定义的权限策略](https://help.aliyun.com/zh/oss/user-guide/common-examples-of-ram-policies#section-ucu-jv0-zip)。


## 分片上传流程

分片上传（Multipart Upload）分为以下三个步骤：

1.  初始化一个分片上传事件。

    调用Client.InitiateMultipartUpload方法返回OSS创建的全局唯一的uploadID。

2.  上传分片。

    调用Client.UploadPart方法上传分片数据。

    **说明**

    -   对于同一个uploadID，分片号（partNumber）标识了该分片在整个文件内的相对位置。如果使用同一个分片号上传了新的数据，那么OSS上该分片已有的数据将会被覆盖。

    -   OSS将收到的分片数据的MD5值放在ETag头内返回给用户。

    -   OSS计算上传数据的[MD5](https://help.aliyun.com/zh/oss/user-guide/can-i-use-etag-values-as-oss-md5-hashes-to-check-data-consistency)值，并与SDK计算的MD5值比较，如果不一致则返回InvalidDigest错误码。


3.  完成分片上传。

    所有分片上传完成后，调用Client.CompleteMultipartUpload方法将所有分片合并成完整的文件。


## **示例代码**

以下代码展示如何将本地的大文件分割成多个分片文件并发上传到存储空间，然后合并成完整的文件对象。

```
package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"io"
	"log"
	"os"
	"sync"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// 定义全局变量
var (
	region     string // 存储区域
	bucketName string // 源存储空间名称
	objectName string // 源对象名称

)

// init函数用于初始化命令行参数
func init() {
	flag.StringVar(&region, "region", "", "The region in which the bucket is located.")
	flag.StringVar(&bucketName, "bucket", "", "The name of the source bucket.")
	flag.StringVar(&objectName, "object", "", "The name of the source object.")
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 定义上传ID
	var uploadId string

	// 检查源存储空间名称是否为空
	if len(bucketName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, source bucket name required")
	}

	// 检查存储区域是否为空
	if len(region) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, region required")
	}

	// 检查源对象名称是否为空
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

	// 初始化分片上传请求
	initRequest := &oss.InitiateMultipartUploadRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr(objectName),
	}
	initResult, err := client.InitiateMultipartUpload(context.TODO(), initRequest)
	if err != nil {
		log.Fatalf("failed to initiate multipart upload %v", err)
	}

	// 打印初始化分片上传的结果
	log.Printf("initiate multipart upload result:%#v\n", *initResult.UploadId)
	uploadId = *initResult.UploadId

	// 初始化等待组和互斥锁
	var wg sync.WaitGroup
	var parts []oss.UploadPart
	count := 3
	var mu sync.Mutex

	// 读取本地文件内容到内存，将yourLocalFile替换为实际的本地文件名和路径
	file, err := os.Open("yourLocalFile")
	if err != nil {
		log.Fatalf("failed to open local file %v", err)
	}
	defer file.Close()

	bufReader := bufio.NewReader(file)
	content, err := io.ReadAll(bufReader)
	if err != nil {
		log.Fatalf("failed to read local file %v", err)
	}
	log.Printf("file size: %d\n", len(content))

	// 计算每个分片的大小
	chunkSize := len(content) / count
	if chunkSize == 0 {
		chunkSize = 1
	}

	// 启动多个goroutine进行分片上传
	for i := 0; i < count; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == count-1 {
			end = len(content)
		}

		wg.Add(1)
		go func(partNumber int, start, end int) {
			defer wg.Done()

			// 创建分片上传请求
			partRequest := &oss.UploadPartRequest{
				Bucket:     oss.Ptr(bucketName),                 // 目标存储空间名称
				Key:        oss.Ptr(objectName),                 // 目标对象名称
				PartNumber: int32(partNumber),                   // 分片编号
				UploadId:   oss.Ptr(uploadId),                   // 上传ID
				Body:       bytes.NewReader(content[start:end]), // 分片内容
			}

			// 发送分片上传请求
			partResult, err := client.UploadPart(context.TODO(), partRequest)
			if err != nil {
				log.Fatalf("failed to upload part %d: %v", partNumber, err)
			}

			// 记录分片上传结果
			part := oss.UploadPart{
				PartNumber: partRequest.PartNumber,
				ETag:       partResult.ETag,
			}

			// 使用互斥锁保护共享数据
			mu.Lock()
			parts = append(parts, part)
			mu.Unlock()
		}(i+1, start, end)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 完成分片上传请求
	request := &oss.CompleteMultipartUploadRequest{
		Bucket:   oss.Ptr(bucketName),
		Key:      oss.Ptr(objectName),
		UploadId: oss.Ptr(uploadId),
		CompleteMultipartUpload: &oss.CompleteMultipartUpload{
			Parts: parts,
		},
	}
	result, err := client.CompleteMultipartUpload(context.TODO(), request)
	if err != nil {
		log.Fatalf("failed to complete multipart upload %v", err)
	}

	// 打印完成分片上传的结果
	log.Printf("complete multipart upload result:%#v\n", result)
}
```

## **常见使用场景**

### **将指定长度的随机字符串进行分片上传**

以下代码展示了将400kb的随机字符串分割成3个分片文件并发上传到存储空间，然后合并成完整的文件对象。

```
package main

import (
	"bufio"
	"context"
	"flag"
	"io"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// 定义全局变量
var (
	region     string                                                                     // 存储区域
	bucketName string                                                                     // 存储空间名称
	objectName string                                                                     // 对象名称
	letters    = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") // 用于生成随机字符串的字符集
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

	// 定义上传ID
	var uploadId string

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

	// 创建初始化分片上传的请求
	initRequest := &oss.InitiateMultipartUploadRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
	}

	// 执行初始化分片上传的操作并处理结果
	initResult, err := client.InitiateMultipartUpload(context.TODO(), initRequest)
	if err != nil {
		log.Fatalf("failed to initiate multipart upload %v", err)
	}

	// 打印初始化分片上传的结果
	log.Printf("initiate multipart upload result:%#v\n", initResult)
	uploadId = *initResult.UploadId

	// 初始化等待组和互斥锁
	var wg sync.WaitGroup
	var parts []oss.UploadPart
	count := 3
	body := randBody(400000) // 生成400KB的随机字符串
	reader := strings.NewReader(body)
	bufReader := bufio.NewReader(reader)
	content, _ := io.ReadAll(bufReader)
	partSize := len(body) / count
	var mu sync.Mutex

	// 启动多个goroutine进行分片上传
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(partNumber int, partSize int, i int) {
			defer wg.Done()

			// 创建分片上传请求
			partRequest := &oss.UploadPartRequest{
				Bucket:     oss.Ptr(bucketName),                                             // 存储空间名称
				Key:        oss.Ptr(objectName),                                             // 对象名称
				PartNumber: int32(partNumber),                                               // 分片编号
				UploadId:   oss.Ptr(uploadId),                                               // 上传ID
				Body:       strings.NewReader(string(content[i*partSize : (i+1)*partSize])), // 分片内容
			}

			// 发送分片上传请求
			partResult, err := client.UploadPart(context.TODO(), partRequest)
			if err != nil {
				log.Fatalf("failed to upload part %d: %v", partNumber, err)
			}

			// 记录分片上传结果
			part := oss.UploadPart{
				PartNumber: partRequest.PartNumber,
				ETag:       partResult.ETag,
			}

			// 使用互斥锁保护共享数据
			mu.Lock()
			parts = append(parts, part)
			mu.Unlock()
		}(i+1, partSize, i)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 打印分片上传成功的消息
	log.Println("upload part success!")

	// 创建完成分片上传请求
	request := &oss.CompleteMultipartUploadRequest{
		Bucket:   oss.Ptr(bucketName),
		Key:      oss.Ptr(objectName),
		UploadId: oss.Ptr(uploadId),
		CompleteMultipartUpload: &oss.CompleteMultipartUpload{
			Parts: parts,
		},
	}

	// 完成分片上传并处理结果
	result, err := client.CompleteMultipartUpload(context.TODO(), request)
	if err != nil {
		log.Fatalf("failed to complete multipart upload %v", err)
	}
	log.Printf("complete multipart upload result:%#v\n", result)
}

// randBody生成指定长度的随机字符串
func randBody(n int) string {
	b := make([]rune, n)
	randMarker := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letters[randMarker.Intn(len(letters))]
	}
	return string(b)
}
```

### **取消指定的分片上传事件**

当您遇到以下场景时，可以使用`Client.AbortMultipartUpload`方法取消分片上传事件。

1.  **文件出错**：

    -   如果在上传过程中发现文件有错误，如文件损坏或包含恶意代码，您可以选择取消上传以避免潜在的风险。

2.  **网络不稳定**：

    -   当网络连接不稳定或中断时，可能会导致上传过程中的分片丢失或损坏，您可以选择取消上传并重新开始，以确保数据的完整性和一致性。

3.  **资源限制**：

    -   当您的存储空间有限，而上传的文件过大，您可以取消上传以释放存储资源，将资源分配给其他更重要的任务。

4.  **误操作：**

    -   当不小心启动了一个不必要的上传任务，或者上传了一个错误的文件版本，您可以取消此次上传事件


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
	bucketName string // 源存储空间名称
	objectName string // 源对象名称

)

// init函数用于初始化命令行参数
func init() {
	flag.StringVar(&region, "region", "", "The region in which the bucket is located.")
	flag.StringVar(&bucketName, "bucket", "", "The name of the source bucket.")
	flag.StringVar(&objectName, "object", "", "The name of the source object.")
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 定义上传ID
	var uploadId string

	// 检查源存储空间名称是否为空
	if len(bucketName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, source bucket name required")
	}

	// 检查存储区域是否为空
	if len(region) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, region required")
	}

	// 检查源对象名称是否为空
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

	// 初始化分片上传请求
	initRequest := &oss.InitiateMultipartUploadRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr(objectName),
	}

	// 执行初始化分片上传请求
	initResult, err := client.InitiateMultipartUpload(context.TODO(), initRequest)
	if err != nil {
		log.Fatalf("failed to initiate multipart upload %v", err)
	}

	// 打印初始化分片上传的结果
	log.Printf("initiate multipart upload result:%#v\n", *initResult.UploadId)
	uploadId = *initResult.UploadId

	// 创建一个AbortMultipartUploadRequest请求
	request := &oss.AbortMultipartUploadRequest{
		Bucket:   oss.Ptr(bucketName), // 存储空间名称
		Key:      oss.Ptr(objectName), // 对象名称
		UploadId: oss.Ptr(uploadId),   // 上传uploadId
	}
	// 执行请求并处理结果
	result, err := client.AbortMultipartUpload(context.TODO(), request)
	if err != nil {
		log.Fatalf("failed to abort multipart upload %v", err)
	}
	log.Printf("abort multipart upload result:%#v\n", result)

}
```

### **列举指定的分片上传事件中已经成功上传的分片**

当您遇到以下场景时，可以使用`Client.NewListPartsPaginator`分页器列举某个分片上传事件中已经成功上传的分片。

**监控上传进度：**

1.  **大型文件上传**：

    -   当上传非常大的文件时，您通过列举已上传的分片，确保上传过程按照预期进行，并及时发现问题。

2.  **断点续传**：

    -   在网络不稳定或上传过程中发生中断时，您可以通过查看已上传的分片来决定是否需要重试上传未完成的部分，从而实现断点续传。

3.  **故障排除**：

    -   如果上传过程中出现错误，通过检查已上传的分片，您可以快速定位问题所在，比如某个特定分片上传失败，然后针对性地解决问题。

4.  **资源管理**：

    -   对于需要严格控制资源使用情况的场景，通过监控上传进度，可以更好地管理存储空间和带宽资源，确保资源的有效利用。


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
	bucketName string // 源存储空间名称
	objectName string // 源对象名称

)

// init函数用于初始化命令行参数
func init() {
	flag.StringVar(&region, "region", "", "The region in which the bucket is located.")
	flag.StringVar(&bucketName, "bucket", "", "The name of the source bucket.")
	flag.StringVar(&objectName, "object", "", "The name of the source object.")
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 定义上传ID
	var uploadId string

	// 检查源存储空间名称是否为空
	if len(bucketName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, source bucket name required")
	}

	// 检查存储区域是否为空
	if len(region) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, region required")
	}

	// 检查源对象名称是否为空
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

	// 初始化分片上传请求
	initRequest := &oss.InitiateMultipartUploadRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr(objectName),
	}

	// 执行初始化分片上传请求
	initResult, err := client.InitiateMultipartUpload(context.TODO(), initRequest)
	if err != nil {
		log.Fatalf("failed to initiate multipart upload %v", err)
	}

	// 打印初始化分片上传的结果
	log.Printf("initiate multipart upload result:%#v\n", *initResult.UploadId)
	uploadId = *initResult.UploadId

	// 创建列出分片的请求
	request := &oss.ListPartsRequest{
		Bucket:   oss.Ptr(bucketName), // 存储空间名称
		Key:      oss.Ptr(objectName), // 对象名称
		UploadId: oss.Ptr(uploadId),   // 上传uploadId
	}

	// 创建分页器
	p := client.NewListPartsPaginator(request)

	// 初始化页码计数器
	var i int
	log.Println("List Parts:")

	// 遍历分页器中的每一页
	for p.HasNext() {
		i++

		// 获取下一页的数据
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Fatalf("failed to get page %v, %v", i, err)
		}

		// 打印该页中的每个分片的信息
		for _, part := range page.Parts {
			log.Printf("Part Number: %v, ETag: %v, Last Modified: %v, Size: %v, HashCRC64: %v\n",
				part.PartNumber,
				oss.ToString(part.ETag),
				oss.ToTime(part.LastModified),
				part.Size,
				oss.ToString(part.HashCRC64))
		}
	}

}
```

### **列举分片上传事件**

当您遇到以下场景时，可以使用`Client.NewListMultipartUploadsPaginator`分页器列举某个存储空间所有进行中的分片上传事件。

**监控场景：**

1.  **批量文件上传管理**：

    -   当需要上传大量文件时，为了确保所有文件都能正确完成分片上传，您可以使用`ListMultipartUploads`方法来实时监控所有的分片上传活动。

2.  **故障检测与恢复**：

    -   在上传过程中如果遇到网络问题或其他故障，可能导致部分分片未能成功上传。通过监控正在进行中的分片上传事件，可以及时发现这些问题，并采取措施恢复上传。

3.  **资源优化与管理**：

    -   在大规模的文件上传过程中，监控正在进行中的分片上传事件可以帮助优化资源分配，例如根据上传进度调整带宽使用或优化上传策略。

4.  **数据迁移：**

    -   在进行大规模的数据迁移项目时，监控所有正在进行的分片上传事件可以确保迁移任务的顺利进行，及时发现并解决任何潜在的问题。


#### **参数设置**

| **参数** | **说明** |
| --- | --- |
| Delimiter | 用于对Object名字进行分组的字符。所有名字包含指定的前缀且第一次出现Delimiter字符之间的Object作为一组元素。 |
| MaxUploads | 限定此次返回分片上传事件的最大数目，默认值和最大值均为1000。 |
| KeyMarker | 所有文件名称的字母序大于KeyMarker参数值的分片上传事件，可以与UploadIDMarker参数一同使用来指定返回结果的起始位置。 |
| Prefix | 限定返回的文件名称必须以指定的Prefix作为前缀。注意使用Prefix查询时，返回的文件名称中仍会包含Prefix。 |
| UploadIDMarker | 与KeyMarker参数一同使用来指定返回结果的起始位置。 - 如果KeyMarker参数未设置，则OSS忽略该参数。 - 如果KeyMarker参数被设置，查询结果中包含： - 所有Object名字的字典序大于KeyMarker参数值的分片上传事件。 - Object名字等于KeyMarker参数值，但是UploadID比UploadIDMarker参数值大的分片上传事件。 |

-   指定前缀为file且最多返回100条结果数据

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
    	bucketName string // 源存储空间名称
    	objectName string // 源对象名称
    
    )
    
    // init函数用于初始化命令行参数
    func init() {
    	flag.StringVar(&region, "region", "", "The region in which the bucket is located.")
    	flag.StringVar(&bucketName, "bucket", "", "The name of the source bucket.")
    	flag.StringVar(&objectName, "object", "", "The name of the source object.")
    }
    
    func main() {
    	// 解析命令行参数
    	flag.Parse()
    
    	// 检查源存储空间名称是否为空
    	if len(bucketName) == 0 {
    		flag.PrintDefaults()
    		log.Fatalf("invalid parameters, source bucket name required")
    	}
    
    	// 检查存储区域是否为空
    	if len(region) == 0 {
    		flag.PrintDefaults()
    		log.Fatalf("invalid parameters, region required")
    	}
    
    	// 检查源对象名称是否为空
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
    
    	// 创建列出分片上传的请求
    	request := &oss.ListMultipartUploadsRequest{
    		Bucket:     oss.Ptr(bucketName), // 存储空间名称
    		MaxUploads: 100,                 // 指定最多返回100条结果数据
    		Prefix:     oss.Ptr("file"),     // 指定前缀为file
    	}
    
    	// 创建分页器
    	p := client.NewListMultipartUploadsPaginator(request)
    
    	var i int
    	log.Println("List Multipart Uploads:")
    
    	// 遍历分页器中的每一页
    	for p.HasNext() {
    		i++
    
    		// 获取下一页的数据
    		page, err := p.NextPage(context.TODO())
    		if err != nil {
    			log.Fatalf("failed to get page %v, %v", i, err)
    		}
    
    		// 打印该页中的每个分片上传的信息
    		for _, u := range page.Uploads {
    			log.Printf("Upload key: %v, upload id: %v, initiated: %v\n", oss.ToString(u.Key), oss.ToString(u.UploadId), oss.ToTime(u.Initiated))
    		}
    	}
    
    }
    ```


### 分片上传并设置回调

以下代码实现了将一个 400KB 大小的文件分割成 3 个片段并发上传到阿里云 OSS，上传完成后将这些片段合并成完整的文件对象，并在合并完成后触发回调通知。

```
package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"io"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

var (
	region     string
	bucketName string
	objectName string
	letters    = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func init() {
	flag.StringVar(&region, "region", "", "The region in which the bucket is located.")
	flag.StringVar(&bucketName, "bucket", "", "The name of the bucket.")
	flag.StringVar(&objectName, "object", "", "The name of the object.")
}

func main() {
	flag.Parse()
	if len(bucketName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, bucket name required")
	}

	if len(region) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, region required")
	}

	if len(objectName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, object name required")
	}

	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	client := oss.NewClient(cfg)

	initRequest := &oss.InitiateMultipartUploadRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr(objectName),
	}
	initResult, err := client.InitiateMultipartUpload(context.TODO(), initRequest)

	// 定义回调参数
	callbackMap := map[string]string{
		"callbackUrl":      "https://example.com:23450",                                                                  // 设置回调服务器的URL，例如https://example.com:23450。
		"callbackBody":     "bucket=${bucket}&object=${object}&size=${size}&my_var_1=${x:my_var1}&my_var_2=${x:my_var2}", // 设置回调请求体。
		"callbackBodyType": "application/x-www-form-urlencoded",                                                          //设置回调请求体类型。
	}

	// 将回调参数转换为JSON并进行Base64编码，以便将其作为回调参数传递
	callbackStr, err := json.Marshal(callbackMap)
	if err != nil {
		log.Fatalf("failed to marshal callback map: %v", err)
	}
	callbackBase64 := base64.StdEncoding.EncodeToString(callbackStr)

	callbackVarMap := map[string]string{}
	callbackVarMap["x:my_var1"] = "thi is var 1"
	callbackVarMap["x:my_var2"] = "thi is var 2"
	callbackVarStr, err := json.Marshal(callbackVarMap)
	if err != nil {
		log.Fatalf("failed to marshal callback var: %v", err)
	}
	callbackVarBase64 := base64.StdEncoding.EncodeToString(callbackVarStr)

	var wg sync.WaitGroup
	var parts []oss.UploadPart
	count := 3
	body := randBody(400000)
	reader := strings.NewReader(body)
	bufReader := bufio.NewReader(reader)
	content, _ := io.ReadAll(bufReader)
	partSize := len(body) / count
	var mu sync.Mutex
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(partNumber int, partSize int, i int) {
			defer wg.Done()
			partRequest := &oss.UploadPartRequest{
				Bucket:     oss.Ptr(bucketName),
				Key:        oss.Ptr(objectName),
				PartNumber: int32(partNumber),
				UploadId:   initResult.UploadId,
				Body:       strings.NewReader(string(content[i*partSize : (i+1)*partSize])),
			}
			partResult, err := client.UploadPart(context.TODO(), partRequest)
			if err != nil {
				log.Fatalf("failed to upload part %d: %v", partNumber, err)
			}
			part := oss.UploadPart{
				PartNumber: partRequest.PartNumber,
				ETag:       partResult.ETag,
			}
			mu.Lock()
			parts = append(parts, part)
			mu.Unlock()
		}(i+1, partSize, i)
	}
	wg.Wait()

	request := &oss.CompleteMultipartUploadRequest{
		Bucket:   oss.Ptr(bucketName),
		Key:      oss.Ptr(objectName),
		UploadId: initResult.UploadId,
		CompleteMultipartUpload: &oss.CompleteMultipartUpload{
			Parts: parts,
		},
		Callback:    oss.Ptr(callbackBase64), // 填写回调参数
		CallbackVar: oss.Ptr(callbackVarBase64),
	}
	result, err := client.CompleteMultipartUpload(context.TODO(), request)
	if err != nil {
		log.Fatalf("failed to complete multipart upload %v", err)
	}
	log.Printf("complete multipart upload result:%#v\n", result)
}

func randBody(n int) string {
	b := make([]rune, n)
	randMarker := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letters[randMarker.Intn(len(letters))]
	}
	return string(b)
}
```

### **分片上传并显示进度条**

```
package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// 定义全局变量
var (
	region     string // 存储区域
	bucketName string // 源存储空间名称
	objectName string // 源对象名称

)

// init函数用于初始化命令行参数
func init() {
	flag.StringVar(&region, "region", "", "The region in which the bucket is located.")
	flag.StringVar(&bucketName, "bucket", "", "The name of the source bucket.")
	flag.StringVar(&objectName, "object", "", "The name of the source object.")
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 定义上传ID
	var uploadId string

	// 检查源存储空间名称是否为空
	if len(bucketName) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, source bucket name required")
	}

	// 检查存储区域是否为空
	if len(region) == 0 {
		flag.PrintDefaults()
		log.Fatalf("invalid parameters, region required")
	}

	// 检查源对象名称是否为空
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

	// 初始化分段上传请求
	initRequest := &oss.InitiateMultipartUploadRequest{
		Bucket:  oss.Ptr(bucketName),
		Key:     oss.Ptr(objectName),
	}

	// 执行初始化分段上传请求
	initResult, err := client.InitiateMultipartUpload(context.TODO(), initRequest)
	if err != nil {
		log.Fatalf("failed to initiate multipart upload %v", err)
	}

	// 打印初始化分段上传的结果
	log.Printf("initiate multipart upload result:%#v\n", *initResult.UploadId)
	uploadId = *initResult.UploadId

	// 初始化等待组和互斥锁
	var wg sync.WaitGroup
	var parts []oss.UploadPart
	count := 5
	var mu sync.Mutex

	// 读取本地文件内容到内存，将yourLocalFile替换为实际的本地文件名和路径
	file, err := os.Open("/Users/yourLocalPath/yourFileName")
	if err != nil {
		log.Fatalf("failed to open local file %v", err)
	}
	defer file.Close()

	bufReader := bufio.NewReader(file)
	content, err := io.ReadAll(bufReader)
	if err != nil {
		log.Fatalf("failed to read local file %v", err)
	}
	log.Printf("file size: %d\n", len(content))

	// 计算每个分片的大小
	chunkSize := len(content) / count
	if chunkSize == 0 {
		chunkSize = 1
	}

	// 启动多个goroutine进行分片上传
	for i := 0; i < count; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == count-1 {
			end = len(content)
		}

		wg.Add(1)
		go func(partNumber int, start, end int) {
			defer wg.Done()

			// 创建分片上传请求
			partRequest := &oss.UploadPartRequest{
				Bucket:     oss.Ptr(bucketName),                 // 目标存储空间名称
				Key:        oss.Ptr(objectName),                 // 目标对象名称
				PartNumber: int32(partNumber),                   // 分片编号
				UploadId:   oss.Ptr(uploadId),                   // 上传ID
				Body:       bytes.NewReader(content[start:end]), // 分片内容
				ProgressFn: func(increment, transferred, total int64) {
					fmt.Printf("increment:%v, transferred:%v, total:%v\n", increment, transferred, total)
				}, // 进度回调函数，用于显示上传进度
			}

			// 发送分片上传请求
			partResult, err := client.UploadPart(context.TODO(), partRequest)
			if err != nil {
				log.Fatalf("failed to upload part %d: %v", partNumber, err)
			}

			log.Printf("successfully uploaded part %d (start: %d, end: %d)", partNumber, start, end)

			// 记录分片上传结果
			part := oss.UploadPart{
				PartNumber: partRequest.PartNumber,
				ETag:       partResult.ETag,
			}

			// 使用互斥锁保护共享数据
			mu.Lock()
			parts = append(parts, part)
			mu.Unlock()
		}(i+1, start, end)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 完成分段上传请求
	request := &oss.CompleteMultipartUploadRequest{
		Bucket:   oss.Ptr(bucketName),
		Key:      oss.Ptr(objectName),
		UploadId: oss.Ptr(uploadId),
		CompleteMultipartUpload: &oss.CompleteMultipartUpload{
			Parts: parts,
		},
	}
	result, err := client.CompleteMultipartUpload(context.TODO(), request)
	if err != nil {
		log.Fatalf("failed to complete multipart upload %v", err)
	}

	// 打印完成分段上传的versionId
	log.Printf("complete multipart upload result versionId:%#v\n", result)
}
```

## 相关文档

-   关于分片上传的完整示例代码，请参见[GitHub示例](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2/blob/master/sample/upload_part.go)。

-   分片上传的完整实现涉及三个API接口，详情如下：

    -   关于初始化分片上传事件的API接口说明，请参见[InitiateMultipartUpload](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Client.InitiateMultipartUpload)。

    -   关于分片上传Part的API接口说明，请参见[UploadPart](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Client.UploadPart)。

    -   关于完成分片上传的API接口说明，请参见[CompleteMultipartUpload](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Client.CompleteMultipartUpload)。

-   关于取消分片上传事件的API接口说明，请参见[AbortMultipartUpload](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Client.AbortMultipartUpload)。

-   关于列举已上传分片的API接口说明，请参见[NewListPartsPaginator](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Client.NewListPartsPaginator)。

-   关于列举所有执行中的分片上传事件（即已初始化但尚未完成或已取消的分片上传事件）的API接口说明，请参见[NewListMultipartUploadsPaginator](https://pkg.go.dev/github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss#Client.NewListMultipartUploadsPaginator)。