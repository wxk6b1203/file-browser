您可以将任何类型的文件上传到OSS的Bucket中，包括图片、文档、视频等。当您将文件上传到OSS时，文件将作为OSS Object进行存储。Object包含文件数据本身和描述该对象的元数据。在一个Bucket中，您可以上传大量的Object。

## **上传方式**

OSS提供以下文件上传方式：

-   [简单上传](https://help.aliyun.com/zh/oss/user-guide/simple-upload)：适用于上传小文件，文件大小不超过5 GB，操作简单，通过调用OSS提供的PutObject接口一次性上传整个文件，无需特殊配置。

-   [分片上传](https://help.aliyun.com/zh/oss/user-guide/multipart-upload)：适用于上传大文件，文件大小不超过48.8 TB，通过调用OSS提供的多个接口，包括InitiateMultipartUpload、UploadPart、CompleteMultipartUpload，将文件分割成多个分片并行上传，然后在上传完成后合并最终上传整个文件。如果因为网络环境不稳定等情况导致上传中断，客户端需要手动记录哪些分片上传失败以进行重传。

-   [追加上传](https://help.aliyun.com/zh/oss/user-guide/append-upload-11)：适用于上传需要持续添加数据的文件，例如视频流，文件大小不超过5 GB，通过调用OSS提供AppendObject接口上传文件，并生成Appendable类型的Object。Appendable类型Object后面允许直接追加内容，且每次追加上传的数据都能够即时可读。非Appendable类型的Object不支持追加上传。

-   [断点续传上传](https://help.aliyun.com/zh/oss/user-guide/resumable-upload)：适用于在网络环境不稳定的情况下上传大文件，文件大小不超过48.8 TB，通过调用OSS SDK基于分片上传封装的方法，例如Java SDK的`uploadFile`，实现在客户端本地自动记录上传进度，然后在中断后从上次停止的地方继续上传。

-   [表单上传](https://help.aliyun.com/zh/oss/user-guide/form-upload)：适用于让用户在HTML网页中上传Object，文件大小不超过5 GB，通过发起HTTP POST请求上传文件到OSS。您可以借助服务端生成的PostPolicy限制客户端上传的文件，例如限制文件大小、文件类型。

-   [使用预签名URL上传文件](https://help.aliyun.com/zh/oss/user-guide/upload-files-using-presigned-urls)：适用于授权第三方上传文件的场景，文件大小不超过5GB。文件拥有者生成具有时效的预签名URL，他人无需密钥即可安全上传，过期自动失效。


## **相关文档**

-   如果您希望在上传文件时监控并显示数据传输的进度，您可以利用OSS SDK提供的进度监听功能实现一个进度条来反馈实时的上传状态。更多信息，请参见[上传进度条](https://help.aliyun.com/zh/oss/user-guide/upload-progress-bar)。

-   在文件上传到OSS后，您可以通过上传回调向指定的应用服务器发起回调请求。更多信息，请参见[上传回调](https://help.aliyun.com/zh/oss/user-guide/upload-callbacks-12#concept-ywd-dlb-5db)。

-   如果需要控制上传的文件的缓存、下载、数据处理等行为，您可以在上传时携带Object Meta信息，例如Content-Type等标准HTTP头。更多信息，请参见[设置文件元数据](https://help.aliyun.com/zh/oss/user-guide/manage-object-metadata-10/#concept-lkf-swy-5db)。

-   推荐使用客户端直传的方式将文件上传到OSS。相对于服务端代理上传，客户端直传避免了业务服务器中转文件，提高了上传速度，节省了服务器资源。更多信息，请参见[客户端直传](https://help.aliyun.com/zh/oss/user-guide/uploading-objects-to-oss-directly-from-clients/)。

-   如果您希望对已上传的图片进行添加图片水印、转换格式、获取图片信息等操作，请参见[图片处理](https://help.aliyun.com/zh/oss/user-guide/overview-17/)。

-   如果您希望对上传的视频等进行视频转码、视频截帧等处理，请参见[音视频处理](https://help.aliyun.com/zh/oss/user-guide/introduction-2/)。

-   如果您希望对上传的PPT、Word等格式的文档进行在线预览或在线编辑，请参见[WebOffice在线预览](https://help.aliyun.com/zh/oss/user-guide/online-object-preview)和[WebOffice在线编辑](https://help.aliyun.com/zh/oss/user-guide/online-object-editing)。

-   文件上传完成后，您可以生成签名URL，以便将该URL转给第三方实现授权访问。更多信息，请参见[使用预签名URL下载或预览文件](https://help.aliyun.com/zh/oss/user-guide/how-to-obtain-the-url-of-a-single-object-or-the-urls-of-multiple-objects)。