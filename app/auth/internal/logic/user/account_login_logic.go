package user

import (
	"context"

	"schisandra-album-cloud-microservices/app/auth/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AccountLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAccountLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AccountLoginLogic {
	return &AccountLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AccountLoginLogic) AccountLogin(req *types.AccountLoginRequest) (resp *types.LoginResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
