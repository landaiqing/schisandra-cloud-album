package storage

import (
	"context"
	"io"
	"time"
)

type BucketProperties struct {
	Name             *string
	Location         *string
	CreationDate     *time.Time
	StorageClass     *string
	ExtranetEndpoint *string
	IntranetEndpoint *string
	Region           *string
	ResourceGroupId  *string
}

// 通用存储桶统计信息
type BucketStat struct {
	Storage             int64
	ObjectCount         int64
	LastModified        int64
	StandardStorage     int64
	StandardObjectCount int64
}

// 通用存储桶信息
type BucketInfo struct {
	Name         string
	Location     string
	CreationDate *time.Time
}
type PutObjectResult struct {
	ContentMD5     *string
	ETag           *string
	HashCRC64      *string
	VersionId      *string
	CallbackResult map[string]any
}
type CompleteMultipartUploadResult struct {
	VersionId      *string
	HashCRC64      *string
	EncodingType   *string
	Location       *string
	Bucket         *string
	Key            *string
	ETag           *string
	CallbackResult map[string]any
}

type ObjectProperties struct {
	Key            *string
	Type           *string
	Size           int64
	ETag           *string
	LastModified   *time.Time
	StorageClass   *string
	RestoreInfo    *string
	TransitionTime *time.Time
}

// Service 定义存储服务接口
type Service interface {
	CreateBucket(ctx context.Context, name string) (string, error)
	ListBuckets(ctx context.Context, prefix string, maxKeys int32, marker string) ([]BucketProperties, error)
	ListBucketsPage(ctx context.Context) ([]BucketProperties, error)
	IsBucketExist(ctx context.Context, name string) (bool, error)
	GetBucketStat(ctx context.Context, name string) (*BucketStat, error)
	GetBucketInfo(ctx context.Context, name string) (*BucketInfo, error)
	DeleteBucket(ctx context.Context, name string) int
	UploadFileSimple(ctx context.Context, bucketName string, objectName string, fileData io.Reader, metadata map[string]string) (*PutObjectResult, error)
	MultipartUpload(ctx context.Context, bucketName, objectName string, filePath string) (*CompleteMultipartUploadResult, error)
	IsObjectExist(ctx context.Context, bucket string, objectName string) (bool, error)
	ListObjects(ctx context.Context, bucketName string, maxKeys int32) ([]ObjectProperties, error)
	DeleteObject(ctx context.Context, bucketName, objectName string) (int, error)
	RenameObject(ctx context.Context, destBucketName, destObjectName, srcObjectName, srcBucketName string) (int, error)
	PresignedURL(ctx context.Context, bucketName, objectKey string, expires time.Duration) (string, error)
}
