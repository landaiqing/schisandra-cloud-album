package client

import (
	"context"
	"net/http"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/errors"
	"time"

	"github.com/ccpwcn/kgo"

	"github.com/zeromicro/go-zero/core/logx"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
)

type GenerateClientIdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateClientIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateClientIdLogic {
	return &GenerateClientIdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateClientIdLogic) GenerateClientId(clientIP string) (resp string, err error) {
	clientId := l.svcCtx.RedisClient.Get(l.ctx, constant.UserClientPrefix+clientIP).Val()

	if clientId != "" {
		return clientId, nil
	}
	simpleUuid := kgo.SimpleUuid()
	if err = l.svcCtx.RedisClient.SetEx(l.ctx, constant.UserClientPrefix+clientIP, simpleUuid, time.Hour*24*7).Err(); err != nil {
		return "", errors.New(http.StatusInternalServerError, err.Error())
	}
	return simpleUuid, nil
}
