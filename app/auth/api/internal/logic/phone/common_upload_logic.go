package phone

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
)

type CommonUploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommonUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommonUploadLogic {
	return &CommonUploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CommonUploadLogic) CommonUpload() error {
	// todo: add your logic here and delete this line

	return nil
}
