package storage

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/common/encrypt"
	"schisandra-album-cloud-microservices/common/storage/config"
)

type UploadFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadFileLogic {
	return &UploadFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadFileLogic) UploadFile(r *http.Request) (resp string, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return "", errors.New("user_id not found")
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		return "", errors.New("file not found")
	}
	defer file.Close()
	//formValue := r.PostFormValue("result")
	//
	//var result types.File
	//err = json.Unmarshal([]byte(formValue), &result)
	//if err != nil {
	//	return "", errors.New("invalid result")
	//}
	//fmt.Println(result)
	ossConfig := l.svcCtx.DB.ScaStorageConfig
	dbConfig, err := ossConfig.Where(ossConfig.UserID.Eq(uid)).First()
	if err != nil {
		return "", errors.New("oss config not found")
	}
	accessKey, err := encrypt.Decrypt(dbConfig.AccessKey, l.svcCtx.Config.Encrypt.Key)
	if err != nil {
		return "", errors.New("decrypt access key failed")
	}
	secretKey, err := encrypt.Decrypt(dbConfig.SecretKey, l.svcCtx.Config.Encrypt.Key)
	if err != nil {
		return "", errors.New("decrypt secret key failed")
	}
	storageConfig := &config.StorageConfig{
		Provider:   dbConfig.Type,
		Endpoint:   dbConfig.Endpoint,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		BucketName: dbConfig.Bucket,
		Region:     dbConfig.Region,
	}
	service, err := l.svcCtx.StorageManager.GetStorage(uid, storageConfig)
	if err != nil {
		return "", errors.New("get storage failed")
	}
	result, err := service.UploadFileSimple(l.ctx, dbConfig.Bucket, header.Filename, file, map[string]string{})
	if err != nil {
		return "", errors.New("upload file failed")
	}
	return *result.ContentMD5, nil
}
