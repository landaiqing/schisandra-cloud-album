package client

import (
	"context"
	"time"

	"github.com/ccpwcn/kgo"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *GenerateClientIdLogic) GenerateClientId(clientIP string) (resp *types.Response, err error) {
	clientId := l.svcCtx.RedisClient.Get(l.ctx, constant.UserClientPrefix+clientIP).Val()

	if clientId != "" {
		return response.SuccessWithData(clientId), nil
	}
	simpleUuid := kgo.SimpleUuid()
	if err = l.svcCtx.RedisClient.SetEx(l.ctx, constant.UserClientPrefix+clientIP, simpleUuid, time.Hour*24*7).Err(); err != nil {
		return response.Error(), err
	}
	return response.SuccessWithData(simpleUuid), nil
}
