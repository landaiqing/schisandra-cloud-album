package storage

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/encrypt"
)

type SetStorageConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetStorageConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetStorageConfigLogic {
	return &SetStorageConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetStorageConfigLogic) SetStorageConfig(req *types.StorageConfigRequest) (resp string, err error) {

	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return "", errors.New("user_id not found")
	}
	accessKey, err := encrypt.Encrypt(req.AccessKey, l.svcCtx.Config.Encrypt.Key)
	if err != nil {
		return "", err
	}
	secretKey, err := encrypt.Encrypt(req.SecretKey, l.svcCtx.Config.Encrypt.Key)
	if err != nil {
		return "", err
	}
	ossConfig := &model.ScaStorageConfig{
		UserID:    uid,
		Type:      req.Type,
		Endpoint:  req.Endpoint,
		Bucket:    req.Bucket,
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    req.Region,
	}
	err = l.svcCtx.DB.ScaStorageConfig.Create(ossConfig)
	if err != nil {
		return "", err
	}
	return "success", nil
}
