package user

import (
	"context"

	"schisandra-album-cloud-microservices/app/auth/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PhoneLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPhoneLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PhoneLoginLogic {
	return &PhoneLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PhoneLoginLogic) PhoneLogin(req *types.PhoneLoginRequest) (resp *types.LoginResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
