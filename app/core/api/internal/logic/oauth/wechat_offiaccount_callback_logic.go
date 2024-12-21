package oauth

import (
	"context"
	"encoding/json"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/contract"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/messages"
	models2 "github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/models"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount/server/handlers/models"
	"net/http"
	"schisandra-album-cloud-microservices/app/core/api/common/encrypt"
	"schisandra-album-cloud-microservices/app/core/api/common/i18n"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/internal/logic/websocket"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
)

type WechatOffiaccountCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}
type MessageData struct {
	Openid   string `json:"openid"`
	ClientId string `json:"client_id"`
}

func NewWechatOffiaccountCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WechatOffiaccountCallbackLogic {
	return &WechatOffiaccountCallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WechatOffiaccountCallbackLogic) WechatOffiaccountCallback(r *http.Request) (*http.Response, error) {
	rs, err := l.svcCtx.WechatOfficial.Server.Notify(r, func(event contract.EventInterface) interface{} {
		switch event.GetMsgType() {
		case models2.CALLBACK_MSG_TYPE_EVENT:
			switch event.GetEvent() {
			case models.CALLBACK_EVENT_SUBSCRIBE:
				msg := models.EventSubscribe{}
				err := event.ReadMessage(&msg)
				if err != nil {
					logx.Error(err.Error())
					return "error"
				}
				key := strings.TrimPrefix(msg.EventKey, "qrscene_")
				err = l.SendMessage(msg.FromUserName, key)
				if err != nil {
					return messages.NewText(i18n.FormatText(l.ctx, "login.loginFailed"))
				}
				return messages.NewText(i18n.FormatText(l.ctx, "login.loginSuccess"))

			case models.CALLBACK_EVENT_UNSUBSCRIBE:
				msg := models.EventUnSubscribe{}
				err := event.ReadMessage(&msg)
				if err != nil {
					logx.Error(err.Error())
					return "error"
				}
				return messages.NewText("ok")

			case models.CALLBACK_EVENT_SCAN:
				msg := models.EventScan{}
				err := event.ReadMessage(&msg)
				if err != nil {
					logx.Error(err.Error())
					return "error"
				}
				err = l.SendMessage(msg.FromUserName, msg.EventKey)
				if err != nil {
					return messages.NewText(i18n.FormatText(l.ctx, "login.loginFailed"))
				}
				return messages.NewText(i18n.FormatText(l.ctx, "login.loginSuccess"))

			}

		case models2.CALLBACK_MSG_TYPE_TEXT:
			msg := models.MessageText{}
			err := event.ReadMessage(&msg)
			if err != nil {
				logx.Error(err.Error())
				return "error"
			}
		}
		return messages.NewText("ok")
	})
	if err != nil {
		return nil, err
	}
	return rs, nil
}

// SendMessage 发送消息到客户端
func (l *WechatOffiaccountCallbackLogic) SendMessage(openId string, clientId string) error {
	encryptClientId, err := encrypt.Encrypt(clientId, l.svcCtx.Config.Encrypt.Key, l.svcCtx.Config.Encrypt.IV)
	if err != nil {
		return err
	}
	encryptOpenId, err := encrypt.Encrypt(openId, l.svcCtx.Config.Encrypt.Key, l.svcCtx.Config.Encrypt.IV)
	if err != nil {
		return err
	}
	messageData := MessageData{
		Openid:   encryptOpenId,
		ClientId: encryptClientId,
	}
	jsonData, err := json.Marshal(response.SuccessWithData(messageData))
	if err != nil {
		return err
	}
	err = websocket.QrcodeWebSocketHandler.SendMessageToClient(clientId, jsonData)
	if err != nil {
		return err
	}
	return nil
}
