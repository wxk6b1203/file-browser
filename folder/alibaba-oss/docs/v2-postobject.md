OSS表单上传允许网页应用通过标准HTML表单直接将文件上传至OSS。本文介绍如何使用Go SDK V2生成Post签名和Post Policy等信息，并调用HTTP Post方法上传文件到OSS。

## **注意事项**

-   本文示例代码以华东1（杭州）的地域ID`cn-hangzhou`为例，默认使用外网Endpoint，如果您希望通过与OSS同地域的其他阿里云产品访问OSS，请使用内网Endpoint。关于OSS支持的Region与Endpoint的对应关系，请参见[OSS地域和访问域名](https://help.aliyun.com/zh/oss/user-guide/regions-and-endpoints#concept-zt4-cvy-5db)。

-   本文以从环境变量读取访问凭证为例。如何配置访问凭证，请参见[配置访问凭证](https://help.aliyun.com/zh/oss/developer-reference/configure-access-credentials-by-using-oss-sdk-for-go-v2)。

-   通过表单上传的方式上传的Object大小不能超过5 GB。


## **示例代码**

以下代码示例实现了表单上传的完整过程，主要步骤如下：

1.  **创建****Post Policy**：定义上传请求的有效时间和条件，包括存储桶名称、签名版本、凭证信息、请求日期和请求体长度范围。

2.  **序列化并编码Policy**：将Policy序列化为JSON字符串，并进行Base64编码。

3.  **生成签名密钥**：使用HMAC-SHA256算法生成签名密钥，包括日期、区域、产品和请求类型。

4.  **计算签名**：使用生成的密钥对Base64编码后的Policy字符串进行签名，并将签名结果转换为十六进制字符串。

5.  **构建请求体**：创建一个multipart表单写入器，添加对象键、策略、签名版本、凭证信息、请求日期和签名到表单中，并将要上传的数据写入表单。

6.  **创建并执行请求**：创建一个HTTP POST请求，设置请求头，并发送请求，检查响应状态码确保请求成功。


```
package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"hash"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

var (
	region     string // 存储区域
	bucketName string // 存储桶名称
	objectName string // 对象名称
	product    = "oss"
)

// init 函数用于初始化命令行参数
func init() {
	flag.StringVar(&region, "region", "", "The region in which the bucket is located.")
	flag.StringVar(&bucketName, "bucket", "", "The name of the bucket.")
	flag.StringVar(&objectName, "object", "", "The name of the object.")
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 检查存储桶名称是否为空
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

	// 创建凭证提供者
	credentialsProvider := credentials.NewEnvironmentVariableCredentialsProvider()
	cred, err := credentialsProvider.GetCredentials(context.TODO())
	if err != nil {
		log.Fatalf("GetCredentials fail, err:%v", err)
	}

	// 填写要上传的内容
	content := "hi oss"

	// 构建Post Policy
	utcTime := time.Now().UTC()
	date := utcTime.Format("20060102")
	expiration := utcTime.Add(1 * time.Hour)
	policyMap := map[string]any{
		"expiration": expiration.Format("2006-01-02T15:04:05.000Z"),
		"conditions": []any{
			map[string]string{"bucket": bucketName},
			map[string]string{"x-oss-signature-version": "OSS4-HMAC-SHA256"},
			map[string]string{"x-oss-credential": fmt.Sprintf("%v/%v/%v/%v/aliyun_v4_request",
				cred.AccessKeyID, date, region, product)}, // 凭证
			map[string]string{"x-oss-date": utcTime.Format("20060102T150405Z")},
			// 其他条件
			[]any{"content-length-range", 1, 1024},
			// []any{"eq", "$success_action_status", "201"},
			// []any{"starts-with", "$key", "user/eric/"},
			// []any{"in", "$content-type", []string{"image/jpg", "image/png"}},
			// []any{"not-in", "$cache-control", []string{"no-cache"}},
		},
	}

	// 将Post Policy序列化为JSON字符串
	policy, err := json.Marshal(policyMap)
	if err != nil {
		log.Fatalf("json.Marshal fail, err:%v", err)
	}

	// 将Post Policy编码为Base64字符串
	stringToSign := base64.StdEncoding.EncodeToString([]byte(policy))

	// 生成签名密钥
	hmacHash := func() hash.Hash { return sha256.New() }
	signingKey := "aliyun_v4" + cred.AccessKeySecret
	h1 := hmac.New(hmacHash, []byte(signingKey))
	io.WriteString(h1, date)
	h1Key := h1.Sum(nil)

	h2 := hmac.New(hmacHash, h1Key)
	io.WriteString(h2, region)
	h2Key := h2.Sum(nil)

	h3 := hmac.New(hmacHash, h2Key)
	io.WriteString(h3, product)
	h3Key := h3.Sum(nil)

	h4 := hmac.New(hmacHash, h3Key)
	io.WriteString(h4, "aliyun_v4_request")
	h4Key := h4.Sum(nil)

	// 计算Post签名
	h := hmac.New(hmacHash, h4Key)
	io.WriteString(h, stringToSign)
	signature := hex.EncodeToString(h.Sum(nil))

	// 构建Post请求体
	bodyBuf := &bytes.Buffer{}

	// 创建一个multipart表单写入器，用于构建请求体
	bodyWriter := multipart.NewWriter(bodyBuf)

	// 设置对象信息，包括键和元数据
	bodyWriter.WriteField("key", objectName) // 设置对象键（文件名）
	// bodyWriter.WriteField("x-oss-", value) // 设置元数据（可选）

	// 设置Base64编码后的策略字符串
	bodyWriter.WriteField("policy", stringToSign)

	// 设置签名版本
	bodyWriter.WriteField("x-oss-signature-version", "OSS4-HMAC-SHA256")

	// 设置凭证信息
	bodyWriter.WriteField("x-oss-credential", fmt.Sprintf("%v/%v/%v/%v/aliyun_v4_request", cred.AccessKeyID, date, region, product))

	// 设置请求日期
	bodyWriter.WriteField("x-oss-date", utcTime.Format("20060102T150405Z"))

	// 设置签名
	bodyWriter.WriteField("x-oss-signature", signature)

	// 创建一个名为"file"的表单字段，用于上传文件内容
	w, _ := bodyWriter.CreateFormField("file")

	// 将要上传的数据写入表单字段
	w.Write([]byte(content))

	// 关闭表单写入器，确保所有数据都被正确写入请求体
	bodyWriter.Close()

	// 创建Post请求
	req, _ := http.NewRequest("POST", fmt.Sprintf("http://%v.oss-%v.aliyuncs.com/", bucketName, region), bodyBuf)
	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	// 执行Post请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Do fail, err:%v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode/100 != 2 {
		log.Fatalf("Post Object Fail, status code:%v, reason:%v", resp.StatusCode, resp.Status)
	}

	log.Printf("post object done, status code:%v, request id:%v\n", resp.StatusCode, resp.Header.Get("X-Oss-Request-Id"))
}
```

## **相关文档**

-   关于表单上传的完整示例，请参见[Github示例](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2/blob/master/sample/post_object.go)。