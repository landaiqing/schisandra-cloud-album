package miniox

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zeromicro/go-zero/core/logx"
	"schisandra-album-cloud-microservices/common/constant"
)

func NewMinio(endpoint, accessKeyID, secretAccessKey string, useSSL bool) *minio.Client {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		panic(err)
	}
	// 初始化存储桶
	thumbnailBucketExists, err := client.BucketExists(context.Background(), constant.ThumbnailBucketName)
	if err != nil || !thumbnailBucketExists {
		err = client.MakeBucket(context.Background(), constant.ThumbnailBucketName, minio.MakeBucketOptions{Region: "us-east-1", ObjectLocking: true})
		if err != nil {
			logx.Errorf("Failed to create MinIO bucket: %v", err)
			panic(err)
		}
	}
	faceBucketExists, err := client.BucketExists(context.Background(), constant.FaceBucketName)
	if err != nil || !faceBucketExists {
		err = client.MakeBucket(context.Background(), constant.FaceBucketName, minio.MakeBucketOptions{Region: "us-east-1", ObjectLocking: true})
		if err != nil {
			logx.Errorf("Failed to create MinIO bucket: %v", err)
			panic(err)
		}
	}
	commentImagesBucketExists, err := client.BucketExists(context.Background(), constant.CommentImagesBucketName)
	if err != nil || !commentImagesBucketExists {
		err = client.MakeBucket(context.Background(), constant.CommentImagesBucketName, minio.MakeBucketOptions{Region: "us-east-1", ObjectLocking: true})
		if err != nil {
			logx.Errorf("Failed to create MinIO bucket: %v", err)
			panic(err)
		}
	}
	shareImagesBucketExists, err := client.BucketExists(context.Background(), constant.ShareImagesBucketName)
	if err != nil || !shareImagesBucketExists {
		err = client.MakeBucket(context.Background(), constant.ShareImagesBucketName, minio.MakeBucketOptions{Region: "us-east-1", ObjectLocking: true})
		if err != nil {
			logx.Errorf("Failed to create MinIO bucket: %v", err)
			panic(err)
		}
	}
	return client
}
