package user

import (
	"context"
	"encoding/json"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/basicService/qrCode/response"
	"net/http"
	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	response2 "schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/common/utils"
	"time"

	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWechatOffiaccountQrcodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetWechatOffiaccountQrcodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWechatOffiaccountQrcodeLogic {
	return &GetWechatOffiaccountQrcodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetWechatOffiaccountQrcodeLogic) GetWechatOffiaccountQrcode(r *http.Request, req *types.OAuthWechatRequest) (resp *types.Response, err error) {
	ip := utils.GetClientIP(r) // 使用工具函数获取客户端IP
	key := constant.UserQrcodePrefix + ip

	// 从Redis获取二维码数据
	qrcode := l.svcCtx.RedisClient.Get(l.ctx, key).Val()
	if qrcode != "" {
		data := new(response.ResponseQRCodeCreate)
		if err = json.Unmarshal([]byte(qrcode), data); err != nil {
			return nil, err
		}
		return response2.SuccessWithData(data.Url), nil
	}

	// 生成临时二维码
	data, err := l.svcCtx.WechatOfficial.QRCode.Temporary(l.ctx, req.Client_id, 7*24*3600)
	if err != nil {
		return nil, err
	}

	// 序列化数据并存储到Redis
	serializedData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if err = l.svcCtx.RedisClient.Set(l.ctx, key, serializedData, time.Hour*24*7).Err(); err != nil {
		return nil, err
	}

	return response2.SuccessWithData(data.Url), nil
}
