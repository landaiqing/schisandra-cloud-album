package auth

import (
	"context"
	"errors"
	"schisandra-album-cloud-microservices/common/constant"

	"github.com/zeromicro/go-zero/core/logx"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutLogic) Logout() (resp string, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return "", errors.New("user_id not found")
	}
	cacheKey := constant.UserTokenPrefix + uid
	err = l.svcCtx.RedisClient.Del(l.ctx, cacheKey).Err()
	if err != nil {
		l.Logger.Error("logout failed")
		return "", errors.New("logout failed")
	}
	return "logout success", nil
}
