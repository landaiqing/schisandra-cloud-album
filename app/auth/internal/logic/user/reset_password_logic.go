package user

import (
	"context"

	"schisandra-album-cloud-microservices/app/auth/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPasswordLogic {
	return &ResetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetPasswordLogic) ResetPassword(req *types.ResetPasswordRequest) (resp string, err error) {
	// todo: add your logic here and delete this line

	return
}
