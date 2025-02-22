package storage

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"io"
	"log"
	"os"
	"schisandra-album-cloud-microservices/common/storage/config"
	"schisandra-album-cloud-microservices/common/storage/events"
	"sync"
	"time"
)

type AliOSS struct {
	client     *oss.Client
	bucket     string
	dispatcher events.Dispatcher
}

// NewAliOSS 创建阿里云 OSS 实例
func NewAliOSS(config *config.StorageConfig, dispatcher events.Dispatcher) (*AliOSS, error) {
	credentialsProvider := credentials.NewStaticCredentialsProvider(config.AccessKey, config.SecretKey)
	cfg := oss.NewConfig().WithCredentialsProvider(credentialsProvider).
		WithEndpoint(config.Endpoint).
		WithRegion(config.Region).WithInsecureSkipVerify(false)
	client := oss.NewClient(cfg)
	return &AliOSS{client: client, bucket: config.BucketName, dispatcher: dispatcher}, nil
}

// CreateBucket 创建存储桶
func (a *AliOSS) CreateBucket(ctx context.Context, bucketName string) (string, error) {
	request := &oss.PutBucketRequest{
		Bucket: oss.Ptr(bucketName),
	}
	result, err := a.client.PutBucket(ctx, request)
	if err != nil {
		return "", fmt.Errorf("failed to put bucket, error: %v", err)
	}
	return result.Status, nil
}

// ListBucketsPage 列出所有存储桶
func (a *AliOSS) ListBucketsPage(ctx context.Context) ([]BucketProperties, error) {
	request := &oss.ListBucketsRequest{}
	// 定义一个函数来处理 PaginatorOptions
	modifyOptions := func(opts *oss.PaginatorOptions) {
		// 在这里可以修改opts的值，比如设置每页返回的存储空间数量上限
		// 示例：opts.Limit = 5，即每页返回5个存储空间
		opts.Limit = 5
	}
	p := a.client.NewListBucketsPaginator(request, modifyOptions)
	var buckets []BucketProperties
	for p.HasNext() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list buckets, error: %v", err)
		}
		for _, b := range page.Buckets {
			buckets = append(buckets, BucketProperties{
				Name:             b.Name,
				CreationDate:     b.CreationDate,
				Location:         b.Location,
				Region:           b.Region,
				StorageClass:     b.StorageClass,
				ExtranetEndpoint: b.ExtranetEndpoint,
				IntranetEndpoint: b.IntranetEndpoint,
				ResourceGroupId:  b.ResourceGroupId,
			})
		}
	}
	return buckets, nil
}

// ListBuckets 列出所有存储桶
func (a *AliOSS) ListBuckets(ctx context.Context, prefix string, maxKeys int32, marker string) ([]BucketProperties, error) {
	request := &oss.ListBucketsRequest{
		Prefix:  oss.Ptr(prefix),
		MaxKeys: maxKeys,
		Marker:  oss.Ptr(marker),
	}
	var buckets []BucketProperties
	for {
		lsRes, err := a.client.ListBuckets(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("failed to list buckets, error: %v", err)
		}

		for _, bucket := range lsRes.Buckets {
			buckets = append(buckets, BucketProperties{
				Name:             bucket.Name,
				CreationDate:     bucket.CreationDate,
				Location:         bucket.Location,
				Region:           bucket.Region,
				StorageClass:     bucket.StorageClass,
				ExtranetEndpoint: bucket.ExtranetEndpoint,
				IntranetEndpoint: bucket.IntranetEndpoint,
				ResourceGroupId:  bucket.ResourceGroupId,
			})
		}

		if !lsRes.IsTruncated {
			break
		}
		marker = *lsRes.NextMarker
	}

	return buckets, nil
}

// IsBucketExist 检查存储桶是否存在
func (a *AliOSS) IsBucketExist(ctx context.Context, bucketName string) (bool, error) {
	exist, err := a.client.IsBucketExist(ctx, bucketName)
	if err != nil {
		return false, fmt.Errorf("failed to check bucket exist, error: %v", err)
	}
	return exist, nil
}

// GetBucketStat 获取存储桶容量
func (a *AliOSS) GetBucketStat(ctx context.Context, bucketName string) (*BucketStat, error) {
	request := &oss.GetBucketStatRequest{
		Bucket: oss.Ptr(bucketName),
	}
	result, err := a.client.GetBucketStat(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket stat, error: %v", err)
	}
	return &BucketStat{
		Storage:             result.Storage,
		ObjectCount:         result.ObjectCount,
		LastModified:        result.LastModifiedTime,
		StandardStorage:     result.StandardStorage,
		StandardObjectCount: result.StandardObjectCount,
	}, nil
}

// GetBucketInfo 获取存储桶信息
func (a *AliOSS) GetBucketInfo(ctx context.Context, bucketName string) (*BucketInfo, error) {
	request := &oss.GetBucketInfoRequest{
		Bucket: oss.Ptr(bucketName),
	}
	result, err := a.client.GetBucketInfo(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket info, error: %v", err)
	}
	return &BucketInfo{
		Name:         *result.BucketInfo.Name,
		Location:     *result.BucketInfo.Location,
		CreationDate: result.BucketInfo.CreationDate,
	}, nil
}

// DeleteBucket 删除存储桶
func (a *AliOSS) DeleteBucket(ctx context.Context, bucketName string) int {
	request := &oss.DeleteBucketRequest{
		Bucket: oss.Ptr(bucketName),
	}
	result, err := a.client.DeleteBucket(ctx, request)
	if err != nil {
		log.Fatalf("failed to delete bucket %v", err)
	}
	return result.StatusCode
}

// UploadFileSimple 上传文件
func (a *AliOSS) UploadFileSimple(ctx context.Context, bucketName, objectName string, fileData io.Reader, metadata map[string]string) (*PutObjectResult, error) {
	putRequest := &oss.PutObjectRequest{
		Bucket:               oss.Ptr(bucketName),      // 存储空间名称
		Key:                  oss.Ptr(objectName),      // 对象名称
		StorageClass:         oss.StorageClassStandard, // 指定对象的存储类型为标准存储
		Acl:                  oss.ObjectACLPrivate,     // 指定对象的访问权限为私有访问
		Metadata:             metadata,                 // 指定对象的元数据
		Body:                 fileData,                 // 使用文件流
		ServerSideEncryption: oss.Ptr("AES256"),
	}
	result, err := a.client.PutObject(ctx, putRequest)

	if err != nil {
		return nil, fmt.Errorf("failed to upload file, error: %v", err)
	}
	return &PutObjectResult{
		ContentMD5:     result.ContentMD5,
		ETag:           result.ETag,
		HashCRC64:      result.HashCRC64,
		VersionId:      result.VersionId,
		CallbackResult: result.CallbackResult,
	}, nil
}

// MultipartUpload 分片上传文件
func (a *AliOSS) MultipartUpload(ctx context.Context, bucketName, objectName string, filePath string) (*CompleteMultipartUploadResult, error) {
	initRequest := &oss.InitiateMultipartUploadRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr(objectName),
	}
	initResult, err := a.client.InitiateMultipartUpload(ctx, initRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate multipart upload, error: %v", err)
	}
	uploadId := *initResult.UploadId

	var wg sync.WaitGroup
	var parts []oss.UploadPart
	count := 3
	var mu sync.Mutex

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("failed to open local file %v", err)
	}
	defer file.Close()

	bufReader := bufio.NewReader(file)
	content, err := io.ReadAll(bufReader)
	if err != nil {
		log.Fatalf("failed to read local file %v", err)
	}

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
			partResult, err := a.client.UploadPart(context.TODO(), partRequest)
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
	result, err := a.client.CompleteMultipartUpload(context.TODO(), request)
	if err != nil {
		log.Fatalf("failed to complete multipart upload %v", err)
	}
	return &CompleteMultipartUploadResult{
		VersionId:      result.VersionId,      // 版本号
		ETag:           result.ETag,           // 对象的ETag
		HashCRC64:      result.HashCRC64,      // 对象的Hash值
		EncodingType:   result.EncodingType,   // 对象的编码格式
		Location:       result.Location,       // 对象的存储位置
		Bucket:         result.Bucket,         // 对象的存储空间名称
		Key:            result.Key,            // 对象的名称
		CallbackResult: result.CallbackResult, // 回调结果
	}, nil
}

// DownloadFile 下载文件
func (a *AliOSS) DownloadFile(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
	}
	result, err := a.client.GetObject(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to download file, error: %v", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content, error: %v", err)
	}
	return data, nil
}

// IsObjectExist 检查对象是否存在
func (a *AliOSS) IsObjectExist(ctx context.Context, bucket string, objectName string) (bool, error) {
	result, err := a.client.IsObjectExist(ctx, bucket, objectName)
	if err != nil {
		return false, fmt.Errorf("failed to check object exist, error: %v", err)
	}
	return result, nil
}

// ListObjects 列出存储桶中的对象
func (a *AliOSS) ListObjects(ctx context.Context, bucketName string, maxKeys int32) ([]ObjectProperties, error) {
	var continueToken = ""
	request := &oss.ListObjectsV2Request{
		Bucket:            oss.Ptr(bucketName),
		ContinuationToken: &continueToken,
		MaxKeys:           maxKeys,
	}
	var objects []ObjectProperties
	for {
		// 执行列举所有文件的操作
		lsRes, err := a.client.ListObjectsV2(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects, error: %v", err)
		}
		// 打印列举结果
		for _, object := range lsRes.Contents {
			objects = append(objects, ObjectProperties{
				Key:            object.Key,
				Type:           object.Type,
				Size:           object.Size,
				LastModified:   object.LastModified,
				ETag:           object.ETag,
				StorageClass:   object.StorageClass,
				RestoreInfo:    object.RestoreInfo,
				TransitionTime: object.TransitionTime,
			})
		}

		// 如果还有更多对象需要列举，则更新continueToken标记并继续循环
		if lsRes.IsTruncated {
			continueToken = *lsRes.NextContinuationToken
		} else {
			break // 如果没有更多对象，退出循环
		}
	}
	return objects, nil
}

// DeleteObject 删除对象
func (a *AliOSS) DeleteObject(ctx context.Context, bucketName, objectName string) (int, error) {
	request := &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
	}
	result, err := a.client.DeleteObject(ctx, request)
	if err != nil {
		return -1, fmt.Errorf("failed to delete object, error: %v", err)
	}
	return result.StatusCode, nil
}

// RenameObject 重命名对象
func (a *AliOSS) RenameObject(ctx context.Context, destBucketName, destObjectName, srcObjectName, srcBucketName string) (int, error) {
	// 创建文件拷贝器
	c := a.client.NewCopier() // 构建拷贝对象的请求
	copyRequest := &oss.CopyObjectRequest{
		Bucket:       oss.Ptr(destBucketName),  // 目标存储空间名称
		Key:          oss.Ptr(destObjectName),  // 目标对象名称
		SourceKey:    oss.Ptr(srcObjectName),   // 源对象名称
		SourceBucket: oss.Ptr(srcBucketName),   // 源存储空间名称
		StorageClass: oss.StorageClassStandard, // 指定存储类型为归档类型
	}
	// 执行拷贝对象的操作
	_, err := c.Copy(ctx, copyRequest)
	if err != nil {
		return -1, fmt.Errorf("failed to copy object, error: %v", err)
	}

	// 构建删除对象的请求
	deleteRequest := &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(srcBucketName), // 存储空间名称
		Key:    oss.Ptr(srcObjectName), // 要删除的对象名称
	}
	// 执行删除对象的操作
	deleteResult, err := a.client.DeleteObject(ctx, deleteRequest)
	if err != nil {
		return -1, fmt.Errorf("failed to delete object, error: %v", err)
	}
	return deleteResult.StatusCode, nil
}

// PresignedURL 生成预签名URL
func (a *AliOSS) PresignedURL(ctx context.Context, bucketName, objectKey string, expires time.Duration) (string, error) {

	// 生成预签名URL
	presignedResult, err := a.client.Presign(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr(objectKey),
	}, oss.PresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL, error: %v", err)
	}

	return presignedResult.URL, nil
}
