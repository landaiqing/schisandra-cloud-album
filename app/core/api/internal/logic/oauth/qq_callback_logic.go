package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/yitter/idgenerator-go/idgen"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/model"
)

type QqCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}
type AuthQQme struct {
	ClientID string `json:"client_id"`
	OpenID   string `json:"openid"`
}
type QQToken struct {
	AccessToken  string `json:"access_token"`
	ExpireIn     string `json:"expire_in"`
	RefreshToken string `json:"refresh_token"`
}

type QQUserInfo struct {
	City            string `json:"city"`
	Figureurl       string `json:"figureurl"`
	Figureurl1      string `json:"figureurl_1"`
	Figureurl2      string `json:"figureurl_2"`
	FigureurlQq     string `json:"figureurl_qq"`
	FigureurlQq1    string `json:"figureurl_qq_1"`
	FigureurlQq2    string `json:"figureurl_qq_2"`
	Gender          string `json:"gender"`
	GenderType      int    `json:"gender_type"`
	IsLost          int    `json:"is_lost"`
	IsYellowVip     string `json:"is_yellow_vip"`
	IsYellowYearVip string `json:"is_yellow_year_vip"`
	Level           string `json:"level"`
	Msg             string `json:"msg"`
	Nickname        string `json:"nickname"`
	Province        string `json:"province"`
	Ret             int    `json:"ret"`
	Vip             string `json:"vip"`
	Year            string `json:"year"`
	YellowVipLevel  string `json:"yellow_vip_level"`
}

func NewQqCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QqCallbackLogic {
	return &QqCallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QqCallbackLogic) QqCallback(w http.ResponseWriter, r *http.Request, req *types.OAuthCallbackRequest) error {

	tokenAuthUrl := l.GetQQTokenAuthUrl(req.Code)
	token, err := l.GetQQToken(tokenAuthUrl)
	if err != nil {

		return err
	}
	if token == nil {
		return nil
	}

	// 通过 token 获取 openid
	authQQme, err := l.GetQQUserOpenID(token)
	if err != nil {
		return err
	}

	// 通过 token 和 openid 获取用户信息
	userInfo, err := l.GetQQUserUserInfo(token, authQQme.OpenID)
	if err != nil {
		return err
	}

	// 处理用户信息
	userInfoBytes, err := json.Marshal(userInfo)
	if err != nil {
		return err
	}
	var qqUserInfo QQUserInfo
	err = json.Unmarshal(userInfoBytes, &qqUserInfo)
	if err != nil {
		return err
	}

	tx := l.svcCtx.DB.Begin()

	userSocial := l.svcCtx.DB.ScaAuthUserSocial
	socialUser, err := tx.ScaAuthUserSocial.Where(userSocial.OpenID.Eq(authQQme.OpenID), userSocial.Source.Eq(constant.OAuthSourceQQ), userSocial.Deleted.Eq(constant.NotDeleted)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if socialUser == nil {
		// 创建用户
		uid := idgen.NextId()
		uidStr := strconv.FormatInt(uid, 10)

		notDeleted := constant.NotDeleted
		male := constant.Male
		addUser := &model.ScaAuthUser{
			UID:      uidStr,
			Avatar:   qqUserInfo.FigureurlQq1,
			Username: authQQme.OpenID,
			Nickname: qqUserInfo.Nickname,
			Deleted:  notDeleted,
			Gender:   male,
		}
		err = tx.ScaAuthUser.Create(addUser)
		if err != nil {
			_ = tx.Rollback()
			return err
		}

		githubUser := constant.OAuthSourceQQ
		newSocialUser := &model.ScaAuthUserSocial{
			UserID: uidStr,
			OpenID: authQQme.OpenID,
			Source: githubUser,
		}
		err = tx.ScaAuthUserSocial.Create(newSocialUser)
		if err != nil {
			_ = tx.Rollback()
			return err
		}

		if res, err := l.svcCtx.CasbinEnforcer.AddRoleForUser(uidStr, constant.User); !res || err != nil {
			_ = tx.Rollback()
			return err
		}

		if err = HandleOauthLoginResponse(addUser, l.svcCtx, r, w, l.ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	} else {
		authUser := l.svcCtx.DB.ScaAuthUser

		authUserInfo, err := tx.ScaAuthUser.Where(authUser.UID.Eq(socialUser.UserID), authUser.Deleted.Eq(constant.NotDeleted)).First()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			_ = tx.Rollback()
			return err
		}

		if err = HandleOauthLoginResponse(authUserInfo, l.svcCtx, r, w, l.ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

// GetQQTokenAuthUrl 通过code获取token认证url
func (l *QqCallbackLogic) GetQQTokenAuthUrl(code string) string {
	clientId := l.svcCtx.Config.OAuth.QQ.ClientID
	clientSecret := l.svcCtx.Config.OAuth.QQ.ClientSecret
	redirectURI := l.svcCtx.Config.OAuth.QQ.RedirectURI
	return fmt.Sprintf(
		"https://graph.qq.com/oauth2.0/token?grant_type=authorization_code&client_id=%s&client_secret=%s&code=%s&redirect_uri=%s&fmt=json",
		clientId, clientSecret, code, redirectURI,
	)
}

// GetQQToken 获取 token
func (l *QqCallbackLogic) GetQQToken(url string) (*QQToken, error) {

	// 形成请求
	var req *http.Request
	var err error
	if req, err = http.NewRequest(http.MethodGet, url, nil); err != nil {
		return nil, err
	}

	// 发送请求并获得响应
	var httpClient = http.Client{}
	var res *http.Response
	if res, err = httpClient.Do(req); err != nil {
		return nil, err
	}
	// 将响应体解析为 token，并返回
	var token QQToken
	if err = json.NewDecoder(res.Body).Decode(&token); err != nil {
		return nil, err
	}
	return &token, nil
}

// GetQQUserOpenID 获取用户 openid
func (l *QqCallbackLogic) GetQQUserOpenID(token *QQToken) (*AuthQQme, error) {

	// 形成请求
	var userInfoUrl = "https://graph.qq.com/oauth2.0/me?access_token=" + token.AccessToken + "&fmt=json"
	var req *http.Request
	var err error
	if req, err = http.NewRequest(http.MethodGet, userInfoUrl, nil); err != nil {
		return nil, err
	}
	// 发送请求并获取响应
	var client = http.Client{}
	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return nil, err
	}

	// 将响应体解析为 AuthQQme，并返回
	var authQQme AuthQQme
	if err = json.NewDecoder(res.Body).Decode(&authQQme); err != nil {
		return nil, err
	}
	return &authQQme, nil
}

// GetQQUserUserInfo 获取用户信息
func (l *QqCallbackLogic) GetQQUserUserInfo(token *QQToken, openId string) (map[string]interface{}, error) {

	clientId := l.svcCtx.Config.OAuth.QQ.ClientID
	// 形成请求
	var userInfoUrl = "https://graph.qq.com/user/get_user_info?access_token=" + token.AccessToken + "&oauth_consumer_key=" + clientId + "&openid=" + openId
	var req *http.Request
	var err error
	if req, err = http.NewRequest(http.MethodGet, userInfoUrl, nil); err != nil {
		return nil, err
	}
	// 发送请求并获取响应
	var client = http.Client{}
	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return nil, err
	}

	// 将响应的数据写入 userInfo 中，并返回
	var userInfo = make(map[string]interface{})
	if err = json.NewDecoder(res.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	return userInfo, nil
}
