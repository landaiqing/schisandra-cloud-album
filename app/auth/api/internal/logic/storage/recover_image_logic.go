package storage

import (
	"context"
	"errors"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecoverImageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRecoverImageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecoverImageLogic {
	return &RecoverImageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RecoverImageLogic) RecoverImage(req *types.RecoverImageRequest) (resp string, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return "", errors.New("user_id not found")
	}
	storageInfo := l.svcCtx.DB.ScaStorageInfo

	info, err := storageInfo.Where(
		storageInfo.UserID.Eq(uid),
		storageInfo.ID.Eq(req.ID),
		storageInfo.Provider.Eq(req.Provider),
		storageInfo.Bucket.Eq(req.Bucket),
	).Update(storageInfo.DeletedAt, nil)
	if err != nil {
		return "", err
	}
	if info.RowsAffected == 0 {
		return "", errors.New("image not found")
	}

	return "success", nil
}
