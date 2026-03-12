在对象存储中，对象（Object）就像文件一样，是存储数据的基本单位。上传的每个文件（如文档、图片、视频等）都会作为Object保存在[存储空间（Bucket）](https://help.aliyun.com/zh/oss/user-guide/oss-bucket-overview)中，方便后续管理。

## **Object组成**

-   **Key：**Object的名称，用于查询Object，具有唯一性。

-   **Data：**Object的实际数据，由任意长度的字节组成。

-   **Object Meta：**Object元数据，是一组键值对，用于描述Object的属性，如最后修改时间、大小等，也可以存储自定义信息。

-   **Version ID：**若要保存一个文件的多个版本，需启用[版本控制](https://help.aliyun.com/zh/oss/user-guide/manage-objects-in-a-versioning-enabled-bucket)功能。每次上传同名Object时，OSS会为新版本分配一个唯一的Version ID，用于访问和管理特定版本的Object。


## **Object命名（Key）**

每个Object都有唯一的名称，称为Key。例如，Key为"exampledir/example.jpg"表示exampledir目录下的example.jpg文件。

> 在SDK文档中，Key也被称为ObjectKey、ObjectName或对象名称，这些术语含义相同。

#### **命名规范**

-   使用UTF-8编码。

-   长度必须在1~1023字节之间。

-   不能以正斜线（/）或者反斜线（\\）开头。

-   区分大小写。


#### 命名建议

-   **使用有意义的名称**，例如文件名、日期、用户ID等，方便查找和理解。

-   **使用前缀组织数据**，按日期、用户ID、地域等作为前缀，创建层次结构。

-   **确保名称唯一性**，可以包含随机数或UUID，避免命名冲突。


#### 命名示例

| **Object在Bucket中的位置** | **Key的表示方法** |
| --- | --- |
| 存储空间examplebucket根目录下存放了名为exampleobject.txt的Object | exampleobject.txt |
| 存储空间examplebucket根目录下的destdir目录中存放了exampleobject.jpg的Object | destdir/exampleobject.jpg |

## **OSS目录呈现**

在对象存储OSS中，其实并不具备传统文件系统中的文件和文件夹概念，但为了符合用户的使用习惯，OSS通过在Key中添加“/”来模拟文件夹，例如“exampledir/example.jpg”。此时，“exampledir”就被模拟成了一个文件夹，“example.jpg”则模拟成“exampledir”文件夹下的文件名。在OSS控制台和ossbrowser等图形化工具中可以看到这种文件夹结构。但实际上，Object的Key仍然是“exampledir/example.jpg”。详情请参见[管理目录](https://help.aliyun.com/zh/oss/user-guide/manage-directories)。

## **Object类型**

Object包含以下四种类型。

**重要**

不支持在不同类型的 Object 之间相互转换。例如，Normal类型的文件无法转换为Multipart或者Appendable类型。

| **类型** | **定义** | **说明** |
| --- | --- | --- |
| Normal | 通过[简单上传](https://help.aliyun.com/zh/oss/user-guide/simple-upload#concept-bws-3bb-5db)生成的Object类型。 | ![image](https://help-static-aliyun-doc.aliyuncs.com/assets/img/zh-CN/5925512671/CAEQVRiBgIDx3azKrBkiIDFhZTMyNjQxZGFmMzRmMzA5YTA4ZDIyNzQ4ZTBkYTkw4940965_20250220112557.441.svg) - 在版本控制未开启或暂停时，若有多个用户上传同名Object，后上传完成的Object会覆盖前一个，并只保留最后上传完成的Object。例如，用户A在用户B之后完成上传，最终将保留用户A上传的Object。 - 版本控制开启时，每次上传同名Object都会创建新版本，并用版本ID（Version ID）标识。OSS根据上传的开始时间确定最新版本。例如，用户B在用户A之后开始上传，则用户B上传的Object将被标记为最新版本。 |
| Multipart | 通过[分片上传](https://help.aliyun.com/zh/oss/user-guide/multipart-upload#concept-wzs-2gb-5db)生成的Object类型。 | - 在版本控制未开启或暂停时，分片上传完成后会合并为一个完整Object，且新的完整的Object会覆盖同名的旧Object，保留最后完成合并的Object。 - 版本控制开启时，每次分片上传完成并合并后的同名Object都会创建新版本，用Version ID标识，最后完成合并的Object会被标记为最新版本。 |
| Appendable | 通过[追加上传](https://help.aliyun.com/zh/oss/user-guide/append-upload-11#concept-ls5-yhb-5db)生成的Object类型。 | 无论版本控制的状态如何，Appendable类型的Object在追加数据时，不会因为追加操作而生成新的版本，所有追加的数据都只会附加到原始文件之后。 |
| Symlink | 通过[PutSymlink](https://help.aliyun.com/zh/oss/developer-reference/putsymlink)生成的软链接。 | 设置[软链接](https://help.aliyun.com/zh/oss/user-guide/configure-symbolic-links)以便快速访问常用Object。 |

**Delete Marker（删除标记）**

除上述四种Object类型外，OSS还存在一种特殊的对象标识——**Delete Marker（删除标记）**。它并不是传统的Object类型，而是在版本控制已开启或暂停的Bucket中执行[DeleteObject](https://help.aliyun.com/zh/oss/developer-reference/deleteobject)（删除）操作时生成的标记。如果在删除Object时未指定版本ID（Version ID），OSS会创建一个“删除标记”作为最新版本。Object看似已被删除，实际上数据仍然保留，后续仍可恢复。若删除Object时指定了版本ID，则会删除指定版本的Object，而不生成删除标记。

## **功能指引**

## 上传文件

| **操作** | **说明** |
| --- | --- |
| [简单上传](https://help.aliyun.com/zh/oss/user-guide/simple-upload) | 简单上传指的是使用OSS API中的PutObject接口上传小于5 GB的单个文件，适用于一次HTTP请求交互即可完成上传的场景。 |
| [分片上传](https://help.aliyun.com/zh/oss/user-guide/multipart-upload) | 将待上传的文件分成多个碎片（Part）分别上传，上传完成后再调用CompleteMultipartUpload接口将这些Part组合成一个Object。 |
| [断点续传上传](https://help.aliyun.com/zh/oss/user-guide/resumable-upload) | 通过Checkpoint文件指定断点记录点。上传过程中，如果出现网络异常或程序崩溃导致文件上传失败时，将从断点记录处继续上传未上传完成的部分。 |
| [上传回调](https://help.aliyun.com/zh/oss/user-guide/upload-callbacks-12) | OSS在完成文件上传时可以提供回调（Callback）给应用服务器。只需在发送给OSS的请求中携带相应的Callback参数，即可实现回调。 |
| [客户端直传](https://help.aliyun.com/zh/oss/user-guide/uploading-objects-to-oss-directly-from-clients/) | 客户端直接上传文件到对象存储OSS，避免了业务服务器中转文件，提高了上传速度，节省了服务器资源。 |
| [表单上传](https://help.aliyun.com/zh/oss/user-guide/form-upload) | 使用OSS API中的PostObject请求来完成文件上传，上传的文件不能超过5 GB。 |
| [追加上传](https://help.aliyun.com/zh/oss/user-guide/append-upload-11) | 在已上传的Appendable类型Object后面直接追加内容。 |
| [RTMP推流上传](https://help.aliyun.com/zh/oss/developer-reference/rtmp-based-stream-ingest) | OSS支持使用RTMP协议推送H264编码的视频流和AAC编码的音频流到OSS。推送到OSS的音视频数据可以点播播放；在对延迟不敏感的应用场景，也可以做直播用途。 |

## 下载文件

| **操作** | **说明** |
| --- | --- |
| [简单下载](https://help.aliyun.com/zh/oss/user-guide/simple-download-1) | 使用OSS API的GetObject接口，下载已上传的文件，适用于一次HTTP请求交互即可完成下载的场景。 |
| [断点续传下载](https://help.aliyun.com/zh/oss/user-guide/oss-resumable-download) | 从指定位置开始下载文件。在下载大文件时，可以分多次下载。如果下载中断，重启时也可以从上次完成的位置开始继续下载。 |
| [使用签名URL下载文件](https://help.aliyun.com/zh/oss/user-guide/share-objects-by-url) | 文件上传至Bucket后，可以通过文件URL将文件分享给第三方预览或下载。 |
| [限定条件下载](https://help.aliyun.com/zh/oss/user-guide/conditional-download) | 通过检测最后修改时间或ETag条件，只有在文件内容发生变化时才会触发下载，避免重复下载。 |

## 管理文件

| **操作** | **说明** |
| --- | --- |
| [列举文件](https://help.aliyun.com/zh/oss/user-guide/list-objects-18) | Bucket内的Object默认按照字母序排列。可以结合实际场景列举当前Bucket的所有Object、指定前缀的Object、指定个数的Object等。 |
| [拷贝文件](https://help.aliyun.com/zh/oss/user-guide/copy-objects-16) | 在不改变文件内容的情况下，将同一地域下的源Bucket内的文件复制到目标Bucket。 |
| [重命名文件](https://help.aliyun.com/zh/oss/user-guide/rename-objects) | 在同一个Bucket内重命名某个Object。 |
| [搜索文件](https://help.aliyun.com/zh/oss/user-guide/search-for-objects) | 当Bucket上传了大量的Object时，OSS支持通过指定文件名前缀快速搜索并定位目标文件。 |
| [解冻文件](https://help.aliyun.com/zh/oss/user-guide/restore-objects-for-access) | 冷归档存储或者深度冷归档存储类型的Object需要解冻（Restore）之后才能读取。 |
| [删除文件](https://help.aliyun.com/zh/oss/user-guide/delete-objects-18) | 通过多种方式删除Bucket中不再需要保留的Object。 |
| [对象标签](https://help.aliyun.com/zh/oss/user-guide/object-tagging-8) | OSS支持使用标签对Bucket中的Object进行分类，可以针对同标签的Object设置生命周期规则、访问权限等。 |
| [软链接](https://help.aliyun.com/zh/oss/user-guide/configure-symbolic-links) | 软链接功能用于快速访问Bucket内的常用Object。设置软链接后，可以通过软链接文件快速打开Object。 |
| [管理文件元数据](https://help.aliyun.com/zh/oss/user-guide/manage-object-metadata-10/) | 文件元信息是对文件的属性描述，包括HTTP标准属性（HTTP Header）和用户自定义元数据（User Meta）两种。可以通过设置文件HTTP头来自定义HTTP请求的策略，例如文件缓存策略、强制下载策略等，也可以通过设置用户自定义元数据来标识Object的用途或属性等。 |
| [单链接限速](https://help.aliyun.com/zh/oss/user-guide/single-connection-bandwidth-throttling-4) | 客户端访问OSS内的文件时会占用较大带宽，可能会对其他应用造成影响。可以通过OSS提供的单链接限速功能在上传、下载文件等操作中进行流量控制，以保证其他应用的网络带宽。 |
| [查询文件](https://help.aliyun.com/zh/oss/user-guide/query-objects) | 使用SelectObject对目标文件执行SQL语句，返回执行结果。 |
| [删除碎片](https://help.aliyun.com/zh/oss/user-guide/delete-parts) | 在分片上传过程中，如果某些Part不再需要，可以调用AbortMultipartUpload接口删除这些未完成的Part。 |
| [通过云存储网关挂载OSS](https://help.aliyun.com/zh/oss/user-guide/use-csg-to-attach-oss-buckets-to-ecs-instances) | 通过云存储网关挂载OSS，您可以将OSS映射为一个共享的文件存储系统，实现多个用户在不同地点和设备上共享访问OSS数据。挂载完成后，您可以像使用本地文件夹和磁盘一样操作OSS资源。 |

## 常见问题

-   [上传文件常见问题](https://help.aliyun.com/zh/oss/usage-query/)

-   [下载文件常见问题](https://help.aliyun.com/zh/oss/download-file-faq/)

-   [客户端直传常见问题](https://help.aliyun.com/zh/oss/direct-transmission-faq/)

-   [管理文件常见问题](https://help.aliyun.com/zh/oss/manage-files-faq/)