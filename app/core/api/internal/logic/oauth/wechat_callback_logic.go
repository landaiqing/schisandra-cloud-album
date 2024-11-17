package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerLibs/v3/http/helper"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/contract"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/messages"
	models2 "github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/models"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount/server/handlers/models"
	"github.com/yitter/idgenerator-go/idgen"
	"github.com/zeromicro/go-zero/core/logx"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/i18n"
	randomname "schisandra-album-cloud-microservices/app/core/api/common/random_name"
	"schisandra-album-cloud-microservices/app/core/api/common/utils"
	"schisandra-album-cloud-microservices/app/core/api/internal/logic/user"
	"schisandra-album-cloud-microservices/app/core/api/internal/logic/websocket"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent/scaauthuser"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent/scaauthusersocial"
)

type WechatCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWechatCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WechatCallbackLogic {
	return &WechatCallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WechatCallbackLogic) WechatCallback(w http.ResponseWriter, r *http.Request) error {
	_, err := l.svcCtx.WechatPublic.Server.VerifyURL(r)
	if err != nil {
		return err
	}
	rs, err := l.svcCtx.WechatPublic.Server.Notify(r, func(event contract.EventInterface) interface{} {
		switch event.GetMsgType() {
		case models2.CALLBACK_MSG_TYPE_EVENT:
			switch event.GetEvent() {
			case models.CALLBACK_EVENT_SUBSCRIBE:
				msg := models.EventSubscribe{}
				err = event.ReadMessage(&msg)
				if err != nil {
					println(err.Error())
					return "error"
				}
				key := strings.TrimPrefix(msg.EventKey, "qrscene_")
				err = l.HandlerWechatLogin(msg.FromUserName, key, w, r)
				if err != nil {
					return messages.NewText(i18n.FormatText(l.ctx, "login.loginFailed"))
				}
				return messages.NewText(i18n.FormatText(l.ctx, "login.loginSuccess"))

			case models.CALLBACK_EVENT_UNSUBSCRIBE:
				msg := models.EventUnSubscribe{}
				err := event.ReadMessage(&msg)
				if err != nil {
					println(err.Error())
					return "error"
				}
				return messages.NewText("ok")

			case models.CALLBACK_EVENT_SCAN:
				msg := models.EventScan{}
				err = event.ReadMessage(&msg)
				if err != nil {
					println(err.Error())
					return "error"
				}
				err = l.HandlerWechatLogin(msg.FromUserName, msg.EventKey, w, r)
				if err != nil {
					return messages.NewText(i18n.FormatText(l.ctx, "login.loginFailed"))
				}
				return messages.NewText(i18n.FormatText(l.ctx, "login.loginSuccess"))

			}

		case models2.CALLBACK_MSG_TYPE_TEXT:
			msg := models.MessageText{}
			err := event.ReadMessage(&msg)
			if err != nil {
				println(err.Error())
				return "error"
			}
		}
		return messages.NewText("ok")

	})
	if err != nil {
		return err
	}
	err = helper.HttpResponseSend(rs, w)
	if err != nil {
		return err
	}
	return nil
}

// HandlerWechatLogin 处理微信登录
func (l *WechatCallbackLogic) HandlerWechatLogin(openId string, clientId string, w http.ResponseWriter, r *http.Request) error {
	if openId == "" {
		return errors.New("openId is empty")
	}
	socialUser, err := l.svcCtx.MySQLClient.ScaAuthUserSocial.Query().
		Where(scaauthusersocial.OpenID(openId),
			scaauthusersocial.Source(constant.OAuthSourceWechat),
			scaauthusersocial.Deleted(constant.NotDeleted)).
		First(l.ctx)

	if err != nil && !ent.IsNotFound(err) {
		return err
	}
	tx, err := l.svcCtx.MySQLClient.Tx(l.ctx)
	if err != nil {
		return err
	}
	if ent.IsNotFound(err) {
		// 创建用户
		uid := idgen.NextId()
		uidStr := strconv.FormatInt(uid, 10)
		avatar := utils.GenerateAvatar(uidStr)
		name := randomname.GenerateName()

		addUser, fault := l.svcCtx.MySQLClient.ScaAuthUser.Create().
			SetUID(uidStr).
			SetAvatar(avatar).
			SetUsername(openId).
			SetNickname(name).
			SetDeleted(constant.NotDeleted).
			SetGender(constant.Male).
			Save(l.ctx)
		if fault != nil {
			return tx.Rollback()
		}

		if err = l.svcCtx.MySQLClient.ScaAuthUserSocial.Create().
			SetUserID(uidStr).
			SetOpenID(openId).
			SetSource(constant.OAuthSourceWechat).
			Exec(l.ctx); err != nil {
			return tx.Rollback()
		}

		if res, err := l.svcCtx.CasbinEnforcer.AddRoleForUser(uidStr, constant.User); !res || err != nil {
			return tx.Rollback()
		}

		data, result := user.HandleUserLogin(addUser, l.svcCtx, true, r, w, l.ctx)
		if !result {
			return tx.Rollback()
		}
		marshal, fault := json.Marshal(data)
		if fault != nil {
			return tx.Rollback()
		}
		err = websocket.QrcodeWebSocketHandler.SendMessageToClient(clientId, marshal)
		if err != nil {
			return tx.Rollback()
		}

	} else {
		sacAuthUser, fault := l.svcCtx.MySQLClient.ScaAuthUser.Query().
			Where(scaauthuser.UID(socialUser.UserID), scaauthuser.Deleted(constant.NotDeleted)).
			First(l.ctx)
		if fault != nil {
			return tx.Rollback()
		}

		data, result := user.HandleUserLogin(sacAuthUser, l.svcCtx, true, r, w, l.ctx)
		if !result {
			return tx.Rollback()
		}
		marshal, fault := json.Marshal(data)
		if fault != nil {
			return tx.Rollback()
		}
		err = websocket.QrcodeWebSocketHandler.SendMessageToClient(clientId, marshal)
		if err != nil {
			return tx.Rollback()
		}
	}
	if err := tx.Commit(); err != nil {
		return tx.Rollback()
	}
	return nil

}
